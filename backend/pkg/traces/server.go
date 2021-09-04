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

package traces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/backend/pkg/common"
	"github.com/apiclarity/speculator/pkg/spec"
)

type HandleTraceFunc func(trace *spec.SCNTelemetry) error

type HttpTracesServer struct {
	traceHandleFunc HandleTraceFunc
	server          *http.Server
}

func CreateHttpTracesServer(port int, traceHandleFunc HandleTraceFunc) *HttpTracesServer {
	s := &HttpTracesServer{
		server:          &http.Server{Addr: ":" + strconv.Itoa(port)},
		traceHandleFunc: traceHandleFunc,
	}

	http.HandleFunc("/publish", s.httpTracesHandler)

	return s
}

func (o *HttpTracesServer) Start(errChan chan struct{}) {
	log.Infof("Starting traces server")

	go func() {
		if err := o.server.ListenAndServe(); err != nil {
			log.Errorf("Failed to serve traces server: %v", err)
			errChan <- common.Empty
		}
	}()
}

func (s *HttpTracesServer) Stop() {
	log.Infof("Stopping traces server")
	if s.server != nil {
		if err := s.server.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown server: %v", err)
		}
	}
}

func readHttpTraceBodyData(req *http.Request) (*spec.SCNTelemetry, error) {
	decoder := json.NewDecoder(req.Body)
	var bodyData *spec.SCNTelemetry
	err := decoder.Decode(&bodyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode trace: %v", err)
	}

	return bodyData, nil
}

func (s *HttpTracesServer) httpTracesHandler(w http.ResponseWriter, r *http.Request) {
	trace, err := readHttpTraceBodyData(r)
	if err != nil || trace == nil {
		log.Errorf("Invalid trace. err=%v, trace=%+s", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	traceB, _ := json.Marshal(trace)
	log.Infof("Trace was received: %s", traceB)
	err = s.traceHandleFunc(trace)
	if err != nil {
		log.Errorf("Failed to handle trace. err=%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof("Trace was handled successfully")
	w.WriteHeader(http.StatusAccepted)
}
