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
	"reflect"
	"testing"

	"github.com/go-openapi/strfmt"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func Test_getModelsSpecDiffTime(t *testing.T) {
	time := strfmt.NewDateTime()
	generalDiff := models.DiffTypeGENERALDIFF
	zombieDiff := models.DiffTypeZOMBIEDIFF
	shadowDiff := models.DiffTypeSHADOWDIFF
	noDiff := models.DiffTypeNODIFF
	type args struct {
		latestDiffs []database.APIEvent
	}
	tests := []struct {
		name string
		args args
		want []*models.SpecDiffTime
	}{
		{
			name: "different spec diff type",
			args: args{
				latestDiffs: []database.APIEvent{
					{
						ID:           1,
						Time:         time,
						SpecDiffType: generalDiff,
						HostSpecName: "host1",
					},
					{
						ID:           2,
						Time:         time,
						SpecDiffType: zombieDiff,
						HostSpecName: "host2",
					},
					{
						ID:           3,
						Time:         time,
						SpecDiffType: shadowDiff,
						HostSpecName: "host3",
					},
					{
						ID:           4,
						Time:         time,
						SpecDiffType: noDiff,
						HostSpecName: "host4",
					},
				},
			},
			want: []*models.SpecDiffTime{
				{
					APIEventID:  1,
					APIHostName: "host1",
					DiffType:    &generalDiff,
					Time:        time,
				},
				{
					APIEventID:  2,
					APIHostName: "host2",
					DiffType:    &zombieDiff,
					Time:        time,
				},
				{
					APIEventID:  3,
					APIHostName: "host3",
					DiffType:    &shadowDiff,
					Time:        time,
				},
				{
					APIEventID:  4,
					APIHostName: "host4",
					DiffType:    &noDiff,
					Time:        time,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getModelsSpecDiffTime(tt.args.latestDiffs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getModelsSpecDiffTime() = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}
