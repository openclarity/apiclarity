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

package rest

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	oapispec "github.com/getkin/kin-openapi/openapi3"

	"github.com/openclarity/apiclarity/api/server/models"
	speculatorspec "github.com/openclarity/speculator/pkg/spec"
)

func Test_createModelsReviewPathItem(t *testing.T) {
	type args struct {
		speculatorReviewPathItem *speculatorspec.ReviewPathItem
		pathToPathItem           map[string]*oapispec.PathItem
	}
	tests := []struct {
		name string
		args args
		want *models.ReviewPathItem
	}{
		{
			name: "3 paths with 3 different methods",
			args: args{
				pathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": {
						Get: &oapispec.Operation{
							OperationID: "get",
						},
					},
					"/api/2/foo": {
						Put: &oapispec.Operation{
							OperationID: "put",
						},
					},
					"/api/3/foo": {
						Post: &oapispec.Operation{
							OperationID: "post",
						},
					},
				},
				speculatorReviewPathItem: &speculatorspec.ReviewPathItem{
					ParameterizedPath: "/api/{param1}/foo",
					Paths: map[string]bool{
						"/api/1/foo": true,
						"/api/2/foo": true,
						"/api/3/foo": true,
					},
				},
			},
			want: &models.ReviewPathItem{
				APIEventsPaths: []*models.APIEventPathAndMethods{
					{
						Methods: []models.HTTPMethod{models.HTTPMethodGET},
						Path:    "/api/1/foo",
					},
					{
						Methods: []models.HTTPMethod{models.HTTPMethodPUT},
						Path:    "/api/2/foo",
					},
					{
						Methods: []models.HTTPMethod{models.HTTPMethodPOST},
						Path:    "/api/3/foo",
					},
				},
				SuggestedPath: "/api/{param1}/foo",
			},
		},
		{
			name: "2 paths, one has 2 methods",
			args: args{
				pathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": {
						Get: &oapispec.Operation{
							OperationID: "get",
						},
						Put: &oapispec.Operation{
							OperationID: "put",
						},
					},
					"/api/2/foo": {
						Post: &oapispec.Operation{
							OperationID: "post",
						},
					},
				},
				speculatorReviewPathItem: &speculatorspec.ReviewPathItem{
					ParameterizedPath: "/api/{param1}/foo",
					Paths: map[string]bool{
						"/api/1/foo": true,
						"/api/2/foo": true,
					},
				},
			},
			want: &models.ReviewPathItem{
				APIEventsPaths: []*models.APIEventPathAndMethods{
					{
						Methods: []models.HTTPMethod{models.HTTPMethodGET, models.HTTPMethodPUT},
						Path:    "/api/1/foo",
					},
					{
						Methods: []models.HTTPMethod{models.HTTPMethodPOST},
						Path:    "/api/2/foo",
					},
				},
				SuggestedPath: "/api/{param1}/foo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createModelsReviewPathItem(tt.args.speculatorReviewPathItem, tt.args.pathToPathItem)
			sort.Slice(tt.want.APIEventsPaths, func(i, j int) bool {
				return tt.want.APIEventsPaths[i].Path < tt.want.APIEventsPaths[j].Path
			})
			sort.Slice(got.APIEventsPaths, func(i, j int) bool {
				return got.APIEventsPaths[i].Path < got.APIEventsPaths[j].Path
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createReviewPathItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createApprovedReviewForSpeculator(t *testing.T) {
	type args struct {
		review         *models.ApprovedReview
		pathToPathItem map[string]*oapispec.PathItem
	}
	tests := []struct {
		name string
		args args
		want *speculatorspec.ApprovedSpecReview
	}{
		{
			name: "1 review path item",
			args: args{
				pathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				review: &models.ApprovedReview{
					ReviewPathItems: []*models.ReviewPathItem{
						{
							APIEventsPaths: []*models.APIEventPathAndMethods{
								{
									Methods: []models.HTTPMethod{http.MethodPost},
									Path:    "/api/1/foo",
								},
								{
									Methods: []models.HTTPMethod{http.MethodGet},
									Path:    "/api/2/foo",
								},
							},
							SuggestedPath: "/api/{param1}/foo",
						},
					},
				},
			},
			want: &speculatorspec.ApprovedSpecReview{
				PathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				PathItemsReview: []*speculatorspec.ApprovedSpecReviewPathItem{
					{
						ReviewPathItem: speculatorspec.ReviewPathItem{
							ParameterizedPath: "/api/{param1}/foo",
							Paths:             map[string]bool{"/api/1/foo": true, "/api/2/foo": true},
						},
					},
				},
			},
		},
		{
			name: "2 review path item",
			args: args{
				pathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				review: &models.ApprovedReview{
					ReviewPathItems: []*models.ReviewPathItem{
						{
							APIEventsPaths: []*models.APIEventPathAndMethods{
								{
									Methods: []models.HTTPMethod{http.MethodPost},
									Path:    "/api/1/foo",
								},
							},
							SuggestedPath: "/api/1/foo",
						},
						{
							APIEventsPaths: []*models.APIEventPathAndMethods{
								{
									Methods: []models.HTTPMethod{http.MethodGet},
									Path:    "/api/2/foo",
								},
							},
							SuggestedPath: "/api/{param1}/foo",
						},
					},
				},
			},
			want: &speculatorspec.ApprovedSpecReview{
				PathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				PathItemsReview: []*speculatorspec.ApprovedSpecReviewPathItem{
					{
						ReviewPathItem: speculatorspec.ReviewPathItem{
							ParameterizedPath: "/api/{param1}/foo",
							Paths:             map[string]bool{"/api/2/foo": true},
						},
					},
					{
						ReviewPathItem: speculatorspec.ReviewPathItem{
							ParameterizedPath: "/api/1/foo",
							Paths:             map[string]bool{"/api/1/foo": true},
						},
					},
				},
			},
		},
		{
			name: "few methods",
			args: args{
				pathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				review: &models.ApprovedReview{
					ReviewPathItems: []*models.ReviewPathItem{
						{
							APIEventsPaths: []*models.APIEventPathAndMethods{
								{
									Methods: []models.HTTPMethod{http.MethodPost, http.MethodGet},
									Path:    "/api/1/foo",
								},
							},
							SuggestedPath: "/api/1/foo",
						},
						{
							APIEventsPaths: []*models.APIEventPathAndMethods{
								{
									Methods: []models.HTTPMethod{http.MethodGet},
									Path:    "/api/2/foo",
								},
							},
							SuggestedPath: "/api/{param1}/foo",
						},
					},
				},
			},
			want: &speculatorspec.ApprovedSpecReview{
				PathToPathItem: map[string]*oapispec.PathItem{
					"/api/1/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
					"/api/2/foo": &speculatorspec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
				},
				PathItemsReview: []*speculatorspec.ApprovedSpecReviewPathItem{
					{
						ReviewPathItem: speculatorspec.ReviewPathItem{
							ParameterizedPath: "/api/{param1}/foo",
							Paths:             map[string]bool{"/api/2/foo": true},
						},
					},
					{
						ReviewPathItem: speculatorspec.ReviewPathItem{
							ParameterizedPath: "/api/1/foo",
							Paths:             map[string]bool{"/api/1/foo": true},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createApprovedReviewForSpeculator(tt.args.review, tt.args.pathToPathItem)
			sort.Slice(tt.want.PathItemsReview, func(i, j int) bool {
				return tt.want.PathItemsReview[i].ParameterizedPath < tt.want.PathItemsReview[j].ParameterizedPath
			})
			sort.Slice(got.PathItemsReview, func(i, j int) bool {
				return got.PathItemsReview[i].ParameterizedPath < got.PathItemsReview[j].ParameterizedPath
			})

			// path uuid is autogenerated inside the function - need to copy it to the expected
			for i := range got.PathItemsReview {
				got.PathItemsReview[i].PathUUID = tt.want.PathItemsReview[i].PathUUID
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createApprovedReviewForSpeculator() = %v, want %v", got, tt.want)
			}
		})
	}
}
