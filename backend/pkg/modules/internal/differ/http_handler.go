package differ

import (
	"net/http"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/common"
)

type httpHandler struct{}

func (h *httpHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	common.HttpResponse(w, http.StatusOK, &oapicommon.ModuleVersion{
		Version: "1",
	})
}

func (h *httpHandler) Start(w http.ResponseWriter, r *http.Request) {
	common.HttpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Start Success"})
}

func (h *httpHandler) Stop(w http.ResponseWriter, r *http.Request) {
	common.HttpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Stop Success"})
}
