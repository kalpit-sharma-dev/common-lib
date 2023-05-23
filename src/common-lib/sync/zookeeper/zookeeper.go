package zookeeper

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/sync"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

var connection Connection

const ignoreME = "IGNORE-ME"

type zookeeper struct {
	config sync.Config
}

// Instance : Returns an instance of Zookeper implamentation for Sync Service
func Instance(config sync.Config) sync.Service {
	return zookeeper{
		config: config,
	}
}

func (z zookeeper) connect() error {
	if connection == nil {
		logger.Get().Debug("", "Creating Connection for Servers : %s ", z.config.Servers)
		conn, _, err := zk.Connect(z.config.Servers, (time.Duration(z.config.SessionTimeoutInSecond) * time.Second))

		if err != nil {
			return err
		}
		connection = conn
	}
	return nil
}

func (z zookeeper) Send(path, data string) error {
	if err := z.connect(); err != nil {
		return err
	}
	return z.send(path, data, connection)
}

func (z zookeeper) send(path, data string, conn Connection) error {
	_, err := z.createNode(path, data, conn)
	if err != nil {
		logger.Get().Error("", "zookeeper:send:createNode", "Create Node Error : %v", err)
		return err
	}

	s, err := conn.Set(path, []byte(data), -1)
	if err != nil {
		return err
	}

	logger.Get().Trace("", "Sending Data %s on Path : %s at Version : %d", data, path, s.Version)
	return nil
}

func (z zookeeper) createNode(path, data string, conn Connection) (*zk.Stat, error) {
	flags := int32(zk.FlagEphemeral)
	acl := zk.WorldACL(zk.PermAll)

	found, s, err := conn.Exists(path)

	if err != nil {
		logger.Get().Error("", "zookeeper:createNode", "Listen Find Path Error : %v", err)
		return nil, err
	}

	if !found {
		_, err = conn.Create(path, []byte(data), flags, acl)
	}
	return s, err
}

func (z zookeeper) Listen(path string, c chan sync.Response) error {
	if err := z.connect(); err != nil {
		logger.Get().Error("", "zookeeper:Listen", "Listen Connect Error : %v", err)
		return err
	}

	defer connection.Close()

	for {
		z.listen(path, connection, c)
	}
}

func (z zookeeper) listen(path string, conn Connection, c chan sync.Response) {
	data, _, ech, err := conn.GetW(path)
	if err == zk.ErrNoNode {
		_, err = z.createNode(path, ignoreME, conn)
		if err != nil {
			logger.Get().Error("", "zookeeper:listen:node", "Create Node Error : %v", err)
			c <- sync.Response{Error: err}
		}
		return
	}

	if err != nil {
		logger.Get().Error("", "zookeeper:listen", "Listen Get Error : %v", err)
		c <- sync.Response{Error: err}
		return
	}

	d := string(data)
	if d != ignoreME {
		logger.Get().Trace("", "listen Data : %s", d)
		c <- sync.Response{Data: d}
	}
	<-ech
}

// Gets the data stored in the provided node.
// If node doesnot exist then it creates a node with the path and default as a blank string
func (z zookeeper) Get(path string) ([]byte, error) {
	if err := z.connect(); err != nil {
		logger.Get().Error("", "zookeeper:Get", "Connect Error : %v", err)
		return nil, err
	}
	data, _, err := connection.Get(path)
	if err != nil {
		if err == zk.ErrNoNode {
			_, err = z.createNode(path, ignoreME, connection)
			if err != nil {
				logger.Get().Error("", "zookeeper:Get", "Create Node Error : %v", err)
				return nil, err
			}
		} else {
			logger.Get().Error("", "zookeeper:Get", "Cannot be executed successfully, Reason : %v", err)
			return nil, err
		}
	}
	return data, nil
}

func (z zookeeper) connectionState() (string, error) {
	if err := z.connect(); err != nil {
		logger.Get().Error("", "zookeeper:connectionState", "Listen Connect Error : %v", err)
		return "", err
	}
	return connection.State().String(), nil
}

func (z zookeeper) Health() rest.Statuser {
	return zookeeperStatus{
		config:  z.config,
		service: z,
	}
}

// Zookeeper session state
const (
	stateConnected  = "StateConnected"
	stateHasSession = "StateHasSession"
)

type zookeeperStatus struct {
	config  sync.Config
	service zookeeper
}

func (s zookeeperStatus) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = "Zookeeper"
	conn.ConnectionURLs = s.config.Servers
	state, err := s.service.connectionState()
	conn.ConnectionStatus = rest.ConnectionStatusUnavailable

	// Statuses are not exported from "github.com/samuel/go-zookeeper/zk" for direct comparison
	if err == nil || state == stateConnected || state == stateHasSession {
		conn.ConnectionStatus = rest.ConnectionStatusActive
	}
	return &conn
}
