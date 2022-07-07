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

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

const (
	JWTNoAlgField        = "JWT_NO_ALG_FIELD"
	JWTAlgFieldNone      = "JWT_ALG_FIELD_NONE"
	JWTNotRecommendedAlg = "JWT_NOT_RECOMMENDED_ALG"
	JWTNoExpireClaim     = "JWT_NO_EXPIRE_CLAIM"
	JWTExpTooFar         = "JWT_EXP_TOO_FAR"
	//nolint:gosec
	JWTWeakSymetricSecret = "JWT_WEAK_SYMETRIC_SECRET"
	JWTSensitiveContent   = "JWT_SENSITIVE_CONTENT"
)

type AnnotationNoAlgField struct{}

func NewAnnotationNoAlgField() *AnnotationNoAlgField {
	return &AnnotationNoAlgField{}
}
func (a *AnnotationNoAlgField) Name() string { return JWTNoAlgField }
func (a *AnnotationNoAlgField) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationNoAlgField(path, method)
}
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

type APIAnnotationNoAlgField struct {
	utils.BaseTraceAnalyzerAPIAnnotation
}

func NewAPIAnnotationNoAlgField(path, method string) *APIAnnotationNoAlgField {
	return &APIAnnotationNoAlgField{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationNoAlgField) Name() string { return JWTNoAlgField }
func (a *APIAnnotationNoAlgField) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	_, valid := ann.(*AnnotationNoAlgField)
	if !valid {
		panic("invalid type")
	}

	return false
}

func (a APIAnnotationNoAlgField) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationNoAlgField) Deserialize(serialized []byte) error {
	var tmp APIAnnotationNoAlgField
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationNoAlgField) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}
func (a *APIAnnotationNoAlgField) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT has no algorithm specified",
		DetailedDesc: "The JOSE header of the JWT header does not contain an 'alg' field",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationNoAlgField) ToAPIFinding() oapicommon.APIFinding {
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "JWT has no algorithm specified",
		Description: "The JOSE header of the JWT header does not contain an 'alg' field",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: nil,
	}
}

type AnnotationAlgFieldNone struct{}

func NewAnnotationAlgFieldNone() *AnnotationAlgFieldNone {
	return &AnnotationAlgFieldNone{}
}
func (a *AnnotationAlgFieldNone) Name() string { return JWTAlgFieldNone }
func (a *AnnotationAlgFieldNone) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return nil
}
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

type APIAnnotationAlgFieldNone struct {
	utils.BaseTraceAnalyzerAPIAnnotation
}

func NewAPIAnnotationAlgFieldNone(path, method string) *APIAnnotationAlgFieldNone {
	return &APIAnnotationAlgFieldNone{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationAlgFieldNone) Name() string { return JWTAlgFieldNone }
func (a *APIAnnotationAlgFieldNone) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	_, valid := ann.(*AnnotationAlgFieldNone)
	if !valid {
		panic("invalid type")
	}

	return false
}

func (a APIAnnotationAlgFieldNone) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationAlgFieldNone) Deserialize(serialized []byte) error {
	var tmp APIAnnotationAlgFieldNone
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationAlgFieldNone) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}
func (a *APIAnnotationAlgFieldNone) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "'alg' field set to None",
		DetailedDesc: fmt.Sprintf("The JOSE header of the JWT header contains an 'alg' field but it's set to none"),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationAlgFieldNone) ToAPIFinding() oapicommon.APIFinding {
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "'alg' field set to None",
		Description: fmt.Sprintf("The JOSE header of the JWT header contains an 'alg' field but it's set to none"),

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: nil,
	}
}

type AnnotationNotRecommendedAlg struct {
	Algorithm       string   `json:"algorithm"`
	RecommendedAlgs []string `json:"recommended_algs"`
}

func NewAnnotationNotRecommendedAlg(alg string, recommended []string) *AnnotationNotRecommendedAlg {
	return &AnnotationNotRecommendedAlg{Algorithm: alg, RecommendedAlgs: recommended}
}
func (a *AnnotationNotRecommendedAlg) Name() string { return JWTNotRecommendedAlg }
func (a *AnnotationNotRecommendedAlg) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationNotRecommendedAlg(path, method)
}
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

type APIAnnotationNotRecommendedAlg struct {
	utils.BaseTraceAnalyzerAPIAnnotation
	NotRecommendedAlgs map[string]bool `json:"not_recommended"`
}

func NewAPIAnnotationNotRecommendedAlg(path, method string) *APIAnnotationNotRecommendedAlg {
	return &APIAnnotationNotRecommendedAlg{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
		NotRecommendedAlgs:             map[string]bool{},
	}
}
func (a *APIAnnotationNotRecommendedAlg) Name() string { return JWTNotRecommendedAlg }
func (a *APIAnnotationNotRecommendedAlg) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	apiAnn, valid := ann.(*AnnotationNotRecommendedAlg)
	if !valid {
		panic("invalid type")
	}
	initialSize := len(a.NotRecommendedAlgs)

	a.NotRecommendedAlgs[apiAnn.Algorithm] = true

	return initialSize != len(a.NotRecommendedAlgs)
}

func (a APIAnnotationNotRecommendedAlg) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationNotRecommendedAlg) Deserialize(serialized []byte) error {
	var tmp APIAnnotationNotRecommendedAlg
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationNotRecommendedAlg) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}
func (a *APIAnnotationNotRecommendedAlg) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "Not a recommanded JWT signing algorithm",
		DetailedDesc: "Signing algorithms that are not recommended were used",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationNotRecommendedAlg) ToAPIFinding() oapicommon.APIFinding {
	var additionalInfo *map[string]interface{}
	description := "Signing algorithms that are not recommended were used"
	if len(a.NotRecommendedAlgs) > 0 {
		not_recommended := []string{}
		for name := range a.NotRecommendedAlgs {
			not_recommended = append(not_recommended, name)
		}
		additionalInfo = &map[string]interface{}{
			"not_recommended": not_recommended,
		}
		description = fmt.Sprintf("Signing algorithms (%s) are not recommended", strings.Join(not_recommended, ","))
	}
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "Not a recommanded JWT signing algorithm",
		Description: description,

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: additionalInfo,
	}
}

type AnnotationNoExpireClaim struct{}

func NewAnnotationNoExpireClaim() *AnnotationNoExpireClaim {
	return &AnnotationNoExpireClaim{}
}
func (a *AnnotationNoExpireClaim) Name() string { return JWTNoExpireClaim }
func (a *AnnotationNoExpireClaim) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationNoExpireClaim(path, method)
}
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

type APIAnnotationNoExpireClaim struct {
	utils.BaseTraceAnalyzerAPIAnnotation
}

func NewAPIAnnotationNoExpireClaim(path, method string) *APIAnnotationNoExpireClaim {
	return &APIAnnotationNoExpireClaim{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationNoExpireClaim) Name() string { return JWTNoExpireClaim }
func (a *APIAnnotationNoExpireClaim) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	_, valid := ann.(*AnnotationNoExpireClaim)
	if !valid {
		panic("invalid type")
	}

	return false
}

func (a APIAnnotationNoExpireClaim) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationNoExpireClaim) Deserialize(serialized []byte) error {
	var tmp APIAnnotationNoExpireClaim
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationNoExpireClaim) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}
func (a *APIAnnotationNoExpireClaim) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT does not have any expire claims",
		DetailedDesc: "JWT does not have any expire claims",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationNoExpireClaim) ToAPIFinding() oapicommon.APIFinding {
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "JWT does not have any expire claims",
		Description: "JWT does not have any expire claims",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: nil,
	}
}

type AnnotationExpTooFar struct {
	ExpireAt time.Time     `json:"expire_at"`
	ExpireIn time.Duration `json:"expire_in"`
}

func expireString(expireIn time.Duration) string {
	var expireString string
	daysToExpiration := int(expireIn.Hours() / 24)
	if daysToExpiration > 2 {
		expireString = fmt.Sprintf("%d days", daysToExpiration)
	} else {
		expireString = expireIn.String()
	}

	return expireString
}

func NewAnnotationExpTooFar(expireAt time.Time) *AnnotationExpTooFar {
	return &AnnotationExpTooFar{
		ExpireAt: expireAt,
		ExpireIn: time.Until(expireAt),
	}
}
func (a *AnnotationExpTooFar) Name() string { return JWTExpTooFar }
func (a *AnnotationExpTooFar) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationExpTooFar(path, method)
}
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
	return utils.Finding{
		ShortDesc:    "JWT expire too far in the future",
		DetailedDesc: fmt.Sprintf("The JWT expire in %s", expireString(a.ExpireIn)),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type APIAnnotationExpTooFar struct {
	utils.BaseTraceAnalyzerAPIAnnotation
	ExpireInExample time.Duration `json:"expire_in_example"`
}

func NewAPIAnnotationExpTooFar(path, method string) *APIAnnotationExpTooFar {
	return &APIAnnotationExpTooFar{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationExpTooFar) Name() string { return JWTExpTooFar }
func (a *APIAnnotationExpTooFar) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	eventAnn, valid := ann.(*AnnotationExpTooFar)
	if !valid {
		panic("invalid type")
	}

	if a.ExpireInExample == 0 {
		a.ExpireInExample = eventAnn.ExpireIn
		return true
	}

	return false
}

func (a APIAnnotationExpTooFar) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationExpTooFar) Deserialize(serialized []byte) error {
	var tmp APIAnnotationExpTooFar
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationExpTooFar) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}

func (a *APIAnnotationExpTooFar) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT expire too far in the future",
		DetailedDesc: fmt.Sprintf("It has been observed JWT which expire far in the future (ex: %s)", expireString(a.ExpireInExample)),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationExpTooFar) ToAPIFinding() oapicommon.APIFinding {
	additionalInfo := &map[string]interface{}{
		"expire_in_example": uint64(a.ExpireInExample.Seconds()),
	}
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "JWT does not have any expire claims",
		Description: fmt.Sprintf("It has been observed JWT which expire far in the future (ex: %s)", expireString(a.ExpireInExample)),

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: additionalInfo,
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
func (a *AnnotationWeakSymetricSecret) Name() string { return JWTWeakSymetricSecret }
func (a *AnnotationWeakSymetricSecret) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationWeakSymetricSecret(path, method)
}
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

type APIAnnotationWeakSymetricSecret struct {
	utils.BaseTraceAnalyzerAPIAnnotation
}

func NewAPIAnnotationWeakSymetricSecret(path, method string) *APIAnnotationWeakSymetricSecret {
	return &APIAnnotationWeakSymetricSecret{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationWeakSymetricSecret) Name() string { return JWTWeakSymetricSecret }
func (a *APIAnnotationWeakSymetricSecret) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	_, valid := ann.(*AnnotationWeakSymetricSecret)
	if !valid {
		panic("invalid type")
	}
	return false
}

func (a APIAnnotationWeakSymetricSecret) Serialize() ([]byte, error) { return json.Marshal(a) }

func (a *APIAnnotationWeakSymetricSecret) Deserialize(serialized []byte) error {
	var tmp APIAnnotationWeakSymetricSecret
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationWeakSymetricSecret) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}

func (a *APIAnnotationWeakSymetricSecret) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT signed with a weak key",
		DetailedDesc: "It has been observed one or more JWT with weak, known keys",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationWeakSymetricSecret) ToAPIFinding() oapicommon.APIFinding {
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "JWT signed with a weak key",
		Description: "It has been observed one or more JWT with weak, known keys",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: nil,
	}
}

type AnnotationSensitiveContent struct {
	SensitiveWordsInHeaders []string `json:"sensitive_words_in_headers"`
	SensitiveWordsInClaims  []string `json:"sensitive_words_in_claims"`
}

func NewAnnotationSensitiveContent(sensitiveInHeaders, sensitiveInClaims []string) *AnnotationSensitiveContent {
	return &AnnotationSensitiveContent{
		SensitiveWordsInHeaders: sensitiveInHeaders,
		SensitiveWordsInClaims:  sensitiveInClaims,
	}
}
func (a *AnnotationSensitiveContent) Name() string { return JWTSensitiveContent }
func (a *AnnotationSensitiveContent) NewAPIAnnotation(path, method string) utils.TraceAnalyzerAPIAnnotation {
	return NewAPIAnnotationSensitiveContent(path, method)
}
func (a *AnnotationSensitiveContent) Severity() string           { return utils.SeverityMedium }
func (a *AnnotationSensitiveContent) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *AnnotationSensitiveContent) Deserialize(serialized []byte) error {
	var tmp AnnotationSensitiveContent
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a AnnotationSensitiveContent) Redacted() utils.TraceAnalyzerAnnotation {
	redactedHeaders := make([]string, len(a.SensitiveWordsInHeaders))
	for i, r := range a.SensitiveWordsInHeaders {
		// Only provide the first character
		redactedHeaders[i] = r[:1] + "...[redacted]"
	}
	redactedClaims := make([]string, len(a.SensitiveWordsInClaims))
	for i, r := range a.SensitiveWordsInClaims {
		// Only provide the first character
		redactedClaims[i] = r[:1] + "...[redacted]"
	}

	return NewAnnotationSensitiveContent(redactedHeaders, redactedClaims)
}
func (a *AnnotationSensitiveContent) ToFinding() utils.Finding {
	words := append(a.SensitiveWordsInHeaders, a.SensitiveWordsInClaims...)
	return utils.Finding{
		ShortDesc:    "JWT claims or headers may contains sensitive content",
		DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", strings.Join(words, ",")),
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}

type APIAnnotationSensitiveContent struct {
	utils.BaseTraceAnalyzerAPIAnnotation
}

func NewAPIAnnotationSensitiveContent(path, method string) *APIAnnotationSensitiveContent {
	return &APIAnnotationSensitiveContent{
		BaseTraceAnalyzerAPIAnnotation: utils.BaseTraceAnalyzerAPIAnnotation{SpecPath: path, SpecMethod: method},
	}
}
func (a *APIAnnotationSensitiveContent) Name() string { return JWTSensitiveContent }
func (a *APIAnnotationSensitiveContent) Aggregate(ann utils.TraceAnalyzerAnnotation) (updated bool) {
	_, valid := ann.(*AnnotationSensitiveContent)
	if !valid {
		panic("invalid type")
	}

	return false
}

func (a *APIAnnotationSensitiveContent) Severity() string   { return utils.SeverityInfo }
func (a *APIAnnotationSensitiveContent) TTL() time.Duration { return 24 * time.Hour }

func (a *APIAnnotationSensitiveContent) Serialize() ([]byte, error) { return json.Marshal(a) }
func (a *APIAnnotationSensitiveContent) Deserialize(serialized []byte) error {
	var tmp APIAnnotationSensitiveContent
	err := json.Unmarshal(serialized, &tmp)
	*a = tmp

	return err
}
func (a APIAnnotationSensitiveContent) Redacted() utils.TraceAnalyzerAPIAnnotation {
	newA := a
	return &newA
}
func (a *APIAnnotationSensitiveContent) ToFinding() utils.Finding {
	return utils.Finding{
		ShortDesc:    "JWT claims or headers may contains sensitive content",
		DetailedDesc: "JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker",
		Severity:     a.Severity(),
		Alert:        utils.SeverityToAlert(a.Severity()),
	}
}
func (a *APIAnnotationSensitiveContent) ToAPIFinding() oapicommon.APIFinding {
	jsonPointer := a.SpecLocation()
	return oapicommon.APIFinding{
		Source: utils.ModuleName,

		Type:        a.Name(),
		Name:        "JWT claims or headers may contains sensitive content",
		Description: "JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker",

		ProvidedSpecLocation:      &jsonPointer,
		ReconstructedSpecLocation: &jsonPointer,

		Severity: oapicommon.INFO,

		AdditionalInfo: nil,
	}
}
