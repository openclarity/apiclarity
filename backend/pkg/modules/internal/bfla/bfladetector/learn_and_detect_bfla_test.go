package bfladetector_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/bfladetector"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
	pluginsmodels "github.com/apiclarity/apiclarity/plugins/api/server/models"
)

const (
	testNamespace = "sock-shop"
	cartsSpec     = `{"paths": {
"/carts/{id}/items": {"get": {"parameters":[{"name": "id", "in": "path"}]}, "post": {"parameters":[{"name": "id", "in": "path"}]}},
"/carts/{id}/merge": {"get": {"parameters":[{"name": "id", "in": "path"}]}, "post": {"parameters":[{"name": "id", "in": "path"}]}}
}}`
)

var (
	mapID2name = map[string]uint{"user": 1, "carts": 2, "catalogue": 3}
	specs      = map[uint]string{mapID2name["carts"]: cartsSpec}
)

// nolint:unparam
func buildTrace(method, path, src, dest, userid string) *bfladetector.CompositeTrace {
	return &bfladetector.CompositeTrace{
		K8SSource:      newClientRef(src),
		K8SDestination: newClientRef(dest),
		DetectedUser:   &bfladetector.DetectedUser{ID: userid},
		Event: &core.Event{
			APIEvent: &database.APIEvent{
				Method:    models.HTTPMethod(method),
				Path:      path,
				APIInfoID: mapID2name[dest],
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

func initBFLADetector(ctrl *gomock.Controller, storedAuthModels map[uint]bfladetector.AuthorizationModel, storedTracesProcessed, storedTracesToLearn map[uint]int) bfladetector.BFLADetector {
	var (
		ctx             = context.Background()
		learnTracesNr   = 100
		openapiProvider = bfladetector.NewMockOpenAPIProvider(ctrl)
		pathResolver    = bfladetector.NewPathResolver(openapiProvider)
		eventAlerter    = bfladetector.NewMockEventAlerter(ctrl)
		backendAccessor = core.NewMockBackendAccessor(ctrl)
		statePersister  = recovery.NewMockStatePersister(ctrl)
	)
	statePersister.EXPECT().Persist(gomock.Any()).AnyTimes()
	openapiProvider.EXPECT().GetOpenAPI(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, apiID uint) (io.Reader, bfladetector.SpecType, error) {
		return bytes.NewBufferString(specs[apiID]), bfladetector.SpecTypeProvided, nil
	}).AnyTimes()
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
	return bfladetector.NewBFLADetector(ctx, learnTracesNr, pathResolver, eventAlerter, statePersister)
}

func Test_learnAndDetectBFLA_BuildAuthzModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storedAuthModels := map[uint]bfladetector.AuthorizationModel{}
	storedTracesProcessed := map[uint]int{}
	storedTracesToLearn := map[uint]int{}
	detector := initBFLADetector(ctrl, storedAuthModels, storedTracesProcessed, storedTracesToLearn)

	tests := []struct {
		name           string
		traces         []*bfladetector.CompositeTrace
		wantAuthModels map[uint]bfladetector.AuthorizationModel
	}{{
		name: "Builds auth model correctly",
		traces: []*bfladetector.CompositeTrace{
			buildTrace("GET", "/carts/61fbce65997a8ede0eea3c57/items", "frontend", "carts", "user1"),
			buildTrace("GET", "/carts/61fbce65997a8ede0eea3c53/items", "frontend", "carts", "user2"),
			buildTrace("POST", "/carts/61fbce65997a8ede0eea3c57/items", "frontend", "carts", "user1"),
			buildTrace("POST", "/addresses", "frontend", "carts", "user3"),
			buildTrace("POST", "/login", "frontend", "user", "user2"),
			buildTrace("POST", "/register", "frontend", "user", "user2"),
			buildTrace("GET", "/catalogue", "frontend", "catalogue", "user3"),
			buildTrace("GET", "/cards", "frontend", "catalogue", "user3"),
		},
		wantAuthModels: authModels(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, trace := range tt.traces {
				trace.APIEvent.APIInfoID = mapID2name[trace.K8SDestination.Uid]
				detector.SendTrace(trace)
				t.Log("trace sent")
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

			detector := initBFLADetector(ctrl, tt.authModels, storedTracesProcessed, storedTracesToLearn)
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
						EndUsers:   bfladetector.EndUsers{{ID: "user1"}},
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

			detector := initBFLADetector(ctrl, tt.authModels, storedTracesProcessed, storedTracesToLearn)
			detector.ApproveTrace("/carts/31231231132/merge", "POST", newClientRef("frontend"), mapID2name["carts"], &bfladetector.DetectedUser{ID: "user1", Source: bfladetector.DetectedUserSourceJWT})
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
