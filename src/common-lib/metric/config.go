package metric

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
)

// Config - Holds all the configuration for Metric object Publishing
type Config struct {
	// Communication - UDP Communication configuration
	// default - Communication : Default UDP Config
	Communication *udp.Config

	// Namespace - Namespace of a Metric collector for unique identification
	// Default - value is <HostName>
	Namespace string

	// CurrentEnv added to have ability to pass this info from services
	// useful for qa and int metrics - since they live on a single dynatrace tenant
	// this is a way to split metrics data
	// Default value is empty
	CurrentEnv string
}

// New - Default configuration object having default values
// values - Address: "localhost", PortNumber: "7000", Namespace : ""
var New = func() *Config {
	return &Config{
		Communication: udp.New(),
		Namespace:     "",
		CurrentEnv:    "",
	}
}

// GetNamespace - Return service name space for metric
func (c *Config) GetNamespace() string {
	ns := c.Namespace
	if ns == "" {
		ns = util.Hostname(util.ProcessName())
	}
	return ns
}
