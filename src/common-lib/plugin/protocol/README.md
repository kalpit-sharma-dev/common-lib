<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Plugin - Protocol

Common lib wrapper module for protocol usage

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/protocol"
```

**Functions**

```go
TransactionID(req *Request) string    //TransactionID is a function to return transaction id from the protocole request
```


```go
NewResponse() (res *Response)    //NewResponse constructs a response object
```


```go
NewRequest() (req *Request)    //NewRequest constructs and returns a Request
```

```go
CreateSuccessResponse(outBytes []byte, brokerPath string, hdrPersistData string, status ResponseStatus, transactionID string) *Response    //CreateSuccessResponse create success response body
```

```go
CreateErrorResponse(status ResponseStatus, brokerPath string, err error) *Response    //CreateErrorResponse create error response body
```

```go
SetError(code ResponseStatus, rootCause error)    //SetError sets up error information on the response object.
```

```go
SetKeyValue(key HeaderKey, value string)    //SetKeyValue sets a key to a value (overwriting if it exists)
```

```go
SetKeyValues(key HeaderKey, values []string)    //SetKeyValue sets a key to a value (overwriting if it exists)
```

```go
GetKeyValue(key HeaderKey) (value string)    //GetKeyValue returns the value for a given key
```

### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
