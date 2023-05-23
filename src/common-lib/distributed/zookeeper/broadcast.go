package zookeeper

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed"
)

const (
	pathForListeners = "broadcast/listeners"
	defaultTimeout   = time.Second
)

var (
	// Broadcast instance
	Broadcast distributed.Broadcast
	// ErrZookeeperNotInit zookeeper is not initialized error
	ErrZookeeperNotInit = errors.New("zookeeper is not initialized use zookeeper.Init")

	mutex = &sync.Mutex{}
)

type broadcastImpl struct {
	instanceID string
	timeout    time.Duration
	handlers   map[string]distributed.BroadcastHandler
}

// InitBroadcast singleton, thread-safe, returns pointer to *Broadcast
// instanceID should be unique for each instance of micro-service
func InitBroadcast(instanceID string, timeout time.Duration) (distributed.Broadcast, error) {
	if Client == nil {
		return nil, ErrZookeeperNotInit
	}

	if Broadcast != nil {
		return Broadcast, nil
	}

	mutex.Lock()
	defer mutex.Unlock()

	if Broadcast != nil {
		return Broadcast, nil
	}

	if timeout <= 0 {
		timeout = defaultTimeout
	}

	Broadcast = &broadcastImpl{
		instanceID: instanceID,
		timeout:    timeout,
		handlers:   make(map[string]distributed.BroadcastHandler),
	}

	return Broadcast, nil
}

func (n *broadcastImpl) AddHandler(name string, handler distributed.BroadcastHandler) {
	n.handlers[name] = handler
}

func (n *broadcastImpl) absolutePath() string {
	return n.listenersPath() + zkSeparator + n.instanceID
}

func (n *broadcastImpl) listenersPath() string {
	return zookeeperBasePath + zkSeparator + pathForListeners
}

func (n *broadcastImpl) subscribe() error {
	exists, _, err := Client.Exists(n.absolutePath())
	if err != nil {
		return err
	}

	if exists {
		return nil
	}
	// we are checking our path, in case if it is absent we are trying to create it
	_, err = Client.CreateRecursive(n.absolutePath(), []byte{}, int32(zk.FlagEphemeral), zk.WorldACL(zk.PermAll))
	return err
}

func (n *broadcastImpl) Listen(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				Logger().Info(defaultTransaction, "stopped by context")
				return
			case <-time.After(n.timeout):
				if err := n.subscribe(); err != nil {
					Logger().Info(defaultTransaction, "subscribe error: %s", err)
				}
				items, err := Queue.GetList(n.instanceID)
				if err != nil && err != zk.ErrNoNode {
					Logger().Info(defaultTransaction, "getList error: %s", err)
					continue
				}
				n.process(items)
			}
		}
	}()
}

func (n *broadcastImpl) process(items []string) {
	for _, item := range items {
		data, err := Queue.GetItemData(n.instanceID, item)
		if err != nil {
			Logger().Error(defaultTransaction, "Queue.GetItemDataFailed", "%v", err)
			continue
		}

		if err := Queue.RemoveItem(n.instanceID, item); err != nil {
			Logger().Error(defaultTransaction, "Queue.RemoveItemFailed", "%v", err)
		}

		e := new(distributed.Event)
		if err := json.Unmarshal(data, e); err != nil {
			Logger().Error(defaultTransaction, "Json.UnmarshalFailed", "%v", err)
			continue
		}

		if handler, ok := n.handlers[e.Type]; ok {
			handler(e)
		}
	}
}

func (n *broadcastImpl) CreateEvent(e distributed.Event) error {
	// getting all subscribers
	subscribers, _, err := Client.Children(n.listenersPath())
	if err != nil {
		return err
	}

	content, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// sending the message to subscribers
	for _, subscriber := range subscribers {
		_, err := Queue.CreateItem(content, subscriber)
		if err != nil {
			return err
		}
	}

	return nil
}
