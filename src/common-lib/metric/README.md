<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Metric

This is a Standard Metric collector implementation used by all the Go projects in the Continuum for Application metric collection.

### Libraties

- [sanitize](../sanitize")
  - **License** Internal
  - **Description** - Package sanitize provides functions for sanitizing text.

### [Example](example/example.go)

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/metric"
```

**Configuration**

```go
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
```

**Default Configuration Object**

```go
// New - Default configuration object having default values
var New = func() *Config {
	return &Config{
		Communication: udp.New(),
		Namespace:     "",
		CurrentEnv:    "",
	}
}
```

**Collector Functions**

```go
// CreateCounter : Create Counter type metric
CreateCounter (name, description string, value int64) *Counter

// CreateGauge : Create Gauge metric object
CreateGauge (name, description string, value int64) *Gauge

// CreateHistogram : Create histogram metric object
CreateHistogram(name, description string, values []float64) *Histogram

// CreateEvent : Create Event metric object
CreateEvent(title string, description string) *Event

// Publish : Publish Metric
Publish(cfg *Config, collector ...Collector) error

// PeriodicPublish : Periodically Publish  Metric
PeriodicPublish(duration time.Duration, cfg *Config, callback func() []Collector, handler func(err error))
```

**Counter Functions**

```go
// Snapshot : Return current Counter Value
Snapshot() int64

// Clear : Clear Counter value
Clear()

// Inc : Increase Counter Value
Inc(value int64)

// MetricType : Metric Type for counter
MetricType() string

// AddProperty : Add a property for counter
AddProperty(key, value string)

// RemoveProperty : remove a property for counter
RemoveProperty(key string)
```

**Gauge Functions**

```go
// Snapshot : Current Gauge value
Snapshot() int64

// Clear : Clear Gauge value
Clear()

// Inc : Increase Gauge value
Inc(value int64)

// Dec : Decrease Gauge value
Dec(value int64)

// MetricType : Metric Type for Gauge
MetricType() string

// AddProperty : Add a property for Gauge
AddProperty(key, value string)

// RemoveProperty : remove a property for Gauge
RemoveProperty(key string)
```

**Histogram Functions**

```go
// Snapshot : Current Gauge value
Snapshot() []float64

// Update : Update Histogram values
Update(values []float64)

// Clear : Clear Histogram values
Clear()

// MetricType : Metric Type for Histogram
MetricType() string

// AddProperty : Add a property for counter
AddProperty(key, value string)

// RemoveProperty : remove a property for counter
RemoveProperty(key string)
```

**Event Functions**

```go
// StartEvent : Set start time for an Event
StartEvent(date time.Time)

// EndEvent : Set end time for an Event
EndEvent(date time.Time)

// MetricType : Metric Type for counter
MetricType() string

// AddProperty : Add a property for event
AddProperty(key, value string)

// RemoveProperty : remove a property for event
RemoveProperty(key string)
```

### Contribution

Any changes in this package should be communicated to Juno Team.
