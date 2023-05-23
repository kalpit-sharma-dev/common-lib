package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

// Request - A http request builder
type Request struct {
	transactionID string
	method        string
	url           string
	body          io.Reader
	header        http.Header
}

// CreateRequest - Create a new request builder instance and return
func CreateRequest(transactionID, method, url string) *Request {
	r := &Request{
		transactionID: transactionID,
		method:        method,
		url:           url,
		header:        http.Header{},
	}

	r.AddHeader(constants.TransactionID, transactionID)
	r.AddHeader(constants.XRequestID, transactionID)
	r.AddHeader(constants.ServiceName, utils.GetServiceName())

	return r
}

// SetBytes - Set request body using byte array
func (r *Request) SetBytes(body []byte) {
	r.body = bytes.NewReader(body)
}

// SetInterface - Set request body by marshling object
func (r *Request) SetInterface(body interface{}) error {
	if body == nil {
		return ErrBlankBody
	}

	b, ok := body.([]byte)

	if ok {
		r.SetBytes(b)
		return nil
	}

	reader, ok := body.(io.Reader)

	if ok {
		r.SetReader(reader)
		return nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	r.body = bytes.NewReader(data)
	return nil
}

// SetReader - Set request body using reader
func (r *Request) SetReader(body io.Reader) {
	r.body = body
}

// AddHeader - Add header value in the request, if exists, it adds another header with duplicate key
func (r *Request) AddHeader(key string, value string) {
	r.header.Add(key, value)
}

// SetHeader - Replaces the header key with the new value
func (r *Request) SetHeader(key string, value string) {
	r.header.Set(key, value)
}

// Build - Convert request object into HTTP Request
func (r *Request) Build() (*http.Request, error) {
	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		return req, err
	}

	req.Header = r.header
	return req, nil
}

// Execute - Execute request using client and return a response object
func (r *Request) Execute(cfg *client.Config, ignoreProxy bool) *Response {
	clt := client.TLS(cfg, ignoreProxy)
	req, err := r.Build()
	if err != nil {
		return &Response{TransactionID: r.transactionID, Err: err}
	}
	res, err := clt.Do(req) //nolint:bodyclose
	return &Response{TransactionID: r.transactionID, HTTPResponse: res, Err: err}
}
