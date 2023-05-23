package cassandraorm

import (
	"errors"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/goc"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		wantErr      bool
		hosts        []string
		keyspace     string
		timeoutValue string
	}{
		{
			name:         "All Empty",
			wantErr:      true,
			hosts:        []string{},
			keyspace:     "",
			timeoutValue: "",
		},
		{
			name:         "Correct Hosts",
			wantErr:      true,
			hosts:        []string{"localhost:9042"},
			keyspace:     "",
			timeoutValue: "",
		},
		{
			name:         "Correct Hosts and KeySpace",
			wantErr:      true,
			hosts:        []string{"localhost:9042"},
			keyspace:     "platform_db",
			timeoutValue: "",
		},
		{
			name:         "All Correct",
			wantErr:      false,
			hosts:        []string{"localhost:9042"},
			keyspace:     "platform_db",
			timeoutValue: "5s",
		},
	}

	oldGocNewSimpleSession := gocNewSimpleSession
	defer func() { gocNewSimpleSession = oldGocNewSimpleSession }()
	gocNewSimpleSession = func(keyspace string, hosts []string, _ time.Duration) (goc.Session, error) {
		if keyspace != "" && len(hosts) > 0 {
			return goc.NewSession(&gocql.Session{}), nil
		}
		return nil, errors.New("some error")
	}

	log, err := logger.Create(logger.Config{Destination: logger.STDOUT})
	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Load(tt.hosts, tt.keyspace, tt.timeoutValue, log)
			if err != nil && tt.wantErr != true {
				t.Errorf("expected err: %t, got: %v", tt.wantErr, err)
			}
		})
	}
}
