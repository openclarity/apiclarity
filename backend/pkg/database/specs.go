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
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
)

func GetAPISpecs(apiID uint32) (*APIInfo, error) {
	apiInfo := APIInfo{}

	if err := GetAPIInventoryTable().Select(reconstructedSpecColumnName, providedSpecColumnName).First(&apiInfo, apiID).Error; err != nil {
		return nil, err
	}

	return &apiInfo, nil
}

func PutProvidedAPISpec(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) error {
	if err := GetAPIInventoryTable().Model(&APIInfo{}).Where("id = ?", params.APIID).Updates(map[string]interface{}{providedSpecColumnName: params.Body.RawSpec, hasProvidedSpecColumnName: true}).Error; err != nil {
		return err
	}

	return nil
}
