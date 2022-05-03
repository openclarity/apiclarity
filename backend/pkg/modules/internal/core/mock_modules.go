// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openclarity/apiclarity/backend/pkg/modules/internal/core (interfaces: Module,BackendAccessor)

// Package core is a generated GoMock package.
package core

import (
	context "context"
	http "net/http"
	reflect "reflect"

	database "github.com/openclarity/apiclarity/backend/pkg/database"
	gomock "github.com/golang/mock/gomock"
	kubernetes "k8s.io/client-go/kubernetes"
)

// MockModule is a mock of Module interface.
type MockModule struct {
	ctrl     *gomock.Controller
	recorder *MockModuleMockRecorder
}

// MockModuleMockRecorder is the mock recorder for MockModule.
type MockModuleMockRecorder struct {
	mock *MockModule
}

// NewMockModule creates a new mock instance.
func NewMockModule(ctrl *gomock.Controller) *MockModule {
	mock := &MockModule{ctrl: ctrl}
	mock.recorder = &MockModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModule) EXPECT() *MockModuleMockRecorder {
	return m.recorder
}

// EventNotify mocks base method.
func (m *MockModule) EventNotify(arg0 context.Context, arg1 *Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EventNotify", arg0, arg1)
}

// EventNotify indicates an expected call of EventNotify.
func (mr *MockModuleMockRecorder) EventNotify(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventNotify", reflect.TypeOf((*MockModule)(nil).EventNotify), arg0, arg1)
}

// HTTPHandler mocks base method.
func (m *MockModule) HTTPHandler() http.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HTTPHandler")
	ret0, _ := ret[0].(http.Handler)
	return ret0
}

// HTTPHandler indicates an expected call of HTTPHandler.
func (mr *MockModuleMockRecorder) HTTPHandler() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HTTPHandler", reflect.TypeOf((*MockModule)(nil).HTTPHandler))
}

// Name mocks base method.
func (m *MockModule) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockModuleMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockModule)(nil).Name))
}

// MockBackendAccessor is a mock of BackendAccessor interface.
type MockBackendAccessor struct {
	ctrl     *gomock.Controller
	recorder *MockBackendAccessorMockRecorder
}

// MockBackendAccessorMockRecorder is the mock recorder for MockBackendAccessor.
type MockBackendAccessorMockRecorder struct {
	mock *MockBackendAccessor
}

// NewMockBackendAccessor creates a new mock instance.
func NewMockBackendAccessor(ctrl *gomock.Controller) *MockBackendAccessor {
	mock := &MockBackendAccessor{ctrl: ctrl}
	mock.recorder = &MockBackendAccessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBackendAccessor) EXPECT() *MockBackendAccessorMockRecorder {
	return m.recorder
}

// CreateAPIEventAnnotations mocks base method.
func (m *MockBackendAccessor) CreateAPIEventAnnotations(arg0 context.Context, arg1 string, arg2 uint, arg3 ...Annotation) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateAPIEventAnnotations", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAPIEventAnnotations indicates an expected call of CreateAPIEventAnnotations.
func (mr *MockBackendAccessorMockRecorder) CreateAPIEventAnnotations(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAPIEventAnnotations", reflect.TypeOf((*MockBackendAccessor)(nil).CreateAPIEventAnnotations), varargs...)
}

// DeleteAPIInfoAnnotations mocks base method.
func (m *MockBackendAccessor) DeleteAPIInfoAnnotations(arg0 context.Context, arg1 string, arg2 uint, arg3 ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteAPIInfoAnnotations", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAPIInfoAnnotations indicates an expected call of DeleteAPIInfoAnnotations.
func (mr *MockBackendAccessorMockRecorder) DeleteAPIInfoAnnotations(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAPIInfoAnnotations", reflect.TypeOf((*MockBackendAccessor)(nil).DeleteAPIInfoAnnotations), varargs...)
}

// GetAPIEventAnnotation mocks base method.
func (m *MockBackendAccessor) GetAPIEventAnnotation(arg0 context.Context, arg1 string, arg2 uint, arg3 string) (*Annotation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAPIEventAnnotation", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*Annotation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAPIEventAnnotation indicates an expected call of GetAPIEventAnnotation.
func (mr *MockBackendAccessorMockRecorder) GetAPIEventAnnotation(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAPIEventAnnotation", reflect.TypeOf((*MockBackendAccessor)(nil).GetAPIEventAnnotation), arg0, arg1, arg2, arg3)
}

// GetAPIEvents mocks base method.
func (m *MockBackendAccessor) GetAPIEvents(arg0 context.Context, arg1 database.GetAPIEventsQuery) ([]*database.APIEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAPIEvents", arg0, arg1)
	ret0, _ := ret[0].([]*database.APIEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAPIEvents indicates an expected call of GetAPIEvents.
func (mr *MockBackendAccessorMockRecorder) GetAPIEvents(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAPIEvents", reflect.TypeOf((*MockBackendAccessor)(nil).GetAPIEvents), arg0, arg1)
}

// GetAPIInfo mocks base method.
func (m *MockBackendAccessor) GetAPIInfo(arg0 context.Context, arg1 uint) (*database.APIInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAPIInfo", arg0, arg1)
	ret0, _ := ret[0].(*database.APIInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAPIInfo indicates an expected call of GetAPIInfo.
func (mr *MockBackendAccessorMockRecorder) GetAPIInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAPIInfo", reflect.TypeOf((*MockBackendAccessor)(nil).GetAPIInfo), arg0, arg1)
}

// GetAPIInfoAnnotation mocks base method.
func (m *MockBackendAccessor) GetAPIInfoAnnotation(arg0 context.Context, arg1 string, arg2 uint, arg3 string) (*Annotation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAPIInfoAnnotation", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*Annotation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAPIInfoAnnotation indicates an expected call of GetAPIInfoAnnotation.
func (mr *MockBackendAccessorMockRecorder) GetAPIInfoAnnotation(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAPIInfoAnnotation", reflect.TypeOf((*MockBackendAccessor)(nil).GetAPIInfoAnnotation), arg0, arg1, arg2, arg3)
}

// K8SClient mocks base method.
func (m *MockBackendAccessor) K8SClient() kubernetes.Interface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "K8SClient")
	ret0, _ := ret[0].(kubernetes.Interface)
	return ret0
}

// K8SClient indicates an expected call of K8SClient.
func (mr *MockBackendAccessorMockRecorder) K8SClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "K8SClient", reflect.TypeOf((*MockBackendAccessor)(nil).K8SClient))
}

// ListAPIEventAnnotations mocks base method.
func (m *MockBackendAccessor) ListAPIEventAnnotations(arg0 context.Context, arg1 string, arg2 uint) ([]*Annotation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAPIEventAnnotations", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*Annotation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAPIEventAnnotations indicates an expected call of ListAPIEventAnnotations.
func (mr *MockBackendAccessorMockRecorder) ListAPIEventAnnotations(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAPIEventAnnotations", reflect.TypeOf((*MockBackendAccessor)(nil).ListAPIEventAnnotations), arg0, arg1, arg2)
}

// ListAPIInfoAnnotations mocks base method.
func (m *MockBackendAccessor) ListAPIInfoAnnotations(arg0 context.Context, arg1 string, arg2 uint) ([]*Annotation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAPIInfoAnnotations", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*Annotation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAPIInfoAnnotations indicates an expected call of ListAPIInfoAnnotations.
func (mr *MockBackendAccessorMockRecorder) ListAPIInfoAnnotations(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAPIInfoAnnotations", reflect.TypeOf((*MockBackendAccessor)(nil).ListAPIInfoAnnotations), arg0, arg1, arg2)
}

// StoreAPIInfoAnnotations mocks base method.
func (m *MockBackendAccessor) StoreAPIInfoAnnotations(arg0 context.Context, arg1 string, arg2 uint, arg3 ...Annotation) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StoreAPIInfoAnnotations", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreAPIInfoAnnotations indicates an expected call of StoreAPIInfoAnnotations.
func (mr *MockBackendAccessorMockRecorder) StoreAPIInfoAnnotations(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreAPIInfoAnnotations", reflect.TypeOf((*MockBackendAccessor)(nil).StoreAPIInfoAnnotations), varargs...)
}
