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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"plugin"
	"sort"
	"strconv"
	"syscall"

	logutils "github.com/Portshift/go-utils/log"
	log "github.com/sirupsen/logrus"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap"
	"github.com/up9inc/mizu/tap/api"
	"github.com/urfave/cli"

	"github.com/openclarity/apiclarity/plugins/api/client/client"
	"github.com/openclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/openclarity/apiclarity/plugins/api/client/models"
	"github.com/openclarity/apiclarity/plugins/common"
	"github.com/openclarity/apiclarity/plugins/common/trace_sampling_client"
	"github.com/openclarity/apiclarity/plugins/taper/config"
	"github.com/openclarity/apiclarity/plugins/taper/monitor"
)

type Agent struct {
	podMonitor          *monitor.PodMonitor
	apiClient           *client.APIClarityPluginsTelemetriesAPI
	traceSamplingClient *trace_sampling_client.Client
}

func run(c *cli.Context) {
	logutils.InitLogs(c, os.Stdout)

	runConfig := config.LoadConfig()
	log.Infof("Loaded config: %+v", runConfig)

	// load http plugin
	extensions, err := loadExtensions()
	if err != nil {
		log.Errorf("Failed to load extensions: %v", err)
		return
	}
	opts := &tap.TapOpts{
		HostMode: true,
	}

	var tlsOptions *common.ClientTLSOptions
	if runConfig.EnableTLS {
		tlsOptions = &common.ClientTLSOptions{
			RootCAFileName: runConfig.RootCertFilePath,
		}
	}
	apiClient, err := common.NewTelemetryAPIClient(runConfig.UpstreamAddress, tlsOptions)
	if err != nil {
		log.Errorf("Failed to create new api client: %v", err)
		return
	}
	agent := &Agent{
		apiClient: apiClient,
	}

	if runConfig.TraceSamplingEnabled {
		TSM, err := trace_sampling_client.Create(false, runConfig.TraceSamplingManagerAddress, common.SamplingInterval)
		if err != nil {
			log.Errorf("Failed to create trace sampling client: %v", err)
			return
		}
		TSM.Start()
		agent.traceSamplingClient = TSM
	}
	// set mizu logger
	logger.InitLoggerStderrOnly(runConfig.MizuLogLevel)

	podMonitor, err := monitor.NewPodMonitor(runConfig.NamespaceToTap)
	if err != nil {
		log.Errorf("Failed to create pod monitor: %v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	agent.podMonitor = podMonitor
	go podMonitor.Start(ctx)

	outputItems := make(chan *api.OutputChannelItem)
	options := &api.TrafficFilteringOptions{}
	tap.StartPassiveTapper(opts, outputItems, extensions, options)

	go agent.startReadOutputItems(ctx, outputItems)

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-sig
	log.Warningf("Received a termination signal: %v", s)
	cancel()
}

func (a *Agent) startReadOutputItems(ctx context.Context, outputItems chan *api.OutputChannelItem) {
	log.Info("Starting to read output items")
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-outputItems:
			go func() {
				log.Infof("Output item received: connection info: %+v", item.ConnectionInfo)
				if a.shouldIgnoreTelemetry(item) {
					return
				}
				telemetry, err := a.createTelemetry(item)
				if err != nil {
					log.Errorf("Failed to create telemetry: %v", err)
					return
				}
				if !a.shouldTrace(telemetry.Request.Host, item.ConnectionInfo.ServerPort) {
					log.Infof("Ignoring host: %v", telemetry.Request.Host)
					return
				}

				params := operations.NewPostTelemetryParams().WithBody(telemetry)

				_, err = a.apiClient.Operations.PostTelemetry(params)
				if err != nil {
					log.Errorf("Failed to post telemetry: %v", err)
					return
				}
				log.Info("Telemetry has been sent")
			}()
		}
	}
}

func (a *Agent) shouldTrace(host, port string) bool {
	if a.traceSamplingClient == nil {
		return true
	}
	if a.traceSamplingClient.ShouldTrace(host, port) {
		return true
	}

	return false
}

// for each connection we will get the pod to pod communication (maybe twice if they are on different nodes) and the pod to service communication.
// the filtering logic is that we only send to API Clarity telemetries where the client is in a monitored namespace
// and the destination is not a pod IP (service IP or external IP). Same behaviour as the wasm filter.
func (a *Agent) shouldIgnoreTelemetry(item *api.OutputChannelItem) bool {
	clientNamespace := a.podMonitor.GetPodNamespaceByIP(item.ConnectionInfo.ClientIP)
	if !a.podMonitor.IsMonitoredNamespace(clientNamespace) {
		// client pod is not on monitored namespace, ignore
		return true
	}
	serverNamespace := a.podMonitor.GetPodNamespaceByIP(item.ConnectionInfo.ServerIP)

	// if we found the namespace of that ip, it means it is a pod, otherwise, it is a service ip
	return serverNamespace != ""
}

func (a *Agent) createTelemetry(item *api.OutputChannelItem) (*models.Telemetry, error) {
	requestHTTPPayload, ok := item.Pair.Request.Payload.(api.HTTPPayload)
	if !ok {
		return nil, fmt.Errorf("failed to convert request Payload into HTTPPayload")
	}

	request, ok := requestHTTPPayload.Data.(*http.Request)
	if !ok {
		return nil, fmt.Errorf("failed in type assertion to http.Request")
	}

	responseHTTPPayload, ok := item.Pair.Response.Payload.(api.HTTPPayload)
	if !ok {
		return nil, fmt.Errorf("failed to convert response Payload into HTTPPayload")
	}

	response, ok := responseHTTPPayload.Data.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("failed in type assertion to http.Response")
	}

	reqBody, truncatedBodyReq, err := common.ReadBody(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	resBody, truncatedBodyRes, err := common.ReadBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	pathAndQuery := common.GetPathWithQuery(request.URL)

	clientNamespace := a.podMonitor.GetPodNamespaceByIP(item.ConnectionInfo.ClientIP)
	host, _ := common.GetHostAndPortFromURL(request.Host, clientNamespace)
	destinationNamespace := common.GetDestinationNamespaceFromHostOrDefault(host, clientNamespace)

	return &models.Telemetry{
		DestinationAddress:   item.ConnectionInfo.ServerIP + ":" + item.ConnectionInfo.ServerPort,
		DestinationNamespace: destinationNamespace,
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: truncatedBodyReq,
				Body:          reqBody,
				Headers:       common.CreateHeaders(request.Header),
				Version:       item.Protocol.Version,
			},
			Host:   host,
			Method: request.Method,
			Path:   pathAndQuery,
		},
		RequestID: common.GetRequestIDFromHeadersOrGenerate(request.Header),
		Response: &models.Response{
			Common: &models.Common{
				TruncatedBody: truncatedBodyRes,
				Body:          resBody,
				Headers:       common.CreateHeaders(response.Header),
				Version:       item.Protocol.Version,
			},
			StatusCode: strconv.Itoa(response.StatusCode),
		},
		Scheme:        item.Protocol.Name,
		SourceAddress: item.ConnectionInfo.ClientIP + ":" + item.ConnectionInfo.ClientPort,
	}, nil
}

func loadExtensions() ([]*api.Extension, error) {
	var extensions []*api.Extension
	files, err := ioutil.ReadDir("/app/extensions")
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %v", err)
	}
	extensions = make([]*api.Extension, len(files))
	for i, file := range files {
		filename := file.Name()
		extension := &api.Extension{
			Path: path.Join("/app/extensions", filename),
		}
		plug, err := plugin.Open(path.Join("/app/extensions", filename))
		if err != nil {
			return nil, fmt.Errorf("failed to open plugin: %v", err)
		}
		extension.Plug = plug
		symDissector, err := plug.Lookup("Dissector")

		var dissector api.Dissector
		var ok bool
		dissector, ok = symDissector.(api.Dissector)
		if err != nil || !ok {
			return nil, fmt.Errorf("failed to load the extension: %s", extension.Path)
		}
		dissector.Register(extension)
		extension.Dissector = dissector
		extensions[i] = extension
	}

	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Protocol.Priority < extensions[j].Protocol.Priority
	})

	for _, extension := range extensions {
		log.Infof("Extension Properties: %+v\n", extension)
	}
	return extensions, nil
}
