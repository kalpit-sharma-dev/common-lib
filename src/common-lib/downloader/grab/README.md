<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Downloader

This is a Standard implementation used by all the Go projects in the google to download a file from any network location.

### Third-Party Libraties

- [Grab](https://github.com/cavaliercoder/grab)
  - **License** [BSD 3-Clause "New" or "Revised" License](https://github.com/cavaliercoder/grab/blob/master/LICENSE)
  - **Description**
    - A download manager package for Go.

### Internal Packages

- [Client](src/http/client)
- [Checksum](src/checksum)

### [Example](example/example.go)

**Import Statement**

```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader/grab"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
)
```

**Grab downloader Instance**

```go
service := grab.GetDownloader(&client.Config{
		MaxIdleConns:                100,
		MaxIdleConnsPerHost:         10,
		IdleConnTimeoutMinute:       1,
		TimeoutMinute:               1,
		DialTimeoutSecond:           100,
		DialKeepAliveSecond:         100,
		TLSHandshakeTimeoutSecond:   100,
		ExpectContinueTimeoutSecond: 100,
	})
```

### Contribution

Any changes in this package should be communicated to [Common-Frameworks](Common-Frameworks@gmail.com) team.
