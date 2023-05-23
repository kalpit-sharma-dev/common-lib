package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging/producer"
)

func main() {
	numMessages := 500000
	printEveryMessages := 500
	done := make(chan bool)

	topic := "test"
	config := producer.NewConfig()
	config.Address = []string{"localhost:9092"}
	config.ProduceChannelSize = 10000
	p, err := producer.NewAsyncProducer(config)
	if err != nil {
		fmt.Printf("Error creating sync producer: %+v\n", err)
		os.Exit(1)
	}

	receivedCounter := int(0)
	go func(drChan <-chan *producer.DeliveryReport) {
		start := time.Now()
		for range drChan {
			receivedCounter++
			if receivedCounter%printEveryMessages == 0 {
				fmt.Printf("DeliveryReport received for %d messages\n", receivedCounter)
			}
		}
		elapsed := time.Since(start)
		fmt.Fprintf(os.Stdout, "DeliveryReports time to receive %d: %v\n", numMessages, elapsed.Seconds())
		close(done)
	}(p.DeliveryReportChannel())

	start := time.Now()
	sentCounter := int(0)
	for i := 1; i <= numMessages; i++ {
		msg := &producer.Message{
			Topic: topic,
			Value: []byte(fmt.Sprintf("message %d", i)),
		}
		p.ProduceChannel() <- msg
		sentCounter++
		if sentCounter%printEveryMessages == 0 {
			fmt.Printf("Sent %d messages\n", sentCounter)
		}
	}
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stdout, "Messages time to send %d: %v\n", numMessages, elapsed.Seconds())

	fmt.Fprintf(os.Stdout, "Flushing any queued messages\n")
	for {
		liq := p.Flush(5 * 1000)
		if liq == 0 {
			break
		}
		fmt.Fprintf(os.Stdout, "%d messages left in queue", liq)
	}

	fmt.Fprintf(os.Stdout, "Closing producer\n")
	p.Close()

	<-done
	fmt.Println("Finished getting all reports")
}
