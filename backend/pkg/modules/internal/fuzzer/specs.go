package fuzzer

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
