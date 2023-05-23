// +build integration

package producer

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/profile"
	"strconv"
	"testing"
	"time"
)

const (
	host                 = "localhost:9092"
	topic                = "fat-producer-topic"
	numPartitions        = 4
	messagesPerPartition = 500
)

func TestProduceDataInSync(t *testing.T) {
	err := createTopic()
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
	producerCfg := NewConfig()
	producerCfg.Address = []string{host}
	producerCfg.EnableIdempotence = true
	producerCfg.QueueBufferingMaxMs = 2 // this value significantly impacts the performance. Default '5' is too much for producers which produce only one message at a time.
	producerService, err := SyncProducer(RegularKafkaProducer, producerCfg)
	if err != nil {
		t.Errorf("producer.SyncProducer() error = %v", err)
	}
	ctx := context.Background()

	defer profile.Start(profile.BlockProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	start := time.Now()

	for partition := 0; partition < numPartitions; partition++ {
		for messageIndex := 0; messageIndex < messagesPerPartition; messageIndex++ {
			err := producerService.Produce(ctx, "", &Message{
				Key:   []byte(strconv.Itoa(partition * messageIndex)),
				Topic: topic,
				Value: []byte("{}"),
			})
			if err != nil {
				t.Errorf("producerService.Produce() error = %v", err)
				return
			}
		}
	}

	fmt.Printf("published %d messages in %q\n", numPartitions*messagesPerPartition, time.Now().Sub(start))

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
