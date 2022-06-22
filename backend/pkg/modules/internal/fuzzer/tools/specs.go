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

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

func GetAPISpecsInfo(ctx context.Context, accessor core.BackendAccessor, apiID uint) (*models.OpenAPISpecs, error) {
	apiInfo, err := accessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("can't access to apiID (%v), error=%v", apiID, err)
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
