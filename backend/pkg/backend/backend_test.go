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

package backend

import (
	"context"
	"net/http"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	"github.com/openclarity/apiclarity/backend/pkg/common"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/k8smonitor"
	"github.com/openclarity/apiclarity/backend/pkg/modules"
	"github.com/openclarity/apiclarity/backend/pkg/speculators"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

func Test_isNonAPI(t *testing.T) {
	type args struct {
		trace *_spec.Telemetry
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "content type is not application/json expected to classify as non API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "non-api",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "REST API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: contentTypeApplicationJSON,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "no headers expected to classify as API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{},
					},
				},
			},
			want: false,
		},
		{
			name: "content type is application/hal+json - classify as API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "application/hal+json",
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonAPI(tt.args.trace); got != tt.want {
				t.Errorf("isNonAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHostname(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no scheme",
			args: args{
				host: "example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "with scheme",
			args: args{
				host: "acap://example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "only host",
			args: args{
				host: "example.com",
			},
			want: "example.com",
		},
		{
			name: "hostname is empty",
			args: args{
				host: "https://",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "failed to parse host",
			args: args{
				host: "1 2 3",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHostname(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostname() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_handleHTTPTrace(t *testing.T) {
	mockCtrlDatabase := gomock.NewController(t)
	defer mockCtrlDatabase.Finish()
	mockDatabase := _database.NewMockDatabase(mockCtrlDatabase)

	mockCtrlAPIEventTable := gomock.NewController(t)
	defer mockCtrlAPIEventTable.Finish()
	mockAPIEventTable := _database.NewMockAPIEventsTable(mockCtrlAPIEventTable)

	mockCtrlAPIInventoryTable := gomock.NewController(t)
	defer mockCtrlAPIInventoryTable.Finish()
	mockAPIInventoryTable := _database.NewMockAPIInventoryTable(mockCtrlAPIInventoryTable)

	mockCtrlModules := gomock.NewController(t)
	defer mockCtrlModules.Finish()
	mockModulesManager := modules.NewMockModulesManager(mockCtrlModules)

	op := spec.NewOperation()
	op.Responses = spec.NewResponses()
	op.Responses["202"] = &spec.ResponseRef{Value: spec.NewResponse().WithDescription("response")}

	speculatorsWithProvidedSpec := speculators.NewMapRepository(_speculator.Config{})
	speculatorsWithProvidedSpec.Get(common.DefaultTraceSourceID).Specs[testSpecKey] = _spec.CreateDefaultSpec(host, port, _spec.OperationGeneratorConfig{})
	err := speculatorsWithProvidedSpec.Get(common.DefaultTraceSourceID).LoadProvidedSpec(testSpecKey, []byte(providedSpecV3), map[string]string{})
	assert.NilError(t, err)

	speculatorsWithApprovedSpec := speculators.NewMapRepository(_speculator.Config{})
	speculatorsWithApprovedSpec.Get(common.DefaultTraceSourceID).Specs[testSpecKey] = _spec.CreateDefaultSpec(host, port, _spec.OperationGeneratorConfig{})
	ApprovedSpecReview := &_spec.ApprovedSpecReview{
		PathToPathItem: map[string]*spec.PathItem{
			"/api/1/foo": &_spec.NewTestPathItem().WithOperation(http.MethodPost, op).PathItem,
			"/api/2/foo": &_spec.NewTestPathItem().WithOperation(http.MethodGet, op).PathItem,
		},
		PathItemsReview: []*_spec.ApprovedSpecReviewPathItem{
			{
				ReviewPathItem: _spec.ReviewPathItem{
					ParameterizedPath: "/api/{param1}/foo",
					Paths:             map[string]bool{"/api/1/foo": true, "/api/2/foo": true},
				},
			},
		},
	}
	err = speculatorsWithApprovedSpec.Get(common.DefaultTraceSourceID).ApplyApprovedReview(testSpecKey, ApprovedSpecReview, _spec.OASv2)
	assert.NilError(t, err)

	type fields struct {
		speculators             *speculators.Repository
		monitor                 *k8smonitor.Monitor
		dbHandler               _database.Database
		expectDatabase          func(database *_database.MockDatabase)
		expectAPIEventTable     func(apiEventTable *_database.MockAPIEventsTable)
		expectAPIInventoryTable func(apiInventoryTable *_database.MockAPIInventoryTable)
	}
	type args struct {
		trace *pluginsmodels.Telemetry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "good run",
			fields: fields{
				speculators: speculators.NewMapRepository(_speculator.Config{}),
				monitor:     nil, // TODO turn monitor into interface so we can use it in tests. for now we assume to run locally (no monitor)
				dbHandler:   mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "Host field is empty, get host from headers",
			fields: fields{
				speculators: speculators.NewMapRepository(_speculator.Config{}),
				monitor:     nil,
				dbHandler:   mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers: []*pluginsmodels.Header{
								{
									Key:   "host",
									Value: host,
								},
							},
							Time:    0,
							Version: "1.1",
						},
						Host:   "",
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "no host name found",
			fields: fields{
				speculators:             speculators.NewMapRepository(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   "",
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid destination address",
			fields: fields{
				speculators:             speculators.NewMapRepository(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   "1.1.1.1",
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid source address",
			fields: fields{
				speculators:             speculators.NewMapRepository(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2",
				},
			},
			wantErr: true,
		},
		{
			name: "non api",
			fields: fields{
				speculators: speculators.NewMapRepository(_speculator.Config{}),
				monitor:     nil,
				dbHandler:   mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().WithIsNonAPI(true).event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers: []*pluginsmodels.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "xml",
								},
							},
							Time:    0,
							Version: "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "has provided spec diff",
			fields: fields{
				speculators: speculatorsWithProvidedSpec,
				monitor:     nil,
				dbHandler:   mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "has reconstructed spec diff",
			fields: fields{
				speculators: speculatorsWithApprovedSpec,
				monitor:     nil,
				dbHandler:   mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
	}
	ctx := context.Background()

	for _, tt := range tests {
		tt.fields.expectDatabase(mockDatabase)
		tt.fields.expectAPIInventoryTable(mockAPIInventoryTable)
		tt.fields.expectAPIEventTable(mockAPIEventTable)
		mockModulesManager.EXPECT().EventNotify(ctx, gomock.Any()).AnyTimes()
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				speculators:    tt.fields.speculators,
				monitor:        tt.fields.monitor,
				dbHandler:      tt.fields.dbHandler,
				modulesManager: mockModulesManager,
			}
			if err := b.handleHTTPTrace(ctx, tt.args.trace, nil); (err != nil) != tt.wantErr {
				t.Errorf("handleHTTPTrace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
