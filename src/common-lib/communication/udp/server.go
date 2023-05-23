package udp

import (
	"context"
	"fmt"
	"net"
)

type serverImpl struct {
	conf     *Config
	shutdown chan bool
	listner  *net.UDPConn
}

// NewServer - Create a New UDP server Instance
var NewServer = func(conf *Config) Server {
	return &serverImpl{conf: conf, shutdown: make(chan bool, 1)}
}

// RequestHandler - request handler function
type RequestHandler func(request Request) []byte

// Receive : Receive message over UDP
func (s *serverImpl) Receive(handler RequestHandler) error {
	defer func() {
		if r := recover(); r != nil {
			handler(Request{Err: fmt.Errorf("receive: Recovered from %s", r)})
		}
	}()
	return s.receive(handler)
}

func (s *serverImpl) receive(handler RequestHandler) error {
	serverAddress := s.conf.ServerAddress()
	address, err := net.ResolveUDPAddr(Network, serverAddress)

	if err != nil {
		return err
	}

	listner, err := net.ListenUDP(Network, address)

	if err != nil {
		return err
	}

	s.listner = listner

	for {
		select {
		case <-s.shutdown:
			close(s.shutdown)
			return nil
		default:
			s.handle(listner, handler)
		}
	}
}

func (s *serverImpl) handle(conn *net.UDPConn, handler RequestHandler) {
	defer func() {
		if r := recover(); r != nil {
			handler(Request{Err: fmt.Errorf("receive:handle: Recovered from %s", r)})
		}
	}()

	buffer := make([]byte, 1024)
	size, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		handler(Request{Err: err})
		return
	}

	address := ""
	if addr != nil {
		address = addr.String()
	}

	data := buffer[:size]

	message := handler(Request{Size: size, Address: address, Body: data, Err: err})
	if message != nil {
		_, err = conn.WriteToUDP(message, addr)
		if err != nil {
			handler(Request{Err: err})
		}
	}
}

// Shutdown - Graceful shutdown support for UDP communication
func (s *serverImpl) Shutdown(ctx context.Context) error {
	if s.listner != nil {
		err := make(chan error, 1)
		go func() {
			s.shutdown <- true
			err <- s.listner.Close()
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-err:
			return e
		}
	}
	return fmt.Errorf("shutdown:Server not initialized for %s", s.conf.ServerAddress())
}
