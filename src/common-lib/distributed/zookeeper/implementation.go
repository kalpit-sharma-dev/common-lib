package zookeeper

import (
	"fmt"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	leaderelection "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/leader-election"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/lock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/queue"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/scheduler"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

const (
	nodePrefix         = "node-"
	queuePrefix        = "queue-"
	locksNode          = "locks"
	queueNode          = "queue"
	leaderElectionNode = "leader-election"
)

// Zookeeper session state
const (
	stateConnected  = "StateConnected"
	stateHasSession = "StateHasSession"
)

var (
	// LeaderElector implementation
	LeaderElector leaderelection.Interface = leaderElectorImpl{}
	// Queue implementation
	Queue queue.Interface = queueImpl{}
	// Scheduler implementation
	Scheduler scheduler.Interface = schedulerImpl{}
	// Client implementation
	Client ZKClient
	// Connect implement connection to Zookeeper
	Connect = implConnect
	// Logger : Logger instance used for logging
	// Defaults to Discard
	Logger = logger.DiscardLogger

	zookeeperBasePath  string
	defaultTransaction = "Zookeeper"
)

type (
	leaderElectorImpl struct{}
	queueImpl         struct{}
	schedulerImpl     struct{}
	lockWrapper       struct {
		zkLock    lock.Locker
		cbEnabled bool
		name      string
	}
)

// Init makes zkClient initializations
func Init(zookeeperHosts []string, basePath string) error {
	return initImpl(zookeeperHosts, basePath)
}

// InitWithLogger makes zkClient initializations with custom logger instance
func InitWithLogger(zookeeperHosts []string, basePath string, logImpl logger.Log) error {
	initLogger(logImpl)
	return initImpl(zookeeperHosts, basePath)
}

func initImpl(zookeeperHosts []string, basePath string) error {
	conn, events, err := Connect(zookeeperHosts, 10*time.Second)
	if err != nil {
		return err
	}

	Client = &zkClient{conn: conn, events: events}

	if len(basePath) < 1 {
		return fmt.Errorf("incorrect base path: %s", basePath)
	}
	zookeeperBasePath = basePath
	return nil
}

// NewLock is a wrapper for creating new lock
func NewLock(name string) lock.Locker {
	path := getLockPath(name)
	acl := zk.WorldACL(zk.PermAll)
	return &lockWrapper{zkLock: Client.NewLock(path, acl), cbEnabled: Client.isCBEnabled(), name: name}
}

// DeleteLock deletes lock parent node
func DeleteLock(name string) error {
	path := getLockPath(name)
	exist, _, err := Client.Exists(path)
	if exist {
		err = Client.Delete(path, -1)
	}
	return err
}

func getLockPath(name string) string {
	return zookeeperBasePath + zkSeparator + locksNode + zkSeparator + name
}

func implConnect(hosts []string, duration time.Duration) (*zk.Conn, <-chan zk.Event, error) {
	return zk.Connect(hosts, duration)
}

func initLogger(logImpl logger.Log) {
	if logImpl == nil {
		return
	}
	Logger = func() logger.Log {
		return logImpl
	}
}

// ConnectionStatus struct for Cassandra status
type ConnectionStatus struct {
	Path  string
	Hosts []string
}

// Status used for getting status of Casandra connection
func (cs *ConnectionStatus) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = "Zookeeper"
	conn.ConnectionURLs = cs.Hosts
	conn.ConnectionStatus = rest.ConnectionStatusUnavailable

	state := Client.State()
	// Statuses are not exported from "github.com/samuel/go-zookeeper/zk" for direct comparison
	if state == stateConnected || state == stateHasSession {
		conn.ConnectionStatus = rest.ConnectionStatusActive
	}

	return &conn
}
