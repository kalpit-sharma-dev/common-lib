<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# HTTP Communication

This is a Standard HTTP communication implementation used by all the Go projects in the google.
ignoreProxy - will decide do we need to include proxy settings while communnication or not

## Import Statement

```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/http/client"
)
```

**[Client Configuration](client)**

## Request builder Instance

```go
// API to create Request builder instance
http.CreateRequest(transactionID, method, url string) *Request
```

## Request Functions

```go
// SetBytes - Set request body using byte array
SetBytes(body []byte)

// SetInterface - Set request body by marshling object
SetInterface(body interface{}) error

// SetReader - Set request body using reader
SetReader(body io.Reader)

// AddHeader - Add header value in the request, if exists, it adds with duplicate key
AddHeader(key string, value string)

// SetHeader - Add header value in the request, if exists, replaces the value
SetHeader(key string, value string)

// Build - Convert request object into HTTP Request
Build() (*http.Request, error)

// Execute - Execute request using client and return a response object
Execute(cfg *client.Config, ignoreProxy bool) *Response
```

## Response Functions

```go
// IsProxyError - Identify waither response is a proxy error or not
IsProxyError() bool

// IsSuccess - Identify waither we received a success response or not
IsSuccess() bool

// HasBody - Identify waither response has body or not
HasBody() bool

// Ignore - Ignore response
Ignore() error

// Status - HTTP Response status
Status() int

// GetBytes - return Response body as byte array
GetBytes() ([]byte, error)

// GetInterface - return Response body as object
GetInterface(data interface{}) error

// GetReader - return Response body as an io.reader
// User need to close the read closer after reading body
GetReader() (io.ReadCloser, error)
```

## Helper Functions

```go
// IsProxyError - Identify waither response is a proxy error or not
IsProxyError(err error, res *http.Response) bool

// IsSuccess - Identify waither we received a success response or not
IsSuccess(res *http.Response) bool

// HasBody - Identify waither response has body or not
HasBody(res *http.Response) bool
```

## Errors
```go
// ErrBlankBody - Error if http response does not have a body
http.ErrBlankBody
```

### Contribution

Any changes in this package should be communicated to Juno Team.
