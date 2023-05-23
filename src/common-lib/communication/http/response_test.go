package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type fakeReadCloser struct{}

// here's a fake ReadFile method that matches the signature of ioutil.ReadFile
func (f fakeReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("Error")
}

func (f fakeReadCloser) Close() error {
	return errors.New("Error")
}

func TestResponse_IsProxyError(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "both nil", want: false},

		// Via Err
		{name: "1 proxyconnect Error", want: true, fields: fields{Err: errors.New("proxyconnect")}},
		{name: "2 proxy1 Error", want: false, fields: fields{Err: errors.New("proxy1")}},
		{name: "3 proxyconnect Error", want: true, fields: fields{Err: errors.New("error proxyconnect Err")}},

		// Via Response
		{name: "4 StatusUseProxy", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "5 StatusUnauthorized", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "6 StatusProxyAuthRequired", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "7 StatusGatewayTimeout", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusGatewayTimeout}}},
		{name: "8 StatusForbidden", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusForbidden}}},
		{name: "9 StatusOK", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "10 StatusCreated", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusCreated}}},
		{name: "11 StatusNoContent", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusNoContent}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if got := r.IsProxyError(); got != tt.want {
				t.Errorf("Response.IsProxyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_IsSuccess(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "nil", want: false},

		// Via Response
		{name: "1 StatusUseProxy", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "2 StatusUnauthorized", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "3 StatusProxyAuthRequired", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "4 StatusGatewayTimeout", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusGatewayTimeout}}},
		{name: "5 StatusForbidden", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusForbidden}}},
		{name: "6 StatusOK", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "7 StatusCreated", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusCreated}}},
		{name: "8 StatusNoContent", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusNoContent}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if got := r.IsSuccess(); got != tt.want {
				t.Errorf("Response.IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_HasBody(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "nil", want: false},
		{name: "No Body", want: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "body", want: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if got := r.HasBody(); got != tt.want {
				t.Errorf("Response.HasBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_Ignore(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "Read Error", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: fakeReadCloser{}}}},
		{name: "Blank Body", wantErr: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "Has Body", wantErr: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if err := r.Ignore(); (err != nil) != tt.wantErr {
				t.Errorf("Response.Ignore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponse_Status(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{name: "StatusOK", want: http.StatusOK, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "StatusPartialContent", want: http.StatusPartialContent, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusPartialContent}}},
		{name: "StatusPreconditionFailed", want: http.StatusPreconditionFailed, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusPreconditionFailed}}},
		{name: "No Response", want: 0, fields: fields{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if got := r.Status(); got != tt.want {
				t.Errorf("Response.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_GetBytes(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{name: "Read Error", wantErr: true, want: []byte{}, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: fakeReadCloser{}}}},
		{name: "Blank Body", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "Has Body", wantErr: false, want: []byte{97, 97}, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			got, err := r.GetBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.GetBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Response.GetBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_GetInterface(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Read Error", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: fakeReadCloser{}}}},
		{name: "Blank Body", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "Has non JSON Body", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
		{name: "Has proper Body", wantErr: false, args: args{data: &map[string]string{}}, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("{}"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}
			if err := r.GetInterface(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Response.GetInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponse_GetReader(t *testing.T) {
	type fields struct {
		TransactionID string
		HTTPResponse  *http.Response
		Err           error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "Read Error", wantErr: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: fakeReadCloser{}}}},
		{name: "Blank Body", wantErr: true, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}},
		{name: "Has non JSON Body", wantErr: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
		{name: "Has proper Body", wantErr: false, fields: fields{HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("{}"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{
				TransactionID: tt.fields.TransactionID,
				HTTPResponse:  tt.fields.HTTPResponse,
				Err:           tt.fields.Err,
			}

			_, err := r.GetReader()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.GetReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
