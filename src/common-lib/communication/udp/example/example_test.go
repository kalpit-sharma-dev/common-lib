package main

import (
	"errors"
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp/mock"
	"github.com/golang/mock/gomock"
)

func Test_setupClient(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "udp.Send Error",
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return errors.New("Error") }
			},
			wantErr: true,
		},
		{
			name: "udp.Send Success",
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return nil }
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := setupClient(); (err != nil) != tt.wantErr {
				t.Errorf("setupClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setupServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	server := mock.NewMockServer(ctrl)

	tests := []struct {
		name    string
		wantErr bool
		setup   func()
	}{
		{name: "1 Recieve Error", wantErr: true, setup: func() {
			udp.NewServer = func(conf *udp.Config) udp.Server { return server }
			server.EXPECT().Receive(gomock.Any()).Return(errors.New("Error"))
		}},
		{name: "2 Recieve Success", wantErr: false, setup: func() {
			udp.NewServer = func(conf *udp.Config) udp.Server { return server }
			server.EXPECT().Receive(gomock.Any()).Return(nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := setupServer(); (err != nil) != tt.wantErr {
				t.Errorf("setupServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_responseHandler(t *testing.T) {
	type args struct {
		response udp.Response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1 No response"},
		{name: "2 Response", args: args{response: udp.Response{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := responseHandler(tt.args.response); (err != nil) != tt.wantErr {
				t.Errorf("responseHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_requestHandler(t *testing.T) {
	type args struct {
		request udp.Request
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "1 No Request", want: []byte{84, 101, 115, 116, 32, 77, 101, 115, 115, 97, 103, 101, 32,
			82, 101, 99, 101, 105, 118, 101, 100, 32, 79, 118, 101, 114, 32, 85, 68, 80}},
		{name: "2 Request", args: args{request: udp.Request{}}, want: []byte{84, 101, 115, 116, 32,
			77, 101, 115, 115, 97, 103, 101, 32, 82, 101, 99, 101, 105, 118, 101, 100, 32, 79, 118, 101, 114, 32, 85, 68, 80}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestHandler(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_main(t *testing.T) {
	ctrl := gomock.NewController(t)
	server := mock.NewMockServer(ctrl)

	tests := []struct {
		name  string
		setup func()
	}{
		{name: "test Server", setup: func() {
			udp.NewServer = func(conf *udp.Config) udp.Server { return server }
			server.EXPECT().Receive(gomock.Any()).Return(errors.New("Error"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			main()
		})
	}
}
