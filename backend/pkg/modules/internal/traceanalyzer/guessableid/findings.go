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

package guessableid

import (
	"fmt"
	"strings"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	GuessableType = "GUESSABLE_ID"
)

type GuessableReason struct {
	Distance             float32 `json:"distance"`
	DistanceThreshold    float32 `json:"distance_threshold"`
	CompressionRatio     float32 `json:"compression_ratio"`
	CompressionThreshold float32 `json:"compression_threshold"`
}

type GuessableParameter struct {
	Name   string          `json:"name"`
	Value  string          `json:"value"`
	Reason GuessableReason `json:"reason"`
}

type AnnotationGuessableID struct {
	Params []GuessableParameter `json:"parameters"`
}

func NewAnnotationGuessableID(path, method string, parameters []GuessableParameter) *AnnotationGuessableID {
	return &AnnotationGuessableID{
		Params: parameters,
	}
}
func (a *AnnotationGuessableID) Name() string { return GuessableType }
func (a *AnnotationGuessableID) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationGuessableID(path, method)
}
func (a *AnnotationGuessableID) Severity() string { return utils.SeverityInfo }
func (a AnnotationGuessableID) Redacted() utils.TraceAnalyzerAnnotation {
	newA := a
	for i := range newA.Params {
		newA.Params[i].Value = "[redacted]"
	}
	return &newA
}

func (a *AnnotationGuessableID) ToFinding() utils.Finding {
	paramNames := []string{}
	for _, p := range a.Params {
		paramNames = append(paramNames, p.Value)
	}

	return utils.Finding{
		ShortDesc:    "Guessable identifier",
		DetailedDesc: fmt.Sprintf("Parameter(s) '%s' seems to be guessable", strings.Join(paramNames, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type APIAnnotationGuessableID struct {
	utils.BaseTraceAnalyzerAPIAnnotation
	ParamNames map[string]bool `json:"parameters"`
}

func NewAPIAnnotationGuessableID(path, method string) *APIAnnotationGuessableID {
	return &APIAnnotationGuessableID{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
		ParamNames:                     make(map[string]bool),
	}
}
func (a *APIAnnotationGuessableID) Name() string { return GuessableType }
func (a *APIAnnotationGuessableID) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	apiAnn, valid := ann.(*AnnotationGuessableID)
	if !valid {
		panic("invalid type")
	}
	initialSize := len(a.ParamNames)
	// Merge parameter
	for _, p := range apiAnn.Params {
		a.ParamNames[p.Name] = true
	}

	return initialSize != len(a.ParamNames)
}

func (a APIAnnotationGuessableID) Severity() string { return utils.SeverityMedium }

func (a APIAnnotationGuessableID) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}

func (a *APIAnnotationGuessableID) ToAPIFinding() oapicommon.APIFinding {
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
		Name:        "Guessable identifier",
		Description: "Parameters seems to be guessable",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.MEDIUM,

		AdditionalInfo: additionalInfo,
	}
}
