package utils

import (
	"os"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
)

const (
	// ServiceNameEnv service name environment variable
	ServiceNameEnv = "SERVICE_NAME"
	// ServiceVersionEnv service version environment variable
	ServiceVersionEnv = "SERVICE_VERSION"
)

// GetServiceName gets service name
// Tries SERVICE_NAME environment variable first, then process name
func GetServiceName() string {
	return GetEnvVar(ServiceNameEnv, util.ProcessName())
}

// GetServiceVersion gets service version via SERVICE_VERSION environment variable
func GetServiceVersion() string {
	return GetEnvVar(ServiceVersionEnv, "")
}

// GetEnvVar returns environment variable value or default provided
func GetEnvVar(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
