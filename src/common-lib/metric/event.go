package metric

import "time"

// EventMetricType : Gauge Metric type
const EventMetricType = "Event"

// CreateEvent : Create Event metric object
func CreateEvent(title string, description string) *Event {
	return &Event{Title: title, Description: description, Properties: map[string]string{}}
}

// StartEvent : Set start time for an Event
func (e *Event) StartEvent(date time.Time) {
	e.Start = date.UTC().UnixNano()
}

// EndEvent : Set end time for an Event
func (e *Event) EndEvent(date time.Time) {
	e.End = date.UTC().UnixNano()
}

// MetricType : Metric Type for counter
func (e *Event) MetricType() string {
	return EventMetricType
}

// AddProperty : Add a property for event
func (e *Event) AddProperty(key, value string) {
	e.Properties[key] = value
}

// RemoveProperty : remove a property for event
func (e *Event) RemoveProperty(key string) {
	delete(e.Properties, key)
}
