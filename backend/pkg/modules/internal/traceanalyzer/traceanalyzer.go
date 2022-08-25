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
	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/guessableid"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/nlid"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/sensitive"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/weakbasicauth"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/weakjwt"
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

	aggregator *APIsFindingsRepo

	accessor core.BackendAccessor
	info     *core.ModuleInfo
}

func newTraceAnalyzer(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	var err error

	p := traceAnalyzer{
		info: &core.ModuleInfo{
			Name:        utils.ModuleName,
			Description: utils.ModuleDescription,
		},
	}
	h := restapi.HandlerWithOptions(&httpHandler{ta: &p}, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + utils.ModuleName})
	p.httpHandler = h
	p.ignoreFindings = map[string]bool{}
	p.accessor = accessor

	p.config = loadConfig()
	log.Debugf("traceanalyzer Configuration: %+v", p.config)

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

	p.aggregator = NewAPIsFindingsRepo(p.accessor)

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
			dictFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, utils.ModuleName, "dictionaries"))
			if err != nil {
				log.Warnf("There was problem while reading the Trace Analyzer assets directory 'dictionaries': %s", err)
			}
		}
		if len(rulesFilenames) == 0 {
			rulesFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, utils.ModuleName, "sensitive_rules"))
			if err != nil {
				log.Warnf("There was problem while reading the Trace Analyzer assets directory 'sensitive_rules': %s", err)
			}
		}
		if len(keywordsFilenames) == 0 {
			keywordsFilenames, err = utils.WalkFiles(filepath.Join(modulesAssets, utils.ModuleName, "sensitive_keywords"))
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

func (p *traceAnalyzer) Info() core.ModuleInfo {
	return *p.info
}

func (p *traceAnalyzer) HTTPHandler() http.Handler {
	return p.httpHandler
}

func (p *traceAnalyzer) EventNotify(ctx context.Context, e *core.Event) {
	event, trace := e.APIEvent, e.Telemetry
	log.Debugf("[traceanalyzer] received a new trace for API(%v) EventID(%v)", event.APIInfoID, event.ID)
	eventAnns := []utils.TraceAnalyzerAnnotation{}

	wbaEventAnns := p.weakBasicAuth.Analyze(trace)
	eventAnns = append(eventAnns, wbaEventAnns...)

	wjtEventAnns := p.weakJWT.Analyze(trace)
	eventAnns = append(eventAnns, wjtEventAnns...)

	sensEventAnns := p.sensitive.Analyze(trace)
	eventAnns = append(eventAnns, sensEventAnns...)

	// If the status code starts with 2, it means that the request has been
	// accepted, hence, the parameters were accepted as well. So, we can look at
	// the parameters to see if they are very similar with the one in previous
	// accepted queries.
	// FIXME: Performance KILLER. For each request, this function is called, which calls the database a deserializes data to get the specinfo
	// FIXME: We MUST create a memory cache map[apiid]specinfo to avoid that
	specPath, pathParams, _, err := p.getParams(ctx, event)
	// if specPath == "" {
	// 	specPath = trace.Request.Path
	// }
	if err == nil && strings.HasPrefix(trace.Response.StatusCode, "2") {
		// Check for guessable IDs
		eventGuessable, _ := p.guessableID.Analyze(specPath, string(event.Method), pathParams, trace)
		eventAnns = append(eventAnns, eventGuessable...)

		// Check for NLIDS
		eventNLIDAnns, _ := p.nlid.Analyze(specPath, string(event.Method), pathParams, trace)
		eventAnns = append(eventAnns, eventNLIDAnns...)
	}

	// Filter ignored findings for event annotations
	filteredEventAnns := []utils.TraceAnalyzerAnnotation{}
	for _, a := range eventAnns {
		if !p.ignoreFindings[a.Name()] {
			filteredEventAnns = append(filteredEventAnns, a)
		}
	}

	if len(filteredEventAnns) > 0 {
		coreEventAnnotations := p.toCoreEventAnnotations(filteredEventAnns, false)
		if err := p.accessor.CreateAPIEventAnnotations(ctx, utils.ModuleName, event.ID, coreEventAnnotations...); err != nil {
			log.Error(err)
		}
		p.setAlertSeverity(ctx, event.ID, filteredEventAnns)

		updatedFindings := p.aggregator.Aggregate(uint64(event.APIInfoID), specPath, trace.Request.Method, filteredEventAnns...)
		// If API findings were updated:
		// - only store the new ones in the database
		// - send notification will ALL API findings
		if len(updatedFindings) > 0 {
			allAPIFindings := p.aggregator.GetAPIFindings(uint64(event.APIInfoID))
			err := p.sendAPIFindingsNotification(ctx, event.APIInfoID, allAPIFindings)
			if err != nil {
				log.Error(err)
			}
			// For each finding type that were updated, take all findings of
			// that type
			findingsToStore := []utils.TraceAnalyzerAPIAnnotation{}
			for _, uFinding := range updatedFindings {
				for _, finding := range allAPIFindings {
					if finding.Name() == uFinding.Name() {
						findingsToStore = append(findingsToStore, finding)
					}
				}
			}
			coreAPIAnnotations := p.toCoreAPIAnnotations(findingsToStore, false)
			if err := p.accessor.StoreAPIInfoAnnotations(ctx, utils.ModuleName, event.APIInfoID, coreAPIAnnotations...); err != nil {
				log.Error(err)
			}
		}
	}
}

func (p *traceAnalyzer) toCoreEventAnnotations(eventAnns []utils.TraceAnalyzerAnnotation, redacted bool) (coreAnnotations []core.Annotation) {
	for _, a := range eventAnns {
		if redacted {
			a = a.Redacted()
		}
		annotation, err := json.Marshal(a)
		if err != nil {
			log.Errorf("unable to serialize annotation: %s", err)
		}
		coreAnnotations = append(coreAnnotations, core.Annotation{Name: a.Name(), Annotation: annotation})
	}
	return coreAnnotations
}

func fromCoreEventAnnotation(coreAnn *core.Annotation) (ann utils.TraceAnalyzerAnnotation, err error) {
	var a utils.TraceAnalyzerAnnotation
	switch coreAnn.Name {
	case weakbasicauth.KindKnownPassword:
		a = &weakbasicauth.AnnotationKnownPassword{}
	case weakbasicauth.KindShortPassword:
		a = &weakbasicauth.AnnotationShortPassword{}
	case weakbasicauth.KindSamePassword:
		a = &weakbasicauth.AnnotationSamePassword{}

	case weakjwt.JWTNoAlgField:
		a = &weakjwt.AnnotationNoAlgField{}
	case weakjwt.JWTAlgFieldNone:
		a = &weakjwt.AnnotationAlgFieldNone{}
	case weakjwt.JWTNotRecommendedAlg:
		a = &weakjwt.AnnotationNotRecommendedAlg{}
	case weakjwt.JWTNoExpireClaim:
		a = &weakjwt.AnnotationNoExpireClaim{}
	case weakjwt.JWTExpTooFar:
		a = &weakjwt.AnnotationExpTooFar{}
	case weakjwt.JWTWeakSymetricSecret:
		a = &weakjwt.AnnotationWeakSymetricSecret{}
	case weakjwt.JWTSensitiveContent:
		a = &weakjwt.AnnotationSensitiveContent{}

	case sensitive.RegexpMatchingType:
		a = &sensitive.AnnotationRegexpMatching{}

	case nlid.NLIDType:
		a = &nlid.AnnotationNLID{}

	case guessableid.GuessableType:
		a = &guessableid.AnnotationGuessableID{}

	default:
		return nil, fmt.Errorf("unknown annotation '%s'", coreAnn.Name)
	}

	err = json.Unmarshal(coreAnn.Annotation, a)
	if err != nil {
		return a, fmt.Errorf("unable to convert trace analyzer annotation in database: %w", err)
	}
	return a, nil
}

func fromCoreEventAnnotations(coreAnns []*core.Annotation) (anns []utils.TraceAnalyzerAnnotation) {
	for _, coreAnn := range coreAnns {
		taAnn, err := fromCoreEventAnnotation(coreAnn)
		if err != nil {
			log.Errorf("Unable to understand annotation: %v", err)
		} else {
			anns = append(anns, taAnn)
		}
	}

	return anns
}

func (p *traceAnalyzer) toCoreAPIAnnotations(anns []utils.TraceAnalyzerAPIAnnotation, redacted bool) (coreAnnotations []core.Annotation) {
	// In order to create Core Annotations, we need to group APIAnnotations by Name of annotations
	groupedAnns := map[string][]utils.TraceAnalyzerAPIAnnotation{}
	for _, a := range anns {
		if redacted {
			a = a.Redacted()
		}
		groupedAnns[a.Name()] = append(groupedAnns[a.Name()], a)
	}
	for annName, anns := range groupedAnns {
		annotation, err := json.Marshal(anns)
		if err != nil {
			log.Errorf("unable to serialize annotation: %s", err)
		}
		coreAnnotations = append(coreAnnotations, core.Annotation{Name: annName, Annotation: annotation})
	}
	return coreAnnotations
}

func fromCoreAPIAnnotation(coreAnn *core.Annotation) (anns []utils.TraceAnalyzerAPIAnnotation, err error) {
	var rawAnnotations []json.RawMessage
	if err := json.Unmarshal(coreAnn.Annotation, &rawAnnotations); err != nil {
		return anns, fmt.Errorf("unable to convert API Annotation to DB representation: %w", err)
	}

	var a utils.TraceAnalyzerAPIAnnotation
	switch coreAnn.Name {
	case weakbasicauth.KindShortPassword:
		a = &weakbasicauth.APIAnnotationShortPassword{}
	case weakbasicauth.KindKnownPassword:
		a = &weakbasicauth.APIAnnotationKnownPassword{}
	case weakbasicauth.KindSamePassword:
		a = &weakbasicauth.APIAnnotationSamePassword{}

	case sensitive.RegexpMatchingType:
		a = &sensitive.APIAnnotationRegexpMatching{}

	case weakjwt.JWTNoAlgField:
		a = &weakjwt.APIAnnotationNoAlgField{}

	case weakjwt.JWTAlgFieldNone:
		a = &weakjwt.APIAnnotationAlgFieldNone{}
	case weakjwt.JWTNotRecommendedAlg:
		a = &weakjwt.APIAnnotationNotRecommendedAlg{}
	case weakjwt.JWTNoExpireClaim:
		a = &weakjwt.APIAnnotationNoExpireClaim{}
	case weakjwt.JWTExpTooFar:
		a = &weakjwt.APIAnnotationExpTooFar{}
	case weakjwt.JWTWeakSymetricSecret:
		a = &weakjwt.APIAnnotationWeakSymetricSecret{}
	case weakjwt.JWTSensitiveContent:
		a = &weakjwt.APIAnnotationSensitiveContent{}

	case nlid.NLIDType:
		a = &nlid.APIAnnotationNLID{}
	case guessableid.GuessableType:
		a = &guessableid.APIAnnotationGuessableID{}

	default:
		return nil, fmt.Errorf("unknown annotation '%s'", coreAnn.Name)
	}

	for _, rawJSON := range rawAnnotations {
		if err := json.Unmarshal(rawJSON, a); err != nil {
			log.Errorf("Unable to unmarshal one of the %s annotations", coreAnn.Name)
			log.Debugf("Unable to unmarshal annotation of type %s %s", coreAnn.Name, string(rawJSON))
		} else {
			anns = append(anns, a)
		}
	}

	return anns, nil
}

func fromCoreAPIAnnotations(coreAnns []*core.Annotation) (anns []utils.TraceAnalyzerAPIAnnotation) {
	for _, coreAnn := range coreAnns {
		taAnns, err := fromCoreAPIAnnotation(coreAnn)
		if err != nil {
			log.Errorf("Unable to understand annotation: %v", err)
		} else {
			anns = append(anns, taAnns...)
		}
	}

	return anns
}

func (p *traceAnalyzer) sendAPIFindingsNotification(ctx context.Context, apiID uint, apiFindings []utils.TraceAnalyzerAPIAnnotation) error {
	apiN := notifications.ApiFindingsNotification{
		NotificationType: "ApiFindingsNotification",
		Items:            &[]oapicommon.APIFinding{},
	}

	for _, finding := range apiFindings {
		*(apiN.Items) = append(*(apiN.Items), finding.ToAPIFinding())
	}

	n := notifications.APIClarityNotification{}
	if err := n.FromApiFindingsNotification(apiN); err != nil {
		return fmt.Errorf("unable serialize notification: %w", err)
	}

	err := p.accessor.Notify(ctx, utils.ModuleName, apiID, n)
	if err != nil {
		return fmt.Errorf("unable to send notification: %w", err)
	}

	return nil
}

func (p *traceAnalyzer) EventAnnotationNotify(modName string, eventID uint, ann core.Annotation) error {
	return nil
}

func (p *traceAnalyzer) APIAnnotationNotify(modName string, apiID uint, annotation *core.Annotation) error {
	return nil
}

func (p *traceAnalyzer) setAlertSeverity(ctx context.Context, eventID uint, anns []utils.TraceAnalyzerAnnotation) {
	maxAlert := core.AlertInfo
	for _, a := range anns {
		alert := utils.SeverityToAlert(a.Severity())
		if alert > maxAlert {
			maxAlert = alert
		}
		// We reach the maximum alert level, not need to go further
		if maxAlert == core.AlertCritical {
			break
		}
	}

	var alertAnn core.Annotation
	switch maxAlert {
	case core.AlertInfo:
		alertAnn = core.AlertInfoAnn
	case core.AlertWarn:
		alertAnn = core.AlertWarnAnn
	case core.AlertCritical:
		alertAnn = core.AlertCriticalAnn
	}

	if err := p.accessor.CreateAPIEventAnnotations(ctx, utils.ModuleName, eventID, alertAnn); err != nil {
		log.Error(err)
	}
}

func (p *traceAnalyzer) getParams(ctx context.Context, event *database.APIEvent) (specPath string, pathParams map[string]string, queryParams map[string]string, err error) {
	apiInfo, err := p.accessor.GetAPIInfo(ctx, event.APIInfoID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("unable to get API Information from database: %w", err)
	}

	// Prefer Provided specification if available
	var serializedSpecInfo *string
	var eventPathID string
	if apiInfo.HasProvidedSpec && apiInfo.ProvidedSpecInfo != "" {
		serializedSpecInfo = &apiInfo.ProvidedSpecInfo
		eventPathID = event.ProvidedPathID
	} else if apiInfo.HasReconstructedSpec && apiInfo.ReconstructedSpecInfo != "" {
		serializedSpecInfo = &apiInfo.ReconstructedSpecInfo
		eventPathID = event.ReconstructedPathID
	} else {
		return specPath, pathParams, queryParams, nil
	}

	var specInfo models.SpecInfo
	if err := json.Unmarshal([]byte(*serializedSpecInfo), &specInfo); err != nil {
		return specPath, pathParams, queryParams, fmt.Errorf("failed to unmarshal spec info for api=%d: %w", event.APIInfoID, err)
	}

	pathParams = make(map[string]string)
	queryParams = make(map[string]string)

	for _, t := range specInfo.Tags {
		for _, path := range t.MethodAndPathList {
			if path.PathID.String() == eventPathID && path.Method == event.Method {
				specPath = path.Path
				pathParams = utils.GetPathParams(path.Path, event.Path)
				// XXX Need to get other parameters
				break
			}
		}
	}

	return specPath, pathParams, queryParams, nil
}

func (p *traceAnalyzer) getAPIFindings(ctx context.Context, apiID uint, sensitive bool) (apiFindings []oapicommon.APIFinding, err error) {
	dbAnns, err := p.accessor.ListAPIInfoAnnotations(ctx, utils.ModuleName, apiID)
	if err != nil {
		return apiFindings, fmt.Errorf("unable to get list of API annotations: %w", err)
	}

	anns := fromCoreAPIAnnotations(dbAnns)
	for _, ann := range anns {
		var f oapicommon.APIFinding
		if sensitive {
			f = ann.ToAPIFinding()
		} else {
			f = ann.Redacted().ToAPIFinding()
		}
		apiFindings = append(apiFindings, f)
	}

	return apiFindings, nil
}

type httpHandler struct {
	ta *traceAnalyzer
}

func (h httpHandler) GetEventAnnotations(w http.ResponseWriter, r *http.Request, eventID int64, params restapi.GetEventAnnotationsParams) {
	dbAnns, err := h.ta.accessor.ListAPIEventAnnotations(r.Context(), utils.ModuleName, uint(eventID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: "Internal error, could not read data from database"})
		return
	}
	annList := []restapi.Annotation{}

	taAnns := fromCoreEventAnnotations(dbAnns)
	for _, a := range taAnns {
		if params.Redacted != nil && *params.Redacted {
			a = a.Redacted()
		}
		f := a.ToFinding()
		annList = append(annList, restapi.Annotation{
			Annotation: f.DetailedDesc,
			Name:       f.ShortDesc,
			Severity:   f.Severity,
			Kind:       a.Name(),
		})
	}

	result := restapi.Annotations{
		Items: &annList,
		Total: len(annList),
	}

	common.HTTPResponse(w, http.StatusOK, result)
}

func (h httpHandler) GetApiFindings(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.GetApiFindingsParams) { //nolint:revive,stylecheck
	// If sensitive parameter is not set, default to false (ie: do not include sensitive data)
	sensitive := params.Sensitive != nil && *params.Sensitive
	apiFindings, err := h.ta.getAPIFindings(r.Context(), uint(apiID), sensitive)
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: "Internal error, could not read data from database"})
		return
	}

	if len(apiFindings) == 0 {
		apiFindings = make([]oapicommon.APIFinding, 0)
	}

	apiFindingsObject := oapicommon.APIFindings{
		Items: &apiFindings,
	}
	common.HTTPResponse(w, http.StatusOK, apiFindingsObject)
}

func (h httpHandler) StartTraceAnalysis(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.ta.accessor.EnableTraces(r.Context(), utils.ModuleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Tracing successfully started for api=%d", apiID)
	common.HTTPResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Trace analysis successfully started for api %d", apiID)})
}

func (h httpHandler) StopTraceAnalysis(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.ta.accessor.DisableTraces(r.Context(), utils.ModuleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Tracing successfully stopped for api=%d", apiID)

	common.HTTPResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Trace analysis stopped for api %d", apiID)})
}

//nolint:revive,stylecheck // Api is not uppercased because it's defined as is in the specification
func (h httpHandler) ResetApiFindings(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	h.ta.aggregator.ResetAPIFindings(uint64(apiID))

	err := h.ta.accessor.DeleteAllAPIInfoAnnotations(r.Context(), utils.ModuleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, oapicommon.ApiResponse{Message: "Internal error, could not delete data from database"})
		return
	}

	log.Infof("API Findings successfully reset for api=%d", apiID)
	common.HTTPResponse(w, http.StatusNoContent, nil)
}

//nolint:gochecknoinits
func init() {
	core.RegisterModule(newTraceAnalyzer)
}
