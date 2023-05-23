package publisher

import (
	"reflect"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

func TestHealth(t *testing.T) {
	type args struct {
		producerType ProducerType
		cfg          *Config
	}
	tests := []struct {
		name string
		args args
		want rest.Statuser
	}{
		{
			name: "Instance BigKafkaProducer", want: status{producerType: BigKafkaProducer, cfg: &Config{}},
			args: args{producerType: BigKafkaProducer, cfg: &Config{}},
		},
		{
			name: "Instance RegularKafkaProducer", want: status{producerType: RegularKafkaProducer, cfg: &Config{}},
			args: args{producerType: RegularKafkaProducer, cfg: &Config{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Health(tt.args.producerType, tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Health() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_status_Status(t *testing.T) {
	tm := time.Now()
	circuit.Register("transaction", "ClosedState", circuit.New(),
		func(transaction, commandName string, state string) {
		})

	type fields struct {
		producerType ProducerType
		cfg          *Config
	}
	type args struct {
		conn rest.OutboundConnectionStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *rest.OutboundConnectionStatus
	}{
		{
			name: "UninitializedState", args: args{conn: rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1"}},
			fields: fields{producerType: RegularKafkaProducer, cfg: &Config{CommandName: "Uninitialized", Address: []string{"12"}}},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1", ConnectionType: "Kafka-regular-Producer",
				ConnectionURLs: []string{"12"}, ConnectionStatus: rest.ConnectionStatusActive},
		},
		{
			name: "ClosedState", args: args{conn: rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1"}},
			fields: fields{producerType: RegularKafkaProducer, cfg: &Config{CommandName: "ClosedState", Address: []string{"12"}}},
			want: &rest.OutboundConnectionStatus{TimeStampUTC: tm, Type: "1", Name: "1", ConnectionType: "Kafka-regular-Producer",
				ConnectionURLs: []string{"12"}, ConnectionStatus: rest.ConnectionStatusActive},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := status{
				producerType: tt.fields.producerType,
				cfg:          tt.fields.cfg,
			}
			if got := k.Status(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("status.Status() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
