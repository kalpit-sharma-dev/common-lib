package consumer

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1. Blank Config", wantErr: true, args: args{cfg: Config{}}},
		{name: "2. Blank Address", wantErr: true, args: args{cfg: Config{Address: []string{}}}},
		{name: "3. Blank Group", wantErr: true, args: args{cfg: Config{Address: []string{"localhost"}, Group: ""}}},
		{name: "4. Blank Topic", wantErr: true, args: args{cfg: Config{Address: []string{"localhost"}, Group: "Group", Topics: []string{}}}},
		{name: "5. Blank Message Handler", wantErr: true, args: args{cfg: Config{Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: nil}}},
		{name: "6. Blank Subscriber", wantErr: true, args: args{cfg: Config{Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "7. Zero Subscriber", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 0, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "8. Negative Subscriber", wantErr: true, args: args{cfg: Config{SubscriberPerCore: -1, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "9. Positive Subscriber", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 1, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "10. Set Offset", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 1,
			OffsetsInitial: OffsetNewest, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "11. Rebalance.Timeout", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 1,
			OffsetsInitial: OffsetNewest, RebalanceTimeout: time.Second, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "12. Timeout", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 1, Timeout: time.Second,
			ConsumerMode: PullOrdered, OffsetsInitial: OffsetNewest, RebalanceTimeout: time.Second, Address: []string{"localhost"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
		{name: "13. Timeout", wantErr: true, args: args{cfg: Config{SubscriberPerCore: 1, Timeout: time.Second,
			ConsumerMode: PullUnOrdered, CommitMode: OnMessageCompletion, OffsetsInitial: OffsetNewest, RebalanceTimeout: time.Second, Address: []string{"localhost:9092"},
			Group: "Group", Topics: []string{"Topic"}, MessageHandler: func(Message) error { return nil }}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_invokeErrorHandler(t *testing.T) {
	type args struct {
		err     error
		message *Message
		cfg     Config
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "nil message", args: args{err: errors.New("Error"), message: nil, cfg: Config{}}},
		{name: "message available", args: args{err: errors.New("Error"), message: &Message{}, cfg: Config{}}},
		{name: "Error handler panic", args: args{err: errors.New("Error"), message: &Message{},
			cfg: Config{ErrorHandler: func(error, *Message) { panic("Error") }}}},
		{name: "Error handler Success", args: args{err: errors.New("Error"), message: &Message{},
			cfg: Config{ErrorHandler: func(error, *Message) {}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invokeErrorHandler(tt.args.err, tt.args.message, tt.args.cfg)
		})
	}
}

func Test_invokeNotificationHandler(t *testing.T) {
	type args struct {
		notification string
		cfg          Config
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Error handler panic", args: args{notification: "aaa",
			cfg: Config{NotificationHandler: func(string) { panic("Error") }}}},
		{name: "Error handler Success", args: args{notification: "aaaa",
			cfg: Config{NotificationHandler: func(string) {}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invokeNotificationHandler(tt.args.notification, tt.args.cfg)
		})
	}
}
