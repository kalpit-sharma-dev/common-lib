// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entitlement "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/entitlement"
)

// MockEntitlementService is a mock of EntitlementService interface.
type MockEntitlementService struct {
	ctrl     *gomock.Controller
	recorder *MockEntitlementServiceMockRecorder
}

// MockEntitlementServiceMockRecorder is the mock recorder for MockEntitlementService.
type MockEntitlementServiceMockRecorder struct {
	mock *MockEntitlementService
}

// NewMockEntitlementService creates a new mock instance.
func NewMockEntitlementService(ctrl *gomock.Controller) *MockEntitlementService {
	mock := &MockEntitlementService{ctrl: ctrl}
	mock.recorder = &MockEntitlementServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEntitlementService) EXPECT() *MockEntitlementServiceMockRecorder {
	return m.recorder
}

// GetPartnerFeatureNames mocks base method.
func (m *MockEntitlementService) GetPartnerFeatureNames(partnerID string) (map[string]bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPartnerFeatureNames", partnerID)
	ret0, _ := ret[0].(map[string]bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPartnerFeatureNames indicates an expected call of GetPartnerFeatureNames.
func (mr *MockEntitlementServiceMockRecorder) GetPartnerFeatureNames(partnerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPartnerFeatureNames", reflect.TypeOf((*MockEntitlementService)(nil).GetPartnerFeatureNames), partnerID)
}

// GetPartnerFeatures mocks base method.
func (m *MockEntitlementService) GetPartnerFeatures(partnerID string) ([]entitlement.Feature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPartnerFeatures", partnerID)
	ret0, _ := ret[0].([]entitlement.Feature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPartnerFeatures indicates an expected call of GetPartnerFeatures.
func (mr *MockEntitlementServiceMockRecorder) GetPartnerFeatures(partnerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPartnerFeatures", reflect.TypeOf((*MockEntitlementService)(nil).GetPartnerFeatures), partnerID)
}

// IsPartnerAuthorized mocks base method.
func (m *MockEntitlementService) IsPartnerAuthorized(partnerID, featureName string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPartnerAuthorized", partnerID, featureName)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsPartnerAuthorized indicates an expected call of IsPartnerAuthorized.
func (mr *MockEntitlementServiceMockRecorder) IsPartnerAuthorized(partnerID, featureName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPartnerAuthorized", reflect.TypeOf((*MockEntitlementService)(nil).IsPartnerAuthorized), partnerID, featureName)
}
