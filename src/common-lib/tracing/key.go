package tracing

// KeyCorrelationId key for transcactionid
const KeyCorrelationId = "CorrelationId"

// KeyType key to represent type
const KeyType = "Type"

// CassandraKey holds Cassandra related Key
const (
	KeyCassandra                = "Cassandra"
	KeyCassandraStartTime       = "StartTime"
	KeyCassandraEndTime         = "EndTime"
	KeyCassandraKeySpace        = "Keyspace"
	KeyCassandraType            = "Type"
	KeyCassandraTypeNormalQuery = "NormalQuery"
	KeyCassandraTypeBatchQuery  = "BatchQuery"
	KeyCassandraStatement       = "Statement"
	KeyCassandraHost            = "Host"
	KeyCassandraHostInfo        = "HostInfo"
	KeyCassandraAttempts        = "Attempts"
	KeyCassandraTotalLatency    = "TotalLatency"
)

// KeyKafkaTopic holds Kafka related key
const (
	KeyKafkaTopic       = "Topic"
	KeyKafkaTopics      = "Topics"
	KeyKafkaMessageSize = "MessageSize"
	KeyConsumerGroup    = "ConsumerGroup"
	KeyMessageHandler   = "MessageHandler"
)

// KafkaConsumer holds Kafka Consumer related key
const (
	KeyConsumerSegmentName    = "KafkaConsumerSegment"
	KeyConsumerSubSegmentName = "KafkaConsumer"
	KeyConsumerType           = "ConsumerType"
	KeyConsumerMode           = "ConsumerMode"
)

// KafkaProducer holds Kafka Producer related key
const (
	KeyProducerSubSegmentName = "KafkaProducer"
	KeyProducerType           = "Type"
	KeyProducerTypeValue      = "KafkaProducer"
	KeyBrokers                = "Brokers"
)
