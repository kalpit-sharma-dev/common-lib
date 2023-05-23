package metric

// CounterType : Counter Metric type
const CounterType = "Counter"

// CreateCounter : Create Counter type metric
var CreateCounter = func(name, description string, value int64) *Counter {
	return &Counter{Name: name, Description: description, Value: value, Properties: map[string]string{}}
}

// Snapshot : Return current Counter Value
func (c *Counter) Snapshot() int64 {
	return c.Value
}

// Clear : Clear Counter value
func (c *Counter) Clear() {
	c.Value = 0
}

// Inc : Increase Counter Value
func (c *Counter) Inc(value int64) {
	c.Value += value
}

// MetricType : Metric Type for counter
func (c *Counter) MetricType() string {
	return CounterType
}

// AddProperty : Add a property for counter
func (c *Counter) AddProperty(key, value string) {
	c.Properties[key] = value
}

// RemoveProperty : remove a property for counter
func (c *Counter) RemoveProperty(key string) {
	delete(c.Properties, key)
}
