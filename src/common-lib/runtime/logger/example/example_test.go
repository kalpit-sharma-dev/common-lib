package main

import (
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func Test_main_function(t *testing.T) {
	t.Run("main function", func(t *testing.T) {
		logger.Create(logger.Config{Name: "Logger-1", MaxSize: 1, Destination: logger.DISCARD, LogFormat: logger.TextFormat})
		logger.Create(logger.Config{Name: "Logger-2", MaxSize: 1, Destination: logger.DISCARD, LogFormat: logger.JSONFormat})
		logger.Create(logger.Config{Name: "Logger-3", MaxSize: 1, Destination: logger.DISCARD, LogFormat: logger.JSONFormat})
		main()
	})

}
