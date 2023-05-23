package with

import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

// Log - Logger implementation for logging
var Log logger.Log

func log() logger.Log {
	if Log != nil {
		return Log
	}
	l, _ := logger.Create(logger.Config{Name: "ExecuteWith", LogLevel: logger.OFF, Destination: logger.DISCARD}) //nolint
	return l
}
