// +build windows

// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/api/win/pdh (interfaces: PerfWinCollector)

// Package mock is a generated GoMock package.
package pdhmock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pdh "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/api/win/pdh"
)

// MockPerfWinCollector is a mock of PerfWinCollector interface
type MockPerfWinCollector struct {
	ctrl     *gomock.Controller
	recorder *MockPerfWinCollectorMockRecorder
}

// MockPerfWinCollectorMockRecorder is the mock recorder for MockPerfWinCollector
type MockPerfWinCollectorMockRecorder struct {
	mock *MockPerfWinCollector
}

// NewMockPerfWinCollector creates a new mock instance
func NewMockPerfWinCollector(ctrl *gomock.Controller) *MockPerfWinCollector {
	mock := &MockPerfWinCollector{ctrl: ctrl}
	mock.recorder = &MockPerfWinCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPerfWinCollector) EXPECT() *MockPerfWinCollectorMockRecorder {
	return m.recorder
}

// GetCountersAndInstances mocks base method
func (m *MockPerfWinCollector) GetCountersAndInstances(arg0 string) ([]string, []string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCountersAndInstances", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].([]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetCountersAndInstances indicates an expected call of GetCountersAndInstances
func (mr *MockPerfWinCollectorMockRecorder) GetCountersAndInstances(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCountersAndInstances", reflect.TypeOf((*MockPerfWinCollector)(nil).GetCountersAndInstances), arg0)
}

// QueryPerformanceData mocks base method
func (m *MockPerfWinCollector) QueryPerformanceData(arg0 []pdh.PerfPaths) ([]pdh.PerfData, uint32) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryPerformanceData", arg0)
	ret0, _ := ret[0].([]pdh.PerfData)
	ret1, _ := ret[1].(uint32)
	return ret0, ret1
}

// QueryPerformanceData indicates an expected call of QueryPerformanceData
func (mr *MockPerfWinCollectorMockRecorder) QueryPerformanceData(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryPerformanceData", reflect.TypeOf((*MockPerfWinCollector)(nil).QueryPerformanceData), arg0)
}
