package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/producer"
)

func main() {
	done := make(chan bool)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	topic := "test"
	config := producer.NewConfig()
	config.Address = []string{"localhost:9092"}
	p, err := producer.NewAsyncProducer(config)
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

	go func(drChan <-chan *producer.DeliveryReport) {
		for dr := range drChan {
			printDeliveryReport(dr)
		}
		close(done)
	}(p.DeliveryReportChannel())

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
				Topic: topic,
				Value: []byte(line),
			}
			p.ProduceChannel() <- msg
		}
	}

	fmt.Fprintf(os.Stdout, "Flushing any queued messages\n")
	leftInQueue := p.Flush(15 * 1000)
	fmt.Fprintf(os.Stdout, "%d messages left in queue", leftInQueue)
	fmt.Fprintf(os.Stdout, "Closing producer\n")
	p.Close()

	<-done
	fmt.Println("Finished getting all reports")
}

func printDeliveryReport(dr *producer.DeliveryReport) {
	fmt.Printf("Delivery report:\n  message: %+v\n  topic: %s/%d/%d\n  error: %+v\n >",
		string(dr.Message.Value),
		dr.TopicPartition.Topic,
		dr.TopicPartition.Partition,
		dr.TopicPartition.Offset,
		dr.Error)
}
