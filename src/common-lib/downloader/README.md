<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Downloader

This is a Standard implementation used by all the Go projects in the google to download a file from any network location. It also has support for mirror sites and retriable error codes.

### Third-Party Libraties

- [Grab](https://github.com/cavaliercoder/grab)
  - **License** [BSD 3-Clause "New" or "Revised" License](https://github.com/cavaliercoder/grab/blob/master/LICENSE)
  - **Description**
    - A download manager package for Go.
- [copier](https://github.com/jinzhu/copier)
  - **License** [BSD 3-Clause "New" or "Revised" License](https://github.com/jinzhu/copier/blob/master/License)
  - **Description**
    - Library is used to copy the content.

### Internal Packages

- [Client](src/http/client)
- [Checksum](src/checksum)

### [Example](grab/example/example.go)

**Import Statement**

```go
import	(
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/downloader/grab"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/http/client"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
)
```

**[Grab Downloader Instance](grab)**

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

**Download a File**

```go
response := service.Download(&downloader.Config{
	URL:              "http://cdn.itsupport247.net/InstallJunoAgent/Plugin/Windows/platform-installation-manager/1.0.216/platform_installation_manager_windows32_1.0.216.zip",
	DownloadLocation: "/home/juno/Desktop/test",
	FileName:         "platform_installation_manager_windows32_1.0.216.zip",
	TransactionID:    "1",
	CheckSumType:     checksum.MD5,
})
```

**Interface**

```go
	// Download - Download and Validate a file by using provided configuration
	// downloading an resource from provided @URL at @DownloadLocation with @FileName
	Download(conf *Config) *Response
```

**Response**

```go
// Response - An error struct to holds any error accured while downloading a file
type Response struct {
	// Error - Any error accured while downloading the required file
	Error error

	// ErrorCode - User defined errorcode, so that we can understand what went wrong
	ErrorCode string

	// Destination - Final download location for a file
	Destination string

	//Size - downloaded file size
	Size int64

	// StatusCode - HTTP status code received while downloading file
	StatusCode int

	// Config - Provided configuration for downloading a file
	Config *Config
}
```

**Configuration**

```go

// Config is a struct provides a download information to Download Service
type Config struct {
	// URL - Download service will be downloading files from this location
	URL string

	// FileName - Download file name; used while saving file on local machine
	FileName string

	// DownloadLocation - Local machine location for file to be saved
	DownloadLocation string

	// TransactionID - Transaction id for tracing
	TransactionID string

	// CheckSum - Checksum value to validate downloaded file
	CheckSum string

	// KeepOriginalName - Do we want to keep orignal file name?
	KeepOriginalName bool

	// CheckSumType - what kind on checksum needs to be applied while validating doenloaded file
	// Possible values are : MD5, SHA1, SHA256, NONE
	CheckSumType checksum.Type

	// UserAgent specifies the User-Agent string which will be set in the
	// headers of all requests made by this client.
	//
	// The user agent string may be overridden in the headers of each request.
	// Default: dwonloader
	UserAgent string

	// BufferSize specifies the size in bytes of the buffer that is used for
	// transferring all requested files. Larger buffers may result in faster
	// throughput but will use more memory and result in less frequent updates
	// to the transfer progress statistics. The BufferSize of each request can
	// be overridden on each Request object. Default: 32KB.
	BufferSize int

	// Header - Header to be passed while downloading a file from provided URL
	Header map[string]string

	// LoggerFunc : Logger instance used for logging
	// defaults to Discard
	LoggerFunc func() logger.Log

	// DownloadRetryCount : Download Retry Count
	DownloadRetryCount int

	// DownloadRetryDelay : Download Retry Delay
	DownloadRetryDelay time.Duration

	// MirrorSites : List of Mirror Sites specific config
	MirrorSites []MirrorSites

	// RetryableStatuses : List of Retryable Statuses if any
	// Default value will be HTTP status 404, 429
	// For Eg.
	// RetryableStatuses = map[int]bool{
	// http.StatusNotFound:        true,
	// http.StatusTooManyRequests: true,
    // }
	RetryableStatuses map[int]bool
}

// MirrorSites : Mirror Sites specific config
type MirrorSites struct {
	// MirrorURL : Mirror URL
	MirrorURL string
}

```

### Note

If we provide any mirror sites then the lib will try to download from the mirror and if it fails then it will directly download from the actual URL

### Contribution

Any changes in this package should be communicated to [Common-Frameworks](Common-Frameworks@gmail.com) team.
