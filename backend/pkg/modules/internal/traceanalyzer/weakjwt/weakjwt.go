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
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	AuthorizationHeader = "authorization"
	BearerAuth          = "Bearer"
	MaxTokenAge         = 5 * 24 * time.Hour
)

const (
	JWTNoAlgField                = "JWT_NO_ALG_FIELD"
	JWTAlgFieldNone              = "JWT_ALG_FIELD_NONE"
	JWTNotRecommendedAlg         = "JWT_NOT_RECOMMENDED_ALG"
	JWTNoExpireClaim             = "JWT_NO_EXPIRE_CLAIM"
	JWTExpTooFar                 = "JWT_EXP_TOO_FAR"
	JWTWeakSymetricSecret        = "JWT_WEAK_SYMETRIC_SECRET"
	JWTSensitiveContentInHeaders = "JWT_SENSITIVE_CONTENT_IN_HEADERS"
	JWTSensitiveContentInClaims  = "JWT_SENSITIVE_CONTENT_IN_CLAIMS"
)

type WeakJWT struct {
	knownWeakKeys     []string
	sensitiveKeywords ahocorasick.AhoCorasick
	maxTokenAge       time.Duration
}

func findJWTToken(trace *models.Telemetry) (*jwt.Token, []core.Annotation) {
	anns := []core.Annotation{}

	index, found := utils.FindHeader(trace.Request.Common.Headers, AuthorizationHeader)
	if !found {
		return nil, anns
	}

	hValue := trace.Request.Common.Headers[index].Value
	splitValue := strings.Split(hValue, " ")
	if len(splitValue) == 2 && splitValue[0] == BearerAuth {
		// We found a Bearer Token !
		parser := jwt.Parser{
			UseJSONNumber:        true,
			SkipClaimsValidation: true,
		}
		token, _, err := parser.ParseUnverified(splitValue[1], jwt.MapClaims{})
		if err != nil {
			if err.(*jwt.ValidationError).Errors == jwt.ValidationErrorUnverifiable {
				anns = append(anns, core.Annotation{Name: JWTNoAlgField})
			}
			return nil, anns
		}
		return token, anns
	}

	return nil, anns
}

func NewWeakJWT(weakKeyList []string, sensitiveKeywords []string) *WeakJWT {
	acBuilder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  true,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  true,
	})

	return &WeakJWT{
		knownWeakKeys:     weakKeyList,
		sensitiveKeywords: acBuilder.Build(sensitiveKeywords),
		maxTokenAge:       MaxTokenAge,
	}
}

func (w *WeakJWT) analyzeAlg(token *jwt.Token) []core.Annotation {
	anns := []core.Annotation{}

	if token.Method == nil {
		a := core.Annotation{
			Name: JWTNoAlgField,
		}
		anns = append(anns, a)
	} else if token.Method == jwt.SigningMethodNone {
		a := core.Annotation{
			Name: JWTAlgFieldNone,
		}
		anns = append(anns, a)
	} else {
		alg := token.Method.Alg()
		recommended := []string{
			"ES256", "RS256", // Asymetric
			"HS256", // Symetric
		}

		haveRecommended := false
		for _, r := range recommended {
			if alg == r {
				haveRecommended = true
				break
			}
		}
		if !haveRecommended {
			a := core.Annotation{
				Name:       JWTNotRecommendedAlg,
				Annotation: []byte(alg),
			}
			anns = append(anns, a)
		}
	}

	return anns
}

func (w *WeakJWT) analyzeExpireClaims(token *jwt.Token) []core.Annotation {
	anns := []core.Annotation{}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// XXX WE MUST LOG SOMETHING, THIS SHOULD NOT HAPPEN
		return anns
	}
	exp := claims["exp"]
	nbf := claims["nbf"]
	iat := claims["iat"]
	if exp == nil && nbf == nil && iat == nil {
		a := core.Annotation{
			Name: JWTNoExpireClaim,
		}
		anns = append(anns, a)
	} else if exp != nil {
		// There are Claims that allow for token expiration.
		// Check that this token is not expiring too far in the future
		var expireAt time.Time
		switch e := exp.(type) {
		case float64:
			expireAt = time.Unix(int64(e), 0)
		case json.Number:
			v, _ := e.Int64()
			expireAt = time.Unix(v, 0)
		}

		if time.Until(expireAt) >= w.maxTokenAge {
			a := core.Annotation{
				Name:       JWTExpTooFar,
				Annotation: []byte(expireAt.UTC().String()),
			}
			anns = append(anns, a)
		}
	}

	return anns
}

func (w *WeakJWT) analyzeSig(token *jwt.Token) []core.Annotation {
	anns := []core.Annotation{}

	if token.Method == nil || !strings.HasPrefix(token.Method.Alg(), "HS") {
		return anns
	}

	parts := strings.Split(token.Raw, ".")
	if len(parts) != 3 {
		// XXX WE MUST LOG SOMETHING, THIS SHOULD NOT HAPPEN
		return anns
	}
	signingString := strings.Join([]string{parts[0], parts[1]}, ".")
	signMethod := token.Method

	for _, secret := range w.knownWeakKeys {
		sig, err := signMethod.Sign(signingString, []byte(secret))
		if err == nil && sig == parts[2] { // We found the secret signing key !
			a := core.Annotation{
				Name:       JWTWeakSymetricSecret,
				Annotation: []byte(secret),
			}
			anns = append(anns, a)
			break
		}
	}

	return anns
}

func (w *WeakJWT) analyzeSensitive(token *jwt.Token) []core.Annotation {
	anns := []core.Annotation{}
	sensitiveHeader := make([]string, 0)
	sensitiveClaims := make([]string, 0)

	for headerK := range token.Header {
		matches := w.sensitiveKeywords.FindAll(headerK)
		if len(matches) > 0 {
			sensitiveHeader = append(sensitiveHeader, headerK)
		}
	}
	if len(sensitiveHeader) > 0 {
		sort.Strings(sensitiveHeader)
		a := core.Annotation{
			Name:       JWTSensitiveContentInHeaders,
			Annotation: []byte(strings.Join(sensitiveHeader, ",")),
		}
		anns = append(anns, a)
	}

	for claimK := range token.Claims.(jwt.MapClaims) {
		matches := w.sensitiveKeywords.FindAll(claimK)
		if len(matches) > 0 {
			sensitiveClaims = append(sensitiveClaims, claimK)
		}
	}
	if len(sensitiveClaims) > 0 {
		sort.Strings(sensitiveClaims)
		a := core.Annotation{
			Name:       JWTSensitiveContentInClaims,
			Annotation: []byte(strings.Join(sensitiveClaims, ",")),
		}
		anns = append(anns, a)
	}
	return anns
}

func (w *WeakJWT) Analyze(trace *models.Telemetry) ([]core.Annotation, []core.Annotation) {
	eventAnns := []core.Annotation{}
	apiAnns := []core.Annotation{}

	JWTToken, eventAnns := findJWTToken(trace)
	if JWTToken != nil {
		eventAnns = append(eventAnns, w.analyzeAlg(JWTToken)...)
		eventAnns = append(eventAnns, w.analyzeExpireClaims(JWTToken)...)
		eventAnns = append(eventAnns, w.analyzeSig(JWTToken)...)
		eventAnns = append(eventAnns, w.analyzeSensitive(JWTToken)...)
	}

	return eventAnns, apiAnns
}
