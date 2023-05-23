package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/consumer"
)

func main() {

	<-time.After(30 * time.Second)
	fmt.Println("waited 30 seconds for kafka server to start")

	kafkaServer := getEnv("KAFKA_SERVER", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "test")

	config := consumer.NewConfig() // use defaults
	config.Address = []string{kafkaServer}
	config.Topics = []string{kafkaTopic}
	config.Group = "ConsumerExample"
	config.OffsetsInitial = consumer.OffsetOldest

	// notifications are sent from the main process
	config.NotificationHandler = func(str string) {
		fmt.Println(str)
	}

	// errors are sent from main process and from workers (this logic must be thread safe)
	config.ErrorHandler = func(ctx context.Context, err error, message *consumer.Message) {
		if message == nil {
			fmt.Printf("system error: %v\n", err)
		} else {
			fmt.Printf("message error: %s at %d/%d ==> %v\n", message.Topic, message.Partition, message.Offset, err)
		}
	}

	// messages are sent from workers (this logic must be thread safe)
	config.MessageHandler = func(ctx context.Context, message consumer.Message) error {
		value := string(message.Message)
		fmt.Printf("message received: %s at %d/%d ==> %s\n", message.Topic, message.Partition, message.Offset, value)
		return nil
	}

	kc, err := consumer.New(config)
	if err != nil {
		panic(fmt.Errorf("failed to create consumer: %+v", err))
	}
	fmt.Println("constructed new kafka consumer successfully")

	exit := make(chan os.Signal)
	killhealthcheck := make(chan bool)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-exit
		fmt.Println("exiting.. waiting for queued messages to process")
		kc.CloseWait()
		killhealthcheck <- true
		return
	}()

	go func() {
		for {
			select {
			case <-killhealthcheck:
				return
			case <-time.After(10 * time.Second):
				health, _ := kc.Health()
				fmt.Printf("healthcheck: %+v\n", health)
			}
		}
	}()

	// start pulling messages
	fmt.Println("gon start pullin'")
	kc.Pull() // blocks

	fmt.Println("done pulling")
	os.Exit(0)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
