package udp

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	got := NewServer(&Config{TimeoutInSeconds: 1, Address: "a.b.c.d"})
	_, ok := got.(*serverImpl)
	if !ok {
		t.Error("Invalid serviceImpl")
	}
}

func Test_serverImpl_Receive(t *testing.T) {
	type fields struct {
		conf     *Config
		shutdown chan bool
		listner  *net.UDPConn
	}
	type args struct {
		handler RequestHandler
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func(s *serverImpl)
		wantErr bool
	}{
		{name: "1 Wrong Address", wantErr: true, fields: fields{conf: &Config{TimeoutInSeconds: 1, Address: "a.b.c.d"}}, setup: func(s *serverImpl) {}},
		{name: "2 Correct Address", wantErr: false,
			fields: fields{conf: &Config{TimeoutInSeconds: 1, Address: "localhost"}, shutdown: make(chan bool, 1)},
			args:   args{handler: func(request Request) []byte { return nil }},
			setup: func(s *serverImpl) {
				go func(s *serverImpl) {
					time.Sleep(7 * time.Second)
					s.Shutdown(context.Background())
				}(s)
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverImpl{
				conf:     tt.fields.conf,
				shutdown: tt.fields.shutdown,
				listner:  tt.fields.listner,
			}
			tt.setup(s)
			if err := s.Receive(tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("serverImpl.Receive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverImpl_Shutdown(t *testing.T) {
	t.Run("No Listener", func(t *testing.T) {
		s := &serverImpl{conf: &Config{}}
		ctx, cn := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cn()
		if err := s.Shutdown(ctx); err == nil {
			t.Errorf("serverImpl.Shutdown() got nil want an error")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		address, _ := net.ResolveUDPAddr(Network, "192.0.0.1")
		listner, _ := net.ListenUDP(Network, address)
		s := &serverImpl{conf: &Config{}, listner: listner}
		ctx, cn := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cn()
		if err := s.Shutdown(ctx); err == nil {
			t.Errorf("serverImpl.Shutdown() got nil want an error")
		}
	})
}
