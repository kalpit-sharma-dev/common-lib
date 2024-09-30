package services

import (
	"testing"

	aModel "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/healthCheck"

	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
)

func TestGetHealthCheckService(t *testing.T) {
	srv := HealthCheckServiceFactoryImpl{}.GetHealthCheckService(nil)
	_, ok := srv.(healthCheckServiceImpl)
	if !ok {
		t.Error("healthCheckServiceImpl is not IMPL")
	}
}

func TestGetHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	healthCheck := model.HealthCheck{
		Version: model.Version{
			SolutionName:    "SolutionName",
			ServiceName:     "ServiceName",
			ServiceProvider: "googleLLC",
			Major:           "1",
			Minor:           "1",
			Patch:           "11",
		},
		ListenPort: ":8081",
	}

	listerMock := mock.NewMockHealthCheckDependencies(ctrl)
	dal := mock.NewMockHealthCheckDal(ctrl)
	existing := aModel.HealthCheck{}
	dal.EXPECT().GetHealthCheck(gomock.Any()).Return(existing, nil)
	listerMock.EXPECT().GetHealthCheckDal(gomock.Any()).Return(dal)
	srv := HealthCheckServiceFactoryImpl{}.GetHealthCheckService(listerMock)
	_, err := srv.GetHealthCheck(healthCheck)
	if err != nil {
		t.Errorf("Expected HealthCheck but Got Error %v ", err)
	}
}
