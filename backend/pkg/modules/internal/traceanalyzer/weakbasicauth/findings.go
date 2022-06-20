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

package weakbasicauth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	//nolint:gosec
	KindShortPassword = "BASIC_AUTH_SHORT_PASSWORD"
	//nolint:gosec
	KindKnownPassword = "BASIC_AUTH_KNOWN_PASSWORD"
	//nolint:gosec
	KindSamePassword = "BASIC_AUTH_SAME_PASSWORD"
)

type AnnotationShortPassword struct {
	Password string `json:"password"`
	Length   int    `json:"length"`
	MinSize  int    `json:"min_size"`
}

func NewAnnotationShortPassword(password string, minSize int) *AnnotationShortPassword {
	return &AnnotationShortPassword{Password: password, Length: len(password), MinSize: minSize}
}
func (a *AnnotationShortPassword) Name() string               { return KindShortPassword }
func (a *AnnotationShortPassword) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationShortPassword) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationShortPassword) Deserialize(serialized []byte) error {
	var tmp AnnotationShortPassword
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationShortPassword) Redacted() utils.TraceAnalyzerAnnotation {
	return &AnnotationShortPassword{"XXX", a.Length, a.MinSize}
}
func (a *AnnotationShortPassword) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Too short Basic Auth password",
		DetailedDesc: fmt.Sprintf("The the length of Basic Auth password is too short (%d)", a.Length),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationKnownPassword struct {
	Password string `json:"password"`
}

func NewAnnotationKnownPassword(password string) *AnnotationKnownPassword {
	return &AnnotationKnownPassword{Password: password}
}
func (a *AnnotationKnownPassword) Name() string               { return KindKnownPassword }
func (a *AnnotationKnownPassword) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationKnownPassword) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationKnownPassword) Deserialize(serialized []byte) error {
	var tmp AnnotationKnownPassword
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a *AnnotationKnownPassword) Redacted() utils.TraceAnalyzerAnnotation {
	return NewAnnotationKnownPassword("XXX")
}
func (a *AnnotationKnownPassword) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Weak Basic Auth password (found in dictionary)",
		DetailedDesc: fmt.Sprintf("The Basic Auth password is too weak because it's too common (%s)", a.Password),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationSamePassword struct {
	User     string   `json:"user"`
	Password string   `json:"password"`
	APIs     []string `json:"apis"`
}

func NewAnnotationSamePassword(user, password string, apis []string) *AnnotationSamePassword {
	return &AnnotationSamePassword{User: user, Password: password, APIs: apis}
}
func (a *AnnotationSamePassword) Name() string               { return KindSamePassword }
func (a *AnnotationSamePassword) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationSamePassword) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationSamePassword) Deserialize(serialized []byte) error {
	var tmp AnnotationSamePassword
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a *AnnotationSamePassword) Redacted() utils.TraceAnalyzerAnnotation {
	return NewAnnotationSamePassword(a.User, "XXX", a.APIs)
}
func (a *AnnotationSamePassword) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Same Basic Auth credentials used for another service",
		DetailedDesc: fmt.Sprintf("The exact same Basic Auth credentials (%s:%s) of this event are used for multiple services (%s)", a.User, a.Password, strings.Join(a.APIs, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
