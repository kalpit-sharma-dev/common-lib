// Package microservice implements the some common resources.
//
// Deprecated: microservice is old implementation of common-resources and should not be used
// except for compatibility with legacy systems.
//
// Use src/web/rest for all common handlers
// This package is frozen and no new functionality will be added.
package microservice

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
	cweb "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

type downloadResource struct {
	cweb.Post405
	cweb.Put405
	cweb.Delete405
	cweb.Others405
	fileInfo model.DownloadFileInfo
}

func (res downloadResource) Get(rc cweb.RequestContext) {
	logger.Get().Debug(res.fileInfo.Name, "Get File response for Info %v and request %v", res.fileInfo, rc)
	w := rc.GetResponse()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Disposition", "attachment; filename="+res.fileInfo.Name)
	w.Header().Set("Content-Type", res.fileInfo.ContentType)
	f, err := os.Open(res.fileInfo.Path)
	if err != nil {
		logger.Get().Error(res.fileInfo.Name, "downloadResource:Get:Open", "%v Error while reading File at %s Path", err, res.fileInfo.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	count, err := io.Copy(w, f)
	if err != nil {
		logger.Get().Error(res.fileInfo.Name, "downloadResource:Get:Copy", "%v Error while writing File at %s Path", err, res.fileInfo.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(count, 10))
}

// CreateDownloadRouteConfig is a function for creating route config for Download url
func CreateDownloadRouteConfig(fileInfo model.DownloadFileInfo, pathSuffix string) *cweb.RouteConfig {
	return &cweb.RouteConfig{
		URLPathSuffix: pathSuffix,
		URLVars:       []string{},
		Res: downloadResource{
			fileInfo: fileInfo,
		},
	}
}
