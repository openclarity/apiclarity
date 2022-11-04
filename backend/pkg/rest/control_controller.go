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
	"net"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/notifications"

	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *Server) PostControlNewDiscoveredAPIs(params operations.PostControlNewDiscoveredAPIsParams) middleware.Responder {
	log.Infof("PostControlNewDiscoveredAPIs controller was invoked")

	// Check the token and retreive the corresponding trace source
	token := []byte("") // FIXME: get the token from the http header
	traceSourceID, err := s.CheckTraceSourceAuth(token)
	if err != nil {
		log.Errorf("Unable to authenticate the Trace Source")
		return operations.NewPostControlNewDiscoveredAPIsDefault(401)
	}

	// Iterate over each hosts and check if it already exists
	for _, h := range params.Body.Hosts {
		host, strPort, err := net.SplitHostPort(h)
		if err != nil {
			log.Errorf("Unable to parse fqdn:port for '%s': %v", h, err)
			continue
		}

		port, err := strconv.Atoi(strPort)
		if err != nil {
			log.Errorf("In '%s', port '%s' is invalid", h, strPort)
			continue
		}

		apiInfo := &_database.APIInfo{
			Type:        models.APITypeINTERNAL,
			Name:        host,
			Port:        int64(port),
			TraceSourceID: traceSourceID,
		}
		created, err := s.dbHandler.APIInventoryTable().FirstOrCreate(apiInfo)
		if err != nil {
			log.Error(err)
			continue
		}
		if created {
			log.Infof("New API '%s' managed by source '%v' was added to inventory", h, traceSourceID)
			_ = s.speculator.InitSpec(host, strconv.Itoa(port))

			if s.notifier != nil {
				apiID := uint32(apiInfo.ID)
				port := int(apiInfo.Port)
				newDiscoveredAPINotification := notifications.NewDiscoveredAPINotification{
					Id:                   &apiID,
					Name:                 &apiInfo.Name,
					Port:                 &port,
					HasReconstructedSpec: &apiInfo.HasReconstructedSpec,
					HasProvidedSpec:      &apiInfo.HasProvidedSpec,
					DestinationNamespace: &apiInfo.DestinationNamespace,
					CreatedBy:            &apiInfo.CreatedBy,
				}
				notification := notifications.APIClarityNotification{}
				err := notification.FromNewDiscoveredAPINotification(newDiscoveredAPINotification)
				if err != nil {
					log.Error("failed to create 'NewDiscoveredAPI' notification, err=(%v)", err)
				} else {
					log.Infof("will send notification 'NewDiscoveredAPI' with (%v)", notification)
					log.Infof("s.notifier (%v)", s.notifier)
					err = s.notifier.Notify(apiInfo.ID, notification)
					if err != nil {
						log.Error("failed to send 'NewDiscoveredAPI' notification, err=(%v)", err)
					} else {
						log.Infof("notification 'NewDiscoveredAPI' successfully sent (%v)", notification)
					}
				}
			}
		}

	}

	return operations.NewPostControlNewDiscoveredAPIsOK()
}
