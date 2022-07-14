package spec_differ

import (
	"crypto/sha256"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/golang/mock/gomock"
	_spec "github.com/openclarity/speculator/pkg/spec"
	"gotest.tools/v3/assert"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
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
		newSpec = "newSpec"
		oldSpec = "newSpec"
		path    = "/some/path"
	)
	var (
		hash                  = sha256.Sum256([]byte(newSpec + oldSpec))
		methodGet             = common.GET
		specTypeReconstructed = common.RECONSTRUCTED
		specTypeProvided      = common.PROVIDED
	)

	type fields struct {
		apiIDToDiffs map[uint]map[diffHash]global.Diff
		totalEvents  int
	}
	type args struct {
		event    *database.APIEvent
		newSpec  string
		oldSpec  string
		diffType models.DiffType
		specType common.SpecType
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		expectAccessor   func(accessor *core.MockBackendAccessor)
		wantApiIDToDiffs map[uint]map[diffHash]global.Diff
		wantTotalEvents  int
	}{
		{
			name: "no diff event - nothing change",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hash: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						NewSpec:       newSpec,
						OldSpec:       oldSpec,
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
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
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
					1: {hash: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						NewSpec:       newSpec,
						OldSpec:       oldSpec,
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
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
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
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(11, 0)),
					Path:      path,
					Method:    models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeReconstructed,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(11, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
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
					2: {hash: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(1234, 0),
						Method:        methodGet,
						NewSpec:       newSpec,
						OldSpec:       oldSpec,
						Path:          path,
						SpecTimestamp: time.Unix(10, 0),
						SpecType:      specTypeReconstructed,
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(11, 0)),
					Path:      path,
					Method:    models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeProvided,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(11, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
					Path:          path,
					Method:        methodGet,
					SpecType:      specTypeProvided,
					SpecTimestamp: time.Unix(13, 0),
				}},
				2: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(1234, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
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
					1: {hash: global.Diff{
						DiffType:      common.GENERALDIFF,
						LastSeen:      time.Unix(11, 0),
						NewSpec:       newSpec,
						OldSpec:       oldSpec,
						Path:          path,
						Method:        methodGet,
						SpecType:      specTypeProvided,
						SpecTimestamp: time.Unix(13, 0),
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event: &database.APIEvent{
					APIInfoID: 1,
					Time:      strfmt.DateTime(time.Unix(12, 0)),
					Path:      path,
					Method:    models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeProvided,
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ReconstructedSpecCreatedAt: strfmt.DateTime(time.Unix(10, 0)),
					ProvidedSpecCreatedAt:      strfmt.DateTime(time.Unix(13, 0)),
				}, nil)
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType:      common.GENERALDIFF,
					LastSeen:      time.Unix(12, 0),
					NewSpec:       newSpec,
					OldSpec:       oldSpec,
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
			}

			p.addDiffToSend(tt.args.newSpec, tt.args.oldSpec, tt.args.diffType, tt.args.specType, tt.args.event)

			assert.Assert(t, reflect.DeepEqual(tt.wantApiIDToDiffs, p.apiIDToDiffs))
			assert.Assert(t, tt.wantTotalEvents == p.totalUniqueDiffs)
		})
	}
}
