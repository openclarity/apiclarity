// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

// From: https://github.com/up9inc/mizu/tree/main/tap/extensions/http

package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

var reqResMatcher = createResponseRequestMatcher() // global

// Key is {client_addr}:{client_port}->{dest_addr}:{dest_port}_{incremental_counter}
type requestResponseMatcher struct {
	openMessagesMap *sync.Map
}

func createResponseRequestMatcher() requestResponseMatcher {
	newMatcher := &requestResponseMatcher{openMessagesMap: &sync.Map{}}
	return *newMatcher
}

func (matcher *requestResponseMatcher) registerRequest(ident string, request *http.Request, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	requestHTTPMessage := api.GenericMessage{
		IsRequest:   true,
		CaptureTime: captureTime,
		Payload: api.HTTPPayload{
			Type: TypeHttpRequest,
			Data: request,
		},
	}

	if response, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		responseHTTPMessage := response.(*api.GenericMessage)
		if responseHTTPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(&requestHTTPMessage, responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &requestHTTPMessage)
	return nil
}

func (matcher *requestResponseMatcher) registerResponse(ident string, response *http.Response, captureTime time.Time) *api.OutputChannelItem {
	split := splitIdent(ident)
	key := genKey(split)

	responseHTTPMessage := api.GenericMessage{
		IsRequest:   false,
		CaptureTime: captureTime,
		Payload: api.HTTPPayload{
			Type: TypeHttpResponse,
			Data: response,
		},
	}

	if request, found := matcher.openMessagesMap.LoadAndDelete(key); found {
		// Type assertion always succeeds because all of the map's values are of api.GenericMessage type
		requestHTTPMessage := request.(*api.GenericMessage)
		if !requestHTTPMessage.IsRequest {
			return nil
		}
		return matcher.preparePair(requestHTTPMessage, &responseHTTPMessage)
	}

	matcher.openMessagesMap.Store(key, &responseHTTPMessage)
	return nil
}

func (matcher *requestResponseMatcher) preparePair(requestHTTPMessage *api.GenericMessage, responseHTTPMessage *api.GenericMessage) *api.OutputChannelItem {
	return &api.OutputChannelItem{
		Protocol:       protocol,
		Timestamp:      requestHTTPMessage.CaptureTime.UnixNano() / int64(time.Millisecond),
		ConnectionInfo: nil,
		Pair: &api.RequestResponsePair{
			Request:  *requestHTTPMessage,
			Response: *responseHTTPMessage,
		},
	}
}

func splitIdent(ident string) []string {
	ident = strings.Replace(ident, "->", " ", -1)
	return strings.Split(ident, " ")
}

func genKey(split []string) string {
	key := fmt.Sprintf("%s:%s->%s:%s,%s", split[0], split[2], split[1], split[3], split[4])
	return key
}
