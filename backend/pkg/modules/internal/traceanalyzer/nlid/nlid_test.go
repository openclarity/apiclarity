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
	"reflect"
	"testing"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	pluginmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

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

type testCase struct {
	description string
	host        string
	path        string
	pathParams  map[string]string
	method      string
	headersReq  []*pluginmodels.Header
	headersRes  []*pluginmodels.Header
	bodyRes     []byte
	wanted      []utils.TraceAnalyzerAnnotation
}

func checkTC(t *testing.T, testCases []testCase) {
	t.Helper()
	analyzer := NewNLID(NLIDRingBufferSize)

	trace := pluginmodels.Telemetry{}
	trace.Request = &pluginmodels.Request{}
	trace.Request.Common = &pluginmodels.Common{}
	trace.Response = &pluginmodels.Response{}
	trace.Response.Common = &pluginmodels.Common{}

	for _, tc := range testCases {
		trace.Request.Host = tc.host
		trace.Request.Common.Headers = tc.headersReq
		trace.Response.Common.Headers = tc.headersRes
		trace.Request.Path = tc.path
		trace.Request.Method = tc.method
		trace.Response.Common.Body = tc.bodyRes

		eventAnns, _ := analyzer.Analyze(tc.path, tc.method, tc.pathParams, &trace)
		if !sameAnns(eventAnns, tc.wanted) {
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

func TestNLIDHeaders(t *testing.T) {
	testcases := []testCase{
		// The API is not known, do nothing
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "12121212"}, {Key: "param2", Value: "36363636"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// Nothing is learnt because IDs are too short
		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "param1", Value: "12"}, {Key: "param2", Value: "36"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "12"}, {Key: "param2", Value: "36"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// The parameter "id" matches keywords, it's checked but not yet learnt
		{
			description: "Value '12' was not previously learnt",
			host:        "example.com",
			path:        "",
			pathParams:  map[string]string{},
			method:      "",
			headersReq:  []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}},
			headersRes:  []*pluginmodels.Header{},
			bodyRes:     []byte{},
			wanted:      []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "12"}})},
		},

		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// Let's start to learn something which is not an ID keyword
		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "param1", Value: "XXXXXXXX"}, {Key: "param2", Value: "YYYYYYYY"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "XXXXXXXX"}, {Key: "param2", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "AAAAAAAA"}, {Key: "param2", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "AAAAAAAA"}})}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param2", Value: "XXXXXXXX"}, {Key: "param4", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "parama", Value: "11111111"}, {Key: "paramb", Value: "22222222"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "11111111"}, {Name: "XXX", Value: "22222222"}})}},

		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "id", Value: "ééé aaAAA"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "ééé aaAAA"}})}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "ééé aaAAA"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{}},

		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "b889200b-5f7e-4da7-b582-fd64f9473328"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "b889200b-5f7e-4da7-b582-fd64f9473328"}})}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "user_id_23654"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("", "", []parameter{{Name: "XXX", Value: "user_id_23654"}})}},
	}

	checkTC(t, testcases)
}

func TestNLIDQueryParams(t *testing.T) {
	testCases := []testCase{
		{host: "example.com", path: "/test", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test?bla=AAAAAAAAAA", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test?bla=AAAAAAAAAA", wanted: []utils.TraceAnalyzerAnnotation{}}, // bla parameter will be learnt
		{host: "example.com", path: "/test", headersReq: []*pluginmodels.Header{{Key: "test", Value: "AAAAAAAAAA"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test", headersReq: []*pluginmodels.Header{{Key: "test", Value: "BBBbbbBBBb"}}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/test", "", []parameter{{Name: "XXX", Value: "BBBbbbBBBb"}})}},
	}

	checkTC(t, testCases)
}

func TestNLIDBody(t *testing.T) {
	testCases := []testCase{
		{host: "example.com", path: "/test1", wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`12`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`12.5`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{}`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{"a":1}`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`[1,2,3,123654987, 1.0654, 1.111111111111111111]`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`"testtesttest"`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{"key":"blablabla", "additionalStatus": [1,2,3, 321456987654, 1.321654]}`), wanted: []utils.TraceAnalyzerAnnotation{}},

		// Here on, the meaningfull id that should have been registered in history are
		// testtesttest
		// 123654987
		// blablabla
		// 321456987654
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "testtesttest"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "123654987"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []utils.TraceAnalyzerAnnotation{}},

		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "testtesttest"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "123654987"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []utils.TraceAnalyzerAnnotation{}},

		{host: "example.com", path: "/testX?newid=blablabla&otherid=testtesttest", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// We don't check for NLIDs in query parameters (we only learn them)
		{host: "example.com", path: "/testX?newid=blablabla&otherid=testtesttest&strange=THISISNOTCHECKED", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// Now let's check for some NLIDs
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234-NLID-5678"}}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/testX", "", []parameter{{Name: "XXX", Value: "1234-NLID-5678"}})}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234-NLID-5678"}, {Key: "id", Value: "123654987"}}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/testX", "", []parameter{{Name: "XXX", Value: "1234-NLID-5678"}})}},

		// The param is too small, it's probably not an ID, don't check for it
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234"}}, wanted: []utils.TraceAnalyzerAnnotation{}},

		// Check if path parameters are NLIDs
		{host: "example.com", method: "GET", path: "/pet/b889200b-5f7e-4da7-b582-fd64f9473328", pathParams: map[string]string{"petID": "b889200b-5f7e-4da7-b582-fd64f9473328"}, headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234"}}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/pet/b889200b-5f7e-4da7-b582-fd64f9473328", "get", []parameter{{Name: "XXX", Value: "b889200b-5f7e-4da7-b582-fd64f9473328"}})}},

		// The "normal" behavior: Create an object with a POST query, get back an ID, Query this id
		{host: "example.com", method: "POST", path: "/pet", bodyRes: []byte(`{"id": 10, "name": "doggie", "category": {"id": 1, "name": "Dogs"}, "photoUrls": ["string"], "tags": [{"id": 0, "name": "string"}], "status": "available"}`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", method: "GET", path: "/pet/10", pathParams: map[string]string{"petID": "10"}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", method: "GET", path: "/pet/11", pathParams: map[string]string{"petID": "11"}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/pet/11", "get", []parameter{{Name: "XXX", Value: "11"}})}},

		// Another "normal" behavior: GET a list of IDs then get details about one of those IDs

		{host: "example.com", method: "GET", path: "/pet", bodyRes: []byte(`{"ids":[12345678,12345679]}`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", method: "GET", path: "/pet/444", pathParams: map[string]string{"petID": "12345678"}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", method: "GET", path: "/pet/445", pathParams: map[string]string{"petID": "12345679"}, wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", method: "GET", path: "/pet/448", pathParams: map[string]string{"petID": "12345670"}, wanted: []utils.TraceAnalyzerAnnotation{NewAnnotationNLID("/pet/448", "get", []parameter{{Name: "XXX", Value: "12345670"}})}},

		{host: "example.com", path: "/test3", bodyRes: []byte(`{"paramA": 123456789, "param": "blablabla"}`), wanted: []utils.TraceAnalyzerAnnotation{}},
		{host: "example.com", path: "/test4", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []utils.TraceAnalyzerAnnotation{}},
	}

	checkTC(t, testCases)
}
