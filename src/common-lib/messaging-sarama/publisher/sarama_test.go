package publisher

import (
	"reflect"
	"testing"
)

func TestSyncProducer(t *testing.T) {
	type args struct {
		producerType ProducerType
		cfg          *Config
	}
	tests := []struct {
		name    string
		args    args
		want    Producer
		wantErr bool
	}{
		{
			name: "BlankCommandName", args: args{producerType: BigKafkaProducer, cfg: &Config{}},
			want: &syncProducer{producerType: BigKafkaProducer, cfg: &Config{}},
		},
		{
			name: "Command Name", args: args{producerType: RegularKafkaProducer, cfg: &Config{CommandName: "Command Name"}},
			want: &syncProducer{producerType: RegularKafkaProducer, cfg: &Config{CommandName: "Command Name"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SyncProducer(tt.args.producerType, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncProducer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SyncProducer() = %v, want %v", got, tt.want)
			}
		})
	}
}
