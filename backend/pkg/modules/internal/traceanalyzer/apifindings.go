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

func (r *APIsFindingsRepo) Aggregate(apiID uint64, path, method string, anns ...utils.TraceAnalyzerAnnotation) (updated bool) {
	for _, ann := range anns {
		u := r.aggregate(apiID, path, method, ann)
		updated = updated || u
	}

	return updated
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

func (r *APIsFindingsRepo) aggregate(apiID uint64, path, method string, ann utils.TraceAnalyzerAnnotation) (updated bool) {
	// Check if we already have an entry for this apiID
	findings, found := r.apis[apiID]
	if !found {
		findings = apiFindings{
			paths: make(map[findingKey]utils.TraceAnalyzerAPIAnnotation),
		}
		r.apis[apiID] = findings
	}

	// Check if we already have an entry for this (path, annotation name) pair
	key := findingKey{path, method, ann.Name()}
	apiAnn, found := findings.paths[key]
	if !found {
		apiAnn = ann.NewAPIAnnotation(path, method)
		if apiAnn != nil {
			updated = true
		}
	}
	if apiAnn != nil {
		findings.paths[key] = apiAnn
		u := apiAnn.Aggregate(ann)
		updated = updated || u
	}

	return updated
}
