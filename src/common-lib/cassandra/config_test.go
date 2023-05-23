package cassandra

import (
	"reflect"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want *DbConfig
	}{
		{
			name: "Default Config", want: &DbConfig{NumConn: 20, TimeoutMillisecond: time.Second, Consistency: gocql.Quorum, ConnectTimeout: 600 * time.Millisecond,
				CircuitBreaker: &circuit.Config{Enabled: false, TimeoutInSecond: 5, MaxConcurrentRequests: 2500,
					ErrorPercentThreshold: 25, RequestVolumeThreshold: 300, SleepWindowInSecond: 10},
				CommandName: "Database-Command", ValidErrors: []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Nil Config", wantErr: true, args: args{conf: nil}},
		{name: "Blank Config", wantErr: true, args: args{conf: &DbConfig{}}},
		{name: "Blank Host", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{}}}},
		{name: "Blank Keyspace", wantErr: true, args: args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: ""}}},
		{name: "#Num Connection 0", wantErr: false, args: args{conf: &DbConfig{
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 0},
		}},
		{name: "Timeout-0", wantErr: false, args: args{conf: &DbConfig{
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Millisecond},
		}},
		{name: "Timeout-nenosecond", wantErr: false, args: args{conf: &DbConfig{
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Nanosecond},
		}},
		{name: "Command name", wantErr: false, args: args{conf: &DbConfig{CommandName: "Command",
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Millisecond},
		}},
		{name: "ValidErrors", wantErr: false, args: args{conf: &DbConfig{CommandName: "Command", ValidErrors: []string{"Error"},
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Millisecond},
		}},
		{name: "ConnectTimout-0", wantErr: false, args: args{conf: &DbConfig{CommandName: "Command", ValidErrors: []string{"Error"},
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Millisecond, ConnectTimeout: 600 * time.Millisecond},
		}},
		{name: "ConnectTimout-nanosecond", wantErr: false, args: args{conf: &DbConfig{CommandName: "Command", ValidErrors: []string{"Error"},
			Hosts: []string{"Sever"}, Keyspace: "test", NumConn: 10, TimeoutMillisecond: time.Millisecond, ConnectTimeout: time.Nanosecond},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
