package webClient

import (
	"errors"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
)

const (
	proxyURL string = "%s://%s:%d"
)

var errCb = errors.New("Server returned error status code")

// httpClientServiceImpl implements HTTPCommandService
type httpClientServiceImpl struct {
	config     ClientConfig
	httpClient *http.Client
}

// Create create client
func (hc *httpClientServiceImpl) Create() {
	if hc.httpClient != nil {
		return
	}

	// Here false value present that we want to use proxy configuration, if it present
	hc.httpClient = client.Basic(clientConfig(hc.config), false)
}

// Do sends Request to the Server
func (hc *httpClientServiceImpl) Do(request *http.Request) (*http.Response, error) {
	hc.Create()

	var (
		response     *http.Response
		err          error
		cbEnabled    bool
		validCBError = isValidError
	)

	commandName := request.URL.Host
	if cbInfo, ok := circuitBreaker[commandName]; ok {
		cbEnabled = cbInfo.enabled
		if cbInfo.validCBError != nil && cbEnabled {
			validCBError = cbInfo.validCBError
		}
	}

	err = circuit.Do(commandName, cbEnabled, func() error {
		response, err = hc.httpClient.Do(request)
		if err != nil {
			err = checkOffline(err)
			return err
		}

		if validCBError(response) {
			return errCb
		}

		return nil
	}, nil)

	if err == errCb {
		return response, nil
	}

	return response, err
}

// SetCheckRedirect set CheckRedirect to client
func (hc *httpClientServiceImpl) SetCheckRedirect(cr func(req *http.Request, via []*http.Request) error) {
	hc.httpClient.CheckRedirect = cr
}

var isValidError = func(response *http.Response) bool {
	if response.StatusCode >= http.StatusInternalServerError {
		return true
	}
	return false
}

func clientConfig(config ClientConfig) *client.Config {
	cfg := &client.Config{
		TimeoutMinute:               config.TimeoutMinute,
		TimeoutMillisecond:          config.TimeoutMillisecond,
		MaxIdleConns:                config.MaxIdleConns,
		MaxIdleConnsPerHost:         config.MaxIdleConnsPerHost,
		MaxConnsPerHost:             0,
		IdleConnTimeoutMinute:       config.IdleConnTimeoutMinute,
		DialTimeoutSecond:           config.DialTimeoutSecond,
		DialKeepAliveSecond:         config.DialKeepAliveSecond,
		TLSHandshakeTimeoutSecond:   config.TLSHandshakeTimeoutSecond,
		ExpectContinueTimeoutSecond: config.ExpectContinueTimeoutSecond,
		UseIEProxy:                  config.UseIEProxy,
		ValidateSSLCertificate:      config.ValidateSSLCertificate,
	}
	cfg.Proxy = client.Proxy{
		Address:  config.ProxySetting.IP,
		Port:     config.ProxySetting.Port,
		UserName: config.ProxySetting.UserName,
		Password: config.ProxySetting.Password,
		Protocol: config.ProxySetting.Protocol,
	}
	cfg.TracingConfig = config.TracingConfig
	return cfg
}
