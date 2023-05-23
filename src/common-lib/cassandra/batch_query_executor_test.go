//go:build !integration
// +build !integration

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

func TestGetBatchQueryExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	session := mock.NewMockSession(ctrl)

	type args struct {
		conf *DbConfig
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    BatchQueryExecutor
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
			wantErr: true,
			want:    nil,
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
			want: &batchQueryExecutorImpl{
				batchConnection: &batchConnection{
					Connection: connection{conf: &DbConfig{
						Hosts:                    []string{"Sever"},
						Keyspace:                 "test",
						Consistency:              gocql.Quorum,
						NumConn:                  10,
						TimeoutMillisecond:       time.Millisecond * 10,
						ConnectTimeout:           600 * time.Millisecond,
						DisableInitialHostLookup: false,
						CircuitBreaker: &circuit.Config{
							Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
							ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
						},
						CommandName: "Database-Command1",
						ValidErrors: []string{"Error1"},
					}, session: session}},
				batch: &gocql.Batch{},
			},
			args: args{conf: &DbConfig{
				Hosts:              []string{"Sever"},
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
				session.EXPECT().NewBatch(gomock.Any()).Return(&gocql.Batch{})
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return session, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := GetBatchQueryExecutor(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchQueryExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatchQueryExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}
