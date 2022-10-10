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
	"strings"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	RegexpMatchingType = "REGEXP_MATCHING"
)

type AnnotationRegexpMatching struct {
	Matches []RuleMatch `json:"matches"`
}

func NewAnnotationRegexpMatching(matches []RuleMatch) *AnnotationRegexpMatching {
	return &AnnotationRegexpMatching{
		Matches: matches,
	}
}
func (a *AnnotationRegexpMatching) Name() string { return RegexpMatchingType }
func (a *AnnotationRegexpMatching) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationRegexpMatching(path, method)
}
func (a *AnnotationRegexpMatching) Severity() string { return utils.SeverityMedium }
func (a AnnotationRegexpMatching) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}

func (a *AnnotationRegexpMatching) ToFinding() utils.Finding {
	matchingRules := []string{}
	for _, r := range a.Matches {
		matchingRules = append(matchingRules, r.Rule.Description)
	}
	return utils.Finding{
		ShortDesc:    "Matching regular expression",
		DetailedDesc: fmt.Sprintf("This event matches sensitive information (rules: %s)", strings.Join(matchingRules, ", ")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type APIAnnotationRegexpMatching struct {
	utils.BaseTraceAnalyzerAPIAnnotation
	MatchingRules map[string]string `json:"matching_rules_id"`
}

func NewAPIAnnotationRegexpMatching(path, method string) *APIAnnotationRegexpMatching {
	return &APIAnnotationRegexpMatching{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
		MatchingRules:                  make(map[string]string),
	}
}
func (a *APIAnnotationRegexpMatching) Name() string { return RegexpMatchingType }
func (a *APIAnnotationRegexpMatching) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	eventAnn, valid := ann.(*AnnotationRegexpMatching)
	if !valid {
		panic("invalid type")
	}

	initialSize := len(a.MatchingRules)
	for _, r := range eventAnn.Matches {
		a.MatchingRules[r.Rule.ID] = r.Rule.Description
	}

	return initialSize != len(a.MatchingRules)
}

func (a APIAnnotationRegexpMatching) Severity() string { return utils.SeverityHigh }

func (a APIAnnotationRegexpMatching) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}

func (a *APIAnnotationRegexpMatching) ToAPIFinding() oapicommon.APIFinding {
	var additionalInfo *map[string]interface{}
	if len(a.MatchingRules) > 0 {
		matchingRules := []string{}
		for r := range a.MatchingRules {
			matchingRules = append(matchingRules, r)
		}
		additionalInfo = &map[string]interface{}{
			"matching_rules": matchingRules,
		}
	}
	jsonPointer := a.SpecLocation()
	matchingDescs := []string{}
	for _, desc := range a.MatchingRules {
		matchingDescs = append(matchingDescs, desc)
	}
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "Matching regular expression",
		Description: fmt.Sprintf("This operation matches sensitive information. The rules that are matching are %s", strings.Join(matchingDescs, ", ")),

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.HIGH,

		AdditionalInfo: additionalInfo,
	}
}
