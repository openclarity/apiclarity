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
	"encoding/json"
	"fmt"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	RegexpMatching = "REGEXP_MATCHING"
)

type AnnotationRegexpMatching struct {
	Matches []RuleMatch `json:"matches"`
}

func NewAnnotationRegexpMatching(matches []RuleMatch) *AnnotationRegexpMatching {
	return &AnnotationRegexpMatching{
		Matches: matches,
	}
}
func (a *AnnotationRegexpMatching) Name() string               { return RegexpMatching }
func (a *AnnotationRegexpMatching) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationRegexpMatching) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationRegexpMatching) Deserialize(serialized []byte) error {
	var tmp AnnotationRegexpMatching
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationRegexpMatching) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationRegexpMatching) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Matching regular expression",
		DetailedDesc: fmt.Sprintf("This event matches sensitive information"),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
