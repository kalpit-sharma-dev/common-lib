package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/producer"
)

func main() {

	<-time.After(30 * time.Second)
	fmt.Println("waited 30 seconds for kafka server to start")

	kafkaServer := getEnv("KAFKA_SERVER", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "test")

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	config := producer.NewConfig()
	config.Address = []string{kafkaServer}
	p, err := producer.NewSyncProducer(config)
	if err != nil {
		fmt.Printf("Error creating sync producer: %+v\n", err)
		os.Exit(1)
	}

	run := true
LOOP:
	for run == true {
		select {
		case sig := <-sigs:
			fmt.Fprintf(os.Stdout, "Terminating on signal %v\n", sig)
			break LOOP

		default:
			msg := &producer.Message{
				Topic: kafkaTopic,
				Value: []byte(time.Now().String()),
			}

			dr, err := p.ProduceWithReport(context.Background(), "id", msg)
			if err != nil {
				fmt.Printf("Error producing messages: %+v\n", err)
				break LOOP
			}

			printDeliveryReport(dr[0])

			select {
			case <-time.After(5 * time.Second):
			case sig := <-sigs:
				fmt.Fprintf(os.Stdout, "Terminating on signal %v\n", sig)
				break LOOP
			}
		}
	}

	fmt.Fprintf(os.Stdout, "Closing\n")
	p.Close()
}

func printDeliveryReport(dr *producer.DeliveryReport) {
	fmt.Printf("Delivery report:\n  message: %+v\n  topic: %s/%d/%d\n  error: %+v\n",
		string(dr.Message.Value),
		dr.TopicPartition.Topic,
		dr.TopicPartition.Partition,
		dr.TopicPartition.Offset,
		dr.Error)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
