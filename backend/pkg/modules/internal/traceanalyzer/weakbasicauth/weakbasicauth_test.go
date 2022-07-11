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
	"reflect"
	"testing"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

func TestFindToken(t *testing.T) {
	testcases := []struct {
		headers []*models.Header
		wanted  [2]string
	}{
		{headers: []*models.Header{{Key: "authorization", Value: "Basic dXNlcjE6cGFzcw=="}}, wanted: [2]string{"user1", "pass"}},
		{headers: []*models.Header{{Key: "AuthoRizaTioN", Value: "Basic dXNlcjE6cGFzcw=="}}, wanted: [2]string{"user1", "pass"}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic dXNlcjE6cGFzcw=="}}, wanted: [2]string{"user1", "pass"}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4Onl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eXl5eQ=="}}, wanted: [2]string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic: dXNlcjE6cGFzcw=="}}, wanted: [2]string{"", ""}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic "}}, wanted: [2]string{"", ""}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic"}}, wanted: [2]string{"", ""}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "basic"}}, wanted: [2]string{"", ""}},
		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "basic dXNlcjE6cGFzcw=="}}, wanted: [2]string{"", ""}},

		{headers: []*models.Header{{Key: "AUTHORIZATION", Value: "Basic BADBASE64=="}}, wanted: [2]string{"", ""}},
	}

	trace := models.Telemetry{}
	trace.Request = &models.Request{}
	trace.Request.Common = &models.Common{}
	for _, tc := range testcases {
		trace.Request.Common.Headers = tc.headers
		gotUser, gotPassword, _ := findBasicAuthToken(&trace)
		user, password := tc.wanted[0], tc.wanted[1]
		if gotUser != user && gotPassword != password {
			t.Errorf("Wanted (%v, %v) got (%v, %v)", user, password, gotUser, gotPassword)
		}
	}
}

func sameAnns(got []utils.TraceAnalyzerAnnotation, expected []utils.TraceAnalyzerAnnotation) bool {
	if len(got) != len(expected) {
		return false
	}
	for _, eo := range expected { // For each wanted observation
		found := false
		for _, o := range got { // Check if it's in the result
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

func TestBasicAuth(t *testing.T) {
	testcases := []struct {
		headers []*models.Header
		wanted  []utils.TraceAnalyzerAnnotation
	}{
		{headers: []*models.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{headers: []*models.Header{{Key: "authorization", Value: "Basic "}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{headers: []*models.Header{{Key: "authorization", Value: "basic"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{headers: []*models.Header{{Key: "authorization", Value: "XXXX     basic"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{headers: []*models.Header{{Key: "authorization", Value: "XXXX     Basic "}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{headers: []*models.Header{{Key: "authorization", Value: "Basic dXNlcjE6cGFzcw=="}}, wanted: []utils.TraceAnalyzerAnnotation{&AnnotationShortPassword{Password: "pass", Length: 4, MinSize: 8}, &AnnotationKnownPassword{Password: "pass"}}},
		{headers: []*models.Header{{Key: "authorization", Value: "Basic dXNlcjE6bG9uZ2xvbmdsb25n"}}, wanted: []utils.TraceAnalyzerAnnotation{&AnnotationKnownPassword{Password: "longlonglong"}}},
	}

	knownPasswords := []string{"pass", "pass123", "123", "1234", "longlonglong"}
	analyzer := NewWeakBasicAuth(knownPasswords)

	trace := models.Telemetry{}
	trace.Request = &models.Request{}
	trace.Request.Common = &models.Common{}
	for _, tc := range testcases {
		trace.Request.Common.Headers = tc.headers
		eventAnns := analyzer.Analyze(&trace)
		if !sameAnns(eventAnns, tc.wanted) {
			for _, ea := range eventAnns {
				t.Logf("Got: %+v", ea)
			}
			for _, w := range tc.wanted {
				t.Logf("Wanted: %+v", w)
			}
			t.Errorf("Wanted: (%+v) got (%+v)", tc.wanted, eventAnns)
		}
	}
}

func TestSamePassword(t *testing.T) {
	testcases := []struct {
		host   string
		auth   string
		wanted []utils.TraceAnalyzerAnnotation
	}{
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example2.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{&AnnotationSamePassword{User: "user1", Password: "passwordmorethan8", APIs: []string{"example1.com", "example2.com"}}}},
		{host: "example2.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example2.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example3.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{&AnnotationSamePassword{User: "user1", Password: "passwordmorethan8", APIs: []string{"example1.com", "example2.com", "example3.com"}}}},
		{host: "example1.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example2.com", auth: "Basic dXNlcjE6cGFzc3dvcmRtb3JldGhhbjg=", wanted: []utils.TraceAnalyzerAnnotation{}},
	}
	var knownPasswords []string
	analyzer := NewWeakBasicAuth(knownPasswords)

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
		trace.Request.Host = tc.host
		eventAnns := analyzer.Analyze(&trace)
		if !sameAnns(eventAnns, tc.wanted) {
			for _, ea := range eventAnns {
				t.Logf("Got: %+v", ea)
			}
			for _, w := range tc.wanted {
				t.Logf("Wanted: %+v", w)
			}
			t.Errorf("Wanted: (%+v) got (%+v)", tc.wanted, eventAnns)
		}
	}
}
