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
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *Server) PostControlNewDiscoveredAPIs(params operations.PostControlNewDiscoveredAPIsParams) middleware.Responder {
	log.Debugf("PostControlNewDiscoveredAPIs controller was invoked")

	// Iterate over each hosts and check if it alreay exists
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
			Type:                 models.APITypeEXTERNAL,
			Name:                 host,
			Port:                 int64(port),
			DestinationNamespace: "",
		}
		created, err := s.dbHandler.APIInventoryTable().FirstOrCreate(apiInfo)
		if err != nil {
			log.Error(err)
			continue
		}
		if created {
			log.Infof("New API '%s' was added to inventory", h)
			_ = s.speculator.InitSpec(host, strconv.Itoa(port))
		}
	}

	return operations.NewPostControlNewDiscoveredAPIsOK()
}
