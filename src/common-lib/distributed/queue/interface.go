package queue

// Interface - Interface to hold Queue data
type Interface interface {
	// Create creates a new empty distributed queue node in zookeeper
	Create(queueName string) (string, error)

	//Exists checks if the distributed queue with provided name already exists in zookeeper
	Exists(queueName string) (bool, error)

	// GetList returns list of children (items) in the queue
	GetList(queueName string) ([]string, error)

	// CreateItem creates new sequence child node or item in the queue
	CreateItem(data []byte, queueName string) (string, error)

	// GetItemData gets data associated with the item in the queue
	GetItemData(queueName, itemName string) ([]byte, error)

	// RemoveItem removes the named item from the queue
	RemoveItem(queueName string, itemName string) error
}
