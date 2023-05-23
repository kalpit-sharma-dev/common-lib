// Package http implements the http downloader.
//
// Deprecated: http is old implementation of file downloader and should not be used
// except for compatibility with legacy systems.
//
// Use src/download/grab for all downloads
// This package is frozen and no new functionality will be added.
package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	errorCodes "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/errorCodePair"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
	cMock "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient/mock"
)

var (
	downloadLocation = "download"
)

func TestGetDownloader(t *testing.T) {

	service := GetDownloader(webClient.BasicClient, webClient.ClientConfig{})
	_, ok := service.(serviceImpl)
	if !ok {
		t.Error("Invalid serviceImpl")
	}
}

func Test_serviceImpl_Download(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	clientMock := cMock.NewMockHTTPClientService(ctrl)
	clientMock.EXPECT().Create()
	clientMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("Error"))
	clientMock.EXPECT().SetCheckRedirect(gomock.Any())

	resp3 := &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("Read"))}
	clientMock3 := cMock.NewMockHTTPClientService(ctrl)
	clientMock3.EXPECT().Create().Times(1)
	clientMock3.EXPECT().Do(gomock.Any()).Return(resp3, nil).Times(1)
	clientMock3.EXPECT().SetCheckRedirect(gomock.Any()).Times(1)

	mockHeaders := map[string]string{"key": "value"}
	clientHeaderMock := cMock.NewMockHTTPClientService(ctrl)
	clientHeaderMock.EXPECT().Create().Times(2)
	clientHeaderMock.EXPECT().Do(gomock.Any()).Return(resp3, nil).Times(2)
	clientHeaderMock.EXPECT().SetCheckRedirect(gomock.Any()).Times(2)

	type fields struct {
		client webClient.HTTPClientService
	}
	type args struct {
		conf downloader.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: ":::>>>", CheckSumType: checksum.NONE}}, //Wrong URL
			wantErr: true,
		},
		{
			name:    "2",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: "http://test", DownloadLocation: downloadLocation, CheckSumType: checksum.NONE}},
			wantErr: true,
		},
		{
			name:    "3",
			fields:  fields{client: clientMock3},
			args:    args{conf: downloader.Config{URL: "http://test", DownloadLocation: downloadLocation, FileName: fileName, CheckSumType: checksum.NONE}},
			wantErr: false,
		},
		{
			name:    "4",
			fields:  fields{client: clientHeaderMock},
			args:    args{conf: downloader.Config{URL: "http://test", DownloadLocation: downloadLocation, FileName: fileName, Header: mockHeaders, CheckSumType: checksum.NONE}},
			wantErr: false,
		},
		{
			name:    "5",
			fields:  fields{client: clientHeaderMock},
			args:    args{conf: downloader.Config{URL: "http://test", DownloadLocation: downloadLocation, FileName: fileName, Header: mockHeaders, CheckSumType: checksum.MD5, CheckSum: "d41d8cd98f00b204e9800998ecf8427e"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serviceImpl{
				client: tt.fields.client,
			}
			previousFileName := tt.args.conf.DownloadLocation + string(os.PathSeparator) + tt.args.conf.FileName
			res := s.Download(&tt.args.conf)
			if (res.Error != nil) != tt.wantErr {
				t.Errorf("serviceImpl.Download() error = %v, wantErr %v", res.Error, tt.wantErr)
			} else if !tt.wantErr && res.Destination == "" {
				t.Error("Got empty Destination, expecting non empty responce")
			} else if !tt.wantErr && res.Destination != previousFileName {
				t.Errorf("File name got change, Expecting : %s : Got : %s", previousFileName, res.Destination)
			}

			defer func() {
				os.RemoveAll(downloadLocation)
				os.RemoveAll(fileName)
			}()
		})
	}
}

func TestGenerateFileName(t *testing.T) {
	t.Run("EmptyName", func(t *testing.T) {
		conf := downloader.Config{FileName: "test"}
		generateFileName(&conf, &http.Response{})
		assert.NotEqual(t, "", conf.FileName)
	})

	t.Run("GenerateName", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: false}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		generateFileName(&conf, resp)
		msg := []byte(resp.Request.URL.String())
		checkSum := utils.GetChecksum(msg)
		assert.Equal(t, checkSum, conf.FileName)
	})

	t.Run("GetExtFromHeader_1", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: false}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment; filename=foo.exe")
		msg := []byte(resp.Request.URL.String())
		checkSum := utils.GetChecksum(msg)
		generateFileName(&conf, resp)
		assert.Equal(t, checkSum+".exe", conf.FileName)
	})

	t.Run("GetExtFromHeader_2", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: false}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo.exe", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment;")
		msg := []byte(resp.Request.URL.String())
		checkSum := utils.GetChecksum(msg)
		generateFileName(&conf, resp)
		assert.Equal(t, checkSum+".exe", conf.FileName)
	})

	t.Run("GetExtFromHeader_WithQueryParams", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: false}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo.exe?query=params", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment;")
		msg := []byte(resp.Request.URL.String())
		checkSum := utils.GetChecksum(msg)
		generateFileName(&conf, resp)
		assert.Equal(t, checkSum+".exe", conf.FileName)
	})

	t.Run("GetExtFromHeader_3", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: false}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment;")
		msg := []byte(resp.Request.URL.String())
		checkSum := utils.GetChecksum(msg)
		generateFileName(&conf, resp)
		assert.Equal(t, checkSum, conf.FileName)
	})

	t.Run("KeepOriginalName", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: true}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		generateFileName(&conf, resp)
		assert.Equal(t, "foo", conf.FileName)
	})

	t.Run("KeepOriginalNameWithHeader_1", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: true}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment; filename=foo.exe")
		generateFileName(&conf, resp)
		assert.Equal(t, "foo.exe", conf.FileName)
	})

	t.Run("KeepOriginalNameWithHeader_2", func(t *testing.T) {
		conf := downloader.Config{FileName: "", KeepOriginalName: true}
		r := httptest.NewRecorder()
		resp := r.Result()
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		resp.Request = req
		resp.Header.Add("content-disposition", "attachment;")
		generateFileName(&conf, resp)
		assert.Equal(t, "foo", conf.FileName)
	})

}

func Test_serviceImpl_downloadFileWithoutProxy(t *testing.T) {
	type fields struct {
		client     webClient.HTTPClientService
		clientConf webClient.ClientConfig
		clientType webClient.ClientType
	}
	type args struct {
		conf *downloader.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "1",
			fields:  fields{},
			args:    args{conf: &downloader.Config{URL: ":::>>>", CheckSumType: checksum.NONE}}, //Wrong URL
			wantErr: true,
		},
		{
			name:    "2",
			fields:  fields{},
			args:    args{conf: &downloader.Config{URL: "http://google.com", DownloadLocation: downloadLocation, CheckSumType: checksum.NONE}},
			wantErr: false,
			want:    http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serviceImpl{
				client:     tt.fields.client,
				clientConf: tt.fields.clientConf,
				clientType: tt.fields.clientType,
			}
			got, err := s.downloadFileWithoutProxy(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("serviceImpl.downloadFileWithoutProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != 0 && got.StatusCode != tt.want {
				t.Errorf("serviceImpl.downloadFileWithoutProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineDownloadErrors(t *testing.T) {
	type args struct {
		dRspErrorcode string
	}
	tests := []struct {
		name              string
		args              args
		wantMainErrorCode string
		wantSubErrorCode  string
	}{
		{
			name:              "when response error code is ErrorClientOffline",
			args:              args{webClient.ErrorClientOffline},
			wantMainErrorCode: errorCodes.Network,
			wantSubErrorCode:  errorCodes.Connection,
		},
		{
			name:              "when response error code is ChecksumServiceCreationFailed",
			args:              args{ChecksumServiceCreationFailed},
			wantMainErrorCode: errorCodes.Internal,
			wantSubErrorCode:  errorCodes.Operational},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMainErrorCode, gotSubErrorCode := DetermineDownloadErrors(tt.args.dRspErrorcode)
			if gotMainErrorCode != tt.wantMainErrorCode {
				t.Errorf("DetermineDownloadErrors() gotMainErrorCode = %v, want %v", gotMainErrorCode, tt.wantMainErrorCode)
			}
			if gotSubErrorCode != tt.wantSubErrorCode {
				t.Errorf("DetermineDownloadErrors() gotSubErrorCode = %v, want %v", gotSubErrorCode, tt.wantSubErrorCode)
			}
		})
	}
}

func Test_DownloadUsingGateway(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	resp := &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("Read"))}
	clientMock := cMock.NewMockHTTPClientService(ctrl)
	clientMock.EXPECT().Create().Times(3)
	clientMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("Error"))
	clientMock.EXPECT().Do(gomock.Any()).Return(resp, nil).Times(2)
	clientMock.EXPECT().SetCheckRedirect(gomock.Any()).Times(3)

	type fields struct {
		client webClient.HTTPClientService
	}
	type args struct {
		conf downloader.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: "http://10.1.1.0/download", CheckSumType: checksum.MD5, CheckSum: "123", MirrorSites: []downloader.MirrorSites{downloader.MirrorSites{MirrorURL: ":::>>>"}}}}, //Wrong URL
			wantErr: true,
		},
		{
			name:    "2",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: "http://10.1.1.0/download", DownloadLocation: downloadLocation, FileName: fileName, CheckSumType: checksum.MD5, CheckSum: "123", MirrorSites: []downloader.MirrorSites{downloader.MirrorSites{MirrorURL: "http://test"}}}}, //Wrong Checksum
			wantErr: true,
		},
		{
			name:    "3",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: "http://10.1.1.0/download", DownloadLocation: downloadLocation, FileName: fileName, CheckSumType: checksum.MD5, CheckSum: "d41d8cd98f00b204e9800998ecf8427e", MirrorSites: []downloader.MirrorSites{downloader.MirrorSites{MirrorURL: "http://test"}}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serviceImpl{
				client: tt.fields.client,
			}
			if res := s.Download(&tt.args.conf); (res.Error != nil) != tt.wantErr {
				t.Errorf("serviceImpl.Download() error = %v, wantErr %v", res.Error, tt.wantErr)
			}
			defer func() {
				os.RemoveAll(downloadLocation)
				os.RemoveAll(fileName)
			}()
		})
	}
}

func Test_Download_MirrorFailureResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	resp := &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader("Read"))}
	clientMock := cMock.NewMockHTTPClientService(ctrl)
	clientMock.EXPECT().Create().AnyTimes()
	first := clientMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("Error")).Times(1)
	second := clientMock.EXPECT().Do(gomock.Any()).Return(resp, nil).Times(1)
	gomock.InOrder(
		first,
		second,
	)
	clientMock.EXPECT().SetCheckRedirect(gomock.Any()).AnyTimes()

	clientMock1 := cMock.NewMockHTTPClientService(ctrl)
	clientMock1.EXPECT().Create().AnyTimes()
	clientMock1.EXPECT().Do(gomock.Any()).Return(nil, errors.New("Error")).Times(2)
	clientMock1.EXPECT().SetCheckRedirect(gomock.Any()).AnyTimes()

	type fields struct {
		client webClient.HTTPClientService
	}
	type args struct {
		conf downloader.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			fields:  fields{client: clientMock},
			args:    args{conf: downloader.Config{URL: "http://10.1.1.0/download", DownloadLocation: downloadLocation, FileName: fileName, CheckSumType: checksum.MD5, CheckSum: "d41d8cd98f00b204e9800998ecf8427e", MirrorSites: []downloader.MirrorSites{downloader.MirrorSites{MirrorURL: "http://test"}}}},
			wantErr: true,
		},
		{
			name:    "2",
			fields:  fields{client: clientMock1},
			args:    args{conf: downloader.Config{URL: "http://10.1.1.0/download", DownloadLocation: downloadLocation, FileName: fileName, CheckSumType: checksum.MD5, CheckSum: "d41d8cd98f00b204e9800998ecf8427e", MirrorSites: []downloader.MirrorSites{downloader.MirrorSites{MirrorURL: "http://test"}}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serviceImpl{
				client: tt.fields.client,
			}
			if res := s.Download(&tt.args.conf); (res.MirrorFailureError != nil) != tt.wantErr {
				t.Errorf("serviceImpl.Download() error = %v, wantErr %v", res.Error, tt.wantErr)
			}
			defer func() {
				os.RemoveAll(downloadLocation)
				os.RemoveAll(fileName)
			}()
		})
	}
}
