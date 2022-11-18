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

	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/trace-sampling-manager/manager/pkg/manager"
	interfacemanager "github.com/openclarity/trace-sampling-manager/manager/pkg/manager/interface"
)

type TraceSamplingManager struct {
	dbHandler       *database.Handler
	samplingManager *manager.Manager
}

func CreateTraceSamplingManager(dbHandler *database.Handler, samplingManager *manager.Manager) (*TraceSamplingManager, error) {
	s := &TraceSamplingManager{}
	s.dbHandler = dbHandler
	s.samplingManager = samplingManager
	return s, nil
}

func (m *TraceSamplingManager) AddHostToTrace(modName string, apiID uint32) error {
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
	m.dbHandler.TraceSamplingTable().AddHostToTrace(apiID, traceSourceID, modName)

	if traceSourceID == externalTraceSourceId {
		// Relay it to the TSM
		m.samplingManager.AddHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: modName,
			},
		)
	}
	return nil
}

func (m *TraceSamplingManager) RemoveHostToTrace(modName string, apiID uint32) error {
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
	m.dbHandler.TraceSamplingTable().DeleteHostToTrace(apiID, traceSourceID, modName)

	if traceSourceID == externalTraceSourceId {
		// Relay it to the TSM
		m.samplingManager.RemoveHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: modName,
			},
		)
	}
	return nil
}

func (m *TraceSamplingManager) ResetForComponent(modName string) error {
	m.dbHandler.TraceSamplingTable().DeleteAll()
	// Relay it to the TSM
	m.samplingManager.SetHostsToTrace(
		&interfacemanager.HostsByComponentID{
			Hosts:       []string{},
			ComponentID: modName,
		},
	)
	return nil
}
