package http

import (
	"net/http"
	"strings"
)

const (
	proxyErrorString = "proxyconnect"
)

var (
	// successStatus - List of status to be considered as Success
	successStatus = map[int]bool{
		http.StatusOK:        true,
		http.StatusCreated:   true,
		http.StatusNoContent: true,
	}

	// proxyErrorStatus - List of status to be considered as proxy Failure
	proxyErrorStatus = map[int]bool{
		http.StatusUseProxy:          true,
		http.StatusUnauthorized:      true,
		http.StatusProxyAuthRequired: true,
		http.StatusGatewayTimeout:    true,
		http.StatusForbidden:         true,
	}
)

// IsProxyError - Identify waither response is a proxy error or not
func IsProxyError(err error, res *http.Response) bool {
	return (err != nil && strings.Contains(err.Error(), proxyErrorString)) ||
		(res != nil && proxyErrorStatus[res.StatusCode])
}

// IsSuccess - Identify waither we received a success response or not
func IsSuccess(res *http.Response) bool {
	return (res != nil && successStatus[res.StatusCode])
}

// HasBody - Identify waither response has body or not
func HasBody(res *http.Response) bool {
	return (res != nil && res.Body != nil)
}
