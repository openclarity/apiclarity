package rest

import (
	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/speculator/pkg/speculator"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func (s *Server) DeleteAPIInventoryAPIIDSpecsProvidedSpec(params operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
	apiInfo := &database.APIInfo{}

	if err := database.GetAPIInventoryTable().First(&apiInfo, params.APIID).Error; err != nil {
		log.Errorf("Failed to get APIInventory table with api id: %v. %v", params.APIID, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.speculator.UnsetProvidedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port)))); err != nil {
		log.Errorf("Failed to unset provided spec. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}
	if err := database.DeleteProvidedAPISpec(params.APIID); err != nil {
		log.Errorf("Failed to delete provided spec with api id: %v from DB. %v", params.APIID, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewDeleteAPIInventoryAPIIDSpecsProvidedSpecOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}

func (s *Server) DeleteAPIInventoryAPIIDSpecsReconstructedSpec(params operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) middleware.Responder {
	apiInfo := &database.APIInfo{}

	if err := database.GetAPIInventoryTable().First(&apiInfo, params.APIID).Error; err != nil {
		log.Errorf("Failed to get APIInventory table with api id: %v. %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.speculator.UnsetApprovedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port)))); err != nil {
		log.Errorf("Failed to unset provided spec. %v", err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	if err := database.DeleteApprovedAPISpec(params.APIID); err != nil {
		log.Errorf("Failed to delete provided spec with api id: %v from DB. %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}
