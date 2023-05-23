package zookeeper

import (
	"fmt"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

// CBCommandName - command name to be used for zookeeper circuit breaker
const CBCommandName = "zookeeper"

var zkCBErrors = map[error]bool{
	zk.ErrConnectionClosed: true,
	zk.ErrClosing:          true,
	zk.ErrNoServer:         true,
}

func validCBError(err error) bool {
	if err == nil {
		return false
	}
	return zkCBErrors[err]
}

// RegisterCircuitBreaker - registers circuit breaker for the zookeeper
// connection
func RegisterCircuitBreaker(cfg *circuit.Config) error {
	if Client == nil {
		return fmt.Errorf("RegisterCircuitBreaker | zk client not initialized")
	}
	Client.setCBEnabled(cfg.Enabled)
	return circuit.Register("", CBCommandName, cfg, nil)
}
