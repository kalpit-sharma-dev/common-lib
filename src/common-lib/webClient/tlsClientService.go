package webClient

import (
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
)

type tlsClientService struct {
	config ClientConfig
	client *http.Client
}

// Create create client
func (s *tlsClientService) Create() {
	if s.client != nil {
		return
	}

	// Here false value present that we want to use proxy configuration, if it present
	s.client = client.TLS(clientConfig(s.config), false)
}

func (s *tlsClientService) Do(req *http.Request) (*http.Response, error) {
	s.Create()

	var (
		response     *http.Response
		err          error
		cbEnabled    bool
		validCBError = isValidError
	)

	commandName := req.URL.Host
	if cbInfo, ok := circuitBreaker[commandName]; ok {
		cbEnabled = cbInfo.enabled
		if cbInfo.validCBError != nil && cbEnabled {
			validCBError = cbInfo.validCBError
		}
	}

	err = circuit.Do(commandName, cbEnabled, func() error {
		response, err = s.client.Do(req)
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
func (s *tlsClientService) SetCheckRedirect(cr func(req *http.Request, via []*http.Request) error) {
	s.client.CheckRedirect = cr
}
