<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# UDP
This is a Standard UDP communication implementation used by all the Go projects in the Continuum.

### [Example](example/example.go)

**Import Statement**
```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
```

**Configuration**
```go
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
```

**Default Configuration Object**
```go
// New - Create a configuration object having default values
// Values are - Address: "localhost", PortNumber: "7000", TimeoutInSeconds : 10
var New = func() *Config {
	return &Config{
		Address:          "localhost",
		PortNumber:       "7000",
		TimeoutInSeconds: 10,
	}
}
```

**Send Message Over UDP**
```go
// Response : Struct to hold UDP response
type Response struct {
	// Size - Recieved bytes
	Size int

	// Address - Client IP-Address
	Address string

	// Body - Recieved Message Content
	Body []byte

	// Err - Processing Error
	Err error
}

type responseHandler func(response Response) error

// Send : Send message over UDP
Send(conf *Config, message []byte, handler responseHandler) error
```

**Recieve Message Over UDP**
```go

// Request :  UDP request
type Request struct {
	// Size - Recieved bytes
	Size int

	// Address - Client IP-Address
	Address string

	// Body - Recieved Message Content
	Body []byte

	// Err - Processing Error
	Err error
}

type requestHandler func(request Request) []byte

// Recieve : Recieve message over UDP
Recieve(conf *Config, handler requestHandler) error {
```

### Contribution
Any changes in this package should be communicated to Juno Team.
