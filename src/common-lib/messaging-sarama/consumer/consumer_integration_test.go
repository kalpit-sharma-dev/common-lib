//go:build integration
// +build integration

package consumer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/publisher"
)

const (
	host                 = "localhost:9092"
	consumerGroup        = "test-fat-consumer-group"
	topic                = "test-integration-topic"
	numPartitions        = 4
	messagesPerPartition = 500
)

func TestConsumeDataPullOrdered(t *testing.T) {
	var (
		brokers = []string{host}
		cfg     = publisher.NewConfig()
	)
	createTopic(t)
	defer deleteTopic(t)

	cfg.Address = brokers

	producerService, err := publisher.SyncProducer(publisher.RegularKafkaProducer, cfg)
	if err != nil {
		t.Errorf("producer.SyncProducer() error = %v", err)
	}

	ctx := context.Background()

	// produce messagesPerPartition in numPartitions (same num per partition)
	wg := sync.WaitGroup{}
	for partition := 0; partition < numPartitions; partition++ {
		for messageIndex := 0; messageIndex < messagesPerPartition; messageIndex++ {
			err := producerService.Publish(ctx, "", &publisher.Message{
				Key:   sarama.ByteEncoder(strconv.Itoa(partition)),
				Topic: topic,
				Value: sarama.StringEncoder("{}"),
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
	consumerCfg.ConsumerMode = PullOrderedWithOffsetReplay
	consumerCfg.Address = []string{host}
	consumerCfg.Group = consumerGroup
	// expectation that processing will be done in numPartitions parallel routines
	consumerCfg.OffsetsInitial = OffsetOldest
	consumerCfg.Topics = []string{topic}
	consumerCfg.RetryDelay = time.Second
	consumerCfg.RetryCount = 3

	mu := sync.Mutex{}
	messagePartitions := make([]int32, 0, numPartitions*messagesPerPartition)
	var start, end time.Time
	// / handle message ad write to result slice
	consumerCfg.MessageHandler = func(message Message) error {
		mu.Lock()
		defer mu.Unlock()
		defer wg.Done()
		messagePartitions = append(messagePartitions, message.Partition)
		if (start == time.Time{}) {
			start = time.Now()
		}
		end = time.Now()
		return nil
	}

	consumerService, err := New(consumerCfg)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	go consumerService.Pull()

	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("processed %d messages, in %+v\n", numPartitions*messagesPerPartition, end.Sub(start))
	fmt.Printf("processing destribution report\n")
	// the report groups partitions from consumed message
	fmt.Printf(printReport(messagePartitions))
}

func printReport(items []int32) string {
	sb := strings.Builder{}
	var blockCount int32 = 1
	for i, current := range items {
		if len(items) == i+1 || items[i+1] != current {
			sb.Write([]byte(fmt.Sprintf("partition[%d], items count[%d]\n", current, blockCount)))
			blockCount = 1
			continue
		}
		blockCount++
	}
	return sb.String()
}

func createTopic(t *testing.T) {
	b := sarama.NewBroker(host)
	cfg := sarama.NewConfig()
	cfg.Version = sarama.MaxVersion
	require.NoError(t, b.Open(cfg))
	defer func(b *sarama.Broker) {
		require.NoError(t, b.Close())
	}(b)

	r := sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     numPartitions,
				ReplicationFactor: int16(1),
			},
		},
	}

	_, err := b.CreateTopics(&r)
	require.NoError(t, err)
}

func deleteTopic(t *testing.T) {
	b := sarama.NewBroker(host)
	cfg := sarama.NewConfig()
	cfg.Version = sarama.MaxVersion
	require.NoError(t, b.Open(cfg))
	defer func(b *sarama.Broker) {
		require.NoError(t, b.Close())
	}(b)

	r := sarama.DeleteTopicsRequest{
		Topics: []string{topic},
	}

	_, err := b.DeleteTopics(&r)
	require.NoError(t, err)
}
