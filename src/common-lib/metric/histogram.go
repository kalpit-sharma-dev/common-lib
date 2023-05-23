package metric

// HistogramType : Histogram metric type
const HistogramType = "Histogram"

// CreateHistogram : Create histogram metric object
var CreateHistogram = func(name, description string, values []float64) *Histogram {
	return &Histogram{Name: name, Description: description, Values: values, Properties: map[string]string{}}
}

// Snapshot : Current Gauge value
func (h *Histogram) Snapshot() []float64 {
	return h.Values
}

// Update : Update Histogram values
func (h *Histogram) Update(values []float64) {
	h.Values = values
}

// Clear : Clear Histogram values
func (h *Histogram) Clear() {
	h.Values = []float64{}
}

// MetricType : Metric Type for Histogram
func (h *Histogram) MetricType() string {
	return HistogramType
}

// AddProperty : Add a property for counter
func (h *Histogram) AddProperty(key, value string) {
	h.Properties[key] = value
}

// RemoveProperty : remove a property for counter
func (h *Histogram) RemoveProperty(key string) {
	delete(h.Properties, key)
}
