// Package microservice implements the some common resources.
//
// Deprecated: microservice is old implementation of common-resources and should not be used
// except for compatibility with legacy systems.
//
// Use src/web/rest for all common handlers
// This package is frozen and no new functionality will be added.
package microservice

import (
	"encoding/json"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
	cweb "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

type healthCheckResource struct {
	cweb.Post405
	cweb.Put405
	cweb.Delete405
	cweb.Others405
	healthCheck model.HealthCheck
	f           model.HealthCheckDependencies
}

func (res healthCheckResource) Get(rc cweb.RequestContext) {
	logger.Get().Debug("", res.healthCheck.Version.ServiceName, "Get HealthCheck response for HealthCheck %v and request %v", res.healthCheck, rc)
	w := rc.GetResponse()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	service := res.f.GetHealthCheckService(res.f)
	response, err := service.GetHealthCheck(res.healthCheck)
	if err != nil {
		logger.Get().Error(res.healthCheck.Version.ServiceName, "healthCheckResource:Get", "%v Error while getting HealthCheck response for %v", err, res.healthCheck)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Get().Error(res.healthCheck.Version.ServiceName, "healthCheckResource:Get", "%v Error while Encoding HealthCheck response %v", err, response)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// CreateHealthCheckRouteConfig is a function for creating route config for healthCheck url
func CreateHealthCheckRouteConfig(healthCheck model.HealthCheck, f model.HealthCheckDependencies) *cweb.RouteConfig {
	return &cweb.RouteConfig{
		URLPathSuffix: "/healthCheck",
		URLVars:       []string{},
		Res: healthCheckResource{
			healthCheck: healthCheck,
			f:           f,
		},
	}
}
