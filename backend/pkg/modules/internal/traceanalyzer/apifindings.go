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

package traceanalyzer

import (
	"context"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)

type findingKey struct {
	path   string
	method string
	name   string
}

type apiFindings struct {
	paths map[findingKey]utils.TraceAnalyzerAPIAnnotation
}

type APIsFindingsRepo struct {
	accessor core.BackendAccessor

	apis map[uint64]apiFindings
}

func NewAPIsFindingsRepo(accessor core.BackendAccessor) *APIsFindingsRepo {
	return &APIsFindingsRepo{
		accessor: accessor,
		apis:     make(map[uint64]apiFindings),
	}
}

func (r *APIsFindingsRepo) Aggregate(apiID uint64, path, method string, anns ...utils.TraceAnalyzerAnnotation) (updatedFindings []utils.TraceAnalyzerAPIAnnotation) {
	// If we don't know the specPath (certainly because there is no provided not
	// reconstructed spec), then make the method an empty string so we group all
	// the unknown path regarless of the method.
	if path == "" {
		method = ""
	}
	for _, ann := range anns {
		ann, updated := r.aggregate(apiID, path, method, ann)
		if updated {
			updatedFindings = append(updatedFindings, ann)
		}
	}
	return
}

func (r *APIsFindingsRepo) GetAPIFindings(apiID uint64) (apiFindings []utils.TraceAnalyzerAPIAnnotation) {
	findings, found := r.apis[apiID]
	if !found {
		return
	}

	for _, f := range findings.paths {
		apiFindings = append(apiFindings, f)
	}

	return
}

func (r *APIsFindingsRepo) ResetAPIFindings(apiID uint64) {
	delete(r.apis, apiID)
}

func (r *APIsFindingsRepo) aggregate(apiID uint64, path, method string, ann utils.TraceAnalyzerAnnotation) (apiFinding utils.TraceAnalyzerAPIAnnotation, updated bool) {
	// Check if we already have an entry for this apiID
	findings, found := r.apis[apiID]
	if !found {
		findings = getPreviousAPIFindings(context.TODO(), r.accessor, apiID)
		r.apis[apiID] = findings
	}

	// Check if we already have an entry for this (path, method, annotation name) pair
	key := findingKey{path, method, ann.Name()}
	apiAnn, found := findings.paths[key]
	if !found {
		apiAnn = ann.NewAPIAnnotation(path, method)
		if apiAnn != nil {
			findings.paths[key] = apiAnn
			updated = true
		}
	}

	u := apiAnn.Aggregate(ann)
	updated = updated || u

	return apiAnn, updated
}

func getPreviousAPIFindings(ctx context.Context, accessor core.BackendAccessor, apiID uint64) apiFindings {
	dbAnns, err := accessor.ListAPIInfoAnnotations(ctx, utils.ModuleName, uint(apiID))
	if err != nil {
		return apiFindings{
			paths: map[findingKey]utils.TraceAnalyzerAPIAnnotation{},
		}
	}

	paths := map[findingKey]utils.TraceAnalyzerAPIAnnotation{}
	taAnns := fromCoreAPIAnnotations(dbAnns)
	for _, taAnn := range taAnns {
		paths[findingKey{path: taAnn.Path(), method: taAnn.Method(), name: taAnn.Name()}] = taAnn
	}

	return apiFindings{
		paths: paths,
	}
}
