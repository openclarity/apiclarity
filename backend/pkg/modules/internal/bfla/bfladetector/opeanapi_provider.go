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

package bfladetector

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
)

type SpecType uint

const (
	SpecTypeNone = iota
	SpecTypeProvided
	SpecTypeReconstructed
)

type OpenAPIProvider interface {
	GetOpenAPI(ctx context.Context, apiID uint) (io.Reader, SpecType, error)
}

func NewBFLAOpenAPIProvider(accessor core.BackendAccessor) OpenAPIProvider {
	return bflaOpenAPIProvider{accessor: accessor}
}

type bflaOpenAPIProvider struct {
	accessor core.BackendAccessor
}

func (d bflaOpenAPIProvider) GetOpenAPI(ctx context.Context, apiID uint) (spec io.Reader, specType SpecType, err error) {
	invInfo, err := d.accessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return nil, SpecTypeNone, fmt.Errorf("unable to get openapi spec: %w", err)
	}

	if invInfo.HasProvidedSpec {
		spec = bytes.NewBufferString(invInfo.ProvidedSpec)
		specType = SpecTypeProvided
	} else if invInfo.HasReconstructedSpec {
		spec = bytes.NewBufferString(invInfo.ReconstructedSpec)
		specType = SpecTypeReconstructed
	} else {
		return nil, SpecTypeNone, fmt.Errorf("unable to find OpenAPI spec for service: %d", apiID)
	}
	return
}
