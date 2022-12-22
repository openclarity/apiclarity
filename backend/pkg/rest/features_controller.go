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
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
)

func (s *Server) GetFeatures(params operations.GetFeaturesParams) middleware.Responder {
	log.Debugf("GetFeatures controller was invoked")

	featureList := &models.APIClarityFeatureList{
		Features: []*models.APIClarityFeature{},
	}

	for _, info := range s.features {
		hostsToTrace := models.HostsToTraceForComponent{
			Component:         info.Name,
			TraceSourcesHosts: []*models.HostsToTraceForTraceSource{},
		}
		if s.samplingManager != nil {
			hostsMap, err := s.samplingManager.GetHostsToTraceByComponent(info.Name)
			if err != nil {
				log.Errorf("failed to retrieve HostsToTraceByComponent for component=%v: %v", info.Name, err)
				continue
			}
			for traceSourceID, hosts := range hostsMap {
				hostsToTrace.TraceSourcesHosts = append(hostsToTrace.TraceSourcesHosts,
					&models.HostsToTraceForTraceSource{
						TraceSourceID: uint32(traceSourceID),
						HostsToTrace:  hosts,
					})
			}
		} else {
			hostsToTrace.TraceSourcesHosts = append(hostsToTrace.TraceSourcesHosts,
				&models.HostsToTraceForTraceSource{
					TraceSourceID: 0,
					HostsToTrace:  []string{"*"},
				})
		}
		featureList.Features = append(featureList.Features,
			&models.APIClarityFeature{
				FeatureDescription: info.Description,
				FeatureName:        models.NewAPIClarityFeatureEnum(models.APIClarityFeatureEnum(info.Name)),
				HostsToTrace:       &hostsToTrace,
			},
		)
	}

	return operations.NewGetFeaturesOK().WithPayload(featureList)
}
