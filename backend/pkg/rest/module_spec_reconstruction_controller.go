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

package rest

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	interfacemanager "github.com/openclarity/trace-sampling-manager/manager/pkg/manager/interface"
)

func (s *Server) PostModulesSpecReconstructionAPIIDStart(params operations.PostModulesSpecReconstructionAPIIDStartParams) middleware.Responder {
	log.Debugf("PostModulesSpecReconstructionAPIIDStart controller was invoked")

	apiID := params.APIID
	modName := "*"

	externalTraceSourceId, err := s.dbHandler.TraceSamplingTable().GetExternalTraceSourceID()
	if err != nil {
		log.Errorf("failed to retrieve external source ID: %v", err)
		return operations.NewPostModulesSpecReconstructionAPIIDStartDefault(http.StatusInternalServerError)
	}

	apiInfo := &database.APIInfo{}
	if err := s.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		log.Errorf("failed to retrieve API info for apiID=%v: %v", apiID, err)
		return operations.NewPostModulesSpecReconstructionAPIIDStartDefault(http.StatusInternalServerError)
	}

	traceSourceID := externalTraceSourceId
	if apiInfo.TraceSourceID != nil {
		traceSourceID = *apiInfo.TraceSourceID
	}
	s.dbHandler.TraceSamplingTable().AddHostToTrace(apiID, traceSourceID, modName)

	if traceSourceID == externalTraceSourceId {
		// Relay it to the TSM
		s.samplingManager.AddHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: modName,
			},
		)
	}

	log.Infof("Tracing successfully started for api=%d", apiID)

	return operations.NewPostModulesSpecReconstructionAPIIDStartNoContent()
}

func (s *Server) PostModulesSpecReconstructionAPIIDStop(params operations.PostModulesSpecReconstructionAPIIDStopParams) middleware.Responder {
	log.Debugf("PostModulesSpecReconstructionAPIIDStop controller was invoked")

	modName := "*"
	apiID := params.APIID

	externalTraceSourceId, err := s.dbHandler.TraceSamplingTable().GetExternalTraceSourceID()
	if err != nil {
		log.Errorf("failed to retrieve external source ID: %v", err)
		return operations.NewPostModulesSpecReconstructionAPIIDStartDefault(http.StatusInternalServerError)
	}

	apiInfo := &database.APIInfo{}
	if err := s.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		log.Errorf("failed to retrieve API info for apiID=%v: %v", apiID, err)
		return operations.NewPostModulesSpecReconstructionAPIIDStartDefault(http.StatusInternalServerError)
	}

	traceSourceID := externalTraceSourceId
	if apiInfo.TraceSourceID != nil {
		traceSourceID = *apiInfo.TraceSourceID
	}
	s.dbHandler.TraceSamplingTable().DeleteHostToTrace(apiID, traceSourceID, modName)

	if traceSourceID == externalTraceSourceId {
		// Relay it to the TSM
		s.samplingManager.RemoveHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{fmt.Sprintf("%s:%d", apiInfo.Name, apiInfo.Port)},
				ComponentID: modName,
			},
		)
	}

	log.Infof("Tracing successfully stoped for api=%d", apiID)

	return operations.NewPostModulesSpecReconstructionAPIIDStopNoContent()
}

func (s *Server) PostModulesSpecReconstructionEnable(params operations.PostModulesSpecReconstructionEnableParams) middleware.Responder {
	log.Debugf("PostModulesSpecReconstructionEnable controller was invoked")

	modName := "*"
	flag := params.Body.Enable
	if !flag {
		s.dbHandler.TraceSamplingTable().DeleteAll()
		// Relay it to the TSM
		s.samplingManager.SetHostsToTrace(
			&interfacemanager.HostsByComponentID{
				Hosts:       []string{},
				ComponentID: modName,
			},
		)
	}

	return operations.NewPostModulesSpecReconstructionEnableNoContent()
}
