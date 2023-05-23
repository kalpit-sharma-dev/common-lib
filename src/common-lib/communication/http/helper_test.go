package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestIsProxyError(t *testing.T) {
	type args struct {
		err error
		res *http.Response
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "both nil", want: false},

		// Via Err
		{name: "1 proxyconnect Error", want: true, args: args{err: errors.New("proxyconnect")}},
		{name: "2 proxy1 Error", want: false, args: args{err: errors.New("proxy1")}},
		{name: "3 proxyconnect Error", want: true, args: args{err: errors.New("error proxyconnect err")}},

		// Via Response
		{name: "4 StatusUseProxy", want: true, args: args{res: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "5 StatusUnauthorized", want: true, args: args{res: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "6 StatusProxyAuthRequired", want: true, args: args{res: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "7 StatusGatewayTimeout", want: true, args: args{res: &http.Response{StatusCode: http.StatusGatewayTimeout}}},
		{name: "8 StatusForbidden", want: true, args: args{res: &http.Response{StatusCode: http.StatusForbidden}}},
		{name: "9 StatusOK", want: false, args: args{res: &http.Response{StatusCode: http.StatusOK}}},
		{name: "10 StatusCreated", want: false, args: args{res: &http.Response{StatusCode: http.StatusCreated}}},
		{name: "11 StatusNoContent", want: false, args: args{res: &http.Response{StatusCode: http.StatusNoContent}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProxyError(tt.args.err, tt.args.res); got != tt.want {
				t.Errorf("IsProxyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSuccess(t *testing.T) {
	type args struct {
		res *http.Response
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "nil", want: false},

		// Via Response
		{name: "1 StatusUseProxy", want: false, args: args{res: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "2 StatusUnauthorized", want: false, args: args{res: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "3 StatusProxyAuthRequired", want: false, args: args{res: &http.Response{StatusCode: http.StatusUnauthorized}}},
		{name: "4 StatusGatewayTimeout", want: false, args: args{res: &http.Response{StatusCode: http.StatusGatewayTimeout}}},
		{name: "5 StatusForbidden", want: false, args: args{res: &http.Response{StatusCode: http.StatusForbidden}}},
		{name: "6 StatusOK", want: true, args: args{res: &http.Response{StatusCode: http.StatusOK}}},
		{name: "7 StatusCreated", want: true, args: args{res: &http.Response{StatusCode: http.StatusCreated}}},
		{name: "8 StatusNoContent", want: true, args: args{res: &http.Response{StatusCode: http.StatusNoContent}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSuccess(tt.args.res); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasBody(t *testing.T) {
	type args struct {
		res *http.Response
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "nil", want: false},
		{name: "No Body", want: false, args: args{res: &http.Response{StatusCode: http.StatusUseProxy}}},
		{name: "body", want: true, args: args{res: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("aa"))}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasBody(tt.args.res); got != tt.want {
				t.Errorf("HasBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
