package grab

import (
	"fmt"
	"time"

	"github.com/cavaliercoder/grab"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
)

type serviceImpl struct {
	conf *client.Config
}

// GetDownloader is a definition returns a HTTP downloader instance
func GetDownloader(conf *client.Config) downloader.Service {
	return &serviceImpl{
		conf: conf,
	}
}

func (s *serviceImpl) Download(conf *downloader.Config) *downloader.Response {
	res, dresp := s.download(conf, true)

	if res != nil && res.HTTPResponse != nil {
		dresp.StatusCode = res.HTTPResponse.StatusCode
	}

	if dresp.Error != nil {
		return dresp
	}

	return s.validate(res, conf)
}

func (s *serviceImpl) download(conf *downloader.Config, store bool) (*grab.Response, *downloader.Response) {
	req, err := s.createRequest(conf, store)
	if err != nil {
		return nil, &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, ErrorCode: "RequestCreationFailed"}
	}

	client := s.createClient(false, conf)
	res := client.Do(req)
	err = s.handleResponse(res, conf)

	if http.IsProxyError(err, res.HTTPResponse) {
		if s.conf.Proxy.Address != "" {
			conf.Logger().Trace(conf.TransactionID, "Failed to execute request with Proxy for url : %s with Error : %v retrying without Proxy ", conf.URL, err)
			req, err = s.createRequest(conf, store)
			if err != nil {
				return nil, &downloader.Response{Error: err, Destination: conf.Destination(),
					Config: conf, ErrorCode: "RequestCreationFailed"}
			}

			client := s.createClient(true, conf)
			res = client.Do(req)
			err = s.handleResponse(res, conf)

			if err != nil {
				return nil, &downloader.Response{Error: err, Destination: conf.Destination(),
					Config: conf, Size: res.BytesComplete(), ErrorCode: "WithoutProxyRequestExecutionFailed"}
			}
		} else {
			return res, &downloader.Response{Error: err, Destination: conf.Destination(),
				Config: conf, Size: res.BytesComplete(), ErrorCode: "WithoutProxyRequestExecutionFailed"}
		}
	} else if err != nil {
		return res, &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, Size: res.BytesComplete(), ErrorCode: "RequestExecutionFailed"}
	}
	return res, &downloader.Response{Destination: conf.Destination(), Size: res.BytesComplete()}
}

func (s serviceImpl) validate(res *grab.Response, conf *downloader.Config) *downloader.Response {
	cfg := &downloader.Config{
		CheckSum:         conf.CheckSum,
		DownloadLocation: conf.DownloadLocation,
		FileName:         fmt.Sprintf("%s.%s", conf.GenerateFileName(), conf.CheckSumType.Name),
		KeepOriginalName: conf.KeepOriginalName,
		TransactionID:    conf.TransactionID,
		URL:              fmt.Sprintf("%s.%s", conf.URL, conf.CheckSumType.Name),
		CheckSumType:     conf.CheckSumType,
	}

	resp, dresp := s.download(cfg, false)

	if res != nil && res.HTTPResponse != nil {
		dresp.StatusCode = res.HTTPResponse.StatusCode
	}

	if dresp.Error != nil {
		return dresp
	}

	sum, err := resp.Bytes()
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, Size: res.BytesComplete(), ErrorCode: "ChecksumDownloadFailed"}
	}

	service, err := checksum.GetService(conf.CheckSumType)
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, Size: res.BytesComplete(), ErrorCode: "ChecksumServiceCreationFailed"}
	}

	file, err := res.Open()
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, Size: res.BytesComplete(), ErrorCode: "FileReadingFailed"}
	}

	defer file.Close() // nolint

	_, err = service.Validate(file, string(sum))
	if err != nil {
		return &downloader.Response{Error: err, Destination: conf.Destination(),
			Config: conf, Size: res.BytesComplete(), ErrorCode: "ChecksumValidationFailed"}
	}

	return &downloader.Response{Destination: conf.Destination(), Config: conf, Size: res.BytesComplete()}
}

func (s serviceImpl) handleResponse(resp *grab.Response, conf *downloader.Config) error {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			if resp.BytesComplete() != resp.Size() {
				conf.Logger().Trace(conf.TransactionID, "File %s => transferred %v / %v bytes (%.2f%%)",
					conf.Destination(), resp.BytesComplete(), resp.Size(), 100*resp.Progress())
			}
		case <-resp.Done:
			// Download is complete; so move out from loop
			break Loop
		}
	}

	return resp.Err()
}

func (s *serviceImpl) createRequest(conf *downloader.Config, store bool) (*grab.Request, error) {
	req, err := grab.NewRequest(conf.Destination(), conf.URL)

	if err != nil {
		return nil, err
	}

	req.NoResume = false
	req.SkipExisting = false
	req.NoStore = !store

	for key, value := range conf.Header {
		req.HTTPRequest.Header.Set(key, value)
	}

	return req, err
}

func (s *serviceImpl) createClient(ignoreProxy bool, cfg *downloader.Config) *grab.Client {
	clnt := client.TLS(s.conf, ignoreProxy)
	client.Redirect(clnt, cfg.Header)

	return &grab.Client{
		HTTPClient: clnt,
		UserAgent:  cfg.UsrAgent(),
		BufferSize: cfg.BuffSize(),
	}
}
