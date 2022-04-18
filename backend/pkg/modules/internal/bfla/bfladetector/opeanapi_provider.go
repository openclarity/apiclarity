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
