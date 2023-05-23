package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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
	reader := bufio.NewReader(os.Stdin)
	stdinChan := make(chan string)
	fmt.Printf(" >")

	go func() {
		for true {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			line = strings.TrimSuffix(line, "\n")
			if len(line) == 0 {
				continue
			}

			stdinChan <- line
		}
		close(stdinChan)
	}()

	run := true

	for run == true {
		select {
		case sig := <-sigs:
			fmt.Fprintf(os.Stdout, "Terminating on signal %v\n", sig)
			run = false

		case line, ok := <-stdinChan:
			if !ok {
				run = false
				break
			}

			msg := &producer.Message{
				Topic: kafkaTopic,
				Value: []byte(line),
			}

			dr, err := p.ProduceWithReport(context.Background(), "id", msg)
			if err != nil {
				fmt.Printf("Error producing messages: %+v\n", err)
				run = false
				break
			}

			printDeliveryReport(dr[0])
			fmt.Printf(" >")
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
