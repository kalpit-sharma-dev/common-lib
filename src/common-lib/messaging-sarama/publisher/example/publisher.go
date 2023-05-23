package main

import (
	"context"
	"fmt"
	"log"
	logg "log"
	"os"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	"github.com/google/uuid"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/publisher"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

// Config is a struct used by Kafka producer
type Config struct {
	Pub            *publisher.Config
	Topics         []string
	Routine        int
	DelayInSeconds int64
}

var mutex = sync.Mutex{}

func main() {
	sarama.Logger = logg.New(os.Stdout, "[Producer] ", (log.Ldate | log.Ltime | log.LUTC | log.Lshortfile))
	log, _ := logger.Create(logger.Config{}) //nolint
	transaction := uuid.New().String()
	pub := publisher.NewConfig()
	pub.Address = []string{"localhost:9092"}
	pub.TimeoutInSecond = 3
	pub.CircuitBreaker = &circuit.Config{Enabled: true, TimeoutInSecond: 6, MaxConcurrentRequests: 400,
		ErrorPercentThreshold: 50, RequestVolumeThreshold: 20, SleepWindowInSecond: 5}

	cfg := &Config{Pub: pub,
		Topics: []string{"test"}, Routine: 50, DelayInSeconds: 6,
	}

	cnt := 1
	for {
		err := produce(cfg, transaction, 1, cnt)
		if err != nil {
			log.Info(transaction, "Failed to execute : %+v", err)
			break
		}
		cnt++
	}
}

func produce(cfg *Config, transaction string, count int, index int) error {
	log := logger.Get()
	publisher.Logger = logger.Get
	log.Info(transaction, "Configuration :: %+v", cfg)

	health(cfg, transaction)

	messages := make([]*publisher.Message, len(cfg.Topics))
	for i, t := range cfg.Topics {
		messages[i] = &publisher.Message{
			Topic: t,
			Key:   publisher.EncodeString(fmt.Sprintf("Test-%v-%v", count, index)),
			Value: publisher.EncodeString(fmt.Sprintf("Test Message %v-%v", count, index)),
		}
	}

	return produceMessage(cfg, transaction, messages...)
}

// This function describe a way to find Kafka producer health
func health(cfg *Config, transaction string) {
	log := logger.Get()
	h := publisher.Health(publisher.RegularKafkaProducer, cfg.Pub)
	log.Info(transaction, "Health :: %+v", h.Status(rest.OutboundConnectionStatus{Name: "Health", TimeStampUTC: time.Now()}))
}

// This function describe a way to produce a message on Kafka
func produceMessage(cfg *Config, transaction string, messages ...*publisher.Message) error {
	producer, err := publisher.SyncProducer(publisher.RegularKafkaProducer, cfg.Pub)
	if err != nil {
		return err
	}

	cntx, cancle := context.WithTimeout(context.Background(), time.Duration(cfg.Pub.TimeoutInSecond)*time.Second)
	defer cancle()
	return producer.Publish(cntx, transaction, messages...)
}
