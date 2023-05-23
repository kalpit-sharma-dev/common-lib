// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/pluginUtils (interfaces: PluginIOReader,PluginIOWriter)

// Package mock is a generated GoMock package.
package mock

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPluginIOReader is a mock of PluginIOReader interface.
type MockPluginIOReader struct {
	ctrl     *gomock.Controller
	recorder *MockPluginIOReaderMockRecorder
}

// MockPluginIOReaderMockRecorder is the mock recorder for MockPluginIOReader.
type MockPluginIOReaderMockRecorder struct {
	mock *MockPluginIOReader
}

// NewMockPluginIOReader creates a new mock instance.
func NewMockPluginIOReader(ctrl *gomock.Controller) *MockPluginIOReader {
	mock := &MockPluginIOReader{ctrl: ctrl}
	mock.recorder = &MockPluginIOReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPluginIOReader) EXPECT() *MockPluginIOReaderMockRecorder {
	return m.recorder
}

// GetReader mocks base method.
func (m *MockPluginIOReader) GetReader() io.Reader {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReader")
	ret0, _ := ret[0].(io.Reader)
	return ret0
}

// GetReader indicates an expected call of GetReader.
func (mr *MockPluginIOReaderMockRecorder) GetReader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReader", reflect.TypeOf((*MockPluginIOReader)(nil).GetReader))
}

// MockPluginIOWriter is a mock of PluginIOWriter interface.
type MockPluginIOWriter struct {
	ctrl     *gomock.Controller
	recorder *MockPluginIOWriterMockRecorder
}

// MockPluginIOWriterMockRecorder is the mock recorder for MockPluginIOWriter.
type MockPluginIOWriterMockRecorder struct {
	mock *MockPluginIOWriter
}

// NewMockPluginIOWriter creates a new mock instance.
func NewMockPluginIOWriter(ctrl *gomock.Controller) *MockPluginIOWriter {
	mock := &MockPluginIOWriter{ctrl: ctrl}
	mock.recorder = &MockPluginIOWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPluginIOWriter) EXPECT() *MockPluginIOWriterMockRecorder {
	return m.recorder
}

// GetWriter mocks base method.
func (m *MockPluginIOWriter) GetWriter() io.Writer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWriter")
	ret0, _ := ret[0].(io.Writer)
	return ret0
}

// GetWriter indicates an expected call of GetWriter.
func (mr *MockPluginIOWriterMockRecorder) GetWriter() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWriter", reflect.TypeOf((*MockPluginIOWriter)(nil).GetWriter))
}