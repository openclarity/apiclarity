package differ

import (
	"crypto/sha256"
	"reflect"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

func Test_pluginDiffer_getSpecDiffsNotifications(t *testing.T) {
	mockCtrlAccessor := gomock.NewController(t)
	defer mockCtrlAccessor.Finish()
	mockAccessor := core.NewMockBackendAccessor(mockCtrlAccessor)

	const (
		newSpec  = "newSpec"
		oldSpec  = "newSpec"
		newSpec2 = "newSpec2"
		oldSpec2 = "newSpec2"
	)
	var (
		hash1 = sha256.Sum256([]byte(newSpec + oldSpec))
		hash2 = sha256.Sum256([]byte(newSpec2 + oldSpec2))
	)

	type fields struct {
		apiIDToDiffs map[uint]map[diffHash]global.Diff
	}
	tests := []struct {
		name           string
		fields         fields
		expectAccessor func(accessor *core.MockBackendAccessor)
		want           []notifications.SpecDiffsNotification
	}{
		{
			name: "2 apis - one with 2 diffs, one with one diff",
			fields: fields{
				apiIDToDiffs: map[uint]map[diffHash]global.Diff{
					1: {hash1: global.Diff{
						DiffType: common.GENERALDIFF,
						LastSeen: 1,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
					},
						hash2: global.Diff{
							DiffType: common.ZOMBIEDIFF,
							LastSeen: 1,
							NewSpec:  newSpec2,
							OldSpec:  oldSpec2,
						}},
					2: {hash1: global.Diff{
						DiffType: common.ZOMBIEDIFF,
						LastSeen: 2,
						NewSpec:  newSpec,
						OldSpec:  oldSpec,
					}},
				},
			},
			want: []notifications.SpecDiffsNotification{
				{
					Diffs: global.APIDiffs{
						ApiInfo: common.ApiInfo{
							DestinationNamespace: stringPtr("bar"),
							HasProvidedSpec:      boolPtr(true),
							HasReconstructedSpec: boolPtr(false),
							Id:                   uint32Ptr(1),
							Name:                 stringPtr("foo"),
							Port:                 intPtr(8080),
						},
						Diffs: []global.Diff{
							{
								DiffType: common.GENERALDIFF,
								LastSeen: 1,
								NewSpec:  newSpec,
								OldSpec:  oldSpec,
							},
							{
								DiffType: common.ZOMBIEDIFF,
								LastSeen: 1,
								NewSpec:  newSpec2,
								OldSpec:  oldSpec2,
							},
						},
					},
				},
				{
					Diffs: global.APIDiffs{
						ApiInfo: common.ApiInfo{
							DestinationNamespace: stringPtr("bar2"),
							HasProvidedSpec:      boolPtr(false),
							HasReconstructedSpec: boolPtr(true),
							Id:                   uint32Ptr(2),
							Name:                 stringPtr("foo2"),
							Port:                 intPtr(8000),
						},
						Diffs: []global.Diff{
							{
								DiffType: common.ZOMBIEDIFF,
								LastSeen: 2,
								NewSpec:  newSpec,
								OldSpec:  oldSpec,
							},
						},
					},
				},
			},
			expectAccessor: func(accessor *core.MockBackendAccessor) {
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(1)).Return(&database.APIInfo{
					ID:                   1,
					Name:                 "foo",
					Port:                 8080,
					HasProvidedSpec:      true,
					HasReconstructedSpec: false,
					DestinationNamespace: "bar",
				}, nil)
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(2)).Return(&database.APIInfo{
					ID:                   2,
					Name:                 "foo2",
					Port:                 8000,
					HasProvidedSpec:      false,
					HasReconstructedSpec: true,
					DestinationNamespace: "bar2",
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expectAccessor(mockAccessor)
			p := &differ{
				apiIDToDiffs: tt.fields.apiIDToDiffs,
				accessor:     mockAccessor,
			}
			got := p.getSpecDiffsNotifications()
			sort.Slice(got, func(i, j int) bool {
				return *got[i].Diffs.ApiInfo.Id > *got[j].Diffs.ApiInfo.Id
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return *tt.want[i].Diffs.ApiInfo.Id > *tt.want[j].Diffs.ApiInfo.Id
			})

			for i := range got {
				goti := got[i]
				sort.Slice(goti.Diffs.Diffs, func(i, j int) bool {
					return goti.Diffs.Diffs[i].DiffType > goti.Diffs.Diffs[j].DiffType
				})
			}

			for i := range tt.want {
				wanti := tt.want[i]
				sort.Slice(wanti.Diffs.Diffs, func(i, j int) bool {
					return wanti.Diffs.Diffs[i].DiffType > wanti.Diffs.Diffs[j].DiffType
				})
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSpecDiffsNotifications() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(val string) *string {
	ret := val

	return &ret
}

func boolPtr(val bool) *bool {
	ret := val

	return &ret
}

func uint32Ptr(val uint32) *uint32 {
	ret := val

	return &ret
}

func intPtr(val int) *int {
	ret := val

	return &ret
}
