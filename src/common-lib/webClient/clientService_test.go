package webClient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

func setup(handler http.Handler) (*http.Client, func()) {

	cbCfg := CircuitBreakerConfig{
		CircuitBreaker: circuit.New(),
		BaseURL:        "https://www.google.com/",
	}

	RegisterCircuitBreaker([]CircuitBreakerConfig{cbCfg})

	s := httptest.NewTLSServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}

func TestValidateRequestNilNoContent(t *testing.T) {
	request := &http.Request{}
	err := ValidateRequest(request)
	if err == nil {
		t.Errorf("Expected error %s, not returned", err)
		return
	}
	if err.Error() != ErrorEmptyContentType {
		t.Errorf("Unexpected error returned, Expected Error: %s, Returned Error: %v", ErrorEmptyContentType, err)
		return
	}
}

func TestValidateRequestNilUrl(t *testing.T) {
	request := &http.Request{}
	request.Header = http.Header{}
	request.Header.Set("Content-Type", "text")
	err := ValidateRequest(request)
	if err == nil {
		t.Errorf("Expected error %s, not returned", err)
		return
	}
	if err.Error() != ErrorNilURL {
		t.Errorf("Unexpected error returned, Expected Error: %s, Returned Error: %v", ErrorNilURL, err)
		return
	}
}

func TestValidateRequestNilMethod(t *testing.T) {
	request := &http.Request{
		URL: &url.URL{},
	}
	request.Header = http.Header{}
	request.Header.Set("Content-Type", "text")
	err := ValidateRequest(request)
	if err == nil {
		t.Errorf("Expected error %s, not returned", err)
		return
	}
	if err.Error() != ErrorBlankHTTPMethod {
		t.Errorf("Unexpected error returned, Expected Error: %s, Returned Error: %v", ErrorBlankHTTPMethod, err)
		return
	}
}

func TestValidateRequestSuccess(t *testing.T) {
	request := &http.Request{
		URL:    &url.URL{},
		Method: "GET",
	}
	request.Header = http.Header{}
	request.Header.Set("Content-Type", "text")
	err := ValidateRequest(request)
	if err != nil {
		t.Errorf("Expected nil error, %s returned", err)
		return
	}
}

func TestGetClientService(t *testing.T) {
	clientFact := ClientFactoryImpl{}
	httpClientFact := HTTPClientFactoryImpl{}
	clientService := clientFact.GetClientService(httpClientFact, ClientConfig{})
	if clientService == nil {
		t.Error("Expected ClientFactory returned nil")
	}
}

func TestGetClientServiceDo(t *testing.T) {
	clientFact := ClientFactoryImpl{}
	httpClientFact := HTTPClientFactoryImpl{}
	clientService := clientFact.GetClientService(httpClientFact, ClientConfig{})
	request, _ := http.NewRequest("GET", "http://local", nil)
	_, err := clientService.Do(request)
	if err == nil {
		t.Error("Expected Error returned nil")
	}
}

func TestGetTLSClientServiceDo(t *testing.T) {

	t.Run("1. Error TLS Client", func(t *testing.T) {
		clientFact := ClientFactoryImpl{}
		clientService := clientFact.GetClientServiceByType(TLSClient, ClientConfig{})
		request, _ := http.NewRequest("GET", "http://local", nil)
		_, err := clientService.Do(request)
		if err == nil {
			t.Error("Expected Error returned nil")
		}

	})

}

func TestTLSClientDo(t *testing.T) {

	t.Run("1. Internal Server Error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		})

		httpClient, teardown := setup(h)
		defer teardown()
		clientService := &tlsClientService{
			config: ClientConfig{},
		}

		clientService.client = httpClient
		request, _ := http.NewRequest(http.MethodGet, "https://www.google.com/", nil)
		_, err := clientService.Do(request)
		if err != nil {
			t.Error("Expected nil but found error")
		}
	})

	t.Run("2. Status 200", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		httpClient, teardown := setup(h)
		defer teardown()
		clientService := &tlsClientService{
			config: ClientConfig{},
		}

		clientService.client = httpClient
		request, _ := http.NewRequest(http.MethodGet, "https://www.google.com/", nil)
		_, err := clientService.Do(request)
		if err != nil {
			t.Error("Expected nil but found error")
		}
	})

}

func TestBasicClientDo(t *testing.T) {

	t.Run("1. Internal Server Error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		})

		httpClient, teardown := setup(h)
		defer teardown()
		clientService := &httpClientServiceImpl{
			config:     ClientConfig{},
			httpClient: httpClient,
		}

		request, _ := http.NewRequest(http.MethodGet, "https://www.google.com/", nil)
		_, err := clientService.Do(request)
		if err != nil {
			t.Error("Expected nil but found error")
		}
	})

	t.Run("2. Status 200", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		httpClient, teardown := setup(h)
		defer teardown()
		clientService := &httpClientServiceImpl{
			config:     ClientConfig{},
			httpClient: httpClient,
		}

		request, _ := http.NewRequest(http.MethodGet, "https://www.google.com/", nil)
		_, err := clientService.Do(request)
		if err != nil {
			t.Error("Expected nil but found error")
		}
	})

}
