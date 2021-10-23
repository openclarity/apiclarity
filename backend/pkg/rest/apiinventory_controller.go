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

package rest

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
)

func (s *Server) GetAPIInventory(params operations.GetAPIInventoryParams) middleware.Responder {
	var apiInventory []*models.APIInfo

	apiInventoryFromDB, total, err := s.dbHandler.APIInventoryTable().GetAPIInventoryAndTotal(params)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIInventory controller was invoked. params=%+v, apiInventoryFromDB=%+v, total=%+v", params, apiInventoryFromDB, total)

	for i := range apiInventoryFromDB {
		apiInventory = append(apiInventory, _database.APIInfoFromDB(&apiInventoryFromDB[i]))
	}

	return operations.NewGetAPIInventoryOK().WithPayload(
		&operations.GetAPIInventoryOKBody{
			Items: apiInventory,
			Total: &total,
		})
}
