package goc

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

const (
	version  = 4
	keyspace = "gockle_test"
)

func TestNewSession(t *testing.T) {
	NewSimpleSession = func(keyspace string, hosts []string, timeout time.Duration) (Session, error) {
		return NewSession(&gocql.Session{}), nil
	}

	defer func() { NewSimpleSession = newSimpleSession }()

	session, err := NewSimpleSessionStatus(keyspace, []string{"localhost:0000"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	status := session.Status(rest.OutboundConnectionStatus{})
	if status.ConnectionStatus != rest.ConnectionStatusActive {
		t.Fatalf("unexpected connection status: expected: %s, actual: %s", rest.ConnectionStatusActive, status.ConnectionStatus)
	}
}

func TestNewSimpleSession(t *testing.T) {
	hosts := []string{"localhost"}
	timeout := 3 * time.Second
	if s, err := NewSimpleSession(keyspace, hosts, timeout); err == nil {
		t.Error("Actual no error, expected error")
	} else if s != nil {
		t.Errorf("Actual session %v, expected nil", s)
		s.Close()
	}

	a, err := NewSimpleSession(keyspace, hosts, timeout)
	switch {
	case err != nil:
		t.Skip(err)
	case a == nil:
		t.Errorf("Actual session nil, expected not nil")
	default:
		a.Close()
	}
}
