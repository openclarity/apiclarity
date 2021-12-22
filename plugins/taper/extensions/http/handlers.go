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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/up9inc/mizu/tap/api"
)

func filterAndEmit(item *api.OutputChannelItem, emitter api.Emitter, options *api.TrafficFilteringOptions) {
	if IsIgnoredUserAgent(item, options) {
		return
	}

	if !options.DisableRedaction {
		FilterSensitiveData(item, options)
	}

	emitter.Emit(item)
}

func handleHTTP2Stream(grpcAssembler *GrpcAssembler, tcpID *api.TcpID, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	streamID, messageHTTP1, err := grpcAssembler.readMessage()
	if err != nil {
		return err
	}

	var item *api.OutputChannelItem

	switch messageHTTP1 := messageHTTP1.(type) {
	case http.Request:
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.SrcIP,
			tcpID.DstIP,
			tcpID.SrcPort,
			tcpID.DstPort,
			streamID,
		)
		item = reqResMatcher.registerRequest(ident, &messageHTTP1, superTimer.CaptureTime)
		if item != nil {
			item.ConnectionInfo = &api.ConnectionInfo{
				ClientIP:   tcpID.SrcIP,
				ClientPort: tcpID.SrcPort,
				ServerIP:   tcpID.DstIP,
				ServerPort: tcpID.DstPort,
				IsOutgoing: true,
			}
		}
	case http.Response:
		ident := fmt.Sprintf(
			"%s->%s %s->%s %d",
			tcpID.DstIP,
			tcpID.SrcIP,
			tcpID.DstPort,
			tcpID.SrcPort,
			streamID,
		)
		item = reqResMatcher.registerResponse(ident, &messageHTTP1, superTimer.CaptureTime)
		if item != nil {
			item.ConnectionInfo = &api.ConnectionInfo{
				ClientIP:   tcpID.DstIP,
				ClientPort: tcpID.DstPort,
				ServerIP:   tcpID.SrcIP,
				ServerPort: tcpID.SrcPort,
				IsOutgoing: false,
			}
		}
	}

	if item != nil {
		item.Protocol = http2Protocol
		filterAndEmit(item, emitter, options)
	}

	return nil
}

func handleHTTP1ClientStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	req, err := http.ReadRequest(b)
	if err != nil {
		return err
	}
	counterPair.Request++

	body, err := ioutil.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.SrcIP,
		tcpID.DstIP,
		tcpID.SrcPort,
		tcpID.DstPort,
		counterPair.Request,
	)
	item := reqResMatcher.registerRequest(ident, req, superTimer.CaptureTime)
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.SrcIP,
			ClientPort: tcpID.SrcPort,
			ServerIP:   tcpID.DstIP,
			ServerPort: tcpID.DstPort,
			IsOutgoing: true,
		}
		filterAndEmit(item, emitter, options)
	}
	return nil
}

func handleHTTP1ServerStream(b *bufio.Reader, tcpID *api.TcpID, counterPair *api.CounterPair, superTimer *api.SuperTimer, emitter api.Emitter, options *api.TrafficFilteringOptions) error {
	res, err := http.ReadResponse(b, nil)
	if err != nil {
		return err
	}
	counterPair.Response++

	body, err := ioutil.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind

	ident := fmt.Sprintf(
		"%s->%s %s->%s %d",
		tcpID.DstIP,
		tcpID.SrcIP,
		tcpID.DstPort,
		tcpID.SrcPort,
		counterPair.Response,
	)
	item := reqResMatcher.registerResponse(ident, res, superTimer.CaptureTime)
	if item != nil {
		item.ConnectionInfo = &api.ConnectionInfo{
			ClientIP:   tcpID.DstIP,
			ClientPort: tcpID.DstPort,
			ServerIP:   tcpID.SrcIP,
			ServerPort: tcpID.SrcPort,
			IsOutgoing: false,
		}
		filterAndEmit(item, emitter, options)
	}
	return nil
}
