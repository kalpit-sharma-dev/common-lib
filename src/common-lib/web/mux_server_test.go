package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetRequestContext(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want RequestContext
	}{
		{
			name: "To Get the Request Context Object",
			args: args{w: &httptest.ResponseRecorder{}, r: &http.Request{}},
			want: &muxRequestContext{
				response:      &httptest.ResponseRecorder{},
				request:       &http.Request{},
				dcDateTimeUTC: time.Now().UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRequestContext(tt.args.w, tt.args.r); got != nil && tt.want != nil &&
				!reflect.DeepEqual(got.GetResponse(), tt.args.w) &&
				!reflect.DeepEqual(got.GetRequest(), tt.args.r) {
				t.Errorf("GetRequestContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
