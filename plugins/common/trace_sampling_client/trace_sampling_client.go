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

package trace_sampling_client

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/plugins/common"
	"github.com/openclarity/trace-sampling-manager/api/client/client"
	"github.com/openclarity/trace-sampling-manager/api/client/client/operations"
)

type Client struct {
	TraceSamplingManagerClient *client.TraceSamplingManager
	// local cache of hosts to trace
	Hosts            map[string]bool
	samplingInterval time.Duration
	lock             sync.RWMutex
}

const allHosts = "*"

func Create(enableTLS bool, host string, samplingInterval time.Duration) (*Client, error) {
	var tlsOptions *common.ClientTLSOptions
	if enableTLS {
		tlsOptions = &common.ClientTLSOptions{
			RootCAFileName: common.CACertFile,
		}
	}
	apiClient, err := common.NewTraceSamplingAPIClient(host, tlsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create new trace sampling api client: %v", err)
	}

	return &Client{
		TraceSamplingManagerClient: apiClient,
		samplingInterval:           samplingInterval,
		Hosts:                      map[string]bool{},
	}, nil
}

func (t *Client) Start() {
	go func() {
		for {
			select {
			case <-time.After(t.samplingInterval):
				params := operations.NewGetHostsToTraceParams()

				response, err := t.TraceSamplingManagerClient.Operations.GetHostsToTrace(params)
				if err != nil {
					log.Errorf("Failed to get hosts to trace: %v", err)
				} else {
					t.setHosts(response.Payload.Hosts)
				}
			}
		}
	}()
}

func (t *Client) ShouldTrace(host, port string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if len(t.Hosts) == 0 {
		return false
	}

	if port != "" {
		host = host + ":" + port
	}

	if t.Hosts[allHosts] || t.Hosts[host] {
		return true
	}

	return false
}

func (t *Client) setHosts(hosts []string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.Hosts = make(map[string]bool)
	for _, host := range hosts {
		t.Hosts[host] = true
	}
}
