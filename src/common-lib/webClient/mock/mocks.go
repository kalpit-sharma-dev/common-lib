// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient (interfaces: ClientFactory,ClientService,HTTPClientFactory,HTTPClientService)

// Package mock is a generated GoMock package.
package mock

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	webClient "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

// MockClientFactory is a mock of ClientFactory interface.
type MockClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockClientFactoryMockRecorder
}

// MockClientFactoryMockRecorder is the mock recorder for MockClientFactory.
type MockClientFactoryMockRecorder struct {
	mock *MockClientFactory
}

// NewMockClientFactory creates a new mock instance.
func NewMockClientFactory(ctrl *gomock.Controller) *MockClientFactory {
	mock := &MockClientFactory{ctrl: ctrl}
	mock.recorder = &MockClientFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientFactory) EXPECT() *MockClientFactoryMockRecorder {
	return m.recorder
}

// GetClientService mocks base method.
func (m *MockClientFactory) GetClientService(arg0 webClient.HTTPClientFactory, arg1 webClient.ClientConfig) webClient.ClientService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClientService", arg0, arg1)
	ret0, _ := ret[0].(webClient.ClientService)
	return ret0
}

// GetClientService indicates an expected call of GetClientService.
func (mr *MockClientFactoryMockRecorder) GetClientService(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClientService", reflect.TypeOf((*MockClientFactory)(nil).GetClientService), arg0, arg1)
}

// GetClientServiceByType mocks base method.
func (m *MockClientFactory) GetClientServiceByType(arg0 webClient.ClientType, arg1 webClient.ClientConfig) webClient.HTTPClientService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClientServiceByType", arg0, arg1)
	ret0, _ := ret[0].(webClient.HTTPClientService)
	return ret0
}

// GetClientServiceByType indicates an expected call of GetClientServiceByType.
func (mr *MockClientFactoryMockRecorder) GetClientServiceByType(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClientServiceByType", reflect.TypeOf((*MockClientFactory)(nil).GetClientServiceByType), arg0, arg1)
}

// MockClientService is a mock of ClientService interface.
type MockClientService struct {
	ctrl     *gomock.Controller
	recorder *MockClientServiceMockRecorder
}

// MockClientServiceMockRecorder is the mock recorder for MockClientService.
type MockClientServiceMockRecorder struct {
	mock *MockClientService
}

// NewMockClientService creates a new mock instance.
func NewMockClientService(ctrl *gomock.Controller) *MockClientService {
	mock := &MockClientService{ctrl: ctrl}
	mock.recorder = &MockClientServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientService) EXPECT() *MockClientServiceMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockClientService) Do(arg0 *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockClientServiceMockRecorder) Do(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockClientService)(nil).Do), arg0)
}

// MockHTTPClientFactory is a mock of HTTPClientFactory interface.
type MockHTTPClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockHTTPClientFactoryMockRecorder
}

// MockHTTPClientFactoryMockRecorder is the mock recorder for MockHTTPClientFactory.
type MockHTTPClientFactoryMockRecorder struct {
	mock *MockHTTPClientFactory
}

// NewMockHTTPClientFactory creates a new mock instance.
func NewMockHTTPClientFactory(ctrl *gomock.Controller) *MockHTTPClientFactory {
	mock := &MockHTTPClientFactory{ctrl: ctrl}
	mock.recorder = &MockHTTPClientFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHTTPClientFactory) EXPECT() *MockHTTPClientFactoryMockRecorder {
	return m.recorder
}

// GetHTTPClient mocks base method.
func (m *MockHTTPClientFactory) GetHTTPClient(arg0 webClient.ClientConfig) webClient.HTTPClientService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHTTPClient", arg0)
	ret0, _ := ret[0].(webClient.HTTPClientService)
	return ret0
}

// GetHTTPClient indicates an expected call of GetHTTPClient.
func (mr *MockHTTPClientFactoryMockRecorder) GetHTTPClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHTTPClient", reflect.TypeOf((*MockHTTPClientFactory)(nil).GetHTTPClient), arg0)
}

// MockHTTPClientService is a mock of HTTPClientService interface.
type MockHTTPClientService struct {
	ctrl     *gomock.Controller
	recorder *MockHTTPClientServiceMockRecorder
}

// MockHTTPClientServiceMockRecorder is the mock recorder for MockHTTPClientService.
type MockHTTPClientServiceMockRecorder struct {
	mock *MockHTTPClientService
}

// NewMockHTTPClientService creates a new mock instance.
func NewMockHTTPClientService(ctrl *gomock.Controller) *MockHTTPClientService {
	mock := &MockHTTPClientService{ctrl: ctrl}
	mock.recorder = &MockHTTPClientServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHTTPClientService) EXPECT() *MockHTTPClientServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockHTTPClientService) Create() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Create")
}

// Create indicates an expected call of Create.
func (mr *MockHTTPClientServiceMockRecorder) Create() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockHTTPClientService)(nil).Create))
}

// Do mocks base method.
func (m *MockHTTPClientService) Do(arg0 *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockHTTPClientServiceMockRecorder) Do(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockHTTPClientService)(nil).Do), arg0)
}

// SetCheckRedirect mocks base method.
func (m *MockHTTPClientService) SetCheckRedirect(arg0 func(*http.Request, []*http.Request) error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCheckRedirect", arg0)
}

// SetCheckRedirect indicates an expected call of SetCheckRedirect.
func (mr *MockHTTPClientServiceMockRecorder) SetCheckRedirect(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCheckRedirect", reflect.TypeOf((*MockHTTPClientService)(nil).SetCheckRedirect), arg0)
}