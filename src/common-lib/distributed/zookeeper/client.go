package zookeeper

import (
	"strings"

	"github.com/samuel/go-zookeeper/zk"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/lock"
)

const zkSeparator = "/"

type (
	// ZKClient describe interface for zookeeper client
	ZKClient interface {
		// State returns the current state of the connection
		State() string
		// Exists checks is item exist
		Exists(path string) (bool, *zk.Stat, error)
		// Get gets data by path
		Get(path string) ([]byte, *zk.Stat, error)
		// Children gets list of item children
		Children(path string) ([]string, *zk.Stat, error)
		// Set sets item data to path
		Set(path string, data []byte, version int32) (*zk.Stat, error)
		// Delete deletes item from zookeeper by its path
		Delete(path string, version int32) error
		// NewLock creates new zookeeper lock
		NewLock(path string, acl []zk.ACL) lock.Locker
		//Create creates a znode in the given path
		Create(path string, data []byte, flag int32, acl []zk.ACL) (string, error)
		// CreateRecursive creates new zookeeper item recursive
		CreateRecursive(childPath string, data []byte, flag int32, acl []zk.ACL) (string, error)
		// Close closes zookeeper client connection
		Close()
		// Events returns chan of client events
		Events() <-chan zk.Event

		isCBEnabled() bool
		setCBEnabled(cbEnabled bool)
	}

	zkClient struct {
		conn      *zk.Conn
		events    <-chan zk.Event
		cbEnabled bool
	}
)

func (client *zkClient) State() string {
	return client.conn.State().String()
}

func (client *zkClient) Exists(path string) (bool, *zk.Stat, error) {
	var (
		err    error
		exists bool
		stat   *zk.Stat
	)
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		exists, stat, err = client.conn.Exists(path)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	return exists, stat, err
}

func (client *zkClient) Get(path string) ([]byte, *zk.Stat, error) {
	var (
		data []byte
		stat *zk.Stat
		err  error
	)
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		data, stat, err = client.conn.Get(path)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	return data, stat, err
}

func (client *zkClient) Children(path string) ([]string, *zk.Stat, error) {
	var (
		data []string
		stat *zk.Stat
		err  error
	)
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		data, stat, err = client.conn.Children(path)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	return data, stat, err
}

func (client *zkClient) Set(path string, data []byte, version int32) (*zk.Stat, error) {
	var (
		stat *zk.Stat
		err  error
	)
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		stat, err = client.conn.Set(path, data, version)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return stat, err
}

func (client *zkClient) Delete(path string, version int32) error {
	var err error
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		err = client.conn.Delete(path, version)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	return err
}

func (client *zkClient) NewLock(path string, acl []zk.ACL) lock.Locker {
	return zk.NewLock(client.conn, path, acl)
}

func (client *zkClient) CreateRecursive(childPath string, data []byte, flag int32, acl []zk.ACL) (path string, err error) {

	path, err = client.Create(childPath, data, flag, acl)
	if err != zk.ErrNoNode {
		return path, err
	}

	// Create parent node.
	parts := strings.Split(childPath, zkSeparator)
	// always skip first argument it should be empty string
	for i := range parts[1:] {
		nPath := strings.Join(parts[:i+2], zkSeparator)

		var exists bool
		exists, _, err = client.Exists(nPath)
		if err != nil {
			return path, err
		}

		if exists {
			continue
		}

		// the last one set real data and flag
		if len(parts)-2 == i {
			path, err = client.Create(nPath, data, flag, acl)
			return path, err
		}

		path, err = client.Create(nPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return path, err
		}
	}
	return path, err
}

func (client *zkClient) Close() {
	client.conn.Close()
}

func (client *zkClient) Events() <-chan zk.Event {
	return client.events
}

func (client *zkClient) isCBEnabled() bool {
	return client.cbEnabled
}

func (client *zkClient) setCBEnabled(cbEnabled bool) {
	client.cbEnabled = cbEnabled
}

func (client *zkClient) Create(npath string, data []byte, flag int32, acl []zk.ACL) (path string, err error) {
	cbErr := circuit.Do(CBCommandName, client.cbEnabled, func() error {
		path, err = client.conn.Create(npath, data, flag, acl)
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	return path, err
}
