package distributed

import (
	"context"
	"sync"
)

type (
	// Event type for event
	Event struct {
		Type    string
		Payload interface{}
	}

	// BroadcastHandler type for event func
	BroadcastHandler func(e *Event)

	// Broadcast define methods for broadcast
	Broadcast interface {
		// AddHandler adding new handler for listening
		AddHandler(name string, handler BroadcastHandler)
		// Listen listening input events
		Listen(ctx context.Context, wg *sync.WaitGroup)
		// CreateEvent creating the new event and send to all subscribers
		CreateEvent(e Event) error
	}
)
