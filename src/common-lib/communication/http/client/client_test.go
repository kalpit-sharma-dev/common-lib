package client

import (
	"net/http"
	"testing"
	"time"
)

func Test_transport(t *testing.T) {
	t.Run("1. No Proxy", func(t *testing.T) {
		got := transport(&Config{DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5}, true)

		if got.Proxy != nil {
			t.Errorf("transport() = Proxy should be nil but got one")
		}

		if got.MaxIdleConns != 1 || got.IdleConnTimeout != time.Minute ||
			got.TLSHandshakeTimeout != (3*time.Second) ||
			got.ExpectContinueTimeout != (4*time.Second) ||
			got.MaxIdleConnsPerHost != 5 {
			t.Errorf("transport() = Configuration missmatch")
		}
	})

	t.Run("2. Has Proxy", func(t *testing.T) {
		got := transport(&Config{DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5, Proxy: Proxy{Address: "test.com", Protocol: "http", Port: 80}}, false)

		if got.Proxy == nil {
			t.Errorf("transport() = Proxy should exist but got nil")
		}

		if got.MaxIdleConns != 1 || got.IdleConnTimeout != time.Minute ||
			got.TLSHandshakeTimeout != (3*time.Second) ||
			got.ExpectContinueTimeout != (4*time.Second) ||
			got.MaxIdleConnsPerHost != 5 {
			t.Errorf("transport() = Configuration missmatch")
		}
	})
}

func TestBasic(t *testing.T) {
	t.Run("Transport TLS should be nil - Timeout in minute/default", func(t *testing.T) {
		got := Basic(&Config{DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5}, false)

		trn := got.Transport
		transport, ok := trn.(*http.Transport)
		if !ok {
			t.Error("TestBasic() = Invalid transport")
			return
		}

		if transport.TLSClientConfig != nil {
			t.Errorf("TestBasic() = TLSClientConfig should not exist but got %v", transport.TLSClientConfig)
		}

		if transport.Proxy == nil {
			t.Errorf("TestBasic() = Proxy should exist but got nil")
		}

		if transport.MaxIdleConns != 1 || transport.IdleConnTimeout != time.Minute ||
			transport.TLSHandshakeTimeout != (3*time.Second) ||
			transport.ExpectContinueTimeout != (4*time.Second) ||
			transport.MaxIdleConnsPerHost != 5 {
			t.Errorf("TestBasic() = Configuration missmatch")
		}
	})

	t.Run("Transport TLS should be nil - Timeout in Second", func(t *testing.T) {
		got := Basic(&Config{TimeoutMillisecond: 3000, DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5}, false)

		trn := got.Transport
		transport, ok := trn.(*http.Transport)
		if !ok {
			t.Error("TestBasic() = Invalid transport")
			return
		}

		if transport.TLSClientConfig != nil {
			t.Errorf("TestBasic() = TLSClientConfig should not exist but got %v", transport.TLSClientConfig)
		}

		if transport.Proxy == nil {
			t.Errorf("TestBasic() = Proxy should exist but got nil")
		}

		if transport.MaxIdleConns != 1 || transport.IdleConnTimeout != time.Minute ||
			transport.TLSHandshakeTimeout != (3*time.Second) ||
			transport.ExpectContinueTimeout != (4*time.Second) ||
			transport.MaxIdleConnsPerHost != 5 {
			t.Errorf("TestBasic() = Configuration missmatch")
		}
	})
}

func TestTLS(t *testing.T) {
	t.Run("Transport TLS should be nil - Timeout in Minute/default", func(t *testing.T) {
		got := TLS(&Config{DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5}, false)

		trn := got.Transport
		transport, ok := trn.(*http.Transport)
		if !ok {
			t.Error("TLS() = Invalid transport")
			return
		}

		if transport.TLSClientConfig == nil {
			t.Errorf("TLS() = TLSClientConfig should exist but got nil")
		}

		if transport.Proxy == nil {
			t.Errorf("TLS() = Proxy should exist but got nil")
		}

		if transport.MaxIdleConns != 1 || transport.IdleConnTimeout != time.Minute ||
			transport.TLSHandshakeTimeout != (3*time.Second) ||
			transport.ExpectContinueTimeout != (4*time.Second) ||
			transport.MaxIdleConnsPerHost != 5 {
			t.Errorf("TLS() =  Configuration missmatch")
		}
	})

	t.Run("Transport TLS should be nil - Timeout in Seconds", func(t *testing.T) {
		got := TLS(&Config{TimeoutMillisecond: 3000, DialTimeoutSecond: 2, DialKeepAliveSecond: 2, MaxIdleConns: 1,
			IdleConnTimeoutMinute: 1, TLSHandshakeTimeoutSecond: 3, ExpectContinueTimeoutSecond: 4,
			MaxIdleConnsPerHost: 5}, false)

		trn := got.Transport
		transport, ok := trn.(*http.Transport)
		if !ok {
			t.Error("TLS() = Invalid transport")
			return
		}

		if transport.TLSClientConfig == nil {
			t.Errorf("TLS() = TLSClientConfig should exist but got nil")
		}

		if transport.Proxy == nil {
			t.Errorf("TLS() = Proxy should exist but got nil")
		}

		if transport.MaxIdleConns != 1 || transport.IdleConnTimeout != time.Minute ||
			transport.TLSHandshakeTimeout != (3*time.Second) ||
			transport.ExpectContinueTimeout != (4*time.Second) ||
			transport.MaxIdleConnsPerHost != 5 {
			t.Errorf("TLS() =  Configuration missmatch")
		}
	})
}

func TestRedirect(t *testing.T) {
	t.Run("Max Redirect 11", func(t *testing.T) {
		client := &http.Client{}
		Redirect(client, map[string]string{"Key": "Value"})
		req := &http.Request{}
		via := []*http.Request{&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{},
			&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}}
		err := client.CheckRedirect(req, via)
		if err == nil {
			t.Error("Redirect() = Expected Error got nil")
		}
	})

	t.Run("Max Redirect 10", func(t *testing.T) {
		client := &http.Client{}
		Redirect(client, map[string]string{"Key": "Value"})
		req := &http.Request{}
		via := []*http.Request{&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{},
			&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}}
		err := client.CheckRedirect(req, via)
		if err == nil {
			t.Error("Redirect() = Expected Error got nil")
		}
	})

	t.Run("Redirect 9", func(t *testing.T) {
		req := &http.Request{Header: http.Header{}}
		client := &http.Client{}
		Redirect(client, map[string]string{"Key": "Value"})
		via := []*http.Request{&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}, &http.Request{},
			&http.Request{}, &http.Request{}, &http.Request{}, &http.Request{}}
		err := client.CheckRedirect(req, via)
		if err != nil {
			t.Error("Redirect() = Expected Error got nil")
		}

		value := req.Header.Get("Key")
		if value != "Value" {
			t.Errorf("Redirect() = Expected : Value but got %v", value)
		}
	})
}
