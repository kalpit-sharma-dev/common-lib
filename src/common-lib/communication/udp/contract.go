package udp

import "context"

//go:generate mockgen -package mock -destination=mock/mocks.go . Server

// Server - An interface to hold all UDP Server related functions
type Server interface {
	// Receive : Receive message over UDP
	Receive(handler RequestHandler) error

	// Shutdown - Graceful shutdown support for UDP communication
	Shutdown(ctx context.Context) error
}

// Request :  UDP request
type Request struct {
	// Size - Received bytes
	Size int

	// Address - Client IP-Address
	Address string

	// Body - Received Message Content
	Body []byte

	// Err - Processing Error
	Err error
}

// Response : Struct to hold UDP response
type Response struct {
	// Size - Received bytes
	Size int

	// Address - Client IP-Address
	Address string

	// Body - Received Message Content
	Body []byte

	// Err - Processing Error
	Err error
}
