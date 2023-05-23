// Code generated by MockGen. DO NOT EDIT.
// Source: producer.go

// Package mock_producer is a generated GoMock package.
package mock_producer

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	producer "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/producer"
)

// MockProducer is a mock of Producer interface.
type MockProducer struct {
	ctrl     *gomock.Controller
	recorder *MockProducerMockRecorder
}

// MockProducerMockRecorder is the mock recorder for MockProducer.
type MockProducerMockRecorder struct {
	mock *MockProducer
}

// NewMockProducer creates a new mock instance.
func NewMockProducer(ctrl *gomock.Controller) *MockProducer {
	mock := &MockProducer{ctrl: ctrl}
	mock.recorder = &MockProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProducer) EXPECT() *MockProducerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockProducer) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockProducerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockProducer)(nil).Close))
}

// Health mocks base method.
func (m *MockProducer) Health() (*producer.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Health")
	ret0, _ := ret[0].(*producer.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Health indicates an expected call of Health.
func (mr *MockProducerMockRecorder) Health() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Health", reflect.TypeOf((*MockProducer)(nil).Health))
}

// Produce mocks base method.
func (m *MockProducer) Produce(ctx context.Context, transaction string, messages ...*producer.Message) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, transaction}
	for _, a := range messages {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Produce", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Produce indicates an expected call of Produce.
func (mr *MockProducerMockRecorder) Produce(ctx, transaction interface{}, messages ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, transaction}, messages...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Produce", reflect.TypeOf((*MockProducer)(nil).Produce), varargs...)
}

// ProduceWithReport mocks base method.
func (m *MockProducer) ProduceWithReport(ctx context.Context, transaction string, messages ...*producer.Message) ([]*producer.DeliveryReport, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, transaction}
	for _, a := range messages {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ProduceWithReport", varargs...)
	ret0, _ := ret[0].([]*producer.DeliveryReport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProduceWithReport indicates an expected call of ProduceWithReport.
func (mr *MockProducerMockRecorder) ProduceWithReport(ctx, transaction interface{}, messages ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, transaction}, messages...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProduceWithReport", reflect.TypeOf((*MockProducer)(nil).ProduceWithReport), varargs...)
}

// MockStatusableProducer is a mock of StatusableProducer interface.
type MockStatusableProducer struct {
	ctrl     *gomock.Controller
	recorder *MockStatusableProducerMockRecorder
}

// MockStatusableProducerMockRecorder is the mock recorder for MockStatusableProducer.
type MockStatusableProducerMockRecorder struct {
	mock *MockStatusableProducer
}

// NewMockStatusableProducer creates a new mock instance.
func NewMockStatusableProducer(ctrl *gomock.Controller) *MockStatusableProducer {
	mock := &MockStatusableProducer{ctrl: ctrl}
	mock.recorder = &MockStatusableProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatusableProducer) EXPECT() *MockStatusableProducerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStatusableProducer) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockStatusableProducerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStatusableProducer)(nil).Close))
}

// GetFatalError mocks base method.
func (m *MockStatusableProducer) GetFatalError() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFatalError")
	ret0, _ := ret[0].(error)
	return ret0
}

// GetFatalError indicates an expected call of GetFatalError.
func (mr *MockStatusableProducerMockRecorder) GetFatalError() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFatalError", reflect.TypeOf((*MockStatusableProducer)(nil).GetFatalError))
}

// Health mocks base method.
func (m *MockStatusableProducer) Health() (*producer.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Health")
	ret0, _ := ret[0].(*producer.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Health indicates an expected call of Health.
func (mr *MockStatusableProducerMockRecorder) Health() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Health", reflect.TypeOf((*MockStatusableProducer)(nil).Health))
}

// Produce mocks base method.
func (m *MockStatusableProducer) Produce(ctx context.Context, transaction string, messages ...*producer.Message) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, transaction}
	for _, a := range messages {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Produce", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Produce indicates an expected call of Produce.
func (mr *MockStatusableProducerMockRecorder) Produce(ctx, transaction interface{}, messages ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, transaction}, messages...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Produce", reflect.TypeOf((*MockStatusableProducer)(nil).Produce), varargs...)
}

// ProduceWithReport mocks base method.
func (m *MockStatusableProducer) ProduceWithReport(ctx context.Context, transaction string, messages ...*producer.Message) ([]*producer.DeliveryReport, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, transaction}
	for _, a := range messages {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ProduceWithReport", varargs...)
	ret0, _ := ret[0].([]*producer.DeliveryReport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProduceWithReport indicates an expected call of ProduceWithReport.
func (mr *MockStatusableProducerMockRecorder) ProduceWithReport(ctx, transaction interface{}, messages ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, transaction}, messages...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProduceWithReport", reflect.TypeOf((*MockStatusableProducer)(nil).ProduceWithReport), varargs...)
}

// MockAsyncProducer is a mock of AsyncProducer interface.
type MockAsyncProducer struct {
	ctrl     *gomock.Controller
	recorder *MockAsyncProducerMockRecorder
}

// MockAsyncProducerMockRecorder is the mock recorder for MockAsyncProducer.
type MockAsyncProducerMockRecorder struct {
	mock *MockAsyncProducer
}

// NewMockAsyncProducer creates a new mock instance.
func NewMockAsyncProducer(ctrl *gomock.Controller) *MockAsyncProducer {
	mock := &MockAsyncProducer{ctrl: ctrl}
	mock.recorder = &MockAsyncProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAsyncProducer) EXPECT() *MockAsyncProducerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockAsyncProducer) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockAsyncProducerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockAsyncProducer)(nil).Close))
}

// DeliveryReportChannel mocks base method.
func (m *MockAsyncProducer) DeliveryReportChannel() <-chan *producer.DeliveryReport {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeliveryReportChannel")
	ret0, _ := ret[0].(<-chan *producer.DeliveryReport)
	return ret0
}

// DeliveryReportChannel indicates an expected call of DeliveryReportChannel.
func (mr *MockAsyncProducerMockRecorder) DeliveryReportChannel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeliveryReportChannel", reflect.TypeOf((*MockAsyncProducer)(nil).DeliveryReportChannel))
}

// Flush mocks base method.
func (m *MockAsyncProducer) Flush(timeoutMs int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", timeoutMs)
	ret0, _ := ret[0].(int)
	return ret0
}

// Flush indicates an expected call of Flush.
func (mr *MockAsyncProducerMockRecorder) Flush(timeoutMs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockAsyncProducer)(nil).Flush), timeoutMs)
}

// Health mocks base method.
func (m *MockAsyncProducer) Health() (*producer.Health, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Health")
	ret0, _ := ret[0].(*producer.Health)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Health indicates an expected call of Health.
func (mr *MockAsyncProducerMockRecorder) Health() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Health", reflect.TypeOf((*MockAsyncProducer)(nil).Health))
}

// ProduceChannel mocks base method.
func (m *MockAsyncProducer) ProduceChannel() chan<- *producer.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProduceChannel")
	ret0, _ := ret[0].(chan<- *producer.Message)
	return ret0
}

// ProduceChannel indicates an expected call of ProduceChannel.
func (mr *MockAsyncProducerMockRecorder) ProduceChannel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProduceChannel", reflect.TypeOf((*MockAsyncProducer)(nil).ProduceChannel))
}