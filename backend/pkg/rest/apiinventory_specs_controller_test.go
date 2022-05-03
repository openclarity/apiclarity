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

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/test"
)

func Test_createTagsListFromRawSpec(t *testing.T) {
	type args struct {
		rawSpec      string
		pathToPathID map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.SpecTag
		wantErr bool
	}{
		{
			name: "no tags",
			args: args{
				rawSpec: test.NewTestSpec().
					WithPathItem("/some/path", test.NewTestPathItem().
						WithOperation(http.MethodGet, test.NewTestOperation().Op).PathItem).String(t),
				pathToPathID: map[string]string{
					"/some/path": "1",
				},
			},
			want: []*models.SpecTag{
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodGet,
							Path:   "/some/path",
							PathID: "1",
						},
					},
					Name: defaultTagName,
				},
			},
			wantErr: false,
		},
		{
			name: "1 path with 2 operations with 2 tags and 2 methods",
			args: args{
				rawSpec: test.NewTestSpec().
					WithPathItem("/some/path", test.NewTestPathItem().
						WithOperation(http.MethodGet, test.NewTestOperation().WithTags([]string{"tag1", "tag2"}).Op).
						WithOperation(http.MethodPut, test.NewTestOperation().WithTags([]string{"tag1", "tag2"}).Op).
						PathItem).String(t),
				pathToPathID: map[string]string{
					"/some/path": "1",
				},
			},
			want: []*models.SpecTag{
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodGet,
							Path:   "/some/path",
							PathID: "1",
						},
						{
							Method: http.MethodPut,
							Path:   "/some/path",
							PathID: "1",
						},
					},
					Name: "tag1",
				},
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodGet,
							Path:   "/some/path",
							PathID: "1",
						},
						{
							Method: http.MethodPut,
							Path:   "/some/path",
							PathID: "1",
						},
					},
					Name: "tag2",
				},
			},
			wantErr: false,
		},
		{
			name: "2 path with 2 operations with 2 methods",
			args: args{
				rawSpec: test.NewTestSpec().
					WithPathItem("/some/path", test.NewTestPathItem().
						WithOperation(http.MethodGet, test.NewTestOperation().WithTags([]string{"tag1", "tag2"}).Op).
						WithOperation(http.MethodPut, test.NewTestOperation().WithTags([]string{"tag1", "tag2"}).Op).
						PathItem).
					WithPathItem("/some/path/2", test.NewTestPathItem().
						WithOperation(http.MethodPost, test.NewTestOperation().WithTags([]string{}).Op).
						WithOperation(http.MethodPut, test.NewTestOperation().WithTags([]string{"tag1"}).Op).
						PathItem).
					String(t),
				pathToPathID: map[string]string{
					"/some/path":   "1",
					"/some/path/2": "2",
				},
			},
			want: []*models.SpecTag{
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodGet,
							Path:   "/some/path",
							PathID: "1",
						},
						{
							Method: http.MethodPut,
							Path:   "/some/path",
							PathID: "1",
						},
						{
							Method: http.MethodPut,
							Path:   "/some/path/2",
							PathID: "2",
						},
					},
					Name: "tag1",
				},
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodGet,
							Path:   "/some/path",
							PathID: "1",
						},
						{
							Method: http.MethodPut,
							Path:   "/some/path",
							PathID: "1",
						},
					},
					Name: "tag2",
				},
				{
					Description: "",
					MethodAndPathList: []*models.MethodAndPath{
						{
							Method: http.MethodPost,
							Path:   "/some/path/2",
							PathID: "2",
						},
					},
					Name: defaultTagName,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createTagsListFromRawSpec(tt.args.rawSpec, tt.args.pathToPathID)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTagsListFromRawSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sortTagList(got)
			sortTagList(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTagsListFromRawSpec() got = %v, want %v", marshal(got), marshal(tt.want))
			}
		})
	}
}

func sortTagList(tags []*models.SpecTag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})
	for _, tag := range tags {
		sort.Slice(tag.MethodAndPathList, func(i, j int) bool {
			return tag.MethodAndPathList[i].Path < tag.MethodAndPathList[j].Path
		})
	}
}
