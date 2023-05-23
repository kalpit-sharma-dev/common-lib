// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/goc (interfaces: Session)

// Package mock_db is a generated GoMock package.
package mock_db

import (
	gocql "github.com/gocql/gocql"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSession is a mock of Session interface
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *MockSessionMockRecorder
}

// MockSessionMockRecorder is the mock recorder for MockSession
type MockSessionMockRecorder struct {
	mock *MockSession
}

// NewMockSession creates a new mock instance
func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &MockSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSession) EXPECT() *MockSessionMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockSession) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockSessionMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSession)(nil).Close))
}

// Closed mocks base method
func (m *MockSession) Closed() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Closed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Closed indicates an expected call of Closed
func (mr *MockSessionMockRecorder) Closed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Closed", reflect.TypeOf((*MockSession)(nil).Closed))
}

// Exec mocks base method
func (m *MockSession) Exec(arg0 string, arg1 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Exec indicates an expected call of Exec
func (mr *MockSessionMockRecorder) Exec(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockSession)(nil).Exec), varargs...)
}

// ExecuteBatch mocks base method
func (m *MockSession) ExecuteBatch(arg0 *gocql.Batch) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteBatch", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ExecuteBatch indicates an expected call of ExecuteBatch
func (mr *MockSessionMockRecorder) ExecuteBatch(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteBatch", reflect.TypeOf((*MockSession)(nil).ExecuteBatch), arg0)
}

// NewBatch mocks base method
func (m *MockSession) NewBatch(arg0 gocql.BatchType) *gocql.Batch {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewBatch", arg0)
	ret0, _ := ret[0].(*gocql.Batch)
	return ret0
}

// NewBatch indicates an expected call of NewBatch
func (mr *MockSessionMockRecorder) NewBatch(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewBatch", reflect.TypeOf((*MockSession)(nil).NewBatch), arg0)
}

// Query mocks base method
func (m *MockSession) Query(arg0 string, arg1 ...interface{}) *gocql.Query {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*gocql.Query)
	return ret0
}

// Query indicates an expected call of Query
func (mr *MockSessionMockRecorder) Query(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockSession)(nil).Query), varargs...)
}

// Select mocks base method
func (m *MockSession) Select(arg0 string, arg1 ...interface{}) ([]map[string]interface{}, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Select", varargs...)
	ret0, _ := ret[0].([]map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Select indicates an expected call of Select
func (mr *MockSessionMockRecorder) Select(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockSession)(nil).Select), varargs...)
}

// SetConsistency mocks base method
func (m *MockSession) SetConsistency(arg0 gocql.Consistency) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetConsistency", arg0)
}

// SetConsistency indicates an expected call of SetConsistency
func (mr *MockSessionMockRecorder) SetConsistency(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConsistency", reflect.TypeOf((*MockSession)(nil).SetConsistency), arg0)
}
