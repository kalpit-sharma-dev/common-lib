// Package http implements the http downloader.
//
// Deprecated: http is old implementation of file downloader and should not be used
// except for compatibility with legacy systems.
//
// Use src/download/grab for all downloads
// This package is frozen and no new functionality will be added.
package http

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	errorCodes "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/errorCodePair"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	communication "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

const (
	fileName            string = "checksumfile"
	proxyErrorSubString string = "proxyconnect"
)

//Error Codes
const (
	ChecksumServiceCreationFailed      = "ChecksumServiceCreationFailed"
	RequestCreationFailed              = "RequestCreationFailed"
	FileReadingFailed                  = "FileReadingFailed"
	ChecksumValidationFailed           = "ChecksumValidationFailed"
	WithoutProxyRequestExecutionFailed = "WithoutProxyRequestExecutionFailed"
	RequestExecutionFailed             = "RequestExecutionFailed"
	HandleResponseFailed               = "HandleResponseFailed"
)

type serviceImpl struct {
	client     webClient.HTTPClientService
	clientConf webClient.ClientConfig
	clientType webClient.ClientType
}

//SuccessStatuses indicates successful HTTP responses
var SuccessStatuses = map[int]bool{
	http.StatusOK:        true,
	http.StatusCreated:   true,
	http.StatusNoContent: true,
}

var proxyErrorStatus = map[int]bool{
	http.StatusUseProxy:          true,
	http.StatusUnauthorized:      true,
	http.StatusProxyAuthRequired: true,
	http.StatusGatewayTimeout:    true,
	http.StatusForbidden:         true,
}

//GetDownloader is a definition returns a HTTP downloader instance
func GetDownloader(clientType webClient.ClientType, config webClient.ClientConfig) downloader.Service {
	return serviceImpl{
		client:     webClient.ClientFactoryImpl{}.GetClientServiceByType(clientType, config),
		clientConf: config,
		clientType: clientType,
	}
}

func (s serviceImpl) Download(conf *downloader.Config) *downloader.Response {

	backupConf := *conf

	var res *downloader.Response
	var mirrorFail bool
	var mirrorFailError error

	if len(conf.MirrorSites) > 0 {
		s.setupMirrorSites(conf)

		for i := 0; i < len(conf.MirrorSites); i++ {
			// Setting Up Gateway URL
			conf.URL = conf.MirrorSites[i].MirrorURL

			res, mirrorFail = s.download(conf)

			if mirrorFail {
				continue
			}
			break
		}
	}

	if mirrorFail {
		mirrorFailError = res.Error
	}

	if res == nil || res.Error != nil {
		// Retry package download without gateway
		res, _, _ = s.downloadFile(&backupConf)
		copier.Copy(conf, &backupConf)
	}

	if res.Error != nil {
		//Set mirrorFail to false as it fails to download from Web as well
		mirrorFail = false
		res.MirrorFailure = mirrorFail
		return res
	}

	downloadRsp := s.validate(conf)

	//Adding mirror result when able to download package from web and not from mirror site
	downloadRsp.MirrorFailure = mirrorFail
	downloadRsp.MirrorFailureError = mirrorFailError
	return downloadRsp
}

func (s serviceImpl) download(conf *downloader.Config) (*downloader.Response, bool) {
	var res *downloader.Response
	var retry, mirrorFail bool

	for i := 0; i < conf.GetDownloadRetryCount(); i++ {
		res, mirrorFail, retry = s.downloadFile(conf)

		// To avoid sleep after last retry
		if retry && (i+1) < conf.DownloadRetryCount {
			time.Sleep(conf.DownloadRetryDelay)
			continue
		}
		break
	}
	return res, mirrorFail
}

func (s serviceImpl) setupMirrorSites(conf *downloader.Config) {

	// Setting Up mirror sites specific headers
	if conf.Header == nil {
		conf.Header = map[string]string{}
	}
	conf.Header["url"] = conf.URL
	conf.Header["checksumValue"] = conf.CheckSum
	conf.Header["checksumType"] = conf.CheckSumType.Name

}

func (s serviceImpl) validate(conf *downloader.Config) *downloader.Response {
	service, err := checksum.GetService(conf.CheckSumType)
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: ChecksumServiceCreationFailed}

	}

	checkSum := conf.CheckSum
	if conf.CheckSum == "" && conf.CheckSumType != checksum.NONE {
		sum, dres := s.getCheckSumFromFile(conf)
		if dres.Error != nil {
			return dres
		}
		checkSum = sum
	}

	location := filepath.Join(conf.DownloadLocation, conf.FileName)
	file, err := os.Open(filepath.Clean(location))
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: FileReadingFailed}
	}

	defer file.Close() // nolint
	_, err = service.Validate(file, checkSum)
	if err != nil {
		fi, er := file.Stat()
		if er == nil {
			return &downloader.Response{Error: err, Destination: conf.Destination(), Size: fi.Size(), ErrorCode: ChecksumValidationFailed}
		}
	}
	return &downloader.Response{Destination: conf.Destination()}
}

func (s serviceImpl) getCheckSumFromFile(conf *downloader.Config) (string, *downloader.Response) {
	newConf := downloader.Config{
		CheckSum:         conf.CheckSum,
		DownloadLocation: conf.DownloadLocation,
		FileName:         fileName,
		KeepOriginalName: conf.KeepOriginalName,
		TransactionID:    conf.TransactionID,
		URL:              fmt.Sprintf("%s.%s", conf.URL, conf.CheckSumType.Name),
		CheckSumType:     conf.CheckSumType,
	}

	dres, _, _ := s.downloadFile(&newConf)
	if dres.Error != nil {
		return "", dres
	}

	path := filepath.Join(newConf.DownloadLocation, newConf.FileName)
	value, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: FileReadingFailed}

	}
	return string(value), &downloader.Response{Destination: conf.Destination()}
}

func (s serviceImpl) createRequest(conf *downloader.Config) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, conf.URL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range conf.Header {
		req.Header.Set(key, value)
	}

	return req, err
}

func (s serviceImpl) downloadFile(conf *downloader.Config) (*downloader.Response, bool, bool) {
	req, err := s.createRequest(conf)
	if err != nil {
		conf.Logger().Trace(conf.TransactionID, "Failed to create request for url : %s with Error : %v", conf.URL, err)
		return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: RequestCreationFailed}, false, false
	}
	s.client.Create()
	s.client.SetCheckRedirect(redirectWithHeaders(conf.Header))

	res, err := s.client.Do(req)

	// Closing response body to make sure that we do not have any leak
	if communication.HasBody(res) {
		body := res.Body
		defer body.Close() // nolint
	}

	if s.isProxyError(err, res) {
		if s.clientConf.ProxySetting.IP != "" {
			conf.Logger().Trace(conf.TransactionID, "Failed to execute request with Proxy for url : %s with Error : %v retrying without Proxy ", conf.URL, err)
			res, err = s.downloadFileWithoutProxy(conf)

			// Closing response body to make sure that we do not have any leak
			if communication.HasBody(res) {
				defer res.Body.Close() // nolint
			}

			if err != nil {
				conf.Logger().Trace(conf.TransactionID, "Failed to execute request without Proxy for url  : %s with Error : %v \n ", conf.URL, err)
				return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: WithoutProxyRequestExecutionFailed}, true, false
			}
		} else {
			conf.Logger().Trace(conf.TransactionID, "Failed to execute request without Proxy for url : %s with Error : %v", conf.URL, err)
			return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: WithoutProxyRequestExecutionFailed}, true, false
		}
	} else if err != nil {
		conf.Logger().Trace(conf.TransactionID, "Failed to execute request with or without proxy for url : %s with Error : %v", conf.URL, err)
		return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: RequestExecutionFailed}, true, false
	}

	retry, err := s.handleResponse(res, conf)
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(), ErrorCode: HandleResponseFailed}, false, retry
	}
	return &downloader.Response{Destination: conf.Destination()}, false, false
}

func (s serviceImpl) handleResponse(res *http.Response, conf *downloader.Config) (bool, error) {
	if res == nil || res.Body == nil {
		return false, fmt.Errorf("Failed to download. Response and body does not exist")
	}

	if conf.GetRetryableStatuses()[res.StatusCode] {
		return true, fmt.Errorf("Failed to download. StatusCode %d", res.StatusCode)
	}

	if !SuccessStatuses[res.StatusCode] {
		return false, fmt.Errorf("Failed to download. StatusCode %d", res.StatusCode)
	}

	err := os.MkdirAll(conf.DownloadLocation, os.ModePerm)
	if err != nil {
		return false, err
	}

	generateFileName(conf, res)
	conf.Logger().Trace(conf.TransactionID, "Received Response with status %s for url : %s", res.Status, conf.URL)
	return false, s.createFile(conf, res.Body)
}

func (s serviceImpl) isProxyError(err error, res *http.Response) bool {
	return (err != nil && strings.Contains(err.Error(), proxyErrorSubString)) ||
		(res != nil && proxyErrorStatus[res.StatusCode])
}

func (s serviceImpl) downloadFileWithoutProxy(conf *downloader.Config) (*http.Response, error) {
	client := webClient.ClientFactoryImpl{}.GetClientServiceByType(s.clientType, webClient.ClientConfig{
		IdleConnTimeoutMinute:       s.clientConf.IdleConnTimeoutMinute,
		MaxIdleConns:                s.clientConf.MaxIdleConns,
		MaxIdleConnsPerHost:         s.clientConf.MaxIdleConnsPerHost,
		TimeoutMinute:               s.clientConf.TimeoutMinute,
		DialKeepAliveSecond:         s.clientConf.DialKeepAliveSecond,
		DialTimeoutSecond:           s.clientConf.DialTimeoutSecond,
		ExpectContinueTimeoutSecond: s.clientConf.ExpectContinueTimeoutSecond,
		TLSHandshakeTimeoutSecond:   s.clientConf.TLSHandshakeTimeoutSecond,
	})
	client.Create()
	client.SetCheckRedirect(redirectWithHeaders(conf.Header))

	req, err := s.createRequest(conf)
	if err != nil {
		conf.Logger().Trace(conf.TransactionID, "Failed to create request for url : %s with Error : %v", conf.URL, err)
		return nil, err
	}

	return client.Do(req)
}

// redirectWithHeaders http client CheckRedirect
var redirectWithHeaders = func(h map[string]string) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}

		for key, value := range h {
			req.Header.Set(key, value)
		}
		return nil
	}
}

//### Should be moved to common file at the time of adding new Downloader like FTP ####
func (s serviceImpl) createFile(conf *downloader.Config, data io.Reader) error {
	dst := conf.DownloadLocation + string(os.PathSeparator) + conf.FileName
	out, err := os.Create(dst)
	if err != nil {
		conf.Logger().Trace(conf.TransactionID, "Failed to create File : %s with Error : %v", dst, err)
		return err
	}
	defer out.Close() //nolint

	w, err := io.Copy(out, data)
	if err != nil {
		conf.Logger().Trace(conf.TransactionID, "Failed to copy File : %s with Error %v", dst, err)
		return err
	}
	conf.Logger().Trace(conf.TransactionID, "%d of bytes copied for File : %s", w, dst)
	return nil
}

func generateFileName(config *downloader.Config, resp *http.Response) {
	if config.FileName != "" {
		return
	}

	reqURL := *resp.Request.URL
	reqURL.RawQuery = ""
	filename := reqURL.String()

	header := resp.Header.Get("content-disposition")
	if header != "" {
		_, params, err := mime.ParseMediaType(header)
		if err == nil && params["filename"] != "" {
			filename = params["filename"]
		}
	}

	if !config.KeepOriginalName {
		h := md5.New()                             // #nosec
		h.Write([]byte(resp.Request.URL.String())) //nolint
		config.FileName = hex.EncodeToString(h.Sum(nil)) + filepath.Ext(filename)
		return
	}
	config.FileName = filepath.Base(filename)
}

// changes for agent autoupdate error standardization START HERE. To be refactored as per common-lib standards for comming rollouts

// DetermineDownloadErrors error code pairs for download failures
func DetermineDownloadErrors(dRspErrorcode string) (mainErrorCode, subErrorCode string) {
	switch dRspErrorcode {
	case ChecksumValidationFailed:
		mainErrorCode, subErrorCode = errorCodes.Download, errorCodes.ChecksumValidationFailed
	case FileReadingFailed:
		mainErrorCode, subErrorCode = errorCodes.FileSystem, errorCodes.FileNotFound
	case RequestCreationFailed, RequestExecutionFailed, webClient.ErrorClientOffline:
		mainErrorCode, subErrorCode = errorCodes.Network, errorCodes.Connection
	case WithoutProxyRequestExecutionFailed:
	default: //including ChecksumServiceCreationFailed, HandleResponseFailed
		mainErrorCode, subErrorCode = errorCodes.Internal, errorCodes.Operational
	}
	return
}

// changes for agent autoupdate error standardization END HERE.
