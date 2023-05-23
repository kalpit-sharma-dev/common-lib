package web

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestCreate(t *testing.T) {
	sd := &ServerConfig{ListenURL: ":8080"}
	sd1 := &ServerConfig{ListenURL: ":8080"}
	expectedHTTPServer := &newMuxConfig{
		serverCfg: sd,
		router:    &gorillaRouter{mux.NewRouter()},
		srv: &http.Server{
			Addr:         sd.ListenURL,
			ReadTimeout:  time.Duration(sd.ReadTimeoutMinute) * time.Minute,
			WriteTimeout: time.Duration(sd.WriteTimeoutMinute) * time.Minute,
		},
	}
	expectedHTTPServer1 := &newMuxConfig{
		serverCfg: sd1,
		router:    &gorillaRouter{mux.NewRouter()},
		srv: &http.Server{
			Addr:         sd1.ListenURL,
			ReadTimeout:  time.Duration(sd1.ReadTimeoutMinute) * time.Minute,
			WriteTimeout: time.Duration(sd1.WriteTimeoutMinute) * time.Minute,
		},
	}

	type args struct {
		cfg *ServerConfig
	}
	tests := []struct {
		name string
		args args
		want HTTPServer
	}{
		{
			name: "get Server Object Success with ListenURL:8080",
			args: args{cfg: sd},
			want: expectedHTTPServer,
		},
		{
			name: "get Server Object Success with ListenURL:8081",
			args: args{cfg: sd1},
			want: expectedHTTPServer1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
