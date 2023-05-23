package http

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

// ErrBlankBody - Error if http response does not have a body
var ErrBlankBody = errors.New("Body does not exist")

// Response - http response wrapper
type Response struct {
	TransactionID string
	HTTPResponse  *http.Response
	Err           error
}

// IsProxyError - Identify waither response is a proxy error or not
func (r *Response) IsProxyError() bool {
	return IsProxyError(r.Err, r.HTTPResponse)
}

// IsSuccess - Identify waither we received a success response or not
func (r *Response) IsSuccess() bool {
	return IsSuccess(r.HTTPResponse)
}

// HasBody - Identify waither response has body or not
func (r *Response) HasBody() bool {
	return HasBody(r.HTTPResponse)
}

// Ignore - Ignore response
func (r *Response) Ignore() error {
	if r.HasBody() {
		defer r.HTTPResponse.Body.Close() //nolint
		_, err := ioutil.ReadAll(r.HTTPResponse.Body)
		return err
	}
	return nil
}

// Status - HTTP Response status
func (r *Response) Status() int {
	if r != nil && r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetBytes - return Response body as byte array
func (r *Response) GetBytes() ([]byte, error) {
	if r.HasBody() {
		defer r.HTTPResponse.Body.Close() //nolint
		return ioutil.ReadAll(r.HTTPResponse.Body)
	}
	return nil, ErrBlankBody
}

// GetInterface - return Response body as object
func (r *Response) GetInterface(data interface{}) error {
	if r.HasBody() {
		defer r.HTTPResponse.Body.Close() //nolint
		result, err := ioutil.ReadAll(r.HTTPResponse.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(result, data)
	}
	return ErrBlankBody
}

// GetReader - return Response body as an io.reader
// User need to close the read closer after reading body
func (r *Response) GetReader() (io.ReadCloser, error) {
	if r.HasBody() {
		return r.HTTPResponse.Body, nil
	}
	return nil, ErrBlankBody
}
