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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/jsonpointer"

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
	SpecLocation string               `json:"spec_location"`
	Params       []GuessableParameter `json:"parameters"`
}

func NewAnnotationGuessableID(path, method string, parameters []GuessableParameter) *AnnotationGuessableID {
	pointerTokens := []string{
		jsonpointer.Escape("paths"),
		jsonpointer.Escape(path),
		jsonpointer.Escape(strings.ToLower(method)),
	}
	pointer := strings.Join(pointerTokens, "/")

	return &AnnotationGuessableID{
		SpecLocation: pointer,
		Params:       parameters,
	}
}
func (a *AnnotationGuessableID) Name() string { return GuessableType }
func (a *AnnotationGuessableID) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return nil
}
func (a *AnnotationGuessableID) Severity() string           { return utils.SeverityInfo }
func (a *AnnotationGuessableID) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationGuessableID) Deserialize(serialized []byte) error {
	var tmp AnnotationGuessableID
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
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
		DetailedDesc: fmt.Sprintf("In call '%s', parameter(s) '%s' seems to be guessable", a.SpecLocation, strings.Join(paramNames, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
