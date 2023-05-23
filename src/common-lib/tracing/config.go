package tracing

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

// Config for tracing
type Config struct {
	// Type represents the tracing type (e.g. "xray")
	Type Type
	// HostPlatform represents the service hosting platform ("ecs" | "ec2")
	HostPlatform Type
	// Enabled tells the client whether tracing is enabled or not
	Enabled bool
	// Address is the tracing daemon's address
	Address string
	// ServiceName is the application's name that served the tracing request
	ServiceName string
	// ServiceVersion is the application's version that served the tracing request
	ServiceVersion string
}

// NewConfig returns Config with default values
func NewConfig() *Config {
	// As Configure(config *Config) method is not compulsory we do not have centeralized
	// place to set the environment variable because of that setting environment variable here
	handleTracingStatus(true)
	return &Config{
		Type:           AwsXrayTracingType,
		HostPlatform:   "",
		Enabled:        true,
		ServiceName:    utils.GetServiceName(),
		ServiceVersion: utils.GetServiceVersion(),
	}
}
