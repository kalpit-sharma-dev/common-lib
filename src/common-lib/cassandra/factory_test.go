package cassandra

import (
	"errors"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

func TestFactoryImpl_GetDbConnector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	msession := mock.NewMockSession(ctrl)
	session = cassandraSession{}

	type args struct {
		cfg *DbConfig
	}
	tests := []struct {
		name    string
		f       FactoryImpl
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Session Creation Error", wantErr: true,
			args: args{cfg: &DbConfig{
				Hosts:    []string{"Sever"},
				Keyspace: "test",
			}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return msession, errors.New("error")
				}
			},
		},
		{
			name: "Session Creation Success", wantErr: false,
			args: args{cfg: &DbConfig{
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
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return msession, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			f := FactoryImpl{}
			_, err := f.GetDbConnector(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("FactoryImpl.GetDbConnector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactoryImpl_GetNewDbConnector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	msession := mock.NewMockSession(ctrl)
	new1Session = cassandraSession{}

	type args struct {
		cfg *DbConfig
	}
	tests := []struct {
		name    string
		f       FactoryImpl
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Session Creation Error", wantErr: true,
			args: args{cfg: &DbConfig{
				Hosts:    []string{"Sever"},
				Keyspace: "test",
			}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return msession, errors.New("error")
				}
			},
		},
		{
			name: "Session Creation Success", wantErr: false,
			args: args{cfg: &DbConfig{
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
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return msession, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			f := FactoryImpl{}
			_, err := f.GetNewDbConnector(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("FactoryImpl.GetNewDbConnector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactoryImpl_GetBatchDbConnector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	msession := mock.NewMockSession(ctrl)
	batchSession = batchCassandraSession{}

	type args struct {
		cfg *DbConfig
	}
	tests := []struct {
		name    string
		f       FactoryImpl
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "Session Creation Error", wantErr: true,
			args: args{cfg: &DbConfig{
				Hosts:    []string{"Sever"},
				Keyspace: "test",
			}},
			setup: func() {
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) {
					return msession, errors.New("error")
				}
			},
		},
		{
			name: "Session Creation Success", wantErr: false,
			args: args{cfg: &DbConfig{
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
				cql.CreateSession = func(cluster *gocql.ClusterConfig) (cql.Session, error) { return msession, nil }
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			f := FactoryImpl{}
			_, err := f.GetBatchDbConnector(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("FactoryImpl.GetBatchDbConnector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
