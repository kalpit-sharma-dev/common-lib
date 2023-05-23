// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token (interfaces: Exchanger)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockExchanger is a mock of Exchanger interface.
type MockExchanger struct {
	ctrl     *gomock.Controller
	recorder *MockExchangerMockRecorder
}

// MockExchangerMockRecorder is the mock recorder for MockExchanger.
type MockExchangerMockRecorder struct {
	mock *MockExchanger
}

// NewMockExchanger creates a new mock instance.
func NewMockExchanger(ctrl *gomock.Controller) *MockExchanger {
	mock := &MockExchanger{ctrl: ctrl}
	mock.recorder = &MockExchangerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExchanger) EXPECT() *MockExchangerMockRecorder {
	return m.recorder
}

// Exchange mocks base method.
func (m *MockExchanger) Exchange(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exchange", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exchange indicates an expected call of Exchange.
func (mr *MockExchangerMockRecorder) Exchange(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exchange", reflect.TypeOf((*MockExchanger)(nil).Exchange), arg0, arg1, arg2)
}

// ExchangeEnhanced mocks base method.
func (m *MockExchanger) ExchangeEnhanced(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExchangeEnhanced", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExchangeEnhanced indicates an expected call of ExchangeEnhanced.
func (mr *MockExchangerMockRecorder) ExchangeEnhanced(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeEnhanced", reflect.TypeOf((*MockExchanger)(nil).ExchangeEnhanced), arg0, arg1, arg2)
}
