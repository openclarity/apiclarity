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

package database

import (
	"encoding/json"
	"fmt"

	"github.com/apiclarity/apiclarity/api/server/models"
)

type specType string

const (
	ReconstructedSpecType specType = "ReconstructedSpecType"
	ProvidedSpecType      specType = "ProvidedSpecType"
)

func (a *APIInventoryTableHandler) GetAPISpecs(apiID uint32) (*APIInfo, error) {
	apiInfo := APIInfo{}

	if err := a.tx.Select(reconstructedSpecColumnName, providedSpecColumnName).First(&apiInfo, apiID).Error; err != nil {
		return nil, err
	}

	return &apiInfo, nil
}

func (a *APIInventoryTableHandler) GetAPISpecsInfo(apiID uint32) (*models.OpenAPISpecs, error) {
	apiInfo := APIInfo{}

	if err := a.tx.Select(reconstructedSpecInfoColumnName, providedSpecInfoColumnName).First(&apiInfo, apiID).Error; err != nil {
		return nil, fmt.Errorf("failed to get API info: %v", err)
	}

	specsInfo := &models.OpenAPISpecs{}

	if apiInfo.ProvidedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provided spec info. info=%+v: %v", apiInfo.ProvidedSpecInfo, err)
		}
		specsInfo.ProvidedSpec = &specInfo
	}

	if apiInfo.ReconstructedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reconstructed spec info. info=%+v: %v", apiInfo.ReconstructedSpecInfo, err)
		}
		specsInfo.ReconstructedSpec = &specInfo
	}

	return specsInfo, nil
}

func (a *APIInventoryTableHandler) PutAPISpec(apiID uint, spec string, specInfo *models.SpecInfo, specType specType) error {
	specInfoB, err := json.Marshal(specInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal spec info. info=%+v: %v", specInfo, err)
	}

	var valuesToUpdate map[string]interface{}

	switch specType {
	case ReconstructedSpecType:
		valuesToUpdate = map[string]interface{}{
			reconstructedSpecColumnName:     spec,
			reconstructedSpecInfoColumnName: string(specInfoB),
			hasReconstructedSpecColumnName:  true,
		}
	case ProvidedSpecType:
		valuesToUpdate = map[string]interface{}{
			providedSpecColumnName:     spec,
			providedSpecInfoColumnName: string(specInfoB),
			hasProvidedSpecColumnName:  true,
		}
	}

	if err := a.tx.Model(&APIInfo{}).Where("id = ?", apiID).Updates(valuesToUpdate).Error; err != nil {
		return fmt.Errorf("failed update API info: %v", err)
	}

	return nil
}

func (a *APIInventoryTableHandler) DeleteProvidedAPISpec(apiID uint32) error {
	if err := a.tx.Model(&APIInfo{}).Where("id = ?", apiID).Updates(map[string]interface{}{providedSpecColumnName: "", providedSpecInfoColumnName: "", hasProvidedSpecColumnName: false}).Error; err != nil {
		return err
	}

	return nil
}

func (a *APIInventoryTableHandler) DeleteApprovedAPISpec(apiID uint32) error {
	if err := a.tx.Model(&APIInfo{}).Where("id = ?", apiID).Updates(map[string]interface{}{reconstructedSpecColumnName: "", reconstructedSpecInfoColumnName: "", hasReconstructedSpecColumnName: false}).Error; err != nil {
		return err
	}

	return nil
}
