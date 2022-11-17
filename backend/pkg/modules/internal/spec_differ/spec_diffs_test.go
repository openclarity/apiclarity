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

//nolint: revive,stylecheck
package spec_differ

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi2"
	v3spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/strfmt"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/spec_differ/config"
	_spec "github.com/openclarity/speculator/pkg/spec"
)

func Test_getHighestPrioritySpecDiffType(t *testing.T) {
	type args struct {
		providedDiff      models.DiffType
		reconstructedDiff models.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "Zombie over Shadow",
			args: args{
				providedDiff:      models.DiffTypeZOMBIEDIFF,
				reconstructedDiff: models.DiffTypeSHADOWDIFF,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "Same type",
			args: args{
				providedDiff:      models.DiffTypeGENERALDIFF,
				reconstructedDiff: models.DiffTypeGENERALDIFF,
			},
			want: models.DiffTypeGENERALDIFF,
		},
		{
			name: "reconstructed unknown type",
			args: args{
				providedDiff:      models.DiffTypeNODIFF,
				reconstructedDiff: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "provided unknown type",
			args: args{
				providedDiff:      "unknown type",
				reconstructedDiff: models.DiffTypeNODIFF,
			},
			want: models.DiffTypeNODIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHighestPrioritySpecDiffType(tt.args.providedDiff, tt.args.reconstructedDiff); got != tt.want {
				t.Errorf("getHighestPrioritySpecDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToModelsDiffType(t *testing.T) {
	type args struct {
		diffType _spec.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "unknown type - default DiffTypeNODIFF",
			args: args{
				diffType: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeNoDiff",
			args: args{
				diffType: _spec.DiffTypeNoDiff,
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeZombieDiff",
			args: args{
				diffType: _spec.DiffTypeZombieDiff,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "DiffTypeShadowDiff",
			args: args{
				diffType: _spec.DiffTypeShadowDiff,
			},
			want: models.DiffTypeSHADOWDIFF,
		},
		{
			name: "DiffTypeGeneralDiff",
			args: args{
				diffType: _spec.DiffTypeGeneralDiff,
			},
			want: models.DiffTypeGENERALDIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToModelsDiffType(tt.args.diffType); got != tt.want {
				t.Errorf("convertToModelsDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_differ_addDiffToSend(t *testing.T) {
	mockCtrlAccessor := gomock.NewController(t)
	defer mockCtrlAccessor.Finish()
	mockAccessor := core.NewMockBackendAccessor(mockCtrlAccessor)

	const (
		path      = "/some/path"
		newSpecV2 = "get:\n  responses:\n    \"200\":\n      schema:\n        properties:\n          test:\n            type: string\n        type: object\n"
		oldSpecV2 = "get:\n  responses:\n    \"200\":\n      schema:\n        properties:\n          test:\n            format: int64\n            type: integer\n        type: object\n"
		newSpecV3 = "get:\n  responses:\n    \"200\":\n      content:\n        application/json:\n          schema:\n            properties:\n              test:\n                type: string\n            type: object\n"
		oldSpecV3 = "get:\n  responses:\n    \"200\":\n      content:\n        application/json:\n          schema:\n            properties:\n              test:\n                format: int64\n                type: integer\n            type: object\n"
	)

	var (
		hashV2Reconstructed = sha256.Sum256([]byte(newSpecV2 + oldSpecV2 + common.RECONSTRUCTED))
		hashV2Provided      = sha256.Sum256([]byte(newSpecV2 + oldSpecV2 + common.PROVIDED))
		hashV3Reconstructed = sha256.Sum256([]byte(newSpecV3 + oldSpecV3 + common.RECONSTRUCTED))

		methodGet             = common.GET
		specTypeReconstructed = common.RECONSTRUCTED
		specTypeProvided      = common.PROVIDED
		originalPathItem      = v3spec.PathItem{
			Get: &v3spec.Operation{
				Responses: v3spec.Responses{
					"200": &v3spec.ResponseRef{
						Value: v3spec.NewResponse().
							WithJSONSchemaRef(&v3spec.SchemaRef{
								Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
							},
							),
					},
				},
			},
		}
		modifiedPathItem = v3spec.PathItem{
			Get: &v3spec.Operation{
				Responses: v3spec.Responses{
					"200": &v3spec.ResponseRef{
						Value: v3spec.NewResponse().
							WithJSONSchemaRef(&v3spec.SchemaRef{
								Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewStringSchema()),
							},
							),
					},
				},
			},
		}
	)

	type fields struct {
		apiIDToDiffs map[uint]map[diffHash]global.Diff
		totalEvents  int
	}
	type args struct {
		event            *database.APIEvent
		modifiedPathItem *v3spec.PathItem
		originalPathItem *v3spec.PathItem
		diffType         models.DiffType
		specType         common.SpecType
		version          _spec.OASVersion
		diff             *_spec.APIDiff
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		expectAccessor   func(accessor *core.MockBackendAccessor)
		wantAPIIDToDiffs map[uint]map[diffHash]global.Diff
		wantTotalEvents  int
	}{
		{
			name: "no diff event - nothing change",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hashV2Reconstructed: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						NewSpec:       newSpecV2,
						OldSpec:       oldSpecV2,
						Path:          path,
						Method:        methodGet,
						SpecType:      specTypeReconstructed,
						SpecTimestamp: time.Unix(10, 0),
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event:    &database.APIEvent{},
				diffType: models.DiffTypeNODIFF,
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV2Reconstructed: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeReconstructed,
					SpecTimestamp: time.Unix(10, 0),
				}},
			},
			wantTotalEvents: 1,
			expectAccessor:  func(accessor *core.MockBackendAccessor) {},
		},
		{
			name: "event threshold reached - ignoring event",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hashV2Reconstructed: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						NewSpec:       newSpecV2,
						OldSpec:       oldSpecV2,
						Path:          path,
						Method:        methodGet,
						SpecType:      specTypeReconstructed,
						SpecTimestamp: time.Unix(10, 0),
					}},
				},
				totalEvents: 501,
			},
			args: args{
				event:    &database.APIEvent{},
				diffType: models.DiffTypeZOMBIEDIFF,
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV2Reconstructed: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeReconstructed,
					SpecTimestamp: time.Unix(10, 0),
				}},
			},
			wantTotalEvents: 501,
			expectAccessor:  func(accessor *core.MockBackendAccessor) {},
		},
		{
			name: "event has spec diff - first time for this api id",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{},
				totalEvents:  0,
			},
			args: args{
				diff: &_spec.APIDiff{
					Path: path,
				},
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(11, 0)),
					Method:    models.HTTPMethodGET,
				},
				modifiedPathItem: &modifiedPathItem,
				originalPathItem: &originalPathItem,
				diffType:         models.DiffTypeGENERALDIFF,
				specType:         specTypeReconstructed,
				version:          _spec.OASv2,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV2Reconstructed: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(11, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeReconstructed,
					SpecTimestamp: time.Unix(10, 0),
				}},
			},
			wantTotalEvents: 1,
		},
		{
			name: "event has spec diff - OAS V3",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{},
				totalEvents:  0,
			},
			args: args{
				diff: &_spec.APIDiff{
					Path: path,
				},
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(11, 0)),
					Method:    models.HTTPMethodGET,
				},
				modifiedPathItem: &modifiedPathItem,
				originalPathItem: &originalPathItem,
				diffType:         models.DiffTypeGENERALDIFF,
				specType:         specTypeReconstructed,
				version:          _spec.OASv3,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV3Reconstructed: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(11, 0),
					NewSpec:       newSpecV3,
					OldSpec:       oldSpecV3,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeReconstructed,
					SpecTimestamp: time.Unix(10, 0),
				}},
			},
			wantTotalEvents: 1,
		},
		{
			name: "event has spec diff - first time for this api id - events map is not empty",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					2: {hashV2Reconstructed: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						Method:        methodGet,
						NewSpec:       newSpecV2,
						OldSpec:       oldSpecV2,
						Path:          path,
						SpecTimestamp: time.Unix(10, 0),
						SpecType:      specTypeReconstructed,
					}},
				},
				totalEvents: 1,
			},
			args: args{
				diff: &_spec.APIDiff{
					Path: path,
				},
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(11, 0)),
					Path:      path,
					Method:    models.HTTPMethodGET,
				},
				modifiedPathItem: &modifiedPathItem,
				originalPathItem: &originalPathItem,
				diffType:         models.DiffTypeGENERALDIFF,
				specType:         specTypeProvided,
				version:          _spec.OASv2,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV2Provided: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(11, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeProvided,
					SpecTimestamp: time.Unix(13, 0),
				}},
				2: {hashV2Reconstructed: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecTimestamp: time.Unix(10, 0),
					SpecType:      specTypeReconstructed,
				}},
			},
			wantTotalEvents: 2,
		},
		{
			name: "event has spec diff - already exists for this api id - update last seen - don't increase total",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hashV2Provided: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(11, 0),
						NewSpec:       newSpecV2,
						OldSpec:       oldSpecV2,
						Path:          path,
						Method:        methodGet,
						SpecType:      specTypeProvided,
						SpecTimestamp: time.Unix(13, 0),
					}},
				},
				totalEvents: 1,
			},
			args: args{
				diff: &_spec.APIDiff{
					Path: path,
				},
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(12, 0)),
					Path:      path,
					Method:    models.HTTPMethodGET,
				},
				modifiedPathItem: &modifiedPathItem,
				originalPathItem: &originalPathItem,
				diffType:         models.DiffTypeGENERALDIFF,
				specType:         specTypeProvided,
				version:          _spec.OASv2,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantAPIIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hashV2Provided: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(12, 0),
					NewSpec:       newSpecV2,
					OldSpec:       oldSpecV2,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeProvided,
					SpecTimestamp: time.Unix(13, 0),
				}},
			},
			wantTotalEvents: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expectAccessor(mockAccessor)
			p := &specDiffer{
				apiIDToDiffs:     tt.fields.apiIDToDiffs,
				totalUniqueDiffs: tt.fields.totalEvents,
				accessor:         mockAccessor,
				config:           config.GetConfig(),
			}

			p.addDiffToSend(tt.args.diff, tt.args.modifiedPathItem, tt.args.originalPathItem, tt.args.diffType, tt.args.specType, tt.args.event, tt.args.version)

			assert.DeepEqual(t, tt.wantAPIIDToDiffs, p.apiIDToDiffs)
			assert.Assert(t, tt.wantTotalEvents == p.totalUniqueDiffs)
		})
	}
}

func Test_getPathItemForVersionOrOriginal(t *testing.T) {
	var nilPathItem *v3spec.PathItem
	type args struct {
		v3PathItem *v3spec.PathItem
		version    _spec.OASVersion
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "v2",
			args: args{
				v3PathItem: &v3spec.PathItem{
					Get: &v3spec.Operation{
						Responses: v3spec.Responses{
							"200": &v3spec.ResponseRef{
								Value: v3spec.NewResponse().
									WithJSONSchemaRef(&v3spec.SchemaRef{
										Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
									},
									),
							},
						},
					},
				},
				version: _spec.OASv2,
			},
			want: &openapi2.PathItem{
				Get: &openapi2.Operation{
					Responses: map[string]*openapi2.Response{
						"200": {
							Schema: &v3spec.SchemaRef{
								Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
							},
						},
					},
				},
			},
		},
		{
			name: "v3",
			args: args{
				v3PathItem: &v3spec.PathItem{
					Get: &v3spec.Operation{
						Responses: v3spec.Responses{
							"200": &v3spec.ResponseRef{
								Value: v3spec.NewResponse().
									WithJSONSchemaRef(&v3spec.SchemaRef{
										Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
									},
									),
							},
						},
					},
				},
				version: _spec.OASv3,
			},
			want: &v3spec.PathItem{
				Get: &v3spec.Operation{
					Responses: v3spec.Responses{
						"200": &v3spec.ResponseRef{
							Value: v3spec.NewResponse().
								WithJSONSchemaRef(&v3spec.SchemaRef{
									Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
								},
								),
						},
					},
				},
			},
		},
		{
			name: "unknown",
			args: args{
				v3PathItem: &v3spec.PathItem{
					Get: &v3spec.Operation{
						Responses: v3spec.Responses{
							"200": &v3spec.ResponseRef{
								Value: v3spec.NewResponse().
									WithJSONSchemaRef(&v3spec.SchemaRef{
										Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
									},
									),
							},
						},
					},
				},
				version: _spec.Unknown,
			},
			want: &v3spec.PathItem{
				Get: &v3spec.Operation{
					Responses: v3spec.Responses{
						"200": &v3spec.ResponseRef{
							Value: v3spec.NewResponse().
								WithJSONSchemaRef(&v3spec.SchemaRef{
									Value: v3spec.NewObjectSchema().WithProperty("test", v3spec.NewInt64Schema()),
								},
								),
						},
					},
				},
			},
		},
		{
			name: "nil input",
			args: args{
				v3PathItem: nilPathItem,
				version:    _spec.OASv3,
			},
			want: nilPathItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPathItemForVersionOrOriginal(tt.args.v3PathItem, tt.args.version)
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreUnexported(v3spec.Schema{}))
		})
	}
}
