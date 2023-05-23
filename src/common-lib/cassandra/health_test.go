package cassandra

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

func TestHealth(t *testing.T) {
	type args struct {
		cfg *DbConfig
	}
	tests := []struct {
		name string
		args args
		want rest.Statuser
	}{
		{
			name: "Instance", want: status{conf: &DbConfig{}},
			args: args{cfg: &DbConfig{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Health(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Health() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_status_Status(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)
	tm := time.Now()
	circuit.Register("transaction", "ClosedState", circuit.New(),
		func(transaction, commandName string, state string) {
		})

	type fields struct {
		conf *DbConfig
	}
	type args struct {
		conn rest.OutboundConnectionStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func()
		want   *rest.OutboundConnectionStatus
	}{
		{
			name: "Connection Error", args: args{conn: rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1"}},
			fields: fields{conf: &DbConfig{Hosts: []string{}}},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1", ConnectionType: "Cassandra",
				ConnectionURLs: []string{}, ConnectionStatus: rest.ConnectionStatusUnavailable},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, errors.New("Error") }
			},
		},
		{
			name: "Connection Success", args: args{conn: rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "2", Name: "2"}},
			fields: fields{conf: &DbConfig{Hosts: []string{"test"}, Keyspace: "Test"}},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "2", Name: "2", ConnectionType: "Cassandra",
				ConnectionURLs: []string{"test"}, ConnectionStatus: rest.ConnectionStatusActive},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, nil }
				session.EXPECT().Close()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			c := status{
				conf: tt.fields.conf,
			}
			if got := c.Status(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("status.Status() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
