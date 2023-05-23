package consumer

import (
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

const (
	// OffsetNewest stands for the log head offset, i.e. the offset that will be
	// assigned to the next message that will be produced to the partition. You
	// can send this to a client's GetOffset method to get this offset, or when
	// calling ConsumePartition to start consuming new messages.
	OffsetNewest int64 = -1
	// OffsetOldest stands for the oldest offset available on the broker for a
	// partition. You can send this to a client's GetOffset method to get this
	// offset, or when calling ConsumePartition to start consuming from the
	// oldest offset that is still available on the broker.
	OffsetOldest int64 = -2
	//OnPull means offset will be committed on Pull before message processing starts
	OnPull commitMode = -1
	//OnMessageCompletion means offset will be committed on after message processing completes
	OnMessageCompletion commitMode = -2
	//PullUnOrdered - Messages will be processed in the batches without any order
	PullUnOrdered consumerMode = -1
	//PullOrdered - Note that message will be processed in per partition sequential Order.
	PullOrdered consumerMode = -2
	//PullOrderedWithOffsetReplay -Note that message will be processed in per partition sequential Order
	//and the offset can be reset based on the data returned by the function mentioned in the config:HandleCustomOffsetStash
	PullOrderedWithOffsetReplay consumerMode = -3
)

// Logger : Logger instance used for logging
// Defaults to Discard
var Logger = logger.DiscardLogger

type commitMode int

type consumerMode int

// Config is a struct used by Kafka Consumer
type Config struct {
	//Address : is a Kafka Broker IPs
	Address []string

	//Group : is  a Consumer Group name
	Group string

	//Topics to be consumed
	Topics []string

	//SubscriberPerCore : Number of maximum go routines for processing per Core
	SubscriberPerCore int

	//CommitMode : Specify which commit mode you want to use
	// OnPull means offset will be committed on Pull before message processing starts
	// OnMessageCompletion means offset will be committed on after message processing completes
	CommitMode commitMode

	// ConsumerMode : Specify which consumer mode you want to use
	// PullUnOrdered : Messages will be processed in batches
	// PullOrdered : This Mode claims that message will not lost under any system failure.
	// PullOrderedWithOffsetReplay : This mode adds on replay offset feature on PullOrdered mode
	//
	// NOTE: This needs more partition as messages are processed sequentially / partition.
	// Default is PullUnOrdered
	ConsumerMode consumerMode

	// Timeout After this duration message offset will be committed without releasing worker
	Timeout time.Duration

	// The retention duration for committed offsets.
	// If zero, disabled (in which case the `offsets.retention.minutes` option on the
	// broker will be used).  Kafka only supports precision up to milliseconds;
	// nanoseconds will be truncated. Requires Kafka
	// broker version 0.9.0 or later. (default is 0: disabled).
	Retention time.Duration

	// RebalanceTimeout: The maximum allowed time for each worker to join the group once a rebalance has begun.
	// This is basically a limit on the amount of time needed for all tasks to flush any pending
	// data and commit offsets. If the timeout is exceeded, then the worker will be removed from
	// the group, which will cause offset commit failures (default 60s).
	RebalanceTimeout time.Duration

	//This will ge used as a callback function for message handling
	MessageHandler func(Message) error

	//This function is only used in case of a specific consumer-mode for now.
	//The consumer mode with which this function is used is PullOrderedWithOffsetReplay [-3]
	//Input to the function is the topic and partition for the current consumer
	//The function will return a List of offsetstash for the topic and partition
	//The implementation of this function should ideally include the logic to fetch the offset for the partitions
	//and return the same as the map output defined for the function signature.
	HandleCustomOffsetStash func(string, int32) ([]OffsetStash, error)

	// If available, any errors that occurred while consuming are returned on
	// the Errors Handler (default disabled).
	ErrorHandler func(error, *Message)

	// If available, rebalance notification will be returned on the
	// Notifications Handler (default disabled).
	NotificationHandler func(string)

	// The initial offset to use if no offset was previously committed.
	// Should be OffsetNewest or OffsetOldest. Defaults to OffsetNewest.
	OffsetsInitial int64

	// Whether to maintain a full set of metadata for all topics, or just
	// the minimal set that has been necessary so far. The full set is simpler
	// and usually more convenient, but can take up a substantial amount of
	// memory if you have many topics and partitions. Defaults to true.
	MetadataFull bool

	// RetryDelay: This flag is used in case you want to wait for some time before retrying the the failed message.
	// Defaults to 30s.
	// NOTE: This only Works with PullOrdered Consumer Mode
	RetryDelay time.Duration

	// RetryCount : Specify how meany times a message should be retried before dropping the message
	// Defaults to 10.
	// NOTE: This only Works with PullOrdered Consumer Mode
	RetryCount int64

	//TransactionID - Default transaction id to be used by the consumer
	TransactionID string

	// Partitions - Max Number of partitions to be consumed by consumer
	// Default to 500.
	// NOTE: This number should always be bigger than the number of partitions available in the production
	// for all the topics
	Partitions int

	//KafkaLogConfig is to enable kafka level log if config provided
	KafkaLogConfig *logger.Config
}

// NewConfig is a function to return default consumer configuration
func NewConfig() Config {
	return Config{
		SubscriberPerCore: 20,
		CommitMode:        OnPull,
		ConsumerMode:      PullUnOrdered,
		Retention:         0, // Respect Broker's `offsets.retention.minutes` option
		OffsetsInitial:    OffsetNewest,
		Timeout:           time.Minute,
		MetadataFull:      true,
		RetryCount:        10,
		RetryDelay:        30 * time.Second,
		RebalanceTimeout:  60 * time.Second,
		Partitions:        500,
	}
}

// Service : Service contains all the functions related to Consumer
type Service interface {
	//Pull message from Kafka topic
	Pull()
	//MarkOffset Update offset for partition and topic
	MarkOffset(topic string, partition int32, offset int64)
	//Close consumer connection
	Close() error
	//
	Health() (Health, error)
}

// OffsetStash  this is a model for a specific partition offset for a topic
type OffsetStash struct {
	Topic             string
	Partition         int32
	Offset            int64
	Message           []byte
	Header            map[string]string
	PulledDateTimeUTC time.Time
	TransactionID     string
}

// Message : Represents consumed Kafka Message
type Message struct {
	Message           []byte
	Offset            int64
	Partition         int32
	Topic             string
	PulledDateTimeUTC time.Time
	headers           map[string]string
	transactionID     string
}

// GetHeaders - Retrieve All headers
func (m *Message) GetHeaders() map[string]string {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	newHeaders := make(map[string]string)
	for k, v := range m.headers {
		newHeaders[k] = v
	}
	return newHeaders
}

// GetHeader - Retrieve header value form the Kafka
func (m *Message) GetHeader(key string) string {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	return m.headers[key]
}

// GetTransactionID - Retrieve Transaction ID value form the Kafka Header
func (m *Message) GetTransactionID() string {
	if m.transactionID == "" {
		m.transactionID = utils.GetTransactionID()
		if m.headers != nil {
			transaction, ok := m.headers[constants.TransactionID]
			if ok {
				m.transactionID = transaction
			}
		}
	}
	return m.transactionID
}

// Health : Returns current health of the Kafka connection with Service
type Health struct {
	ConnectionState    map[string]bool
	Topics             []string
	PerTopicPartitions map[string][]int32
	Group              string
	CanCosume          bool
}

// newHealth is a function to return default Health data
func newHealth(cfg Config) *Health {
	health := &Health{
		Topics:    cfg.Topics,
		Group:     cfg.Group,
		CanCosume: false,
	}
	health.ConnectionState = make(map[string]bool)
	for _, address := range cfg.Address {
		health.ConnectionState[address] = false
	}
	health.PerTopicPartitions = make(map[string][]int32)
	return health
}
