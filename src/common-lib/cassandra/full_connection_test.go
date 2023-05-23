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
)

func TestNewFullDbConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)

	conn := connection{
		conf: &DbConfig{
			Hosts:                    []string{"Server"},
			Keyspace:                 "test",
			Consistency:              gocql.Quorum,
			NumConn:                  10,
			TimeoutMillisecond:       time.Millisecond * 10,
			DisableInitialHostLookup: false,
			ConnectTimeout:           time.Millisecond * 600,
			CircuitBreaker: &circuit.Config{
				Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
				ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
			},
			CommandName: "Database-Command1",
			ValidErrors: []string{"Error1"},
		}, session: session,
	}

	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		want    FullDbConnector
		setup   func()
		wantErr bool
	}{
		{
			name:    "Nil Config",
			wantErr: true,
			want:    nil,
			args:    args{conf: nil},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return session, errors.New("error")
				}
			},
		},
		{
			name:    "Blank Config",
			wantErr: true,
			want:    nil,
			args:    args{conf: &DbConfig{}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return session, errors.New("error")
				}
			},
		},
		{
			name:    "Blank Host",
			wantErr: true,
			want:    nil,
			args:    args{conf: &DbConfig{Hosts: []string{}}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return session, errors.New("error")
				}
			},
		},
		{
			name:    "Blank Keyspace",
			wantErr: true,
			want:    nil,
			args:    args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: ""}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return session, errors.New("error")
				}
			},
		},
		{
			name:    "Session Creation Error",
			want:    nil,
			wantErr: true,
			args:    args{conf: &DbConfig{Hosts: []string{"Sever"}, Keyspace: "test"}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return session, errors.New("error")
				}
			},
		},
		{
			name:    "Session Creation Success",
			wantErr: false,
			want: &fullConnection{
				connection: &conn,
				batchConnection: &batchConnection{
					Connection: conn,
				},
			},
			args: args{conf: &DbConfig{
				Hosts:              []string{"Server"},
				Keyspace:           "test",
				NumConn:            10,
				TimeoutMillisecond: time.Millisecond * 10,
				CircuitBreaker: &circuit.Config{
					Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
					ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
				},
				CommandName: "Database-Command1",
				ValidErrors: []string{"Error1"},
			}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := NewFullDbConnection(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDbAndBatchDbConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFullDbConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}
