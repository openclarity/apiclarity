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

package bfladetector_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/bfladetector"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

const testNamespace = "sock-shop"

var mapID2name = map[string]uint{"user": 1, "carts": 2, "catalogue": 3}

// nolint:unparam
func buildTrace(method, path, src, dest, userid string) *bfladetector.CompositeTrace {
	return &bfladetector.CompositeTrace{
		K8SSource:      newClientRef(src),
		K8SDestination: newClientRef(dest),
		DetectedUser:   &bfladetector.DetectedUser{ID: userid},
		Event: &core.Event{
			APIEvent: &database.APIEvent{
				ProvidedPathID: "test",
				Method:         models.HTTPMethod(method),
				Path:           path,
				APIInfoID:      mapID2name[dest],
			},
			Telemetry: &pluginsmodels.Telemetry{
				DestinationNamespace: testNamespace,
				Request: &pluginsmodels.Request{
					Method: method,
					Path:   path,
					Host:   dest,
				},
			},
		},
	}
}

func getAPIInfoWithTags(path string) *database.APIInfo {
	return &database.APIInfo{
		ProvidedSpecInfo: fmt.Sprintf(`{"tags":[{"methodAndPathList":[{"pathId":"test","path":%q}]}]}`, path),
		HasProvidedSpec:  true,
	}
}

func initBFLADetector(ctrl *gomock.Controller, backendAccessor *core.MockBackendAccessor, storedAuthModels map[uint]bfladetector.AuthorizationModel, storedTracesProcessed, storedTracesToLearn map[uint]int) bfladetector.BFLADetector {
	var (
		ctx            = context.Background()
		learnTracesNr  = 100
		eventAlerter   = bfladetector.NewMockEventAlerter(ctrl)
		statePersister = recovery.NewMockStatePersister(ctrl)
	)
	statePersister.EXPECT().Persist(gomock.Any()).AnyTimes()
	eventAlerter.EXPECT().SetEventAlert(ctx, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id string, name uint, severity core.AlertSeverity) error {
		log.Println("alert requested with severity: ", severity)
		return nil
	}).AnyTimes()
	statePersister.EXPECT().AckSubmit(gomock.Any()).AnyTimes()
	statePersister.EXPECT().UseState(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(arg0 uint, arg1 string, arg2 interface{}) (recovery.SetState, bool, error) {
		switch arg1 {
		case bfladetector.AuthzModelAnnotationName:
			val, found := storedAuthModels[arg0]
			reflect.ValueOf(arg2).Elem().Set(reflect.ValueOf(val))
			return func(state interface{}) {
				// nolint:forcetypeassert
				val := state.(bfladetector.AuthorizationModel)
				storedAuthModels[arg0] = val
			}, found, nil
		case bfladetector.AuthzProcessedTracesAnnotationName:
			val, found := storedTracesProcessed[arg0]
			reflect.ValueOf(arg2).Elem().Set(reflect.ValueOf(val))
			return func(state interface{}) {
				// nolint:forcetypeassert
				val := state.(int)
				storedTracesProcessed[arg0] = val
			}, found, nil
		case bfladetector.AuthzTracesToLearnAnnotationName:
			val, found := storedTracesToLearn[arg0]
			reflect.ValueOf(arg2).Elem().Set(reflect.ValueOf(val))
			return func(state interface{}) {
				// nolint:forcetypeassert
				val := state.(int)
				storedTracesToLearn[arg0] = val
			}, found, nil
		}
		panic("unknown annotation name")
	}).AnyTimes()
	backendAccessor.EXPECT().CreateAPIEventAnnotations(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	backendAccessor.EXPECT().GetAPIInfoAnnotation(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return bfladetector.NewBFLADetector(ctx, learnTracesNr, backendAccessor, eventAlerter, statePersister)
}

func Test_learnAndDetectBFLA_BuildAuthzModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	backendAccessor := core.NewMockBackendAccessor(ctrl)

	storedAuthModels := map[uint]bfladetector.AuthorizationModel{}
	storedTracesProcessed := map[uint]int{}
	storedTracesToLearn := map[uint]int{}
	detector := initBFLADetector(ctrl, backendAccessor, storedAuthModels, storedTracesProcessed, storedTracesToLearn)

	type testTrace struct {
		*bfladetector.CompositeTrace
		resolvedPath string
	}
	tests := []struct {
		name           string
		traces         []*testTrace
		wantAuthModels map[uint]bfladetector.AuthorizationModel
	}{{
		name: "Builds auth model correctly",
		traces: []*testTrace{
			{buildTrace("GET", "/carts/61fbce65997a8ede0eea3c57/items", "frontend", "carts", "user1"), "/carts/{id}/items"},
			{buildTrace("GET", "/carts/61fbce65997a8ede0eea3c53/items", "frontend", "carts", "user2"), "/carts/{id}/items"},
			{buildTrace("POST", "/carts/61fbce65997a8ede0eea3c57/items", "frontend", "carts", "user1"), "/carts/{id}/items"},
			{buildTrace("POST", "/addresses", "frontend", "carts", "user3"), "/addresses"},
			{buildTrace("POST", "/login", "frontend", "user", "user2"), "/login"},
			{buildTrace("POST", "/register", "frontend", "user", "user2"), "/register"},
			{buildTrace("GET", "/catalogue", "frontend", "catalogue", "user3"), "/catalogue"},
			{buildTrace("GET", "/cards", "frontend", "catalogue", "user3"), "/cards"},
		},
		wantAuthModels: authModels(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, trace := range tt.traces {
				backendAccessor.EXPECT().GetAPIInfo(context.TODO(), gomock.Any()).DoAndReturn(func(ctx context.Context, apiID uint) (*database.APIInfo, error) {
					return getAPIInfoWithTags(trace.resolvedPath), nil
				}).Times(1)
				trace.APIEvent.APIInfoID = mapID2name[trace.K8SDestination.Uid]
				detector.SendTrace(trace.CompositeTrace)
				time.Sleep(100 * time.Millisecond)
			}
			assert(t, tt.wantAuthModels, storedAuthModels)
		})
	}
}

func assert(t *testing.T, want, got map[uint]bfladetector.AuthorizationModel) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("len(want) = %d len(got) = %d", len(want), len(got))
	}
	for modelKey, authModel := range want {
		if !reflect.DeepEqual(authModel, got[modelKey]) {
			diff := cmp.Diff(authModel, got[modelKey])
			t.Errorf("want = %s got = %s", toJSON(authModel), toJSON(got[modelKey]))
			t.Errorf("diff:\n%v\n", diff)
		}
	}
}

func newClientRef(name string) *k8straceannotator.K8sObjectRef {
	return &k8straceannotator.K8sObjectRef{Namespace: testNamespace, Name: name, Uid: name}
}

func Test_learnAndDetectBFLA_DenyTrace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name                       string
		authModels, wantAuthModels map[uint]bfladetector.AuthorizationModel
	}{{
		name:       "deny trace success",
		authModels: authModels(),
		wantAuthModels: map[uint]bfladetector.AuthorizationModel{
			mapID2name["user"]: {
				Operations: bfladetector.Operations{{
					Method: "POST",
					Path:   "/login",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/register",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
						Authorized: true,
					}},
				}},
			},
			mapID2name["catalogue"]: {
				Operations: bfladetector.Operations{{
					Method: "GET",
					Path:   "/catalogue",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}, {
					Method: "GET",
					Path:   "/cards",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}},
			},
			mapID2name["carts"]: {
				Operations: bfladetector.Operations{{
					Method: "GET",
					Path:   "/carts/{id}/items",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user1"}, {ID: "user2"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/carts/{id}/items",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user1"}},
						Authorized: false,
					}},
				}, {
					Method: "POST",
					Path:   "/addresses",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedTracesProcessed := map[uint]int{}
			storedTracesToLearn := map[uint]int{}

			backendAccessor := core.NewMockBackendAccessor(ctrl)
			detector := initBFLADetector(ctrl, backendAccessor, tt.authModels, storedTracesProcessed, storedTracesToLearn)
			detector.DenyTrace("/carts/{id}/items", "POST", newClientRef("frontend"), mapID2name["carts"], nil)
			time.Sleep(1 * time.Second)
			assert(t, tt.wantAuthModels, tt.authModels)
		})
	}
}

func Test_learnAndDetectBFLA_ApproveTrace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name                       string
		authModels, wantAuthModels map[uint]bfladetector.AuthorizationModel
	}{{
		name:       "approve trace success",
		authModels: authModels(),
		wantAuthModels: map[uint]bfladetector.AuthorizationModel{
			mapID2name["user"]: {
				Operations: bfladetector.Operations{{
					Method: "POST",
					Path:   "/login",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/register",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
						Authorized: true,
					}},
				}},
			},
			mapID2name["catalogue"]: {
				Operations: bfladetector.Operations{{
					Method: "GET",
					Path:   "/catalogue",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}, {
					Method: "GET",
					Path:   "/cards",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}},
			},
			mapID2name["carts"]: {
				Operations: bfladetector.Operations{{
					Method: "GET",
					Path:   "/carts/{id}/items",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user1"}, {ID: "user2"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/carts/{id}/items",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user1"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/addresses",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
						Authorized: true,
					}},
				}, {
					Method: "POST",
					Path:   "/carts/{id}/merge",
					Audience: bfladetector.Audience{{
						K8sObject:  newClientRef("frontend"),
						EndUsers:   bfladetector.EndUsers{{ID: "user1", Source: bfladetector.DetectedUserSourceJWT}},
						Authorized: true,
					}},
				}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedTracesProcessed := map[uint]int{}
			storedTracesToLearn := map[uint]int{}

			backendAccessor := core.NewMockBackendAccessor(ctrl)
			detector := initBFLADetector(ctrl, backendAccessor, tt.authModels, storedTracesProcessed, storedTracesToLearn)
			detector.ApproveTrace("/carts/{id}/merge", "POST", newClientRef("frontend"), mapID2name["carts"], &bfladetector.DetectedUser{ID: "user1", Source: bfladetector.DetectedUserSourceJWT})
			time.Sleep(1 * time.Second)
			assert(t, tt.wantAuthModels, tt.authModels)
		})
	}
}

func toJSON(v interface{}) []byte {
	bb, _ := json.Marshal(v)
	return bb
}

func authModels() map[uint]bfladetector.AuthorizationModel {
	return map[uint]bfladetector.AuthorizationModel{
		mapID2name["user"]: {
			Operations: bfladetector.Operations{{
				Method: "POST",
				Path:   "/login",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
					Authorized: true,
				}},
			}, {
				Method: "POST",
				Path:   "/register",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user2"}},
					Authorized: true,
				}},
			}},
		},
		mapID2name["catalogue"]: {
			Operations: bfladetector.Operations{{
				Method: "GET",
				Path:   "/catalogue",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
					Authorized: true,
				}},
			}, {
				Method: "GET",
				Path:   "/cards",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
					Authorized: true,
				}},
			}},
		},
		mapID2name["carts"]: {
			Operations: bfladetector.Operations{{
				Method: "GET",
				Path:   "/carts/{id}/items",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user1"}, {ID: "user2"}},
					Authorized: true,
				}},
			}, {
				Method: "POST",
				Path:   "/carts/{id}/items",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user1"}},
					Authorized: true,
				}},
			}, {
				Method: "POST",
				Path:   "/addresses",
				Audience: bfladetector.Audience{{
					K8sObject:  newClientRef("frontend"),
					EndUsers:   bfladetector.EndUsers{{ID: "user3"}},
					Authorized: true,
				}},
			}},
		},
	}
}

func Test_Contains(t *testing.T) {
	type args struct {
		items []string
		val   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "is success",
		args: args{
			items: []string{"A", "B", "C", "D", "E"},
			val:   "A",
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bfladetector.Contains(tt.args.items, tt.args.val); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsAll(t *testing.T) {
	type args struct {
		items []string
		vals  []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "is success",
		args: args{
			items: []string{"pets:write", "pets:read"},
			vals:  []string{"pets:write", "pets:read", "admin"},
		},
		want: true,
	}, {
		name: "is failure 1",
		args: args{
			items: []string{"pets:write", "pets:read"},
			vals:  []string{"pets:write"},
		},
		want: false,
	}, {
		name: "is failure 2",
		args: args{
			items: []string{"pets:read"},
			vals:  []string{"pets:write"},
		},
		want: false,
	}, {
		name: "is failure 3",
		args: args{
			items: []string{"pets:write", "pets:read"},
			vals:  []string{"tags:write", "tags:read"},
		},
		want: false,
	}, {
		name: "is failure 4",
		args: args{
			items: []string{"pets:write", "pets:read"},
			vals:  []string{""},
		},
		want: false,
	}, {
		name: "is failure 5",
		args: args{
			items: []string{"pets:write", "pets:read"},
			vals:  []string{},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bfladetector.ContainsAll(tt.args.items, tt.args.vals); got != tt.want {
				t.Errorf("ContainsAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
