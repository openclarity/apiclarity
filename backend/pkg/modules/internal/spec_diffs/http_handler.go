package spec_diffs

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/common"
)

type httpHandler struct{
	differ *differ
}

func (h *httpHandler) StartDiffer(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.differ.accessor.EnableTraces(r.Context(), moduleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HttpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Differ successfully started for apiID=%v", apiID)
	common.HttpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Differ successfully started for apiID=%v", apiID)})
}

func (h *httpHandler) StopDiffer(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	err := h.differ.accessor.DisableTraces(r.Context(), moduleName, uint(apiID))
	if err != nil {
		log.Error(err)
		common.HttpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}

	log.Infof("Differ successfully stopped for apiID=%v", apiID)
	common.HttpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Differ stopped for apiID=%v", apiID)})
}
