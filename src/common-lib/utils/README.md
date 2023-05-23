<p  align="center">

<img  height=70px  src="docs/images/logo.png">

<img  height=70px  src="docs/images/Go-Logo_Blue.png">

</p>

  

# utils

Helper utility functions to perform common operations

**Import Statement**

```go

import  "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"

```
--------------------------------------- 

`ToString()` converts an interface type to string
```go
ToString(v interface{}) string
```
`ToTime` converts an interface type to time  
```go
ToTime(v interface{}) time.Time
```
`ToInt64` converts an interface type to `int64`. `interface{}` holding an `int` will not be type casted to `int64` and will return 0 as the result  
```go
ToInt64(v interface{}) int64
```
`ToInt` converts an interface type to int
```go
ToInt(v interface{}) int
```
`ToFloat64` converts an interface type to `float64`
```go
ToFloat64(v interface{}) float64
```
`ToBool` converts an interface type to bool
```go
ToBool(v interface{}) bool
```
`ToStringArray` converts an interface type to string array
```go
ToStringArray(v interface{}) []string
```
`ToStringMap` converts an interface type to map[string]string
```go
ToStringMap(v interface{}) map[string]string
```
`GetTransactionID` generates new transaction id. The [transaction id](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/tree/master/src/runtime/logger#note) is UUID, which is used to track business transaction in logs.
```go
GetTransactionID() string
```
`GetTransactionIDFromResponse` retrieves transaction id from the http response header
```go
GetTransactionIDFromResponse(res *http.Response) string
```
`GetTransactionIDFromRequest` retrieves transaction if from the http request header
```go
GetTransactionIDFromRequest(req *http.Request) string
```
`GetQueryValuesFromRequest` to get query values from request for given filter
```go
GetQueryValuesFromRequest(req *http.Request, filter string) []string
```
`GetValueFromRequestHeader` retrieves header value for given key from the http request header. `protocol.HeaderKey` is string type.
```go
GetValueFromRequestHeader(req *http.Request, key protocol.HeaderKey) string
```
`GetChecksum` is a function to calculate MD5 hash value of message
```go
GetChecksum(message []byte) string
```
`ValidateMessage` checks if message is corrupted or not, returns calculated checksum
```go
ValidateMessage(message []byte, receievedChecksum string) (bool, string)
```
`DetermineErrorCodePair` determines the error code pair based upon the error message
```go
DetermineErrorCodePair(errMsg string) (mainError, subError string)
```
`EqualFold` is to check if two strings array are same with case insensitive
```go 
EqualFold(source, target []string) bool
### Contribution
Any changes in this package should be communicated to Common Frameworks Team.
```
`Difference` is to return differences between any two entities
```go 
Difference(x, y interface{}) []Change