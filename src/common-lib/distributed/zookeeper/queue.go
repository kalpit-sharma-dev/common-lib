package zookeeper

import "github.com/samuel/go-zookeeper/zk"

// Create creates a new empty distributed queue node in zookeeper
func (queueImpl) Create(queueName string) (string, error) {
	childPath := getQueueZkPath(queueName)
	flag := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	return Client.CreateRecursive(childPath, nil, flag, acl)
}

//Exists checks if the distributed queue with provided name already exists in zookeeper
func (queueImpl) Exists(queueName string) (bool, error) {
	childPath := getQueueZkPath(queueName)
	exists, _, err := Client.Exists(childPath)
	return exists, err
}

// GetList returns list of children in the queue
func (queueImpl) GetList(queueName string) ([]string, error) {
	children, _, err := Client.Children(getQueueZkPath(queueName))
	return children, err
}

// CreateItem creates new sequence child node
func (queueImpl) CreateItem(data []byte, queueName string) (string, error) {
	childPath := getQueueZkPath(queueName) + zkSeparator + queuePrefix
	flag := int32(zk.FlagSequence)
	acl := zk.WorldACL(zk.PermAll)
	return Client.CreateRecursive(childPath, data, flag, acl)
}

// GetItemData gets node data
func (queueImpl) GetItemData(queueName, itemName string) ([]byte, error) {
	b, _, err := Client.Get(getQueueZkPath(queueName) + zkSeparator + itemName)
	return b, err
}

// RemoveItem drop node
func (queueImpl) RemoveItem(queueName, itemName string) error {
	child := getQueueZkPath(queueName) + zkSeparator + itemName
	return Client.Delete(child, 0)
}

func getQueueZkPath(queueName string) string {
	return zookeeperBasePath + zkSeparator + queueNode + zkSeparator + queueName
}
