<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# runtime/util

Helper utility functions to find commonly used values

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
```

ProcessName is a function to return process name for a binary

```go
ProcessName() string
```

InvocationPath is a function to return Invocation path for a binary

```go
InvocationPath() string
```

Hostname is a function to return Hostname for a machine; in case of error it sends default value

```go
Hostname(defaultValue string) string
```

LocalIPAddress returns the non loopback local IP Address of the host, returns blank in case of error

```go
LocalIPAddress() []string
```

NotifyStopSignal - Function to execute callback on reciving a quit signal

```go
NotifyStopSignal(stop <-chan bool, callback func()) error
```

### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
