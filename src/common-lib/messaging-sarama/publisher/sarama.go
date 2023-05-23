// Package publisher provides a Kafka publisher built on sarama.
//
// Deprecated: use messaging/publisher instead of infra-lib/messaging/publisher
package publisher

import (
	"runtime/debug"
	"sync"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

var (
	reConnectInProgress bool
	connected           = make(chan bool, 1)

	producerMutex = &sync.RWMutex{}
	kafkaProducer = map[ProducerType]producer{}
)

// SyncProducer - Return instance of Sync Producer
var SyncProducer = func(producerType ProducerType, cfg *Config) (Producer, error) {
	if kafkaProducer[producerType] == nil {
		producerMutex.Lock()
		defer producerMutex.Unlock()
		if kafkaProducer[producerType] == nil {
			err := register(utils.GetTransactionID(), cfg)
			if err != nil {
				return nil, err
			}
			prod := &syncProducer{producerType: producerType, cfg: cfg}
			kafkaProducer[producerType] = prod
		}
	}
	return kafkaProducer[producerType], nil
}

// register - Register a circuit breaker callback
func register(transaction string, cfg *Config) error {
	return circuit.Register(transaction, cfg.commandName(), cfg.circuitBreaker(),
		func(transaction, commandName string, state string) {
			if state == circuit.Open {
				ReConnect(transaction, cfg)
			} else if state == circuit.Close {
				Connected(transaction, cfg)
			}
		})
}

// ReConnect - Function to Reconnect Kafka nodes
var ReConnect = func(transaction string, cfg *Config) {
	if reConnectInProgress {
		Logger().Info(transaction, "Reconnect to Kafka Already In-Progress...")
		return
	}

	reConnectInProgress = true

	// Wait for the twice a Sleep window before re-trying to connect again on the Kafka
	ticker := time.NewTicker(cfg.reconnectIntervalInSecond())

	// Once Circuit is closed, Stop reconnecting on Kafka;
	// Only check for the Broker Command as most of the traffic is coming on this
LOOP:
	for index := 0; index < cfg.maxReconnectRetry(); index++ {
		reconnect(transaction, cfg)
		select {
		case <-ticker.C:
			Logger().Info(transaction, "Kafka-ReConnect: Trying to reconnect, retry count : %v  ...", index)
		case <-connected:
			Logger().Info(transaction, "Kafka-ReConnect: Moving out of reconnect loop as Circuit is Closed ...")
			break LOOP
		}
	}

	ticker.Stop()
	reConnectInProgress = false
	Logger().Info(transaction, "Reconnected to Kafka ...")
}

func reconnect(transaction string, cfg *Config) {
	defer func() {
		if r := recover(); r != nil {
			Logger().Error(transaction, "kafka.reconnect.recovered", "Kafka-ReConnect: Recovered in reconnect : %v, Trace : %s", r, debug.Stack())
		}
	}()

	Logger().Info(transaction, "Reconnecting to Kafka ...")

	var existingProducer = map[ProducerType]producer{}
	for k, v := range kafkaProducer {
		err := v.reconnect(transaction)
		if err != nil {
			Logger().Error(transaction, "kafka.producer.reconnect.failed", "Failed to create %s producer during reconnect. Error %v", k, err)
			continue
		}
		existingProducer[k] = v
		Logger().Info(transaction, "%s Kafka Producer reconnected...", k)
	}

	ticker := time.NewTicker(cfg.cleanupTimeInSecond())
	<-ticker.C
	ticker.Stop()

	for producerType, v := range existingProducer {
		Logger().Info(transaction, "Closing old %s Kafka Producer", producerType)
		err := v.cleanExisting(transaction)
		if err != nil {
			Logger().Error(transaction, "kafka.connection.close.failed", "Error in closing %s Kafka producer %v", producerType, err)
		}
	}
}

// Connected - Used for stopping reconnect process as soon as we find MS is connected to Kafka and producing messages
func Connected(transaction string, cfg *Config) {
	select {
	case connected <- true:
		Logger().Trace(transaction, "Producer is connected back ...")
	default:
	}
}

// Clean - Clean all Kafka connection, used in case of graceful shutdown
func Clean(transaction string) {
	for k, v := range kafkaProducer {
		err := v.clean(transaction)
		if err != nil {
			Logger().Error(transaction, "kafka.producer.clean.failed", "Failed to clean %s producer during. Error %v", k, err)
		}
	}
}
