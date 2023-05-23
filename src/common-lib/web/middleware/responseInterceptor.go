package middleware

import (
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

// ResponseInterceptor : A wrapper on http response writer that
// intercepts the http response and transforms it.
type ResponseInterceptor struct {
	Request            http.Request
	BaseResponseWriter http.ResponseWriter
	TransformBody      func(response []byte, request http.Request, status int) []byte
	TransformHeader    func(w http.ResponseWriter, request http.Request, status int)
	status             int
}

func (ri *ResponseInterceptor) Write(b []byte) (int, error) {
	if ri.TransformBody != nil {
		//Transform the intercepted response body
		b = ri.TransformBody(b, ri.Request, ri.status)
	}

	return ri.BaseResponseWriter.Write(b)
}

func (ri *ResponseInterceptor) Header() http.Header {
	return ri.BaseResponseWriter.Header()
}

func (ri *ResponseInterceptor) WriteHeader(statusCode int) {
	ri.status = statusCode
	//Transform the intercepted response header
	if ri.TransformHeader != nil {
		ri.TransformHeader(ri.BaseResponseWriter, ri.Request, ri.status)
	}
	ri.BaseResponseWriter.WriteHeader(statusCode)
}

// Intercept : Intercepts and passes a custom response writer to the handler downstream
func Intercept(next func(w http.ResponseWriter, r *http.Request),
	transformResponse func([]byte, http.Request, int) []byte,
	transformHeader func(http.ResponseWriter, http.Request, int)) web.HTTPHandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		//a custom response writer with specific error handling
		ri := &ResponseInterceptor{
			Request:            *req,
			BaseResponseWriter: w,
			TransformBody:      transformResponse,
			TransformHeader:    transformHeader,
		}

		//call downstream with custom response writer
		next(ri, req)
	}
}
