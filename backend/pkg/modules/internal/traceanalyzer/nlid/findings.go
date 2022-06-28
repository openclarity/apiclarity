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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/jsonpointer"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	NLIDType = "NLID"
)

type AnnotationNLID struct {
	SpecLocation string      `json:"spec_location"`
	Params       []parameter `json:"parameters"`
}

func NewAnnotationNLID(path, method string, parameters []parameter) *AnnotationNLID {
	pointerTokens := []string{
		jsonpointer.Escape("paths"),
		jsonpointer.Escape(path),
		jsonpointer.Escape(strings.ToLower(method)),
	}
	pointer := strings.Join(pointerTokens, "/")

	return &AnnotationNLID{
		SpecLocation: pointer,
		Params:       parameters,
	}
}
func (a *AnnotationNLID) Name() string               { return NLIDType }
func (a *AnnotationNLID) Severity() string           { return utils.SeverityInfo }
func (a *AnnotationNLID) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationNLID) Deserialize(serialized []byte) error {
	var tmp AnnotationNLID
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
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
		DetailedDesc: fmt.Sprintf("In call '%s', parameter(s) '%s' were used but not previously retrieved. Potential BOLA.", a.SpecLocation, strings.Join(paramValues, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
