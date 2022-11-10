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

package nlid

import (
	"fmt"
	"strings"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	NLIDType = "NLID"
)

type AnnotationNLID struct {
	Params []parameter `json:"parameters"`
}

func NewAnnotationNLID(path, method string, parameters []parameter) *AnnotationNLID {
	return &AnnotationNLID{
		Params: parameters,
	}
}
func (a *AnnotationNLID) Name() string { return NLIDType }
func (a AnnotationNLID) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationNLID(path, method)
}
func (a *AnnotationNLID) Severity() string { return utils.SeverityInfo }

func (a AnnotationNLID) Redacted() utils.TraceAnalyzerAnnotation {
	newA := a
	for i := range newA.Params {
		newA.Params[i].Value = "[redacted]"
	}
	return &newA
}

func (a *AnnotationNLID) ToFinding() utils.Finding {
	paramValues := []string{}
	for _, p := range a.Params {
		paramValues = append(paramValues, p.Value)
	}

	return utils.Finding{
		ShortDesc:    "NLID (Non learnt Identifier)",
		DetailedDesc: fmt.Sprintf("Parameter(s) '%s' were used but not previously retrieved. Potential BOLA.", strings.Join(paramValues, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type APIAnnotationNLID struct {
	utils.BaseTraceAnalyzerAPIAnnotation
	ParamNames map[string]bool `json:"parameters"`
}

func NewAPIAnnotationNLID(path, method string) *APIAnnotationNLID {
	return &APIAnnotationNLID{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
		ParamNames:                     make(map[string]bool),
	}
}
func (a *APIAnnotationNLID) Name() string { return NLIDType }
func (a *APIAnnotationNLID) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	initialSize := len(a.ParamNames)
	eventAnn, valid := ann.(*AnnotationNLID)
	if !valid {
		panic("invalid type")
	}
	// Merge parameter
	for _, p := range eventAnn.Params {
		a.ParamNames[p.Name] = true
	}

	return initialSize != len(a.ParamNames)
}

func (a APIAnnotationNLID) Severity() string { return utils.SeverityMedium }

func (a APIAnnotationNLID) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}

func (a *APIAnnotationNLID) ToAPIFinding() oapicommon.APIFinding {
	var additionalInfo *map[string]interface{}
	if len(a.ParamNames) > 0 {
		paramNames := []string{}
		for name := range a.ParamNames {
			paramNames = append(paramNames, name)
		}
		additionalInfo = &map[string]interface{}{
			"parameters": paramNames,
		}
	}
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "NLID (Non learnt Identifier)",
		Description: "Parameters were used but not previously retrieved. Potential BOLA",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.MEDIUM,

		AdditionalInfo: additionalInfo,
	}
}
