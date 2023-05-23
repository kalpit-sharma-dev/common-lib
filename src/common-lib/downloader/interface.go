package downloader

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . Service

var log logger.Log

// Service is an interface used for downloading an resource from provided @URL at @DownloadLocation with @FileName
type Service interface {
	// Download - Download and Validate a file by using provided configuration
	Download(conf *Config) *Response
}

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

	//MirrorFailure - Flag provided to indicate download failure
	MirrorFailure bool

	//MirrorFailureError - provide download failure error
	MirrorFailureError error
}
