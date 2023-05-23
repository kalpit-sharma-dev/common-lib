package rest

import (
	"fmt"
	"net/http"
	"strings"
)

const defaultPathSuffix = "/"

var allowedFiles = map[string]bool{"swagger.yaml": true, "readme.md": true}

// FileHandler used for returning static files
func FileHandler(dir string, path string) http.Handler {
	return &fileHandler{staticPath: dir, uriPath: path}
}

// FilePath used for returning static files path
func FilePath(pathPrefix string) string {
	p := defaultPathSuffix
	if pathPrefix != "" {
		p = fmt.Sprintf("%s%s", pathPrefix, defaultPathSuffix)
	}
	return p
}

// fileHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory
// to serve the file form given static directory
type fileHandler struct {
	staticPath string
	uriPath    string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimPrefix(r.URL.Path, h.uriPath)
	ok := allowedFiles[file]
	if !ok {
		http.Error(w, fmt.Sprintf("Permission deinied to access file :: %s", file), http.StatusUnauthorized)
		return
	}

	handler := http.StripPrefix(h.uriPath, http.FileServer(http.Dir(h.staticPath)))
	handler.ServeHTTP(w, r)
}
