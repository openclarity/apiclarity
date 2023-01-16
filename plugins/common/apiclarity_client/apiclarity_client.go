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

package apiclarity_client

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/plugins/api/client/client"
	"github.com/openclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/openclarity/apiclarity/plugins/api/client/models"
	"github.com/openclarity/apiclarity/plugins/common"
)

type Client struct {
	telemetriesAPI *client.APIClarityPluginsTelemetriesAPI
	// local cache of hosts to trace
	HostsToTrace     map[string]bool
	samplingInterval time.Duration
	lock             sync.RWMutex
	token            string
}

const allHosts = "*"

func Create(enableTLS bool, host string, token string, samplingInterval time.Duration) (*Client, error) {
	var tlsOptions *common.ClientTLSOptions
	if enableTLS {
		tlsOptions = &common.ClientTLSOptions{
			RootCAFileName: common.CACertFile,
		}
	}
	apiClient, err := common.NewTelemetryAPIClient(host, tlsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create new telemetry api client: %v", err)
	}

	return &Client{
		telemetriesAPI:   apiClient,
		samplingInterval: samplingInterval,
		HostsToTrace:     map[string]bool{},
		token:            token,
	}, nil
}

func (t *Client) Start() {
	go func() {
		for {
			select {
			case <-time.After(t.samplingInterval):
				_ = t.RefreshHostsToTrace()
			}
		}
	}()
}

func (t *Client) ShouldTrace(host, port string) bool {
	if len(t.HostsToTrace) == 0 {
		return false
	}

	if port != "" {
		host = host + ":" + port
	}

	if t.HostsToTrace[allHosts] || t.HostsToTrace[host] {
		return true
	}

	return false
}

func (t *Client) setHostsToTrace(hosts []string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.HostsToTrace = make(map[string]bool)
	for _, host := range hosts {
		t.HostsToTrace[host] = true
	}
}

func (t *Client) RefreshHostsToTrace() error {
	params := operations.NewGetHostsToTraceParams().
		WithXTraceSourceToken(&t.token)

	response, err := t.telemetriesAPI.Operations.GetHostsToTrace(params)
	if err != nil {
		log.Errorf("Failed to get hosts to trace: %v", err)
		return err
	} else {
		t.setHostsToTrace(response.Payload.Hosts)
	}
	return nil
}

func (t *Client) PostTelemetry(telemetry *models.Telemetry) error {
	params := operations.NewPostTelemetryParams().
		WithBody(telemetry).
		WithXTraceSourceToken(&t.token)

	_, err := t.telemetriesAPI.Operations.PostTelemetry(params)
	if err != nil {
		log.Errorf("Failed to post telemetry: %v", err)
		return err
	}
	return nil
}

func (t *Client) PostNewDiscoveredAPIs(hosts []string) error {
	body := operations.PostControlNewDiscoveredAPIsBody{
		Hosts: hosts,
	}
	params := operations.NewPostControlNewDiscoveredAPIsParams().
		WithBody(body).
		WithXTraceSourceToken(&t.token)

	_, err := t.telemetriesAPI.Operations.PostControlNewDiscoveredAPIs(params)
	if err != nil {
		log.Errorf("Failed to post NewDiscoveredAPIs: %v", err)
		return err
	}
	return nil
}
