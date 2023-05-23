package downloader

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/checksum"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	defaultDownloadRetry = 1
)

// defaultRetryableStatuses indicates retryable HTTP responses
var defaultRetryableStatuses = map[int]bool{
	http.StatusNotFound:        true,
	http.StatusTooManyRequests: true,
}

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

	generatedFileName string

	// DownloadRetryCount : Download Retry Count
	DownloadRetryCount int

	// DownloadRetryDelay : Download Retry Delay
	DownloadRetryDelay time.Duration

	// MirrorSites : Mirror Sites specific config
	MirrorSites []MirrorSites

	// RetryableStatuses : List of Retryable Statuses if any
	// Default value will be HTTP status 404, 429
	RetryableStatuses map[int]bool
}

// MirrorSites : Mirror Sites specific config
type MirrorSites struct {
	// MirrorURL : Mirror URL
	MirrorURL string
}

// NewConfig : Initialize default config
func NewConfig() *Config {
	return &Config{
		DownloadRetryCount: defaultDownloadRetry,
		RetryableStatuses:  defaultRetryableStatuses,
		UserAgent:          "downloader",
		BufferSize:         32 * 1024,
	}
}

// GetDownloadRetryCount - Provide Download Retry Count
func (c *Config) GetDownloadRetryCount() int {
	if c.DownloadRetryCount == 0 {
		return defaultDownloadRetry
	}
	return c.DownloadRetryCount
}

// GetRetryableStatuses - Provide Retryable Statuses
func (c *Config) GetRetryableStatuses() map[int]bool {
	if c.RetryableStatuses == nil {
		return defaultRetryableStatuses
	}
	return c.RetryableStatuses
}

// Logger - Provides logger implementation; default logger is Discard
func (c *Config) Logger() logger.Log {
	if c == nil || c.LoggerFunc != nil {
		return c.LoggerFunc()
	}

	loggerName := "downloader-service"
	if log == nil {
		logger.Create(logger.Config{Name: loggerName, Destination: logger.DISCARD}) // nolint
	}
	return logger.GetViaName(loggerName)
}

// UsrAgent - User Agent
func (c *Config) UsrAgent() string {
	if c == nil || c.UserAgent == "" {
		return "downloader"
	}
	return c.UserAgent
}

// BuffSize - BufferSize
func (c *Config) BuffSize() int {
	if c == nil || c.BufferSize == 0 {
		return 32 * 1024
	}
	return c.BufferSize
}

// Destination - Generates destination based on the
// Download location, File Name and KeepOriginalName
// If KeepOriginalName is false; new file name will look like platform
func (c *Config) Destination() string {
	return c.DownloadLocation + string(os.PathSeparator) + c.GenerateFileName()
}

// GenerateFileName - Generates file name based on the
// File Name and KeepOriginalName
// If KeepOriginalName is false; new file name will look like platform
func (c *Config) GenerateFileName() string {
	if c.generatedFileName == "" {
		c.generatedFileName = "."
		if c != nil && c.FileName != "" {
			c.generatedFileName = c.FileName
		} else if c != nil && !c.KeepOriginalName {
			c.generatedFileName = fmt.Sprintf("platform%v%s", time.Now().UTC().UnixNano(), c.FileName)
		}
	}
	return c.generatedFileName
}
