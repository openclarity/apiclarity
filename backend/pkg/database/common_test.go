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

package database

import (
	"strings"
	"testing"

	"github.com/openclarity/apiclarity/api/server/models"
)

func Test_getSortKeyColumnName(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "models.APIEventSortKeyTime",
			args: args{
				key: string(models.APIEventSortKeyTime),
			},
			want:    timeColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyMethod",
			args: args{
				key: string(models.APIEventSortKeyMethod),
			},
			want:    methodColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyPath",
			args: args{
				key: string(models.APIEventSortKeyPath),
			},
			want:    pathColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyStatusCode",
			args: args{
				key: string(models.APIEventSortKeyStatusCode),
			},
			want:    statusCodeColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeySourceIP",
			args: args{
				key: string(models.APIEventSortKeySourceIP),
			},
			want:    sourceIPColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyDestinationIP",
			args: args{
				key: string(models.APIEventSortKeyDestinationIP),
			},
			want:    destinationIPColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyDestinationPort",
			args: args{
				key: string(models.APIEventSortKeyDestinationPort),
			},
			want:    destinationPortColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeySpecDiffType",
			args: args{
				key: string(models.APIEventSortKeySpecDiffType),
			},
			want:    specDiffTypeColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyHostSpecName",
			args: args{
				key: string(models.APIEventSortKeyHostSpecName),
			},
			want:    hostSpecNameColumnName,
			wantErr: false,
		},
		{
			name: "models.APIEventSortKeyAPIType",
			args: args{
				key: string(models.APIEventSortKeyAPIType),
			},
			want:    eventTypeColumnName,
			wantErr: false,
		},
		{
			name: "models.APIInventorySortKeyName",
			args: args{
				key: string(models.APIInventorySortKeyName),
			},
			want:    nameColumnName,
			wantErr: false,
		},
		{
			name: "models.APIInventorySortKeyPort",
			args: args{
				key: string(models.APIInventorySortKeyPort),
			},
			want:    portColumnName,
			wantErr: false,
		},
		{
			name: "models.APIInventorySortKeyHasReconstructedSpec",
			args: args{
				key: string(models.APIInventorySortKeyHasReconstructedSpec),
			},
			want:    hasReconstructedSpecColumnName,
			wantErr: false,
		},
		{
			name: "models.APIInventorySortKeyHasProvidedSpec",
			args: args{
				key: string(models.APIInventorySortKeyHasProvidedSpec),
			},
			want:    hasProvidedSpecColumnName,
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				key: "unknown",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSortKeyColumnName(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSortKeyColumnName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSortKeyColumnName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSortOrder(t *testing.T) {
	sortDir := "ASC"
	type args struct {
		sortKey string
		sortDir *string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				sortKey: string(models.APIInventorySortKeyHasProvidedSpec),
				sortDir: &sortDir,
			},
			want:    hasProvidedSpecColumnName + " " + strings.ToLower(sortDir),
			wantErr: false,
		},
		{
			name: "unknown sort key",
			args: args{
				sortKey: "unknown",
				sortDir: &sortDir,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateSortOrder(tt.args.sortKey, tt.args.sortDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSortOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateSortOrder() got = %v, want %v", got, tt.want)
			}
		})
	}
}
