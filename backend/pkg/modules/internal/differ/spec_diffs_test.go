package differ

import (
	"crypto/sha256"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	_spec "github.com/openclarity/speculator/pkg/spec"
	"gotest.tools/v3/assert"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/database"
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
	const (
		newSpec = "newSpec"
		oldSpec = "newSpec"
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
		wantApiIDToDiffs map[uint]map[diffHash]global.Diff
		wantTotalEvents  int
	}{
		{
			name: "no diff event - nothing change",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hash: global.Diff{
						DiffType: common.GENERALDIFF,
						LastSeen: 1234,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
						Path:     stringPtr("/some/path"),
						Method:   &methodGet,
						SpecType: &specTypeReconstructed,
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event: &database.APIEvent{
					SpecDiffType: models.DiffTypeNODIFF,
				},
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 1234,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeReconstructed,
				}},
			},
			wantTotalEvents: 1,
		},
		{
			name: "event threshold reached - ignoring event",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hash: global.Diff{
						DiffType: common.GENERALDIFF,
						LastSeen: 1234,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
						Path:     stringPtr("/some/path"),
						Method:   &methodGet,
						SpecType: &specTypeReconstructed,
					}},
				},
				totalEvents: 501,
			},
			args: args{
				event: &database.APIEvent{
					SpecDiffType: models.DiffTypeZOMBIEDIFF,
				},
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 1234,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeReconstructed,
				}},
			},
			wantTotalEvents: 501,
		},
		{
			name: "event has spec diff - first time for this api id",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{},
				totalEvents:  0,
			},
			args: args{
				event: &database.APIEvent{
					HasReconstructedSpecDiff: true,
					HasProvidedSpecDiff:      false,
					APIInfoID:                1,
					Time:                     strfmt.NewDateTime(),
					Path:                     "/some/path",
					Method:                   models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeReconstructed,
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 0,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeReconstructed,
				}},
			},
			wantTotalEvents: 1,
		},
		{
			name: "event has spec diff - first time for this api id - events map is not empty",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					2: {hash: global.Diff{
						DiffType: common.GENERALDIFF,
						LastSeen: 1234,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
						Path:     stringPtr("/some/path"),
						Method:   &methodGet,
						SpecType: &specTypeReconstructed,
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event: &database.APIEvent{
					HasReconstructedSpecDiff: true,
					HasProvidedSpecDiff:      false,
					APIInfoID:                1,
					Time:                     strfmt.NewDateTime(),
					Path:                     "/some/path",
					Method:                   models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeProvided,
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 0,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeProvided,
				}},
				2: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 1234,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeReconstructed,
				}},
			},
			wantTotalEvents: 2,
		},
		{
			name: "event has spec diff - already exists for this api id - update last seen - don't increase total",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hash: global.Diff{
						DiffType: common.GENERALDIFF,
						LastSeen: 0,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
						Path:     stringPtr("/some/path"),
						Method:   &methodGet,
						SpecType: &specTypeProvided,
					}},
				},
				totalEvents: 1,
			},
			args: args{
				event: &database.APIEvent{
					HasReconstructedSpecDiff: false,
					HasProvidedSpecDiff:      true,
					APIInfoID:                1,
					Time:                     strfmt.DateTime(time.Unix(1, 0).UTC()),
					Path:                     "/some/path",
					Method:                   models.HTTPMethodGET,
				},
				newSpec:  newSpec,
				oldSpec:  oldSpec,
				diffType: models.DiffTypeGENERALDIFF,
				specType: specTypeProvided,
			},
			wantApiIDToDiffs: map[uint]map[diffHash]global.Diff{
				1: {hash: global.Diff{
					DiffType: common.GENERALDIFF,
					LastSeen: 1,
					NewSpec:  newSpec,
					OldSpec:  oldSpec,
					Path:     stringPtr("/some/path"),
					Method:   &methodGet,
					SpecType: &specTypeProvided,
				}},
			},
			wantTotalEvents: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &differ{
				apiIDToDiffs:    tt.fields.apiIDToDiffs,
				totalDiffEvents: tt.fields.totalEvents,
			}

			p.addDiffToSend(tt.args.newSpec, tt.args.oldSpec, tt.args.diffType, tt.args.specType, tt.args.event)

			assert.Assert(t, reflect.DeepEqual(tt.wantApiIDToDiffs, p.apiIDToDiffs))
			assert.Assert(t, tt.wantTotalEvents == p.totalDiffEvents)
		})
	}
}
