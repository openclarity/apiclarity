/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rest

import "testing"

var jsonSpec = "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"title\": \"Sample API\",\n    \"description\": \"API description in Markdown.\",\n    \"version\": \"1.0.0\"\n  },\n  \"host\": \"api.example.com\",\n  \"basePath\": \"/v1\",\n  \"schemes\": [\n    \"https\"\n  ],\n  \"paths\": {\n    \"/cats\": {\n      \"get\": {\n        \"summary\": \"Returns a list of cats.\",\n        \"description\": \"Optional extended description in Markdown.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    }\n  }\n}"
var nonValidJsonSpec = "{\n  \"swagger\": \"2.0\",\n  \"host\": \"api.example.com\",\n  \"basePath\": \"/v1\",\n  \"schemes\": [\n    \"https\"\n  ],\n  \"paths\": {\n    \"/cats\": {\n      \"get\": {\n        \"summary\": \"Returns a list of cats.\",\n        \"description\": \"Optional extended description in Markdown.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    }\n  }\n}"
var nonValidJson = "{\n  \"swagger\" \"2.0\",\n  \"info\": {\n    \"title\": \"Sample API\",\n    \"description\": \"API description in Markdown.\",\n    \"version\": \"1.0.0\"\n  },\n  \"host\": \"api.example.com\",\n  \"basePath\": \"/v1\",\n  \"schemes\": [\n    \"https\"\n  ],\n  \"paths\": {\n    \"/cats\": {\n      \"get\": {\n        \"summary\": \"Returns a list of cats.\",\n        \"description\": \"Optional extended description in Markdown.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    }\n  }\n}"

var yamlSpec = "---\nswagger: '2.0'\ninfo:\n  title: Sample API\n  description: API description in Markdown.\n  version: 1.0.0\nhost: api.example.com\nbasePath: \"/v1\"\nschemes:\n- https\npaths:\n  \"/cats\":\n    get:\n      summary: Returns a list of cats.\n      description: Optional extended description in Markdown.\n      produces:\n      - application/json\n      responses:\n        '200':\n          description: OK"
var nonValidYamlSpec = "---\nswagger: '2.0'\nhost: api.example.com\nbasePath: \"/v1\"\nschemes:\n- https\npaths:\n  \"/cats\":\n    get:\n      summary: Returns a list of cats.\n      description: Optional extended description in Markdown.\n      produces:\n      - application/json\n      responses:\n        '200':\n          description: OK\n"
var nonValidYaml = "---\nswagger '2.0'\ninfo:\n  title: Sample API\n  description: API description in Markdown.\n  version: 1.0.0\nhost: api.example.com\nbasePath: \"/v1\"\nschemes:\n- https\npaths:\n  \"/cats\":\n    get:\n      summary: Returns a list of cats.\n      description: Optional extended description in Markdown.\n      produces:\n      - application/json\n      responses:\n        '200':\n          description: OK"


func Test_validateJsonSpec(t *testing.T) {
	type args struct {
		rawSpec []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "json spec - valid",
			args:    args{
				rawSpec: []byte(jsonSpec),
			},
			wantErr: false,
		},
		{
			name:    "yaml spec - not valid",
			args:    args{
				rawSpec: []byte(yamlSpec),
			},
			wantErr: true,
		},
		{
			name:    "not valid - string",
			args:    args{
				rawSpec: []byte("foo"),
			},
			wantErr: true,
		},
		{
			name:    "not a valid json",
			args:    args{
				rawSpec: []byte(nonValidJson),
			},
			wantErr: true,
		},
		{
			name:    "not a valid spec - json",
			args:    args{
				rawSpec: []byte(nonValidJsonSpec),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateJsonSpec(tt.args.rawSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateJsonSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isYamlSpec(t *testing.T) {
	type args struct {
		rawSpec []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid yaml spec",
			args: args{
				rawSpec: []byte(yamlSpec),
			},
			want: true,
		},
		{
			name: "spec is not a valid yaml",
			args: args{
				rawSpec: []byte(nonValidYaml),
			},
			want: false,
		},
		{
			name: "spec not valid, but yaml can marshal into spec",
			args: args{
				rawSpec: []byte(nonValidYamlSpec),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isYamlSpec(tt.args.rawSpec); got != tt.want {
				t.Errorf("isYamlSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
