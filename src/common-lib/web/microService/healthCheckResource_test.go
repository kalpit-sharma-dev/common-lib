// Package microservice implements the some common resources.
//
// Deprecated: microservice is old implementation of common-resources and should not be used
// except for compatibility with legacy systems.
//
// Use src/web/rest for all common handlers
// This package is frozen and no new functionality will be added.
package microservice

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"reflect"

	"github.com/golang/mock/gomock"
	aModel "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/healthCheck"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
	cwmock "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/mock"
)

func TestHealtCheckGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	listerMock := mock.NewMockHealthCheckDependencies(ctrl)
	service := mock.NewMockHealthCheckService(ctrl)
	listerMock.EXPECT().GetHealthCheckService(gomock.Any()).Return(service)
	service.EXPECT().GetHealthCheck(gomock.Any()).Return(aModel.HealthCheck{}, nil)
	mockRsp := httptest.NewRecorder()
	mockRc := cwmock.NewMockRequestContext(ctrl)
	mockRc.EXPECT().GetResponse().Return(mockRsp).AnyTimes()

	healthCheckResource{
		f: listerMock,
		healthCheck: model.HealthCheck{
			Version: model.Version{
				SolutionName:    "SolutionName",
				ServiceName:     "ServiceName",
				ServiceProvider: "googleLLC",
				Major:           "1",
				Minor:           "1",
				Patch:           "11",
			},
			ListenPort: ":8081",
		},
	}.Get(mockRc)

	if mockRsp.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %v", mockRsp.Code)
	}
}

func TestHealtCheckGetErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	listerMock := mock.NewMockHealthCheckDependencies(ctrl)
	service := mock.NewMockHealthCheckService(ctrl)
	listerMock.EXPECT().GetHealthCheckService(gomock.Any()).Return(service)
	service.EXPECT().GetHealthCheck(gomock.Any()).Return(aModel.HealthCheck{}, errors.New("Error"))
	mockRsp := httptest.NewRecorder()
	mockRc := cwmock.NewMockRequestContext(ctrl)
	mockRc.EXPECT().GetResponse().Return(mockRsp).AnyTimes()

	healthCheckResource{
		f: listerMock,
		healthCheck: model.HealthCheck{
			Version: model.Version{
				SolutionName:    "SolutionName",
				ServiceName:     "ServiceName",
				ServiceProvider: "googleLLC",
				Major:           "1",
				Minor:           "1",
				Patch:           "11",
			},
			ListenPort: ":8081",
		},
	}.Get(mockRc)

	if mockRsp.Code == http.StatusOK {
		t.Errorf("Unexpected error code : %v", mockRsp.Code)
	}
}

func TestCreateHealthCheckRouteConfig(t *testing.T) {
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	listerMock := mock.NewMockHealthCheckDependencies(ctrl)
	route := web.RouteConfig{
		URLPathSuffix: "/healthCheck",
		URLVars:       []string{},
		Res: healthCheckResource{
			healthCheck: healthCheck,
			f:           listerMock,
		},
	}

	rout := CreateHealthCheckRouteConfig(healthCheck, listerMock)

	if reflect.DeepEqual(route, rout) {
		t.Errorf("Expected same but got Different %v : %v", route, rout)
	}
}
