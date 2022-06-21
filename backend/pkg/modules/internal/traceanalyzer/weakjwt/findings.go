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

package weakjwt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	JWTNoAlgField        = "JWT_NO_ALG_FIELD"
	JWTAlgFieldNone      = "JWT_ALG_FIELD_NONE"
	JWTNotRecommendedAlg = "JWT_NOT_RECOMMENDED_ALG"
	JWTNoExpireClaim     = "JWT_NO_EXPIRE_CLAIM"
	JWTExpTooFar         = "JWT_EXP_TOO_FAR"
	//nolint:gosec
	JWTWeakSymetricSecret        = "JWT_WEAK_SYMETRIC_SECRET"
	JWTSensitiveContentInHeaders = "JWT_SENSITIVE_CONTENT_IN_HEADERS"
	JWTSensitiveContentInClaims  = "JWT_SENSITIVE_CONTENT_IN_CLAIMS"
)

type AnnotationNoAlgField struct{}

func NewAnnotationNoAlgField() *AnnotationNoAlgField {
	return &AnnotationNoAlgField{}
}
func (a *AnnotationNoAlgField) Name() string               { return JWTNoAlgField }
func (a *AnnotationNoAlgField) Severity() string           { return utils.SeverityHigh }
func (a *AnnotationNoAlgField) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationNoAlgField) Deserialize(serialized []byte) error {
	var tmp AnnotationNoAlgField
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationNoAlgField) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationNoAlgField) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT has no algorithm specified",
		DetailedDesc: fmt.Sprintf("The JOSE header of the JWT header does not contain an 'alg' field"),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationAlgFieldNone struct{}

func NewAnnotationAlgFieldNone() *AnnotationAlgFieldNone {
	return &AnnotationAlgFieldNone{}
}
func (a *AnnotationAlgFieldNone) Name() string               { return JWTAlgFieldNone }
func (a *AnnotationAlgFieldNone) Severity() string           { return utils.SeverityHigh }
func (a *AnnotationAlgFieldNone) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationAlgFieldNone) Deserialize(serialized []byte) error {
	var tmp AnnotationAlgFieldNone
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationAlgFieldNone) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationAlgFieldNone) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "'alg' field set to None",
		DetailedDesc: fmt.Sprintf("The JOSE header of the JWT header contains an 'alg' field but it's set to none"),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationNotRecommendedAlg struct {
	Algorithm       string   `json:"algorithm"`
	RecommendedAlgs []string `json:"recommended_algs"`
}

func NewAnnotationNotRecommendedAlg(alg string, recommended []string) *AnnotationNotRecommendedAlg {
	return &AnnotationNotRecommendedAlg{Algorithm: alg, RecommendedAlgs: recommended}
}
func (a *AnnotationNotRecommendedAlg) Name() string               { return JWTNotRecommendedAlg }
func (a *AnnotationNotRecommendedAlg) Severity() string           { return utils.SeverityHigh }
func (a *AnnotationNotRecommendedAlg) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationNotRecommendedAlg) Deserialize(serialized []byte) error {
	var tmp AnnotationNotRecommendedAlg
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationNotRecommendedAlg) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationNotRecommendedAlg) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Not a recommanded JWT signing algorithm",
		DetailedDesc: fmt.Sprintf("'%s' is not a recommended signing algorithm (recommended are: %s)", a.Algorithm, strings.Join(a.RecommendedAlgs, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationNoExpireClaim struct{}

func NewAnnotationNoExpireClaim() *AnnotationNoExpireClaim {
	return &AnnotationNoExpireClaim{}
}
func (a *AnnotationNoExpireClaim) Name() string               { return JWTNoExpireClaim }
func (a *AnnotationNoExpireClaim) Severity() string           { return utils.SeverityLow }
func (a *AnnotationNoExpireClaim) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationNoExpireClaim) Deserialize(serialized []byte) error {
	var tmp AnnotationNoExpireClaim
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationNoExpireClaim) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationNoExpireClaim) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT does not have any expire claims",
		DetailedDesc: "JWT does not have any expire claims",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationExpTooFar struct {
	ExpireAt time.Time     `json:"expire_at"`
	ExpireIn time.Duration `json:"expire_in"`
}

func NewAnnotationExpTooFar(expireAt time.Time) *AnnotationExpTooFar {
	return &AnnotationExpTooFar{
		ExpireAt: expireAt,
		ExpireIn: time.Until(expireAt),
	}
}
func (a *AnnotationExpTooFar) Name() string               { return JWTExpTooFar }
func (a *AnnotationExpTooFar) Severity() string           { return utils.SeverityLow }
func (a *AnnotationExpTooFar) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationExpTooFar) Deserialize(serialized []byte) error {
	var tmp AnnotationExpTooFar
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationExpTooFar) Redacted() utils.TraceAnalyzerAnnotation {
	return &a
}
func (a *AnnotationExpTooFar) ToFinding() utils.Finding {
	var expireString string
	daysToExpiration := int(a.ExpireIn.Hours() / 24)
	if daysToExpiration > 2 {
		expireString = fmt.Sprintf("%d days", daysToExpiration)
	} else {
		expireString = a.ExpireIn.String()
	}
	return utils.Finding{
		ShortDesc:    "JWT expire too far in the future",
		DetailedDesc: fmt.Sprintf("The JWT expire in %s", expireString),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationWeakSymetricSecret struct {
	WeakKey    []byte `json:"weak_key"`
	WeakKeyLen int    `json:"weak_key_len"`
}

func NewAnnotationWeakSymetricSecret(weakKey []byte) *AnnotationWeakSymetricSecret {
	return &AnnotationWeakSymetricSecret{
		WeakKey:    weakKey,
		WeakKeyLen: len(weakKey),
	}
}
func (a *AnnotationWeakSymetricSecret) Name() string               { return JWTWeakSymetricSecret }
func (a *AnnotationWeakSymetricSecret) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationWeakSymetricSecret) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationWeakSymetricSecret) Deserialize(serialized []byte) error {
	var tmp AnnotationWeakSymetricSecret
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}

const maxRedactedKeyLen = 4

func (a AnnotationWeakSymetricSecret) Redacted() utils.TraceAnalyzerAnnotation {
	redacted := a
	min := utils.Min(maxRedactedKeyLen, len(a.WeakKey))

	redacted.WeakKey = a.WeakKey[:min]
	redacted.WeakKey = append(redacted.WeakKey, "... [redacted]"...)

	return &redacted
}

const maxDisplayKey = 20

func (a *AnnotationWeakSymetricSecret) ToFinding() utils.Finding {
	min := utils.Min(maxDisplayKey, len(a.WeakKey))
	return utils.Finding{
		ShortDesc:    "JWT signed with a weak key",
		DetailedDesc: fmt.Sprintf("The weak signing key is %d bytes long and starts with '%s'", a.WeakKeyLen, a.WeakKey[:min]),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationSensitiveContentInHeaders struct {
	SensitiveWords []string `json:"sensitive_words"`
}

func NewAnnotationSensitiveContentInHeaders(sensitive_words []string) *AnnotationSensitiveContentInHeaders {
	return &AnnotationSensitiveContentInHeaders{
		SensitiveWords: sensitive_words,
	}
}
func (a *AnnotationSensitiveContentInHeaders) Name() string               { return JWTSensitiveContentInHeaders }
func (a *AnnotationSensitiveContentInHeaders) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationSensitiveContentInHeaders) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationSensitiveContentInHeaders) Deserialize(serialized []byte) error {
	var tmp AnnotationSensitiveContentInHeaders
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationSensitiveContentInHeaders) Redacted() utils.TraceAnalyzerAnnotation {
	redacted := make([]string, len(a.SensitiveWords))
	for i, r := range a.SensitiveWords {
		// Only provide the first character
		redacted[i] = r[:1] + "...[redacted]"
	}

	return NewAnnotationSensitiveContentInHeaders(redacted)
}
func (a *AnnotationSensitiveContentInHeaders) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT headers may contain sensitive content",
		DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", strings.Join(a.SensitiveWords, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type AnnotationSensitiveContentInClaims struct {
	SensitiveWords []string `json:"sensitive_words"`
}

func NewAnnotationSensitiveContentInClaims(sensitive_words []string) *AnnotationSensitiveContentInClaims {
	return &AnnotationSensitiveContentInClaims{
		SensitiveWords: sensitive_words,
	}
}
func (a *AnnotationSensitiveContentInClaims) Name() string               { return JWTSensitiveContentInClaims }
func (a *AnnotationSensitiveContentInClaims) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationSensitiveContentInClaims) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationSensitiveContentInClaims) Deserialize(serialized []byte) error {
	var tmp AnnotationSensitiveContentInClaims
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationSensitiveContentInClaims) Redacted() utils.TraceAnalyzerAnnotation {
	redacted := make([]string, len(a.SensitiveWords))
	for i, r := range a.SensitiveWords {
		// Only provide the first character
		redacted[i] = r[:1] + "...[redacted]"
	}

	return NewAnnotationSensitiveContentInClaims(redacted)
}
func (a *AnnotationSensitiveContentInClaims) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT claims may contains sensitive content",
		DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", strings.Join(a.SensitiveWords, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
