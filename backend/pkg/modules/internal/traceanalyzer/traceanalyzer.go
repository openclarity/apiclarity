// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package traceanalyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/guessableid"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/nlid"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/sensitive"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/weakbasicauth"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/weakjwt"
)

const (
	moduleName = "TraceAnalyzer"
)

const (
	dictFilenamesEnvVar  = "TRACE_ANALYZER_DICT_FILENAMES"
	dictFilenamesDefault = ""

	rulesFilenamesEnvVar  = "TRACE_ANALYZER_RULES_FILENAMES"
	rulesFilenamesDefault = ""

	sensitiveKeywordsFilenamesEnvVar  = "TRACE_ANALYZER_SENSITIVE_KEYWORDS_FILENAMES"
	sensitiveKeywordsFilenamesDefault = ""

	ignoreFindingsEnvVar  = "TRACE_ANALYZER_IGNORE_FINDINGS"
	ignoreFindingsDefault = ""
)

// A finding is an interpreted annotation.
type Finding struct {
	ShortDesc    string
	DetailedDesc string
	Severity     string
	Alert        *core.Annotation
}

type ParameterFinding struct {
	Location string      `json:"location"`
	Method   string      `json:"method"`
	Name     string      `json:"name"`
	Value    string      `json:"value"`
	Reason   interface{} `json:"reason"`
}

type traceAnalyzerConfig struct {
	dictFilenames              []string `yaml:"dictFilenames"`
	rulesFilenames             []string `yaml:"rulesFilenames"`
	sensitiveKeywordsFilenames []string `yaml:"keywordsFilenames"`
	ignoreFindings             []string `yaml:"ignoreFindings"`
}

type traceAnalyzer struct {
	httpHandler http.Handler

	config traceAnalyzerConfig

	ignoreFindings map[string]bool

	guessableID   *guessableid.GuessableAnalyzer
	nlid          *nlid.NLID
	weakBasicAuth *weakbasicauth.WeakBasicAuth
	weakJWT       *weakjwt.WeakJWT
	sensitive     *sensitive.Sensitive

	accessor core.BackendAccessor
}

func newTraceAnalyzer(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	var err error

	p := traceAnalyzer{}
	h := HandlerWithOptions(&httpHandler{ta: &p}, ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + moduleName})
	p.httpHandler = h
	p.ignoreFindings = map[string]bool{}
	p.accessor = accessor

	p.config = loadConfig()
	log.Debugf("TraceAnalyzer Configuration: %+v", p.config)

	for _, ifinding := range p.config.ignoreFindings {
		p.ignoreFindings[ifinding] = true
	}

	passwordList, err := utils.ReadDictionaryFiles(p.config.dictFilenames)
	if err != nil {
		return nil, fmt.Errorf("unable to read password files: %w", err)
	}
	weakKeyList, err := utils.ReadDictionaryFiles(p.config.dictFilenames)
	if err != nil {
		return nil, fmt.Errorf("unable to read list of weak keys: %w", err)
	}
	sensitiveKeywords, err := utils.ReadDictionaryFiles(p.config.sensitiveKeywordsFilenames)
	if err != nil {
		return nil, fmt.Errorf("unable to read list of sensitive keywords: %w", err)
	}

	p.guessableID = guessableid.NewGuessableAnalyzer(guessableid.MaxParamHistory)
	p.nlid = nlid.NewNLID(nlid.NLIDRingBufferSize)
	p.weakBasicAuth = weakbasicauth.NewWeakBasicAuth(passwordList)
	p.weakJWT = weakjwt.NewWeakJWT(weakKeyList, sensitiveKeywords)
	if p.sensitive, err = sensitive.NewSensitive(p.config.rulesFilenames); err != nil {
		return nil, fmt.Errorf("unable to initialize Trace Analyzer Regexp Rules: %w", err)
	}

	return &p, nil
}

func parseFilenamesFromEnv(filenames string) []string {
	if filenames == "" {
		return []string{}
	}
	fns := strings.Split(filenames, ":")
	for i := range fns {
		fns[i] = strings.TrimSpace(fns[i])
	}

	return fns
}

func loadConfig() traceAnalyzerConfig {
	viper.SetDefault(dictFilenamesEnvVar, dictFilenamesDefault)
	viper.SetDefault(rulesFilenamesEnvVar, rulesFilenamesDefault)
	viper.SetDefault(sensitiveKeywordsFilenamesEnvVar, sensitiveKeywordsFilenamesDefault)
	viper.SetDefault(ignoreFindingsEnvVar, ignoreFindingsDefault)

	dictFilenames := parseFilenamesFromEnv(viper.GetString(dictFilenamesEnvVar))
	rulesFilenames := parseFilenamesFromEnv(viper.GetString(rulesFilenamesEnvVar))
	keywordsFilenames := parseFilenamesFromEnv(viper.GetString(sensitiveKeywordsFilenamesEnvVar))
	ignoreFindings := viper.GetStringSlice(ignoreFindingsEnvVar)
	modulesAssets := viper.GetString(config.ModulesAssetsEnvVar)

	var err error
	if modulesAssets != "" {
		if len(dictFilenames) == 0 {
			dictFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, moduleName, "dictionaries"))
			if err != nil {
				log.Warnf("There was problem while reading the Trace Analyzer assets directory 'dictionaries': %s", err)
			}
		}
		if len(rulesFilenames) == 0 {
			rulesFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, moduleName, "sensitive_rules"))
			if err != nil {
				log.Warnf("There was problem while reading the Trace Analyzer assets directory 'sensitive_rules': %s", err)
			}
		}
		if len(keywordsFilenames) == 0 {
			keywordsFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, moduleName, "sensitive_keywords"))
			if err != nil {
				log.Warnf("There was problem while reading the Trace Analyzer assets directory 'sensitive_keywords': %s", err)
			}
		}
	}

	c := traceAnalyzerConfig{
		dictFilenames:              dictFilenames,
		rulesFilenames:             rulesFilenames,
		sensitiveKeywordsFilenames: keywordsFilenames,
		ignoreFindings:             ignoreFindings,
	}
	return c
}

func (p *traceAnalyzer) Name() string {
	return moduleName
}

func (p *traceAnalyzer) HTTPHandler() http.Handler {
	return p.httpHandler
}

func (p *traceAnalyzer) EventNotify(ctx context.Context, e *core.Event) {
	event, trace := e.APIEvent, e.Telemetry
	log.Debugf("[TraceAnalyzer] received a new trace for API(%v) EventID(%v)", event.APIInfoID, event.ID)
	eventAnns := []core.Annotation{}
	apiAnns := []core.Annotation{}

	wbaEventAnns, wbaAPIAnns := p.weakBasicAuth.Analyze(trace)
	eventAnns = append(eventAnns, wbaEventAnns...)
	apiAnns = append(apiAnns, wbaAPIAnns...)

	wjtEventAnns, wjtAPIAnns := p.weakJWT.Analyze(trace)
	eventAnns = append(eventAnns, wjtEventAnns...)
	apiAnns = append(apiAnns, wjtAPIAnns...)

	sensEventAnns, sensAPIAnns := p.sensitive.Analyze(trace)
	eventAnns = append(eventAnns, sensEventAnns...)
	apiAnns = append(apiAnns, sensAPIAnns...)

	// If the status code starts with 2, it means that the request has been
	// accepted, hence, the parameters were accepted as well. So, we can look at
	// the parameters to see if they are very similar with the one in previous
	// accepted queries.
	if strings.HasPrefix(trace.Response.StatusCode, "2") {
		// Guessable ID, which is part of the module, not the 3rd party library
		specPath, pathParams, _, _, _ := p.getParams(ctx, event)
		if specPath == "" {
			specPath = trace.Request.Path
		}

		// Check for guessable IDs
		for pName, pValue := range pathParams {
			if guessable, reason := p.guessableID.IsGuessableParam("", pName, pValue); guessable {
				f := ParameterFinding{Location: specPath, Method: string(event.Method), Name: pName, Value: pValue, Reason: reason}
				bytes, err := json.Marshal(f)
				if err == nil {
					apiAnns = append(apiAnns, core.Annotation{Name: "GUESSABLE_ID", Annotation: bytes})
				}
			}
		}

		// Check for NLIDS
		eventNLIDAnns, _ := p.nlid.Analyze(pathParams, trace)
		for _, e := range eventNLIDAnns {
			f := ParameterFinding{Location: specPath, Method: string(event.Method), Name: "", Value: string(e.Annotation), Reason: nlid.Reason{}}
			bytes, err := json.Marshal(f)
			if err == nil {
				eventAnns = append(eventAnns, core.Annotation{Name: "NLID", Annotation: bytes})
			}
		}
	}

	filteredEventAnns := []core.Annotation{}
	for _, a := range eventAnns {
		if !p.ignoreFindings[a.Name] {
			filteredEventAnns = append(filteredEventAnns, a)
		}
	}
	if len(filteredEventAnns) > 0 {
		if err := p.accessor.CreateAPIEventAnnotations(ctx, p.Name(), event.ID, filteredEventAnns...); err != nil {
			log.Error(err)
		}
	}

	filteredAPIAnns := []core.Annotation{}
	for _, a := range apiAnns {
		if !p.ignoreFindings[a.Name] {
			filteredAPIAnns = append(filteredAPIAnns, a)
		}
	}
	if len(filteredAPIAnns) > 0 {
		if err := p.accessor.StoreAPIInfoAnnotations(ctx, p.Name(), event.APIInfoID, filteredAPIAnns...); err != nil {
			log.Error(err)
		}
	}

	p.setAlertSeverity(ctx, event.ID, filteredEventAnns)
}

func (p *traceAnalyzer) EventAnnotationNotify(modName string, eventID uint, ann core.Annotation) error {
	return nil
}

func (p *traceAnalyzer) APIAnnotationNotify(modName string, apiID uint, annotation *core.Annotation) error {
	return nil
}

func (p *traceAnalyzer) setAlertSeverity(ctx context.Context, eventID uint, anns []core.Annotation) {
	for _, a := range anns {
		f := getEventDescription(a)
		if f.Alert != nil {
			if err := p.accessor.CreateAPIEventAnnotations(ctx, p.Name(), eventID, *f.Alert); err != nil {
				log.Error(err)
			} else {
				break
			}
		}
	}
}

func getAPISpecsInfo(ctx context.Context, accessor core.BackendAccessor, apiID uint) (*models.OpenAPISpecs, error) {
	apiInfo, err := accessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("unable to get specification API '%d' for %w", apiID, err)
	}

	specsInfo := &models.OpenAPISpecs{}
	if apiInfo.ProvidedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provided spec info. info=%+v: %v", apiInfo.ProvidedSpecInfo, err)
		}
		specsInfo.ProvidedSpec = &specInfo
	}

	if apiInfo.ReconstructedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reconstructed spec info. info=%+v: %v", apiInfo.ReconstructedSpecInfo, err)
		}
		specsInfo.ReconstructedSpec = &specInfo
	}

	return specsInfo, nil
}

// XXX There are too many parameters to this function. It needs refactoring.
func (p *traceAnalyzer) getParams(ctx context.Context, event *database.APIEvent) (specPath string, pathParams map[string]string, queryParams map[string]string, headerParams map[string]string, bodyParams map[string]string) {
	specInfo, err := getAPISpecsInfo(ctx, p.accessor, event.APIInfoID)
	if err != nil {
		return "", nil, nil, nil, nil
	}

	pathParams = make(map[string]string)
	queryParams = make(map[string]string)
	headerParams = make(map[string]string)
	bodyParams = make(map[string]string)

	var spec *models.SpecInfo
	var eventPathID string
	// Prefer reconstructed spec
	if specInfo.ReconstructedSpec != nil {
		spec = specInfo.ReconstructedSpec
		eventPathID = event.ReconstructedPathID
	} else if specInfo.ProvidedSpec != nil {
		spec = specInfo.ProvidedSpec
		eventPathID = event.ProvidedPathID
	}

	if spec != nil {
		for _, t := range spec.Tags {
			for _, path := range t.MethodAndPathList {
				if path.PathID.String() == eventPathID && path.Method == event.Method {
					specPath = path.Path
					pathParams = utils.GetPathParams(path.Path, event.Path)
					// XXX Need to get other parameters
					break
				}
			}
		}
	}

	return specPath, pathParams, queryParams, headerParams, bodyParams
}

type httpHandler struct {
	ta *traceAnalyzer
}

func (h httpHandler) GetEventAnnotations(w http.ResponseWriter, r *http.Request, eventID int64) {
	dbAnns, err := h.ta.accessor.ListAPIEventAnnotations(r.Context(), moduleName, uint(eventID))
	if err != nil {
		return
	}
	annList := []Annotation{}

	for _, a := range dbAnns {
		f := getEventDescription(*a)

		annList = append(annList, Annotation{
			Annotation: f.DetailedDesc,
			Name:       f.ShortDesc,
			Severity:   f.Severity,
			Kind:       a.Name,
		})
	}
	result := Annotations{
		Items: &annList,
		Total: len(annList),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h httpHandler) GetAPIAnnotations(w http.ResponseWriter, r *http.Request, apiID int64) {
	dbAnns, err := h.ta.accessor.ListAPIInfoAnnotations(r.Context(), moduleName, uint(apiID))
	if err != nil {
		return
	}
	annList := []Annotation{}

	for _, a := range dbAnns {
		f := getAPIDescription(*a)
		annList = append(annList, Annotation{
			Annotation: f.DetailedDesc,
			Name:       f.ShortDesc,
			Severity:   f.Severity,
			Kind:       a.Name,
		})
	}
	result := Annotations{
		Items: &annList,
		Total: len(annList),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h httpHandler) DeleteAPIAnnotations(w http.ResponseWriter, r *http.Request, apiID int64, params DeleteAPIAnnotationsParams) {
	err := h.ta.accessor.DeleteAPIInfoAnnotations(r.Context(), moduleName, uint(apiID), params.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//nolint:gochecknoinits
func init() {
	core.RegisterModule(newTraceAnalyzer)
}
