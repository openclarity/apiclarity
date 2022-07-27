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

//nolint: revive,stylecheck
package spec_differ

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/common"
)

type httpHandler struct {
	differ *specDiffer
}

func (h *httpHandler) StartDiffer(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.differ.accessor.EnableTraces(r.Context(), moduleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Differ successfully started for apiID=%v", apiID)
	common.HTTPResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Differ successfully started for apiID=%v", apiID)})
}

func (h *httpHandler) StopDiffer(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.differ.accessor.DisableTraces(r.Context(), moduleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HTTPResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Differ successfully stopped for apiID=%v", apiID)
	common.HTTPResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Differ stopped for apiID=%v", apiID)})
}
