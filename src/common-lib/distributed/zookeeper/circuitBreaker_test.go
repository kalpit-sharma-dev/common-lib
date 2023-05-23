package zookeeper

import (
	"testing"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

func TestValidCBError(t *testing.T) {
	tests := []struct {
		err    error
		expect bool
	}{
		{zk.ErrConnectionClosed, true},
		{zk.ErrClosing, true},
		{zk.ErrNoServer, true},
		{zk.ErrDeadlock, false},
	}

	for _, test := range tests {
		result := validCBError(test.err)
		if result != test.expect {
			t.Errorf("Expected %v. Got %v", test.expect, result)
		}
	}
}

func TestRegisterCircuitBreaker(t *testing.T) {
	t.Run("Register cb for uninitialized zk client", func(t *testing.T) {
		cbConfig := circuit.New()
		err := RegisterCircuitBreaker(cbConfig)
		if err == nil {
			t.Errorf("Expected error. Got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		Client = &zkClient{}
		cbConfig := circuit.New()
		err := RegisterCircuitBreaker(cbConfig)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		state := circuit.CurrentState(CBCommandName)
		if state != circuit.Close {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
