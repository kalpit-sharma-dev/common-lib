<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# HTTP Client

This is a Standard HTTP client implementation used by all the Go projects in the google.
ignoreProxy - will decide do we need to include proxy settings while communnication or not

### [Example](example/example.go)

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
```

**Client Interfaces**

```go
// API to create TLS client
client.TLS

// API to create Basic HTTP client
client.Basic
```

**Configuration**

```go

//Config - Http Client Configuration for connection
type Config struct {
	// Timeout specifies a time limit for requests made by this
	// Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	//
	// If both TimeoutMinute and TimeoutMillisecond are zero, it means no timeout
	//
	// The Client cancels requests to the underlying Transport
	// as if the Request's Context ended.
	//
	// For compatibility, the Client will also use the deprecated
	// CancelRequest method on Transport if found. New
	// RoundTripper implementations should use the Request's Context
	// for cancelation instead of implementing CancelRequest.
	TimeoutMinute int `json:"timeoutMinute"`

	// Takes timeout value in milliseconds, if this property is set
	// timeout will be considered in milliseconds and TimeoutMinute will
	// be ignored
	// If both TimeoutMinute and TimeoutMillisecond are zero, it means no timeout
	TimeoutMillisecond int `json:"timeoutMillisecond"`

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `json:"maxIdleConns"`

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost"`

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	//
	// For HTTP/2, this currently only controls the number of new
	// connections being created at a time, instead of the total
	// number. In practice, hosts using HTTP/2 only have about one
	// idle connection, though.
	MaxConnsPerHost int `json:"maxConnsPerHost"`

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeoutMinute int `json:"idleConnTimeoutMinute"`

	// DialTimeoutSecond is the maximum amount of time a dial will wait for
	// a connect to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When using TCP and dialing a host name with multiple IP
	// addresses, the timeout may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are`
	// often around 3 minutes.
	DialTimeoutSecond int `json:"dialTimeoutSecond"`

	// DialKeepAliveSecond specifies the keep-alive period for an active
	// network connection.
	// If zero, keep-alives are enabled if supported by the protocol
	// and operating system. Network protocols or operating systems
	// that do not support keep-alives ignore this field.
	// If negative, keep-alives are disabled.
	DialKeepAliveSecond int `json:"dialKeepAliveSecond"`

	// ExpectContinueTimeoutSecond specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeoutSecond int `json:"TLSHandshakeTimeoutSecond"`

	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeoutSecond int `json:"expectContinueTimeoutSecond"`

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	//
	// The proxy type is determined by the URL scheme. "http",
	// "https", and "socks5" are supported. If the scheme is empty,
	// "http" is assumed.
	//
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	Proxy Proxy `json:"proxy"`

	// UseIEProxy if set to true, will retrieve and communicate by using IE Proxy, which
	// is windows specific
	UseIEProxy bool `json:"useIEProxy"`
}
```

Proxy - determined by the URL scheme. "http", "https", and "socks5" are supported. If the scheme is empty, "http" is assumed.
If Proxy is nil or returns a nil \*URL, no proxy is used.

```go
type Proxy struct {
	// Address - Proxy server IP Address
	Address string `json:"address"`

	// Port - Proxy server port
	Port int `json:"port"`

	// UserName - User name to login
	UserName string `json:"userName"`

	// Password - Passward for login
	Password string `json:"password"`

	// Protocol - Proxy server protocol
	Protocol string `json:"protocol"`
}
```

### Contribution

Any changes in this package should be communicated to Juno Team.
