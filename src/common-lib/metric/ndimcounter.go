package metric

// NDIMCounterType : NDIMCounter Metric type
const NDIMCounterType = "NDIMCounter"

// CreateNDIMCounter : Create NDIMCounter type metric
var CreateNDIMCounter = func(name, description string, value map[string]int64) *NDIMCounter {
	return &NDIMCounter{Name: name, Description: description, DimCounters: value, Properties: map[string]string{}}
}

// Snapshot : Return current NDIMCounter Value
func (n *NDIMCounter) Snapshot() map[string]int64 {
	return n.DimCounters
}

// Clear : Clear NDIMCounter value
func (n *NDIMCounter) Clear() {
	for i := range n.DimCounters {
		n.DimCounters[i] = 0
	}
}

// Inc : Increase NDIMCounter Value
func (n *NDIMCounter) Inc(dimension string, value int64) {
	n.DimCounters[dimension] += value
}

// MetricType : Metric Type for counter
func (n *NDIMCounter) MetricType() string {
	return NDIMCounterType
}

// AddProperty : Add a property for counter
func (n *NDIMCounter) AddProperty(key, value string) {
	n.Properties[key] = value
}

// RemoveProperty : remove a property for counter
func (n *NDIMCounter) RemoveProperty(key string) {
	delete(n.Properties, key)
}

func (n *NDIMCounter) AddDimension(dimkey string, value int64) {
	n.DimCounters[dimkey] = value
}

func (n *NDIMCounter) RemoveDimension(dimkey string) {
	delete(n.DimCounters, dimkey)
}
