package cassandraorm

// All events type
const (
	EventBeforeAdd EventType = iota
	EventAfterAdd
	EventBeforeUpdate
	EventAfterUpdate
	EventBeforeDelete
	EventAfterDelete
)

type (
	// EventType describe type for events
	EventType int

	// Observer used for notify on add/update/delete events
	Observer interface {
		OnNotify(eventType EventType, args ...interface{})
	}

	// DefaultObserver is default (empty) implementation of Observer
	DefaultObserver struct{}
)

// OnNotify implements Observer
func (do *DefaultObserver) OnNotify(_ EventType, _ ...interface{}) {

}
