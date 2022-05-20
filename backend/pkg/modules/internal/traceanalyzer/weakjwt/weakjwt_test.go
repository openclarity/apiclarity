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
	"bytes"
	"testing"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

func sameAnns(got []core.Annotation, expected []core.Annotation) bool {
	if len(got) != len(expected) {
		return false
	}
	for _, eo := range expected { // For each wanted observation
		found := false
		for _, o := range got { // Check if it's in the result
			if eo.Name == o.Name && bytes.Equal(eo.Annotation, o.Annotation) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestFindToken(t *testing.T) {
	testcases := []struct {
		headers     []*models.Header
		wantedObs   []core.Annotation
		wantedToken bool
	}{
		{headers: []*models.Header{{Key: "authorization", Value: "123qsdfqsdfXXX"}}, wantedObs: []core.Annotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer 123qsdfqsdfXXX"}}, wantedObs: []core.Annotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ...SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []core.Annotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer XXX00BADeyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []core.Annotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []core.Annotation{}, wantedToken: true},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []core.Annotation{}, wantedToken: true},

		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []core.Annotation{}, wantedToken: true},
		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "Bearer: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []core.Annotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []core.Annotation{}, wantedToken: false},
	}

	trace := models.Telemetry{}
	trace.Request = &models.Request{}
	trace.Request.Common = &models.Common{}
	for _, tc := range testcases {
		trace.Request.Common.Headers = tc.headers
		token, gotObs := findJWTToken(&trace)
		gotToken := token != nil
		if gotToken != tc.wantedToken || !sameAnns(gotObs, tc.wantedObs) {
			t.Errorf("Wanted (%v, %v) got (%v, %v)", tc.wantedToken, tc.wantedObs, gotToken, gotObs)
		}
	}
}

func TestWeakJWT(t *testing.T) {
	testcases := []struct {
		auth   string
		wanted []core.Annotation
	}{
		{auth: "Bearer ", wanted: []core.Annotation{}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.Q6CM1qIz2WTgTlhMzpFL8jI8xbu9FFfj5DY_bGVY98Y", wanted: []core.Annotation{{Name: "JWT_NO_EXPIRE_CLAIM"}}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjowLCJwYXNzd29yZCI6ImJsYSJ9.HoI84Px0J9oVujFgBvY42PF9xaBz0xCDJzuono4qo40", wanted: []core.Annotation{{Name: "JWT_SENSITIVE_CONTENT_IN_CLAIMS", Annotation: []byte("password")}}},
		{auth: "Bearer eyJ0eXAiOiJKV1QifQ.eyJsb2dnZWRJbkFzIjoiYWRtaW4iLCJpYXQiOjE0MjI3Nzk2Mzh9.HoI84Px0J9oVujFgBvY42PF9xaBz0xCDJzuono4qo40", wanted: []core.Annotation{{Name: "JWT_NO_ALG_FIELD"}}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzc24iOiI5OTk5IiwiaXAiOiIxOTIuMS4xLjEiLCJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.bySqmwlwljWpXLWZ4jlkb_ST3VtuPK2Sui79jkGUEIE", wanted: []core.Annotation{{Name: "JWT_SENSITIVE_CONTENT_IN_CLAIMS", Annotation: []byte("ip,ssn")}}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImNvdmlkX3Bvc2l0aXZlIjp0cnVlfQ.eyJzc24iOiI5OTk5IiwiaXAiOiIxOTIuMS4xLjEiLCJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.kiDLC2Kl-3diNJ_8k-LAdQpNjWmPzmJ1YXvh-p2J9T4", wanted: []core.Annotation{{Name: "JWT_SENSITIVE_CONTENT_IN_HEADERS", Annotation: []byte("covid_positive")}, {Name: "JWT_SENSITIVE_CONTENT_IN_CLAIMS", Annotation: []byte("ip,ssn")}}},
		{auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzM4NCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjowfQ.uzPFbqIJ2akC2gmGxN3KlXU_zhMFvE__N5kKwejY19reMaDaaDT21hmy1mMCZZY2", wanted: []core.Annotation{{Name: "JWT_WEAK_SYMETRIC_SECRET", Annotation: []byte("AllYourBase")}, {Name: "JWT_NOT_RECOMMENDED_ALG", Annotation: []byte("HS384")}}},
		{auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5fQ.X5NwJulKmNzdC2vW9J1UOMsaKikgzQbmFBWslfDNqZE", wanted: []core.Annotation{{Name: "JWT_WEAK_SYMETRIC_SECRET", Annotation: []byte("AllYourBase")}, {Name: "JWT_EXP_TOO_FAR", Annotation: []byte("2286-11-20 17:46:39 +0000 UTC")}}},

		{auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzM4NCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5OSwicGFzc3dvcmQiOiJibGEifQ.tCIFaW7882WmxIGednahpwN-1jEqOkkwgS0W1x5F35psVTACPcpbPw-P8K9CfQM3", wanted: []core.Annotation{{Name: "JWT_WEAK_SYMETRIC_SECRET", Annotation: []byte("AllYourBase")}, {Name: "JWT_SENSITIVE_CONTENT_IN_CLAIMS", Annotation: []byte("password")}, {Name: "JWT_NOT_RECOMMENDED_ALG", Annotation: []byte("HS384")}, {Name: "JWT_EXP_TOO_FAR", Annotation: []byte("5138-11-16 09:46:39 +0000 UTC")}}},
		{auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5OSwicGFzc3dvcmQiOiJibGFibHUifQ.SLWwRavOnos1ihyRJUPeG3xjKRy8eIBvUOD6VqW20WU", wanted: []core.Annotation{{Name: "JWT_WEAK_SYMETRIC_SECRET", Annotation: []byte("AllYourBase")}, {Name: "JWT_SENSITIVE_CONTENT_IN_CLAIMS", Annotation: []byte("password")}, {Name: "JWT_EXP_TOO_FAR", Annotation: []byte("5138-11-16 09:46:39 +0000 UTC")}}},
	}
	knownSigningKeys := []string{"pass", "pass123", "123", "1234", "signingkey1", "AllYourBase", "random"}
	sensitiveKeywords := []string{"password", "ip", "ssn", "covid_positive"}
	analyzer := NewWeakJWT(knownSigningKeys, sensitiveKeywords)

	trace := models.Telemetry{}
	trace.Request = &models.Request{}
	trace.Request.Common = &models.Common{}

	for _, tc := range testcases {
		trace.Request.Common.Headers = []*models.Header{
			{
				Key:   "authorization",
				Value: tc.auth,
			},
		}

		eventAnns, _ := analyzer.Analyze(&trace)
		if !sameAnns(eventAnns, tc.wanted) {
			t.Errorf("Wanted: (%v) got (%v)", tc.wanted, eventAnns)
		}
	}
}
