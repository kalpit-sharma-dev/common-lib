// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger (interfaces: Log)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	logger "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// MockLog is a mock of Log interface.
type MockLog struct {
	ctrl     *gomock.Controller
	recorder *MockLogMockRecorder
}

// MockLogMockRecorder is the mock recorder for MockLog.
type MockLogMockRecorder struct {
	mock *MockLog
}

// NewMockLog creates a new mock instance.
func NewMockLog(ctrl *gomock.Controller) *MockLog {
	mock := &MockLog{ctrl: ctrl}
	mock.recorder = &MockLogMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLog) EXPECT() *MockLogMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockLog) Debug(arg0, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLogMockRecorder) Debug(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLog)(nil).Debug), varargs...)
}

// DebugC mocks base method.
func (m *MockLog) DebugC(arg0 context.Context, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "DebugC", varargs...)
}

// DebugC indicates an expected call of DebugC.
func (mr *MockLogMockRecorder) DebugC(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DebugC", reflect.TypeOf((*MockLog)(nil).DebugC), varargs...)
}

// DebugWithLevel mocks base method.
func (m *MockLog) DebugWithLevel(arg0 string, arg1 int, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "DebugWithLevel", varargs...)
}

// DebugWithLevel indicates an expected call of DebugWithLevel.
func (mr *MockLogMockRecorder) DebugWithLevel(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DebugWithLevel", reflect.TypeOf((*MockLog)(nil).DebugWithLevel), varargs...)
}

// Error mocks base method.
func (m *MockLog) Error(arg0, arg1, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLogMockRecorder) Error(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLog)(nil).Error), varargs...)
}

// ErrorC mocks base method.
func (m *MockLog) ErrorC(arg0 context.Context, arg1, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "ErrorC", varargs...)
}

// ErrorC indicates an expected call of ErrorC.
func (mr *MockLogMockRecorder) ErrorC(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ErrorC", reflect.TypeOf((*MockLog)(nil).ErrorC), varargs...)
}

// ErrorWithLevel mocks base method.
func (m *MockLog) ErrorWithLevel(arg0 string, arg1 int, arg2, arg3 string, arg4 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "ErrorWithLevel", varargs...)
}

// ErrorWithLevel indicates an expected call of ErrorWithLevel.
func (mr *MockLogMockRecorder) ErrorWithLevel(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ErrorWithLevel", reflect.TypeOf((*MockLog)(nil).ErrorWithLevel), varargs...)
}

// Fatal mocks base method.
func (m *MockLog) Fatal(arg0, arg1, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fatal", varargs...)
}

// Fatal indicates an expected call of Fatal.
func (mr *MockLogMockRecorder) Fatal(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatal", reflect.TypeOf((*MockLog)(nil).Fatal), varargs...)
}

// FatalC mocks base method.
func (m *MockLog) FatalC(arg0 context.Context, arg1, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "FatalC", varargs...)
}

// FatalC indicates an expected call of FatalC.
func (mr *MockLogMockRecorder) FatalC(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FatalC", reflect.TypeOf((*MockLog)(nil).FatalC), varargs...)
}

// FatalWithLevel mocks base method.
func (m *MockLog) FatalWithLevel(arg0 string, arg1 int, arg2, arg3 string, arg4 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "FatalWithLevel", varargs...)
}

// FatalWithLevel indicates an expected call of FatalWithLevel.
func (mr *MockLogMockRecorder) FatalWithLevel(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FatalWithLevel", reflect.TypeOf((*MockLog)(nil).FatalWithLevel), varargs...)
}

// GetWriter mocks base method.
func (m *MockLog) GetWriter() io.WriteCloser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWriter")
	ret0, _ := ret[0].(io.WriteCloser)
	return ret0
}

// GetWriter indicates an expected call of GetWriter.
func (mr *MockLogMockRecorder) GetWriter() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWriter", reflect.TypeOf((*MockLog)(nil).GetWriter))
}

// Info mocks base method.
func (m *MockLog) Info(arg0, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLogMockRecorder) Info(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLog)(nil).Info), varargs...)
}

// InfoC mocks base method.
func (m *MockLog) InfoC(arg0 context.Context, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "InfoC", varargs...)
}

// InfoC indicates an expected call of InfoC.
func (mr *MockLogMockRecorder) InfoC(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfoC", reflect.TypeOf((*MockLog)(nil).InfoC), varargs...)
}

// InfoWithLevel mocks base method.
func (m *MockLog) InfoWithLevel(arg0 string, arg1 int, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "InfoWithLevel", varargs...)
}

// InfoWithLevel indicates an expected call of InfoWithLevel.
func (mr *MockLogMockRecorder) InfoWithLevel(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfoWithLevel", reflect.TypeOf((*MockLog)(nil).InfoWithLevel), varargs...)
}

// LogEvent mocks base method.
func (m *MockLog) LogEvent(arg0 context.Context, arg1 logger.Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogEvent", arg0, arg1)
}

// LogEvent indicates an expected call of LogEvent.
func (mr *MockLogMockRecorder) LogEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogEvent", reflect.TypeOf((*MockLog)(nil).LogEvent), arg0, arg1)
}

// LogLevel mocks base method.
func (m *MockLog) LogLevel() logger.LogLevel {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogLevel")
	ret0, _ := ret[0].(logger.LogLevel)
	return ret0
}

// LogLevel indicates an expected call of LogLevel.
func (mr *MockLogMockRecorder) LogLevel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogLevel", reflect.TypeOf((*MockLog)(nil).LogLevel))
}

// SetWriter mocks base method.
func (m *MockLog) SetWriter(arg0 io.Writer) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetWriter", arg0)
}

// SetWriter indicates an expected call of SetWriter.
func (mr *MockLogMockRecorder) SetWriter(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWriter", reflect.TypeOf((*MockLog)(nil).SetWriter), arg0)
}

// Sync mocks base method.
func (m *MockLog) Sync() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync")
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockLogMockRecorder) Sync() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockLog)(nil).Sync))
}

// Trace mocks base method.
func (m *MockLog) Trace(arg0, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Trace", varargs...)
}

// Trace indicates an expected call of Trace.
func (mr *MockLogMockRecorder) Trace(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trace", reflect.TypeOf((*MockLog)(nil).Trace), varargs...)
}

// TraceC mocks base method.
func (m *MockLog) TraceC(arg0 context.Context, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "TraceC", varargs...)
}

// TraceC indicates an expected call of TraceC.
func (mr *MockLogMockRecorder) TraceC(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TraceC", reflect.TypeOf((*MockLog)(nil).TraceC), varargs...)
}

// TraceWithLevel mocks base method.
func (m *MockLog) TraceWithLevel(arg0 string, arg1 int, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "TraceWithLevel", varargs...)
}

// TraceWithLevel indicates an expected call of TraceWithLevel.
func (mr *MockLogMockRecorder) TraceWithLevel(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TraceWithLevel", reflect.TypeOf((*MockLog)(nil).TraceWithLevel), varargs...)
}

// Warn mocks base method.
func (m *MockLog) Warn(arg0, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLogMockRecorder) Warn(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLog)(nil).Warn), varargs...)
}

// WarnC mocks base method.
func (m *MockLog) WarnC(arg0 context.Context, arg1 string, arg2 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "WarnC", varargs...)
}

// WarnC indicates an expected call of WarnC.
func (mr *MockLogMockRecorder) WarnC(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WarnC", reflect.TypeOf((*MockLog)(nil).WarnC), varargs...)
}

// WarnWithLevel mocks base method.
func (m *MockLog) WarnWithLevel(arg0 string, arg1 int, arg2 string, arg3 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "WarnWithLevel", varargs...)
}

// WarnWithLevel indicates an expected call of WarnWithLevel.
func (mr *MockLogMockRecorder) WarnWithLevel(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WarnWithLevel", reflect.TypeOf((*MockLog)(nil).WarnWithLevel), varargs...)
}

// With mocks base method.
func (m *MockLog) With(arg0 ...logger.Option) logger.Log {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(logger.Log)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockLogMockRecorder) With(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockLog)(nil).With), arg0...)
}
