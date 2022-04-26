// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package bfladetector

import (
	"net/http"
	"reflect"
	"testing"
)

func TestGetUserID(t *testing.T) {
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    *DetectedUser
		wantErr bool
	}{{
		name: "success jwt",
		args: args{
			headers: http.Header{
				"authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0MCJ9.Go08qgDIwwiCvcWQ9wA2O2-G4urRxGIbvRKGMRu5uyw"},
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceJWT,
			ID:     "test0",
		},
		wantErr: false,
	}, {
		name: "success kong x-customer-id",
		args: args{
			headers: http.Header{
				"x-customer-id": {"test1"},
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceXConsumerIDHeader,
			ID:     "test1",
		},
		wantErr: false,
	}, {
		name: "success basic",
		args: args{
			headers: http.Header{
				"authorization": {"Basic dGVzdDI6cGFzczEK"},
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceBasic,
			ID:     "test2",
		},
		wantErr: false,
	}, {
		name: "no user detected",
		args: args{
			headers: http.Header{},
		},
		want:    nil,
		wantErr: false,
	}, {
		name: "want error",
		args: args{
			headers: http.Header{
				"authorization": {"Bearer 123123123"},
			},
		},
		want:    nil,
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserID(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
