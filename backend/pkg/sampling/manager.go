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

package sampling

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	_globalConfig "github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/trace-sampling-manager/manager/pkg/manager"
	interfacemanager "github.com/openclarity/trace-sampling-manager/manager/pkg/manager/interface"
	restmanager "github.com/openclarity/trace-sampling-manager/manager/pkg/rest"
)

type TraceSamplingManager struct {
	traceSamplingEnabled bool
	dbHandler            *database.Handler
	samplingManager      *manager.Manager
}

func CreateTraceSamplingManager(dbHandler *database.Handler, config *_globalConfig.Config, clientset kubernetes.Interface, errChan chan struct{}) (*TraceSamplingManager, error) {
	s := &TraceSamplingManager{}
	s.dbHandler = dbHandler
	s.traceSamplingEnabled = config.TraceSamplingEnabled

	if !s.traceSamplingEnabled {
		return s, nil
	}

	/*
	* The "old" TSM... we will relay to it all hosts to trace for internal GTW
	* To removed when internal GTW will moved to new ApiClarity HTTP endpoint.
	 */
	samplingManager, err := manager.Create(clientset, &restmanager.Config{
		RestServerPort:             config.HTTPTraceSamplingManagerPort,
		GRPCServerPort:             config.GRPCTraceSamplingManagerPort,
		HostToTraceSecretName:      config.HostToTraceSecretName,
		HostToTraceSecretNamespace: config.HostToTraceSecretNamespace,
		HostToTraceSecretOwnerName: config.HostToTraceSecretOwnerName,
		EnableTLS:                  config.EnableTLS,
		TLSServerCertFilePath:      config.TLSServerCertFilePath,
		TLSServerKeyFilePath:       config.TLSServerKeyFilePath,
		RootCertFilePath:           config.RootCertFilePath,
		RestServerTLSPort:          config.HTTPSTraceSamplingManagerPort,
	})
	if err != nil {
		log.Errorf("Failed to create a trace sampling manager: %v", err)
		return nil, err
	}
	if err := samplingManager.Start(errChan); err != nil {
		log.Errorf("Failed to start trace sampling manager: %v", err)
		return nil, err
	}
	s.samplingManager = samplingManager
	/*
		End of stuff to remove when internal GTW will moved to new ApiClarity HTTP endpoint.
	*/

	return s, nil
}

func (m *TraceSamplingManager) AddHostToTrace(component string, apiID uint32) error {
	externalTraceSourceId, err := m.dbHandler.TraceSamplingTable().GetExternalTraceSourceID()
	if err != nil {
		log.Errorf("failed to retrieve external source ID: %v", err)
		return err
	}

	apiInfo := &database.APIInfo{}
	if err := m.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		log.Errorf("failed to retrieve API info for apiID=%v: %v", apiID, err)
		return err
	}

	traceSourceID := apiInfo.TraceSourceID
	m.dbHandler.TraceSamplingTable().AddApiToTrace(component, traceSourceID, apiID)

	if traceSourceID == externalTraceSourceId && m.samplingManager != nil {
		// Relay it to the TSM
		m.samplingManager.AddHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: component,
			},
		)
	}
	return nil
}

func (m *TraceSamplingManager) RemoveHostToTrace(component string, apiID uint32) error {
	externalTraceSourceId, err := m.dbHandler.TraceSamplingTable().GetExternalTraceSourceID()
	if err != nil {
		log.Errorf("failed to retrieve external source ID: %v", err)
		return err
	}

	apiInfo := &database.APIInfo{}
	if err := m.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		log.Errorf("failed to retrieve API info for apiID=%v: %v", apiID, err)
		return err
	}

	traceSourceID := apiInfo.TraceSourceID
	m.dbHandler.TraceSamplingTable().DeleteApiToTrace(component, traceSourceID, apiID)

	if traceSourceID == externalTraceSourceId && m.samplingManager != nil {
		// Relay it to the TSM
		m.samplingManager.RemoveHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: component,
			},
		)
	}
	return nil
}

func (m *TraceSamplingManager) GetHostsToTrace(component string, traceSourceID uint) ([]string, error) {
	return m.GetHostsToTraceByTraceSource(component, traceSourceID)
}

func (m *TraceSamplingManager) GetHostsToTraceByComponent(component string) (map[uint][]string, error) {
	hostsMap, err := m.dbHandler.TraceSamplingTable().HostsToTraceByComponent(component)
	if err != nil {
		log.Errorf("failed to retrieve hosts list for component=%v: %v", component, err)
		return nil, err
	}
	return hostsMap, nil
}

func (m *TraceSamplingManager) GetHostsToTraceByTraceSource(component string, traceSourceID uint) ([]string, error) {
	if !m.traceSamplingEnabled {
		return []string{"*"}, nil
	}

	hosts, err := m.dbHandler.TraceSamplingTable().HostsToTraceByTraceSource(component, traceSourceID)
	if err != nil {
		log.Errorf("failed to retrieve hosts list for component=%v: %v", component, err)
		return nil, err
	}
	return hosts, nil
}

func (m *TraceSamplingManager) ResetForComponent(component string) error {
	m.dbHandler.TraceSamplingTable().DeleteAll()
	if m.samplingManager != nil {
		// Relay it to the TSM
		m.samplingManager.SetHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{},
				ComponentID: component,
			},
		)
	}
	return nil
}
