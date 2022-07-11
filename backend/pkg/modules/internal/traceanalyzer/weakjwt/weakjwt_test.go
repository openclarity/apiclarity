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
	"reflect"
	"testing"
	"time"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

func dirtyTimeHack(a utils.TraceAnalyzerAnnotation) {
	if expTooFar, ok := a.(*AnnotationExpTooFar); ok {
		expTooFar.ExpireIn = 0
	}
}

func sameAnns(got []utils.TraceAnalyzerAnnotation, expected []utils.TraceAnalyzerAnnotation) bool {
	if len(got) != len(expected) {
		return false
	}
	for _, eo := range expected { // For each wanted observation
		dirtyTimeHack(eo)
		found := false
		for _, o := range got { // Check if it's in the result
			dirtyTimeHack(o)
			if reflect.DeepEqual(o, eo) {
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
		wantedObs   []utils.TraceAnalyzerAnnotation
		wantedToken bool
	}{
		{headers: []*models.Header{{Key: "authorization", Value: "123qsdfqsdfXXX"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer 123qsdfqsdfXXX"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ...SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer XXX00BADeyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: true},
		{headers: []*models.Header{{Key: "authorization", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: true},

		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: true},
		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "Bearer: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
		{headers: []*models.Header{{Key: "AuthoriZatioN", Value: "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.GYnXnoBNfM69A-9Eml5-e0ICwRbRAcEAZ9gYexivLFg"}}, wantedObs: []utils.TraceAnalyzerAnnotation{}, wantedToken: false},
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
		wanted []utils.TraceAnalyzerAnnotation
	}{
		{auth: "Bearer ", wanted: []utils.TraceAnalyzerAnnotation{}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.Q6CM1qIz2WTgTlhMzpFL8jI8xbu9FFfj5DY_bGVY98Y", wanted: []utils.TraceAnalyzerAnnotation{&AnnotationNoExpireClaim{}}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjowLCJwYXNzd29yZCI6ImJsYSJ9.HoI84Px0J9oVujFgBvY42PF9xaBz0xCDJzuono4qo40", wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationSensitiveContent([]string{}, []string{"password"})}},
		{auth: "Bearer eyJ0eXAiOiJKV1QifQ.eyJsb2dnZWRJbkFzIjoiYWRtaW4iLCJpYXQiOjE0MjI3Nzk2Mzh9.HoI84Px0J9oVujFgBvY42PF9xaBz0xCDJzuono4qo40", wanted: []utils.TraceAnalyzerAnnotation{&AnnotationNoAlgField{}}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzc24iOiI5OTk5IiwiaXAiOiIxOTIuMS4xLjEiLCJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.bySqmwlwljWpXLWZ4jlkb_ST3VtuPK2Sui79jkGUEIE", wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationSensitiveContent([]string{}, []string{"ip", "ssn"})}},
		{auth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImNvdmlkX3Bvc2l0aXZlIjp0cnVlfQ.eyJzc24iOiI5OTk5IiwiaXAiOiIxOTIuMS4xLjEiLCJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.kiDLC2Kl-3diNJ_8k-LAdQpNjWmPzmJ1YXvh-p2J9T4", wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationSensitiveContent([]string{"covid_positive"}, []string{"ip", "ssn"})}},
		{auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzM4NCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjowfQ.uzPFbqIJ2akC2gmGxN3KlXU_zhMFvE__N5kKwejY19reMaDaaDT21hmy1mMCZZY2", wanted: []utils.TraceAnalyzerAnnotation{&AnnotationWeakSymetricSecret{WeakKey: []byte("AllYourBase"), WeakKeyLen: 11}, &AnnotationNotRecommendedAlg{Algorithm: "HS384", RecommendedAlgs: []string{"ES256", "RS256", "HS256"}}}},
		{
			auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5fQ.X5NwJulKmNzdC2vW9J1UOMsaKikgzQbmFBWslfDNqZE",
			wanted: []utils.TraceAnalyzerAnnotation{
				&AnnotationWeakSymetricSecret{WeakKey: []byte("AllYourBase"), WeakKeyLen: 11},
				&AnnotationExpTooFar{ExpireAt: time.Unix(9999999999, 0)},
			},
		},

		{
			auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzM4NCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5OSwicGFzc3dvcmQiOiJibGEifQ.tCIFaW7882WmxIGednahpwN-1jEqOkkwgS0W1x5F35psVTACPcpbPw-P8K9CfQM3",
			wanted: []utils.TraceAnalyzerAnnotation{
				NewAnnotationWeakSymetricSecret([]byte("AllYourBase")),
				NewAnnotationSensitiveContent([]string{}, []string{"password"}),
				&AnnotationNotRecommendedAlg{Algorithm: "HS384", RecommendedAlgs: []string{"ES256", "RS256", "HS256"}},
				&AnnotationExpTooFar{ExpireAt: time.Unix(99999999999, 0)},
			},
		},
		{
			auth: "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo5OTk5OTk5OTk5OSwicGFzc3dvcmQiOiJibGFibHUifQ.SLWwRavOnos1ihyRJUPeG3xjKRy8eIBvUOD6VqW20WU",
			wanted: []utils.TraceAnalyzerAnnotation{
				&AnnotationWeakSymetricSecret{WeakKey: []byte("AllYourBase"), WeakKeyLen: 11},
				NewAnnotationSensitiveContent([]string{}, []string{"password"}),
				&AnnotationExpTooFar{ExpireAt: time.Unix(99999999999, 0)},
			},
		},
	}
	knownSigningKeys := []string{"pass", "pass123", "123", "1234", "signingkey1", "AllYourBase", "random"}
	sensitiveKeywords := []string{"password", "ip", "ssn", "covid_positive"}
	analyzer := NewWeakJWT(knownSigningKeys, sensitiveKeywords)

	trace := models.Telemetry{}
	trace.Request = &models.Request{}
	trace.Request.Common = &models.Common{}

	for i, tc := range testcases {
		trace.Request.Common.Headers = []*models.Header{
			{
				Key:   "authorization",
				Value: tc.auth,
			},
		}

		eventAnns := analyzer.Analyze(&trace)
		if !sameAnns(eventAnns, tc.wanted) {
			for _, ea := range eventAnns {
				t.Logf("   Got: %+v", ea)
			}
			for _, w := range tc.wanted {
				t.Logf("Wanted: %+v", w)
			}
			t.Errorf("^^^ Test Case number %d failed", i)
		}
	}
}
