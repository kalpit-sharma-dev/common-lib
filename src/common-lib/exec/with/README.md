<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# With

Helper functions to execute functions

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exec/with"
```

Recover is a function to call any function with recovery and gives callback to @Handler by passing Error Message and Stack Trace

```go
Recover(name string, transaction string, fn func(), handler func(transaction string, err error))
```

Context - Helper function to Execute function with context, so that GO-routine creation can be avoided in MS

```go
Context(ctx context.Context, name string, transaction string, fn func() error) error
```

### Contribution

Any changes in this package should be communicated to Common Frameworks.
