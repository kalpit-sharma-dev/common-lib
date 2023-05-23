package http

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
)

func TestCreateRequest(t *testing.T) {
	type args struct {
		transactionID string
		method        string
		url           string
		serviceName   string
	}
	tests := []struct {
		name string
		args args
		want *Request
	}{
		{
			name: "1 Create POST", args: args{transactionID: "1", method: "POST", url: "test", serviceName: "http.test"},
			want: &Request{transactionID: "1", method: "POST", url: "test", header: http.Header{}},
		},
		{
			name: "2 Create GET", args: args{transactionID: "2", method: "GET", url: "test1", serviceName: "http.test"},
			want: &Request{transactionID: "2", method: "GET", url: "test1", header: http.Header{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.header.Add(constants.TransactionID, tt.args.transactionID)
			tt.want.header.Add(constants.XRequestID, tt.args.transactionID)
			tt.want.header.Add(constants.ServiceName, tt.args.serviceName)
			if got := CreateRequest(tt.args.transactionID, tt.args.method, tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_SetBytes(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name string
		args args
		want io.Reader
	}{
		{name: "1 test", want: bytes.NewReader([]byte("test")), args: args{body: []byte("test")}},
		{name: "2 test1", want: bytes.NewReader([]byte("test1")), args: args{body: []byte("test1")}},
		{name: "3 test2", want: bytes.NewReader([]byte("test2")), args: args{body: []byte("test2")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{}
			r.SetBytes(tt.args.body)

			if !reflect.DeepEqual(r.body, tt.want) {
				t.Errorf("Create() = %v, want %v", r.body, tt.want)
			}
		})
	}
}

func TestRequest_SetInterface(t *testing.T) {
	type args struct {
		body interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1 test", wantErr: false, args: args{body: "test"}},
		{name: "2 nil", wantErr: true, args: args{body: nil}},
		{name: "3 byte Array", wantErr: false, args: args{body: []byte("test")}},
		{name: "4 io.Reader", wantErr: false, args: args{body: bytes.NewReader([]byte("test"))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{}
			if err := r.SetInterface(tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("Request.SetInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRequest_Build(t *testing.T) {
	t.Run("Request Error", func(t *testing.T) {
		r := &Request{method: ";;;", url: "url"}
		_, err := r.Build()
		if err == nil {
			t.Error("Request.Build() expected error got nil")
		}
	})

	t.Run("Request Success", func(t *testing.T) {
		r := &Request{method: http.MethodPost, url: "url"}
		_, err := r.Build()
		if err != nil {
			t.Errorf("Request.Build() expected nil got %v", err)
		}
	})
}

func TestRequest_Execute(t *testing.T) {
	t.Run("Request Creation Error", func(t *testing.T) {
		r := &Request{method: ";;;", url: "url"}
		got := r.Execute(&client.Config{}, true)
		if got != nil && got.Err == nil {
			t.Errorf("Request.Execute() = expected nil, got %v", got)
		}
	})

	t.Run("Request URL Error", func(t *testing.T) {
		r := &Request{method: http.MethodPost, url: "url"}
		got := r.Execute(&client.Config{}, true)
		if got != nil && got.Err == nil {
			t.Errorf("Request.Execute() = expected nil, got %v", got)
		}
	})
}

func TestRequest_SetHeader(t *testing.T) {
	const newValue = "valueReplace"
	t.Run("SetHeader - positive", func(t *testing.T) {
		r := &Request{transactionID: "1", method: "POST", url: "test", header: http.Header{}}
		r.SetHeader("key1", "valueOne")
		r.SetHeader("key1", newValue)
		got := r.header.Get("key1")
		if got != newValue {
			t.Errorf("Request.SetHeader() = expected %v, got %v", newValue, got)
		}
	})

}
