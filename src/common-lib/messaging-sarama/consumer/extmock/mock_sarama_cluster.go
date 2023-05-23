// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bsm/sarama-cluster (interfaces: PartitionConsumer)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	sarama "github.com/Shopify/sarama"
	gomock "github.com/golang/mock/gomock"
)

// MockPartitionConsumer is a mock of PartitionConsumer interface.
type MockPartitionConsumer struct {
	ctrl     *gomock.Controller
	recorder *MockPartitionConsumerMockRecorder
}

// MockPartitionConsumerMockRecorder is the mock recorder for MockPartitionConsumer.
type MockPartitionConsumerMockRecorder struct {
	mock *MockPartitionConsumer
}

// NewMockPartitionConsumer creates a new mock instance.
func NewMockPartitionConsumer(ctrl *gomock.Controller) *MockPartitionConsumer {
	mock := &MockPartitionConsumer{ctrl: ctrl}
	mock.recorder = &MockPartitionConsumerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPartitionConsumer) EXPECT() *MockPartitionConsumerMockRecorder {
	return m.recorder
}

// AsyncClose mocks base method.
func (m *MockPartitionConsumer) AsyncClose() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AsyncClose")
}

// AsyncClose indicates an expected call of AsyncClose.
func (mr *MockPartitionConsumerMockRecorder) AsyncClose() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AsyncClose", reflect.TypeOf((*MockPartitionConsumer)(nil).AsyncClose))
}

// Close mocks base method.
func (m *MockPartitionConsumer) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPartitionConsumerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPartitionConsumer)(nil).Close))
}

// Errors mocks base method.
func (m *MockPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Errors")
	ret0, _ := ret[0].(<-chan *sarama.ConsumerError)
	return ret0
}

// Errors indicates an expected call of Errors.
func (mr *MockPartitionConsumerMockRecorder) Errors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errors", reflect.TypeOf((*MockPartitionConsumer)(nil).Errors))
}

// HighWaterMarkOffset mocks base method.
func (m *MockPartitionConsumer) HighWaterMarkOffset() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HighWaterMarkOffset")
	ret0, _ := ret[0].(int64)
	return ret0
}

// HighWaterMarkOffset indicates an expected call of HighWaterMarkOffset.
func (mr *MockPartitionConsumerMockRecorder) HighWaterMarkOffset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HighWaterMarkOffset", reflect.TypeOf((*MockPartitionConsumer)(nil).HighWaterMarkOffset))
}

// InitialOffset mocks base method.
func (m *MockPartitionConsumer) InitialOffset() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitialOffset")
	ret0, _ := ret[0].(int64)
	return ret0
}

// InitialOffset indicates an expected call of InitialOffset.
func (mr *MockPartitionConsumerMockRecorder) InitialOffset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitialOffset", reflect.TypeOf((*MockPartitionConsumer)(nil).InitialOffset))
}

// IsPaused mocks base method.
func (m *MockPartitionConsumer) IsPaused() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPaused")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsPaused indicates an expected call of IsPaused.
func (mr *MockPartitionConsumerMockRecorder) IsPaused() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPaused", reflect.TypeOf((*MockPartitionConsumer)(nil).IsPaused))
}

// MarkOffset mocks base method.
func (m *MockPartitionConsumer) MarkOffset(arg0 int64, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "MarkOffset", arg0, arg1)
}

// MarkOffset indicates an expected call of MarkOffset.
func (mr *MockPartitionConsumerMockRecorder) MarkOffset(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkOffset", reflect.TypeOf((*MockPartitionConsumer)(nil).MarkOffset), arg0, arg1)
}

// Messages mocks base method.
func (m *MockPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Messages")
	ret0, _ := ret[0].(<-chan *sarama.ConsumerMessage)
	return ret0
}

// Messages indicates an expected call of Messages.
func (mr *MockPartitionConsumerMockRecorder) Messages() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Messages", reflect.TypeOf((*MockPartitionConsumer)(nil).Messages))
}

// Partition mocks base method.
func (m *MockPartitionConsumer) Partition() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Partition")
	ret0, _ := ret[0].(int32)
	return ret0
}

// Partition indicates an expected call of Partition.
func (mr *MockPartitionConsumerMockRecorder) Partition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Partition", reflect.TypeOf((*MockPartitionConsumer)(nil).Partition))
}

// Pause mocks base method.
func (m *MockPartitionConsumer) Pause() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Pause")
}

// Pause indicates an expected call of Pause.
func (mr *MockPartitionConsumerMockRecorder) Pause() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockPartitionConsumer)(nil).Pause))
}

// ResetOffset mocks base method.
func (m *MockPartitionConsumer) ResetOffset(arg0 int64, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ResetOffset", arg0, arg1)
}

// ResetOffset indicates an expected call of ResetOffset.
func (mr *MockPartitionConsumerMockRecorder) ResetOffset(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetOffset", reflect.TypeOf((*MockPartitionConsumer)(nil).ResetOffset), arg0, arg1)
}

// Resume mocks base method.
func (m *MockPartitionConsumer) Resume() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Resume")
}

// Resume indicates an expected call of Resume.
func (mr *MockPartitionConsumerMockRecorder) Resume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resume", reflect.TypeOf((*MockPartitionConsumer)(nil).Resume))
}

// Topic mocks base method.
func (m *MockPartitionConsumer) Topic() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Topic")
	ret0, _ := ret[0].(string)
	return ret0
}

// Topic indicates an expected call of Topic.
func (mr *MockPartitionConsumerMockRecorder) Topic() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Topic", reflect.TypeOf((*MockPartitionConsumer)(nil).Topic))
}