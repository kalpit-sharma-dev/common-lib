package grab

import (
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
)

func TestGetDownloader(t *testing.T) {
	type args struct {
		conf *client.Config
	}
	tests := []struct {
		name string
		args args
		want downloader.Service
	}{
		{name: "Instance", args: args{conf: &client.Config{}}, want: &serviceImpl{conf: &client.Config{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDownloader(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDownloader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceImpl_Download(t *testing.T) {
	type fields struct {
		conf *client.Config
	}
	type args struct {
		conf *downloader.Config
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		ErrorCode string
	}{
		{
			name: "1. create Request Error", fields: fields{conf: &client.Config{}},
			args:    args{conf: &downloader.Config{URL: ":::>>>"}},
			wantErr: true, ErrorCode: "RequestCreationFailed",
		},
		{
			name: "2. execute Request Error", fields: fields{conf: &client.Config{}},
			args:    args{conf: &downloader.Config{URL: "test"}},
			wantErr: true, ErrorCode: "RequestExecutionFailed",
		},
		{
			name: "3. Proxy Error", fields: fields{conf: &client.Config{DialTimeoutSecond: 1,
				Proxy: client.Proxy{Address: "10.0.0.1", Port: 9000, Protocol: "http"}}},
			args:    args{conf: &downloader.Config{URL: "http://test"}},
			wantErr: true, ErrorCode: "WithoutProxyRequestExecutionFailed",
		},
		// Currect URL to download File
		{
			name: "4. Currect URL to download File with Success", fields: fields{conf: &client.Config{}},
			args: args{conf: &downloader.Config{
				URL:              "http://cdn.itsupport247.net/InstallJunoAgent/Plugin/Windows/platform-installation-manager/1.0.216/platform_installation_manager_windows32_1.0.216.zip",
				DownloadLocation: "./",
				FileName:         "test.zip",
				TransactionID:    "1",
				CheckSumType:     checksum.MD5,
			}},
			wantErr: false, ErrorCode: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serviceImpl{
				conf: tt.fields.conf,
			}
			got := s.Download(tt.args.conf)

			if got == nil || (got.Error != nil) != tt.wantErr || got.ErrorCode != tt.ErrorCode {
				t.Errorf("serviceImpl.Download() = %v, want %v and Code %v", got, tt.wantErr, tt.ErrorCode)
			}
		})
	}
}
