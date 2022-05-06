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

package sensitive

import (
	"fmt"
	"os"
	"regexp"

	yaml "gopkg.in/yaml.v3"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	AuthorizationHeader = "sensitive"

	SearchInRequestHeaders  = "RequestHeaders"
	SearchInResponseHeaders = "ResponseHeaders"
	SearchInRequestBody     = "RequestBody"
	SearchInResponseBody    = "ResponseBody"
)

const (
	RexgexpMatchingRequestBody     = "REGEXP_MATCHING_REQUEST_BODY"
	RexgexpMatchingResponseBody    = "REGEXP_MATCHING_RESPONSE_BODY"
	RegexpMatchingRequestHeaders   = "REGEXP_MATCHING_REQUEST_HEADERS"
	RexgexpMatchingResponseHeaders = "REGEXP_MATCHING_RESPONSE_HEADERS"
)

type Rule struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Regex       string   `yaml:"regex"`
	SearchIn    []string `yaml:"searchIn"`

	CompiledRegex *regexp.Regexp
}

type Sensitive struct {
	Rules []Rule `yaml:"rules"`
}

func isValidSearchIn(searchIn []string) bool {
	for _, s := range searchIn {
		if s != SearchInRequestHeaders &&
			s != SearchInResponseHeaders &&
			s != SearchInRequestBody &&
			s != SearchInResponseBody {
			return false
		}
	}

	return true
}

func loadRules(filename string) (rules []Rule, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file '%s': %w", filename, err)
	}
	defer f.Close()

	if err = yaml.NewDecoder(f).Decode(&rules); err != nil {
		return nil, fmt.Errorf("rule file '%s' is not a valid yaml file: '%w'", filename, err)
	}

	// Check validity and compile all regexs
	for i, r := range rules {
		if !isValidSearchIn(r.SearchIn) {
			return nil, fmt.Errorf("in rule file '%s', the searchIn Location (%v) is not valid", filename, r.SearchIn)
		}
		rules[i].CompiledRegex, err = regexp.Compile(r.Regex)
		if err != nil {
			return nil, fmt.Errorf("in rule file '%s', unable to compile regexp: %w", filename, err)
		}
	}

	return rules, nil
}

func NewSensitive(rulesFilenames []string) (*Sensitive, error) {
	allRules := []Rule{}
	for _, filename := range rulesFilenames {
		rules, err := loadRules(filename)
		if err != nil {
			return nil, err
		}
		allRules = append(allRules, rules...)
	}

	return &Sensitive{
		Rules: allRules,
	}, nil
}

func (w *Sensitive) applyRuleHeaders(headers []*models.Header, rule Rule) bool {
	for _, h := range headers {
		for _, value := range []string{h.Key, h.Value} {
			if rule.CompiledRegex.MatchString(value) {
				return true
			}
		}
	}

	return false
}

func (w *Sensitive) applyRule(trace *models.Telemetry, rule Rule) []core.Annotation {
	anns := []core.Annotation{}

	for _, where := range rule.SearchIn {
		switch where {
		case SearchInRequestBody:
			if rule.CompiledRegex.Match(trace.Request.Common.Body) {
				anns = append(anns, core.Annotation{
					Name:       RexgexpMatchingRequestBody,
					Annotation: []byte(rule.ID + ":" + rule.Description),
				})
			}
		case SearchInResponseBody:
			if rule.CompiledRegex.Match(trace.Response.Common.Body) {
				anns = append(anns, core.Annotation{
					Name:       RexgexpMatchingResponseBody,
					Annotation: []byte(rule.ID + ":" + rule.Description),
				})
			}
		case SearchInRequestHeaders:
			if w.applyRuleHeaders(trace.Request.Common.Headers, rule) {
				anns = append(anns, core.Annotation{
					Name:       RegexpMatchingRequestHeaders,
					Annotation: []byte(rule.ID + ":" + rule.Description),
				})
			}
		case SearchInResponseHeaders:
			if w.applyRuleHeaders(trace.Response.Common.Headers, rule) {
				anns = append(anns, core.Annotation{
					Name:       RexgexpMatchingResponseHeaders,
					Annotation: []byte(rule.ID + ":" + rule.Description),
				})
			}
		}
	}

	return anns
}

func (w *Sensitive) analyzeSensitive(trace *models.Telemetry) (anns []core.Annotation) {
	for _, rule := range w.Rules {
		obs := w.applyRule(trace, rule)
		anns = append(anns, obs...)
	}

	return anns
}

func (w *Sensitive) Analyze(trace *models.Telemetry) (eventAnns []core.Annotation, apiAnns []core.Annotation) {
	eventAnns = append(eventAnns, w.analyzeSensitive(trace)...)

	return eventAnns, apiAnns
}
