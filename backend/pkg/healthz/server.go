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

package healthz

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/common"
)

type Server struct {
	listenAddress string
	isReady       bool
}

func (s *Server) SetIsReady(isReady bool) {
	s.isReady = isReady
}

// Start starts the server run.
func (s *Server) Start(errChan chan struct{}) {
	log.Infof("Starting healthz server. listenAddr=%v", s.listenAddress)

	http.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if s.isReady {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	go func() {
		if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
			log.WithError(err).Error("Failed to serve.")
			errChan <- common.Empty
		}
	}()
}

func NewHealthServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
	}
}
