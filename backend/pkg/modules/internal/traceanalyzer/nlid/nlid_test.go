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
	"bytes"
	"testing"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	pluginmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

func sameObs(got []core.Annotation, expected []core.Annotation) bool {
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

type testCase struct {
	host       string
	path       string
	pathParams map[string]string
	method     string
	headersReq []*pluginmodels.Header
	headersRes []*pluginmodels.Header
	bodyRes    []byte
	wanted     []core.Annotation
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

		eventAnns, _ := analyzer.Analyze(tc.pathParams, &trace)
		if !sameObs(eventAnns, tc.wanted) {
			t.Errorf("Wanted: (%v) got (%v)", tc.wanted, eventAnns)
		}
	}
}

func TestNLIDHeaders(t *testing.T) {
	testcases := []testCase{
		// The API is not known, do nothing
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "12121212"}, {Key: "param2", Value: "36363636"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},

		// Nothing is learnt because IDs are too short
		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "param1", Value: "12"}, {Key: "param2", Value: "36"}}, wanted: []core.Annotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "12"}, {Key: "param2", Value: "36"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},

		// The parameter "id" matches keywords, it's checked but not yet learnt
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("12")}}},

		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}}, wanted: []core.Annotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "id", Value: "12"}, {Key: "test", Value: "36"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},

		// Let's start to learn something which is not an ID keyword
		{host: "example.com", headersReq: []*pluginmodels.Header{}, headersRes: []*pluginmodels.Header{{Key: "param1", Value: "XXXXXXXX"}, {Key: "param2", Value: "YYYYYYYY"}}, wanted: []core.Annotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "XXXXXXXX"}, {Key: "param2", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param1", Value: "AAAAAAAA"}, {Key: "param2", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("AAAAAAAA")}}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "param2", Value: "XXXXXXXX"}, {Key: "param4", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "parama", Value: "11111111"}, {Key: "paramb", Value: "22222222"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("11111111")}, {Name: "NLID", Annotation: []byte("22222222")}}},

		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "id", Value: "ééé aaAAA"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("ééé aaAAA")}}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "ééé aaAAA"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{}},

		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "b889200b-5f7e-4da7-b582-fd64f9473328"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("b889200b-5f7e-4da7-b582-fd64f9473328")}}},
		{host: "example.com", headersReq: []*pluginmodels.Header{{Key: "blabla", Value: "user_id_23654"}, {Key: "test", Value: "YYYYYYYY"}}, headersRes: []*pluginmodels.Header{}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("user_id_23654")}}},
	}

	checkTC(t, testcases)
}

func TestNLIDQueryParams(t *testing.T) {
	testCases := []testCase{
		{host: "example.com", path: "/test", wanted: []core.Annotation{}},
		{host: "example.com", path: "/test?bla=AAAAAAAAAA", wanted: []core.Annotation{}},
		{host: "example.com", path: "/test?bla=AAAAAAAAAA", wanted: []core.Annotation{}}, // bla parameter will be learnt
		{host: "example.com", path: "/test", headersReq: []*pluginmodels.Header{{Key: "test", Value: "AAAAAAAAAA"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/test", headersReq: []*pluginmodels.Header{{Key: "test", Value: "BBBbbbBBBb"}}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("BBBbbbBBBb")}}},
	}

	checkTC(t, testCases)
}

func TestNLIDBody(t *testing.T) {
	testCases := []testCase{
		{host: "example.com", path: "/test1", wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`12`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`12.5`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{}`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{"a":1}`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`[1,2,3,123654987, 1.0654, 1.111111111111111111]`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`"testtesttest"`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", bodyRes: []byte(`{"key":"blablabla", "additionalStatus": [1,2,3, 321456987654, 1.321654]}`), wanted: []core.Annotation{}},

		// Here on, the meaningfull id that should have been registered in history are
		// testtesttest
		// 123654987
		// blablabla
		// 321456987654
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "testtesttest"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "123654987"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/test1", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []core.Annotation{}},

		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "testtesttest"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "123654987"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []core.Annotation{}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []core.Annotation{}},

		{host: "example.com", path: "/testX?newid=blablabla&otherid=testtesttest", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []core.Annotation{}},

		// We don't check for NLIDs in query parameters (we only learn them)
		{host: "example.com", path: "/testX?newid=blablabla&otherid=testtesttest&strange=THISISNOTCHECKED", headersReq: []*pluginmodels.Header{{Key: "param", Value: "321456987654"}}, wanted: []core.Annotation{}},

		// Now let's check for some NLIDs
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234-NLID-5678"}}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("1234-NLID-5678")}}},
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234-NLID-5678"}, {Key: "id", Value: "123654987"}}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("1234-NLID-5678")}}},

		// The param is too small, it's probably not an ID, don't check for it
		{host: "example.com", path: "/testX", headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234"}}, wanted: []core.Annotation{}},

		// Check if path parameters are NLIDs
		{host: "example.com", method: "GET", path: "/pet/b889200b-5f7e-4da7-b582-fd64f9473328", pathParams: map[string]string{"petID": "b889200b-5f7e-4da7-b582-fd64f9473328"}, headersReq: []*pluginmodels.Header{{Key: "param", Value: "1234"}}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("b889200b-5f7e-4da7-b582-fd64f9473328")}}},

		// The "normal" behavior: Create an object with a POST query, get back an ID, Query this id
		{host: "example.com", method: "POST", path: "/pet", bodyRes: []byte(`{"id": 10, "name": "doggie", "category": {"id": 1, "name": "Dogs"}, "photoUrls": ["string"], "tags": [{"id": 0, "name": "string"}], "status": "available"}`), wanted: []core.Annotation{}},
		{host: "example.com", method: "GET", path: "/pet/10", pathParams: map[string]string{"petID": "10"}, wanted: []core.Annotation{}},
		{host: "example.com", method: "GET", path: "/pet/11", pathParams: map[string]string{"petID": "11"}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("11")}}},

		// Another "normal" behavior: GET a list of IDs then get details about one of those IDs

		{host: "example.com", method: "GET", path: "/pet", bodyRes: []byte(`{"ids":[12345678,12345679]}`), wanted: []core.Annotation{}},
		{host: "example.com", method: "GET", path: "/pet/444", pathParams: map[string]string{"petID": "12345678"}, wanted: []core.Annotation{}},
		{host: "example.com", method: "GET", path: "/pet/445", pathParams: map[string]string{"petID": "12345679"}, wanted: []core.Annotation{}},
		{host: "example.com", method: "GET", path: "/pet/448", pathParams: map[string]string{"petID": "12345670"}, wanted: []core.Annotation{{Name: "NLID", Annotation: []byte("12345670")}}},

		{host: "example.com", path: "/test3", bodyRes: []byte(`{"paramA": 123456789, "param": "blablabla"}`), wanted: []core.Annotation{}},
		{host: "example.com", path: "/test4", headersReq: []*pluginmodels.Header{{Key: "param", Value: "blablabla"}}, wanted: []core.Annotation{}},
	}

	checkTC(t, testCases)
}
