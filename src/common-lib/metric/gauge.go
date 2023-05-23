package metric

// GaugeType : Gauge Metric type
const GaugeType = "Gauge"

// CreateGauge : Create Gauge metric object
var CreateGauge = func(name, description string, value int64) *Gauge {
	return &Gauge{Name: name, Description: description, Value: value, Properties: map[string]string{}}
}

// Snapshot : Current Gauge value
func (g *Gauge) Snapshot() int64 {
	return g.Value
}

// Clear : Clear Gauge value
func (g *Gauge) Clear() {
	g.Value = 0
}

// Inc : Increase Gauge value
func (g *Gauge) Inc(value int64) {
	g.Value += value
}

// Dec : Decrease Gauge value
func (g *Gauge) Dec(value int64) {
	g.Value -= value
}

// MetricType : Metric Type for Gauge
func (g *Gauge) MetricType() string {
	return GaugeType
}

// AddProperty : Add a property for Gauge
func (g *Gauge) AddProperty(key, value string) {
	g.Properties[key] = value
}

// RemoveProperty : remove a property for Gauge
func (g *Gauge) RemoveProperty(key string) {
	delete(g.Properties, key)
}
