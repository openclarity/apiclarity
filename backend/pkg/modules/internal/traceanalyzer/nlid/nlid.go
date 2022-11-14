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
	"container/ring"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	NLIDRingBufferSize = 1024
)

const (
	MinIDValueLength = 8
	MaxIDValueLength = 40 // This is at least a UUID. every string larger that than is considered low chance of being an ID
)

var IDKeys = [...]string{"id", "ids", "identifier", "identifiers"}

type params = map[string]bool

type Reason map[string]interface{}

type parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NLID struct {
	historySize   int
	paramsHistory map[utils.API]*ring.Ring
}

func NewNLID(historySize int) *NLID {
	return &NLID{
		historySize:   historySize,
		paramsHistory: make(map[utils.API]*ring.Ring),
	}
}

func (n *NLID) Analyze(path, method string, pathParams map[string]string, trace *pluginsmodels.Telemetry) (eventAnns []utils.TraceAnalyzerAnnotation, apiAnns []utils.TraceAnalyzerAPIAnnotation) {
	if n.skipTrace(trace) {
		return
	}

	params := n.getNLIDS(pathParams, *trace)
	if len(params) > 0 {
		eventAnns = append(eventAnns, NewAnnotationNLID(path, method, params))
	}

	n.learnIDs(*trace)

	return eventAnns, apiAnns
}

// We check if a variable is a NLID if, for a Request:
// - There is already an history of variables for this API
// - It's in the Request header AND looks like an ID
// - XXX: it in the path parameter AND it looks like an ID

func (n *NLID) getNLIDS(pathParams map[string]string, trace pluginsmodels.Telemetry) (NLIDparams []parameter) {
	api := getAPI(trace)

	ph, ok := n.paramsHistory[api]
	if !ok { // There is no history for this API yet
		return
	}

	// Get all parameters of the Request

	// Get all parameters
	reqParamsList := map[string][]parameter{}

	// - Header
	for _, h := range trace.Request.Common.Headers {
		if maybeID(h.Key, h.Value) {
			reqParamsList[h.Value] = append(reqParamsList[h.Value], parameter{Name: h.Key, Value: h.Value})
		}
	}

	// - Path
	for k, v := range pathParams {
		if maybeID(k, v) {
			reqParamsList[v] = append(reqParamsList[v], parameter{Name: k, Value: v})
		}
	}

	// - Query Params, XXX Don't do it right now
	// - Body, XXX Don't do it right now

	r := ph
	r.Do(func(p interface{}) {
		if p == nil {
			return
		}
		prevParams, ok := p.(params)
		if !ok {
			return
		}

		for param := range reqParamsList {
			if prevParams[param] {
				// This parameter was already present, that OK, it's not an NLID
				// Remove it from the parameters to checks
				delete(reqParamsList, param)
			}
		}
	})

	// Here, if reqParams is empty, this means that all parameters from the
	// Request were found in the history, there is no observation to
	// return. The parameters that are left in reqParams are the one which
	// were not found in the history, meaning that they are non learnt IDs
	for value := range reqParamsList {
		NLIDparams = append(NLIDparams, reqParamsList[value]...)
	}

	return NLIDparams
}

// We store an id if:
// - It's in the Response headers AND it looks like an ID.
// - It's in the Query parameters AND it looks like an ID.
// - It's in the flattened Response Body AND it's a number or a string that looks like an ID.
func (n *NLID) learnIDs(trace pluginsmodels.Telemetry) {
	params := make(params)
	api := getAPI(trace)

	// Get all parameters of the response
	// - Header
	for _, h := range trace.Response.Common.Headers {
		if maybeID(h.Key, h.Value) {
			params[h.Value] = true
		}
	}

	// Query Parameters
	u, err := url.Parse(trace.Request.Path)
	if err == nil {
		m, _ := url.ParseQuery(u.RawQuery)
		for k, v := range m {
			if maybeID(k, v[0]) {
				params[v[0]] = true // XXX: only get the first value of each query parameter
			}
		}
	}
	// - Learn Response Body parameters
	if !trace.Response.Common.TruncatedBody && len(trace.Response.Common.Body) > 0 {
		if err = getBodyParams(params, trace.Response.Common.Body); nil != err {
			// Log the problem, but continue anyway, it's not blocking.
			log.Debugf("unable to get parameters from body: %v", err)
		}
	}

	if _, found := n.paramsHistory[api]; !found {
		n.paramsHistory[api] = ring.New(n.historySize)
	}
	n.paramsHistory[api].Value = params
	n.paramsHistory[api] = n.paramsHistory[api].Next()
}

func (n *NLID) skipTrace(trace *pluginsmodels.Telemetry) bool {
	return false
}

func getBodyParams(params params, body []byte) error {
	var parsed interface{}

	// We deserialize the JSON object this way because json.Unmarshal doesn't
	// distinguish between int and floats. Here, thanks to d.UseNumber(), we
	// can switch on json.Number and then try to cast to Int64.
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	if err := d.Decode(&parsed); err != nil {
		return fmt.Errorf("unable to decode body: %w", err)
	}

	getDecodedBodyParams(params, parsed, "")

	return nil
}

func getDecodedBodyParams(o params, val interface{}, prefix string) {
	switch val := val.(type) {
	case bool:
		// Nothing, it's not interresting
	case json.Number:
		// We only add integers to the possible list of identifier.
		// Floats are unlikely indentifiers.
		if n, err := val.Int64(); err == nil {
			s := fmt.Sprintf("%d", n)
			if maybeID(prefix, s) {
				(o)[s] = true
			}
		}
	case string:
		if maybeID(prefix, val) {
			(o)[val] = true
		}
	case []interface{}:
		for i, v := range val {
			getDecodedBodyParams(o, v, fmt.Sprintf("%s[%d]", prefix, i))
		}
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			getDecodedBodyParams(o, val[k], fmt.Sprintf("%s.%s", prefix, k))
		}
	case nil:
		// nothing
	default:
		// nothing
	}
}

func maybeID(key string, value string) bool {
	// Check if the key looks like it's an identifier
	keyLower := strings.ToLower(key)
	for _, id := range IDKeys {
		if strings.HasSuffix(keyLower, id) {
			return true
		}
	}

	if len(value) < MinIDValueLength || len(value) > MaxIDValueLength {
		return false
	}

	for _, c := range value {
		if !(('a' <= c && c <= 'z') ||
			('A' <= c && c <= 'Z') ||
			('0' <= c && c <= '9') ||
			c == '_' ||
			c == '-') {
			return false
		}
	}

	// If we passed all the check, then it's maybe an ID
	return true
}

func getAPI(trace pluginsmodels.Telemetry) utils.API {
	api := trace.Request.Host

	return api
}
