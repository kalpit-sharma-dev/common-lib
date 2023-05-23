package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"

	ieproxy "github.com/mattn/go-ieproxy"
)

func transport(config *Config, ignoreProxy bool) *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(config.DialTimeoutSecond) * time.Second,
			KeepAlive: time.Duration(config.DialKeepAliveSecond) * time.Second,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		IdleConnTimeout:       time.Duration(config.IdleConnTimeoutMinute) * time.Minute,
		TLSHandshakeTimeout:   time.Duration(config.TLSHandshakeTimeoutSecond) * time.Second,
		ExpectContinueTimeout: time.Duration(config.ExpectContinueTimeoutSecond) * time.Second,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
	}

	if config.UseIEProxy {
		transport.Proxy = ieproxy.GetProxyFunc()
	} else if !ignoreProxy {
		transport.Proxy = proxy(config)
	}
	return transport
}

func proxy(config *Config) func(*http.Request) (*url.URL, error) {
	proxy := http.ProxyFromEnvironment
	if config.Proxy.Address != "" {
		// Address should be in the following format : http://127.0.0.1:9999
		address := fmt.Sprintf("%s://%s:%d", config.Proxy.Protocol, config.Proxy.Address, config.Proxy.Port)
		proxyURL, err := url.Parse(address)
		proxyURL.User = url.UserPassword(config.Proxy.UserName, config.Proxy.Password)
		if err == nil {
			proxy = http.ProxyURL(proxyURL)
		}
	}
	return proxy
}

// Basic - Create a basic http client without TLS configuration
// ignoreProxy - Should we ignore proxy configuration while creating client?
func Basic(config *Config, ignoreProxy bool) *http.Client {
	client := &http.Client{}
	if config.TimeoutMillisecond > 0 {
		client.Timeout = time.Duration(config.TimeoutMillisecond) * time.Millisecond
	} else {
		client.Timeout = time.Duration(config.TimeoutMinute) * time.Minute
	}
	client.Transport = transport(config, ignoreProxy)

	instrumentedClient, err := tracing.HTTPClient(config.TracingConfig, client)
	if err != nil {
		return client
	}
	return instrumentedClient
}

// TLS - Create a http client having TLS configuration
// ignoreProxy - Should we ignore proxy configuration while creating client?
func TLS(config *Config, ignoreProxy bool) *http.Client {
	transport := transport(config, ignoreProxy)
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify:     !config.ValidateSSLCertificate,
		SessionTicketsDisabled: false,
		ClientSessionCache:     tls.NewLRUClientSessionCache(1),
	}

	client := &http.Client{}
	if config.TimeoutMillisecond > 0 {
		client.Timeout = time.Duration(config.TimeoutMillisecond) * time.Millisecond
	} else {
		client.Timeout = time.Duration(config.TimeoutMinute) * time.Minute
	}
	client.Transport = transport

	instrumentedClient, err := tracing.HTTPClient(config.TracingConfig, client)
	if err != nil {
		return client
	}
	return instrumentedClient
}

// ErrRedirectLoop ...
var ErrRedirectLoop = errors.New("stopped after 10 redirects")

// Redirect ...
func Redirect(c *http.Client, header map[string]string) {
	if c != nil {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// This this used for maximum number of redirect / hops are allowed
			// for an URL during communication
			if len(via) >= 10 {
				return ErrRedirectLoop
			}

			for key, value := range header {
				req.Header.Set(key, value)
			}
			return nil
		}
	}
}
