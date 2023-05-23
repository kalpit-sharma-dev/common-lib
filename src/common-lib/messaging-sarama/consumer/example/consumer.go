package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/consumer"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gopkg.in/urfave/cli.v1"
)

//nolint:lll
func main() {
	app := cli.NewApp()
	app.Name = "Message Consumer"
	app.Version = "1.0.0"
	app.Compiled = time.Now()
	app.Copyright = "(c) 2018 Continuum IT Managed Platform"
	app.Usage = "demonstrate consumer APP"
	app.UsageText = "demonstrating the available consumer API"
	app.ArgsUsage = "[args and such]"
	app.HideHelp = false
	app.HideVersion = false

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{Name: "address, A", Value: &cli.StringSlice{"localhost:9092"}, Usage: "Kafka IP Address"},
		cli.StringSliceFlag{Name: "topic, T", Value: &cli.StringSlice{"test"}, Usage: "List of topics to be consumed"},
		cli.StringFlag{Name: "group, G", Value: "Jain", Usage: "Consumer Group"},
		cli.IntFlag{Name: "subscriberPerCore, C", Value: 45, Usage: "Number of go routines per Core"},
		cli.DurationFlag{Name: "processingWait, W", Value: time.Second, Usage: "Message offset will be committed without releasing worker"},
		cli.IntFlag{Name: "commitMode, M", Value: int(consumer.OnMessageCompletion), Usage: "Message Commit Mode {-1 (OnPull), -2 (OnMessageCompletion)}"},
		cli.IntFlag{Name: "consumerMode, CM", Value: int(consumer.PullOrdered), Usage: "Message Consumer Mode {-1 (PullUnOrdered), -2 (PullOrdered)}"},
		cli.Int64Flag{Name: "initialOffset, I", Value: consumer.OffsetNewest, Usage: "Message Initial Offset {-1 (OffsetNewest), -2 (OffsetOldest)}"},
		cli.BoolFlag{Name: "errorHandler, E", Usage: "Handle any errors that occurred while consuming Message"},
		cli.BoolFlag{Name: "notificationHandler, N", Usage: "Handle notification that occurred on reblancing"},
		cli.BoolFlag{Name: "health, H", Usage: "Handle any errors that occurred while consuming Message"},
		cli.DurationFlag{Name: "processingDelay, D", Value: time.Second, Usage: "Message Processing Delay"},
		cli.DurationFlag{Name: "healthSchedule, S", Value: time.Second * 30, Usage: "Delay between two health calls"},
		cli.DurationFlag{Name: "retryDelay, RD", Value: time.Second * 5, Usage: "Delay before retrying the failed message again"},
		cli.Int64Flag{Name: "retryCount, RC", Value: 5, Usage: "How many times message should be retried before dropping"},
	}

	log, err := logger.Create(logger.Config{Destination: logger.STDOUT})
	if err != nil {
		fmt.Println(err)
		return
	}

	app.Action = createConsumer
	err = app.Run(os.Args)
	if err != nil {
		log.Info(utils.GetTransactionID(), "Failed to execute : %+v", err)
	}
}

//nolint:dupl
func createConsumer(ctx *cli.Context) error {
	consumer.Logger = logger.Get

	cfg := consumer.NewConfig()
	cfg.Address = ctx.GlobalStringSlice("address")
	cfg.Topics = ctx.GlobalStringSlice("topic")
	cfg.Group = ctx.GlobalString("group")
	cfg.SubscriberPerCore = ctx.GlobalInt("subscriberPerCore")
	cfg.Timeout = ctx.GlobalDuration("processingWait")
	cfg.RetryDelay = ctx.GlobalDuration("retryDelay")
	cfg.RetryCount = ctx.GlobalInt64("retryCount")

	// Choose commit Mode, By Default is OnPull
	cfg.CommitMode = consumer.OnPull
	if ctx.GlobalInt("commitMode") == int(consumer.OnMessageCompletion) {
		cfg.CommitMode = consumer.OnMessageCompletion
	}

	// Pull messages one by one, by default is PullParallel
	// NOTE: Messages are processed one by on, hence performance is slow,
	// totally dependes how fast your service is processing message.
	// This Mode claims that message will be not lost under any system failure
	cfg.ConsumerMode = consumer.PullUnOrdered
	if ctx.GlobalInt("consumerMode") == int(consumer.PullOrdered) {
		cfg.ConsumerMode = consumer.PullOrdered
	}

	cfg.OffsetsInitial = consumer.OffsetNewest
	if ctx.GlobalInt64("initialOffset") == consumer.OffsetOldest {
		cfg.OffsetsInitial = consumer.OffsetOldest
	}

	handler := &handler{ctx: ctx}

	cfg.MessageHandler = handler.message
	cfg.ErrorHandler = handler.error

	if ctx.GlobalBool("notificationHandler") {
		cfg.NotificationHandler = handler.notification
	}
	cfg.TransactionID = utils.GetTransactionID()
	srvc, err := consumer.New(cfg)
	if err != nil {
		logger.Get().Info(cfg.TransactionID, "Failed to create consumer instance :: %+v", err)
		return err
	}

	if ctx.GlobalBool("health") {
		go handler.health(srvc)
	}
	srvc.Pull()

	return nil
}

type handler struct {
	ctx *cli.Context
}

// Message will not be marked as read if error is returned, will return same message again
// You must take care of your panics on your owm
func (h *handler) message(msg consumer.Message) error {
	logger.Get().Debug(msg.GetTransactionID(), "Topic : %s at %d/%d ==> Message : %s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Message))
	time.Sleep(h.ctx.GlobalDuration("processingDelay"))
	return nil
}

func (h *handler) error(err error, message *consumer.Message) {
	transaction := utils.GetTransactionID()
	if message != nil {
		transaction = message.GetTransactionID()
	}

	logger.Get().Info(transaction, "Handle Error : %v\n", err, message)
}

func (h *handler) notification(notification string) {
	logger.Get().Info("notification", "Handle Notification : %s\n", notification)
}

func (h *handler) health(srvc consumer.Service) {
	for {
		time.Sleep(h.ctx.GlobalDuration("healthSchedule"))
		health, err := srvc.Health()
		if err != nil {
			logger.Get().Info("notification", "Failed to find health :: %+v", err)
		} else {
			logger.Get().Info("notification", "Consumer Health :: %+v", health)
		}
	}
}
