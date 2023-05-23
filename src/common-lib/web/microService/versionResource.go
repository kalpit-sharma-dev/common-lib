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

type versionResource struct {
	cweb.Post405
	cweb.Put405
	cweb.Delete405
	cweb.Others405
	version model.Version
	f       model.VersionDependencies
}

func (res versionResource) Get(rc cweb.RequestContext) {
	logger.Get().Debug(res.version.ServiceName, "versionResource:Get", "Get Version response for version %v and request %v", res.version, rc)
	w := rc.GetResponse()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	service := res.f.GetVersionService()
	response := service.GetVersion(res.version)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Get().Error(res.version.ServiceName, "versionResource:Get:Encode", "%v Error while Encoding Version response %v", err, response)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// CreateVersionRouteConfig is a function for creating route config for Version url
func CreateVersionRouteConfig(version model.Version, f model.VersionDependencies) *cweb.RouteConfig {
	return &cweb.RouteConfig{
		URLPathSuffix: "/version",
		URLVars:       []string{},
		Res: versionResource{
			version: version,
			f:       f,
		},
	}
}
