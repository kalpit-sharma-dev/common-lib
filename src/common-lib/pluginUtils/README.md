<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Plugin Utils

Common lib wrapper module for plugin utils usage

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/pluginUtils"
```

**Functions**

```go
GetReader() io.Reader    //GetReader is an implementation of interface IOReaderFactory and returns io.Reader
```

```go
GetWriter() io.Writer    //GetWriter is an implementation of interface IOWriterFactory and returns io.Writer
```

```go
Is64BitOS() bool    //Is64BitOS verifies os architecture
```

```go
DisableRedirection(OldValue *uintptr)    //DisableRedirection disables file system redirection for the calling thread. File system redirection is enabled by default.
```

```go
RevertRedirection(OldValue *uintptr)    //RevertRedirection restores file system redirection for the calling thread.
```

### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
