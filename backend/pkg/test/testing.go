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
	"net/http"
	"testing"

	oapi_spec "github.com/getkin/kin-openapi/openapi3"
	"gotest.tools/v3/assert"
)

type Spec struct {
	Spec *oapi_spec.T
}

func NewTestSpec() *Spec {
	return &Spec{
		Spec: &oapi_spec.T{
			OpenAPI: "3.0.3",
			Info:    createDefaultSwaggerInfo(),
			Paths:   map[string]*oapi_spec.PathItem{},
		},
	}
}

func createDefaultSwaggerInfo() *oapi_spec.Info {
	return &oapi_spec.Info{
		Description:    "This is a generated Open API Spec",
		Title:          "Swagger",
		TermsOfService: "https://swagger.io/terms/",
		Contact: &oapi_spec.Contact{
			Email: "apiteam@swagger.io",
		},
		License: &oapi_spec.License{
			Name: "Apache 2.0",
			URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
		},
		Version: "1.0.0",
	}
}

func (ts *Spec) WithPathItem(path string, pathItem *oapi_spec.PathItem) *Spec {
	ts.Spec.Paths[path] = pathItem
	return ts
}

func (ts *Spec) String(t *testing.T) string {
	t.Helper()
	B, err := json.Marshal(ts.Spec)
	assert.NilError(t, err)
	return string(B)
}

type PathItem struct {
	PathItem *oapi_spec.PathItem
}

func NewTestPathItem() *PathItem {
	return &PathItem{
		PathItem: &oapi_spec.PathItem{},
	}
}

func (t *PathItem) WithOperation(method string, op *oapi_spec.Operation) *PathItem {
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

type Operation struct {
	Op *oapi_spec.Operation
}

func NewTestOperation() *Operation {
	return &Operation{
		Op: &oapi_spec.Operation{
			Responses: oapi_spec.NewResponses(),
		},
	}
}

func (o *Operation) WithTags(tags []string) *Operation {
	o.Op.Tags = tags
	return o
}
