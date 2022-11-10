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
	"testing"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

func createTrace(headersReq, headersRes []*models.Header, bodyReq, bodyRes []byte) *models.Telemetry {
	trace := models.Telemetry{
		Request: &models.Request{
			Common: &models.Common{
				Headers: headersReq,
				Body:    bodyReq,
			},
		},
		Response: &models.Response{
			Common: &models.Common{
				Headers: headersRes,
				Body:    bodyRes,
			},
		},
	}

	return &trace
}

func sameMatch(a, b RuleMatch) bool {
	if a.InRequestBody != b.InRequestBody ||
		a.InResponseBody != b.InResponseBody ||
		a.InRequestHeaders != b.InRequestHeaders ||
		a.InResponseHeaders != b.InResponseHeaders {
		return false
	}

	if a.Rule.ID != b.Rule.ID {
		return false
	}

	return true
}

func sameRegexpMatch(got, expected AnnotationRegexpMatching) bool {
	if len(got.Matches) != len(expected.Matches) {
		return false
	}
	for i := range expected.Matches {
		if !sameMatch(expected.Matches[i], got.Matches[i]) {
			return false
		}
	}

	return true
}

func sameRegexpMatches(got []utils.TraceAnalyzerAnnotation, expected []AnnotationRegexpMatching) bool {
	if len(got) != len(expected) {
		return false
	}

	for _, eo := range expected { // For each wanted observation
		found := false
		for _, o := range got { // Check if it's in the result
			if toRegexpMatch, ok := o.(*AnnotationRegexpMatching); ok {
				if sameRegexpMatch(*toRegexpMatch, eo) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}

var testRulesFiles = []string{"./sensitive_test_01.yaml"}

type testCase struct {
	description string
	headersReq  []*models.Header
	headersRes  []*models.Header
	bodyReq     []byte
	bodyRes     []byte

	wanted []AnnotationRegexpMatching
}

func TestSensitive(t *testing.T) {
	testCases := []testCase{
		{
			description: "Simple test, nothing matches",
			headersReq:  []*models.Header{{Key: "header1_req", Value: "header1_req_value"}},
			headersRes:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			bodyReq:     []byte(""),
			bodyRes:     []byte(""),
			wanted:      []AnnotationRegexpMatching{},
		},
		{
			description: "Username should match in request headers",
			headersReq:  []*models.Header{{Key: "username", Value: "XXX"}},
			headersRes:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			bodyReq:     []byte(""),
			bodyRes:     []byte(""),
			wanted:      []AnnotationRegexpMatching{{Matches: []RuleMatch{{Rule: &Rule{ID: "simple-001"}, InRequestHeaders: true}}}},
		},
		{
			description: "Username should match in response headers",
			headersReq:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			headersRes:  []*models.Header{{Key: "username", Value: "XXX"}},
			bodyReq:     []byte(""),
			bodyRes:     []byte(""),
			wanted:      []AnnotationRegexpMatching{{Matches: []RuleMatch{{Rule: &Rule{ID: "simple-001"}, InResponseHeaders: true}}}},
		},
		{
			description: "Username should match in request body",
			headersReq:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			headersRes:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			bodyReq:     []byte("... username ..."),
			bodyRes:     []byte(""),
			wanted:      []AnnotationRegexpMatching{{Matches: []RuleMatch{{Rule: &Rule{ID: "simple-001"}, InRequestBody: true}}}},
		},
		{
			description: "Username should match in response body",
			headersReq:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			headersRes:  []*models.Header{{Key: "header1_res", Value: "header1_res_value"}},
			bodyReq:     []byte(""),
			bodyRes:     []byte("... username ..."),
			wanted:      []AnnotationRegexpMatching{{Matches: []RuleMatch{{Rule: &Rule{ID: "simple-001"}, InResponseBody: true}}}},
		},
		{
			description: "Username should match everywhere",
			headersReq:  []*models.Header{{Key: "username", Value: "XXX"}},
			headersRes:  []*models.Header{{Key: "username", Value: "XXX"}},
			bodyReq:     []byte("... username ..."),
			bodyRes:     []byte("... username ..."),
			wanted:      []AnnotationRegexpMatching{{Matches: []RuleMatch{{Rule: &Rule{ID: "simple-001"}, InRequestHeaders: true, InResponseHeaders: true, InRequestBody: true, InResponseBody: true}}}},
		},
		{
			description: "Two rules are matching at the same time",
			headersReq:  []*models.Header{{Key: "username", Value: "XXX"}, {Key: "api-key", Value: "XXX"}},
			headersRes:  []*models.Header{{Key: "username", Value: "XXX"}},
			bodyReq:     []byte(""),
			bodyRes:     []byte(""),
			wanted: []AnnotationRegexpMatching{
				{Matches: []RuleMatch{
					{Rule: &Rule{ID: "simple-001"}, InRequestHeaders: true, InResponseHeaders: true},
					{Rule: &Rule{ID: "core-002"}, InRequestHeaders: true},
				}},
			},
		},
	}
	sensitive, err := NewSensitive(testRulesFiles)
	if err != nil {
		t.Fatalf("unable to create sensitive analyzer: %s", err)
	}

	for _, tc := range testCases {
		trace := createTrace(tc.headersReq, tc.headersRes, tc.bodyReq, tc.bodyRes)
		eventAnns := sensitive.Analyze(trace)
		if !sameRegexpMatches(eventAnns, tc.wanted) {
			for _, ea := range eventAnns {
				t.Logf("   Got: %+v", ea)
			}
			for _, w := range tc.wanted {
				t.Logf("Wanted: %+v", w)
			}
			t.Errorf("^^^ %s", tc.description)
		}
	}
}
