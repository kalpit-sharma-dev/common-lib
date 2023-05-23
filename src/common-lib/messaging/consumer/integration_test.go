//go:build integration
// +build integration

package consumer

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/producer"
)

const (
	host                 = "localhost:9092"
	consumerGroup        = "test-fat-consumer-group"
	topic                = "fat-topic"
	numPartitions        = 10
	messagesPerPartition = 500
)

func TestConsumeDataPullOrdered(t *testing.T) {
	err := deleteTopic()
	if err != nil {
		t.Logf("cant delete topic: error = %v. This is fine if dopic does not exist", err)
	}
	fmt.Printf("topic deleted\n")

	err = createTopic()
	if err != nil {
		t.Errorf("cant create topic: error = %v", err)
	}
	fmt.Printf("topic created\n")

	defer func() {
		err := deleteTopic()
		if err != nil {
			t.Errorf("cant delete topic: error = %v", err)
		}
		fmt.Printf("topic deleted\n")
	}()

	// producer
	producerCfg := producer.NewConfig()
	producerCfg.Address = []string{host}
	producerCfg.EnableIdempotence = true
	producerCfg.QueueBufferingMaxMs = 1
	producerService, err := producer.SyncProducer(producer.RegularKafkaProducer, producerCfg)
	if err != nil {
		t.Errorf("producer.SyncProducer() error = %v", err)
	}
	ctx := context.Background()

	// produce messagesPerPartition in numPartitions (same num per partition)
	wg := sync.WaitGroup{}
	for partition := 0; partition < numPartitions; partition++ {
		for messageIndex := 0; messageIndex < messagesPerPartition; messageIndex++ {
			err := producerService.Produce(ctx, "", &producer.Message{
				Key:   []byte(strconv.Itoa(rand.Int())),
				Topic: topic,
				Value: []byte("{}"),
			})
			if err != nil {
				t.Errorf("producerService.Produce() error = %v", err)
				return
			}
			wg.Add(1)
		}
	}

	fmt.Printf("published %d messages\n", numPartitions*messagesPerPartition)

	// consumer
	consumerCfg := NewConfig()
	consumerCfg.CommitMode = OnMessageCompletion
	consumerCfg.ConsumerMode = PullOrdered
	consumerCfg.Address = []string{host}
	consumerCfg.Group = consumerGroup
	// expectation that processing will be done in numPartitions parallel routines
	consumerCfg.MaxQueueSize = numPartitions
	consumerCfg.OffsetsInitial = OffsetOldest
	consumerCfg.Topics = []string{topic}

	mu := sync.Mutex{}
	var start, end time.Time
	// / handle message ad write to result slice
	consumerCfg.MessageHandler = func(message Message) error {
		mu.Lock()
		defer mu.Unlock()
		defer wg.Done()
		if (start == time.Time{}) {
			start = time.Now()
		}
		end = time.Now()
		// fmt.Println("inv")
		return nil
	}

	consumerService, err := New(consumerCfg)
	if err != nil {
		t.Errorf("consumer.New() error = %v", err)
		return
	}

	// defer profile.Start(profile.BlockProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	go consumerService.Pull()

	wg.Wait()
	fmt.Printf("processed %d messages, in %+v\n", numPartitions*messagesPerPartition, end.Sub(start))
	// fmt.Printf("processing distribution report\n")
	// the report groups partitions from consumed message

}

func createTopic() error {
	adminClient, err := createAdminClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	results, err := adminClient.CreateTopics(ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: 1}})
	if err != nil {
		return err
	}

	// Check for specific topic errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("topic creation failed for %s: %v",
				result.Topic, result.Error.String())
		}
	}

	adminClient.Close()
	return nil
}

func deleteTopic() error {
	adminClient, err := createAdminClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	results, err := adminClient.DeleteTopics(ctx, []string{topic})
	if err != nil {
		return err
	}

	// Check for specific topic errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicDeletionDisabled {
			return fmt.Errorf("topic delition failed for %s: %v",
				result.Topic, result.Error.String())
		}
	}

	adminClient.Close()
	return nil
}

func createAdminClient() (*kafka.AdminClient, error) {
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": host,
	})
	if err != nil {
		return nil, err
	}
	return adminClient, nil
}
