package udp

const (
	// Network : Network Name for UDP communication
	Network = "udp"
)

//Config - Holds UDP communication configurations
type Config struct {
	// Address - Hosted UDP server IP Address
	// Default value localhost
	Address string

	// PortNumber - Hosted UDP server port Number
	// Defaut port 7000
	PortNumber string

	// TimeoutInSeconds - Communication timeout between Client and Server
	// Default Timeout 30 Second
	TimeoutInSeconds int64
}

// New - Create a configuration object having default values
// Values are - Address: "localhost", PortNumber: "7000", TimeoutInSeconds : 10
var New = func() *Config {
	return &Config{
		Address:          "localhost",
		PortNumber:       "7000",
		TimeoutInSeconds: 10,
	}
}

// ServerAddress : UDP Communication Address
func (c *Config) ServerAddress() string {
	return c.Address + ":" + c.PortNumber
}
