<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Plugin - Windows Management Instrumentation

Common lib wrapper module for wmi usage

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/protocol/wmi"
```

**Functions**

```go
GetWrapper() StackExchangeWMI    //GetWrapper returns the implementation of WMI Wrapper
```


```go
Query(query string, dst interface{}, connectServerArgs ...interface{}) error    //Query to execute the WMI query and wrap the output to the dst type(struct) which must be passed as reference
```


```go
CreateQuery(src interface{}, where string) string    //CreateQuery is to create the WMI query based on the src type
```


### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
