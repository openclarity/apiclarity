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

package test

import (
	"encoding/json"
	"gotest.tools/v3/assert"
	"net/http"
	"testing"

	oapi_spec "github.com/go-openapi/spec"
)

type TestSpec struct {
	Spec *oapi_spec.Swagger
}

func NewTestSpec() *TestSpec {
	return &TestSpec{
		Spec: &oapi_spec.Swagger{
			SwaggerProps: oapi_spec.SwaggerProps{
				Paths: &oapi_spec.Paths{
					Paths: map[string]oapi_spec.PathItem{},
				},
			},
		},
	}
}

func (t *TestSpec) WithPathItem(path string, pathItem oapi_spec.PathItem) *TestSpec {
	t.Spec.Paths.Paths[path] = pathItem
	return t
}

func (ts *TestSpec) String(t *testing.T) string {
	B, err := json.Marshal(ts.Spec)
	assert.NilError(t, err)
	return string(B)
}

type TestPathItem struct {
	PathItem oapi_spec.PathItem
}

func NewTestPathItem() *TestPathItem {
	return &TestPathItem{
		PathItem: oapi_spec.PathItem{
			PathItemProps: oapi_spec.PathItemProps{},
		},
	}
}

func (t *TestPathItem) WithOperation(method string, op *oapi_spec.Operation) *TestPathItem {
	switch method {
	case http.MethodGet:
		t.PathItem.Get = op
	case http.MethodDelete:
		t.PathItem.Delete = op
	case http.MethodOptions:
		t.PathItem.Options = op
	case http.MethodPatch:
		t.PathItem.Patch = op
	case http.MethodHead:
		t.PathItem.Head = op
	case http.MethodPost:
		t.PathItem.Post = op
	case http.MethodPut:
		t.PathItem.Put = op
	}
	return t
}

type TestOperation struct {
	Op *oapi_spec.Operation
}

func NewTestOperation() *TestOperation {
	return &TestOperation{
		Op: &oapi_spec.Operation{
			OperationProps: oapi_spec.OperationProps{},
		},
	}
}

func (o *TestOperation) WithTags(tags []string) *TestOperation {
	o.Op.Tags = tags
	return o
}
