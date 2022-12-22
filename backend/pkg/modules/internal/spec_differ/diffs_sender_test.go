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

// nolint: revive,stylecheck
package spec_differ

import (
	"crypto/sha256"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/openclarity/apiclarity/api/server/models"
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
		hash1   = sha256.Sum256([]byte(newSpec + oldSpec))
		hash2   = sha256.Sum256([]byte(newSpec2 + oldSpec2))
		apiType = common.INTERNAL
		uuid0   = uuid.MustParse("00000000-0000-0000-0000-000000000000")
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
					1: {
						hash1: global.Diff{
							DiffType:      common.GENERALDIFF,
							LastSeen:      time.Unix(10, 0),
							Method:        common.GET,
							NewSpec:       newSpec,
							OldSpec:       oldSpec,
							Path:          "/some/path",
							SpecTimestamp: time.Unix(1, 0),
							SpecType:      common.PROVIDED,
						},
						hash2: global.Diff{
							DiffType:      common.ZOMBIEDIFF,
							LastSeen:      time.Unix(11, 0),
							Method:        common.POST,
							NewSpec:       newSpec2,
							OldSpec:       oldSpec2,
							Path:          "/some/path/2",
							SpecTimestamp: time.Unix(2, 0),
							SpecType:      common.PROVIDED,
						},
					},
					2: {
						hash1: global.Diff{
							DiffType:      common.ZOMBIEDIFF,
							LastSeen:      time.Unix(12, 0),
							Method:        common.GET,
							NewSpec:       newSpec,
							OldSpec:       oldSpec,
							Path:          "/some/path/3",
							SpecTimestamp: time.Unix(3, 0),
							SpecType:      common.PROVIDED,
						},
					},
				},
			},
			want: []notifications.SpecDiffsNotification{
				{
					Diffs: global.APIDiffs{
						ApiInfo: common.ApiInfoWithType{
							ApiType:              &apiType,
							DestinationNamespace: stringPtr("bar"),
							HasProvidedSpec:      boolPtr(true),
							HasReconstructedSpec: boolPtr(false),
							Id:                   uint32Ptr(1),
							Name:                 stringPtr("foo"),
							Port:                 intPtr(8080),
							TraceSourceId:        &uuid0,
						},
						Diffs: []global.Diff{
							{
								DiffType:      common.GENERALDIFF,
								LastSeen:      time.Unix(10, 0),
								Method:        common.GET,
								NewSpec:       newSpec,
								OldSpec:       oldSpec,
								Path:          "/some/path",
								SpecTimestamp: time.Unix(1, 0),
								SpecType:      common.PROVIDED,
							},
							{
								DiffType:      common.ZOMBIEDIFF,
								LastSeen:      time.Unix(11, 0),
								Method:        common.POST,
								NewSpec:       newSpec2,
								OldSpec:       oldSpec2,
								Path:          "/some/path/2",
								SpecTimestamp: time.Unix(2, 0),
								SpecType:      common.PROVIDED,
							},
						},
					},
				},
				{
					Diffs: global.APIDiffs{
						ApiInfo: common.ApiInfoWithType{
							ApiType:              &apiType,
							DestinationNamespace: stringPtr("bar2"),
							HasProvidedSpec:      boolPtr(false),
							HasReconstructedSpec: boolPtr(true),
							Id:                   uint32Ptr(2),
							Name:                 stringPtr("foo2"),
							Port:                 intPtr(8000),
							TraceSourceId:        &uuid0,
						},
						Diffs: []global.Diff{
							{
								DiffType:      common.ZOMBIEDIFF,
								LastSeen:      time.Unix(12, 0),
								Method:        common.GET,
								NewSpec:       newSpec,
								OldSpec:       oldSpec,
								Path:          "/some/path/3",
								SpecTimestamp: time.Unix(3, 0),
								SpecType:      common.PROVIDED,
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
					Type:                 models.APITypeINTERNAL,
				}, nil)
				accessor.EXPECT().GetAPIInfo(gomock.Any(), uint(2)).Return(&database.APIInfo{
					ID:                   2,
					Name:                 "foo2",
					Port:                 8000,
					HasProvidedSpec:      false,
					HasReconstructedSpec: true,
					DestinationNamespace: "bar2",
					Type:                 models.APITypeINTERNAL,
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expectAccessor(mockAccessor)
			p := &specDiffer{
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
