package publisher

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func TestMain(m *testing.M) {
	Logger = logger.DiscardLogger
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestConfig_timeoutInSecond(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{name: "Default", fields: fields{}, want: 3 * time.Second},
		{name: "3 Second", fields: fields{TimeoutInSecond: 3}, want: 3 * time.Second},
		{name: "10 Second", fields: fields{TimeoutInSecond: 10}, want: 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.timeoutInSecond(); got != tt.want {
				t.Errorf("Config.timeoutInSecond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_maxMessageBytes(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{name: "Default", fields: fields{}, want: 1000},
		{name: "4", fields: fields{MaxMessageBytes: 4}, want: 4},
		{name: "10", fields: fields{MaxMessageBytes: 10}, want: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.maxMessageBytes(); got != tt.want {
				t.Errorf("Config.maxMessageBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_reconnectIntervalInSecond(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{name: "Default", fields: fields{}, want: 12 * time.Second},
		{name: "4 Second", fields: fields{ReconnectIntervalInSecond: 4}, want: 4 * time.Second},
		{name: "10 Second", fields: fields{ReconnectIntervalInSecond: 10}, want: 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.reconnectIntervalInSecond(); got != tt.want {
				t.Errorf("Config.reconnectIntervalInSecond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_maxReconnectRetry(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{name: "Default", fields: fields{}, want: 20},
		{name: "20", fields: fields{MaxReconnectRetry: 20}, want: 20},
		{name: "10", fields: fields{MaxReconnectRetry: 10}, want: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.maxReconnectRetry(); got != tt.want {
				t.Errorf("Config.maxReconnectRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_cleanupTimeInSecond(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{name: "Default", fields: fields{}, want: 15 * time.Second},
		{name: "5 Second", fields: fields{CleanupTimeInSecond: 5}, want: 5 * time.Second},
		{name: "10 Second", fields: fields{CleanupTimeInSecond: 10}, want: 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.cleanupTimeInSecond(); got != tt.want {
				t.Errorf("Config.cleanupTimeInSecond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_circuitBreaker(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   *circuit.Config
	}{
		{name: "Default", fields: fields{}, want: &circuit.Config{Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10}},
		{name: "Disabled", fields: fields{CircuitBreaker: &circuit.Config{}}, want: &circuit.Config{}},
		{name: "Enabled", fields: fields{CircuitBreaker: &circuit.Config{Enabled: true}}, want: &circuit.Config{Enabled: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.circuitBreaker(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.circuitBreaker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_commandName(t *testing.T) {
	type fields struct {
		Address                   []string
		TimeoutInSecond           int64
		MaxMessageBytes           int
		CompressionType           CompressionType
		ReconnectIntervalInSecond int64
		MaxReconnectRetry         int
		CleanupTimeInSecond       int64
		CircuitBreaker            *circuit.Config
		CommandName               string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "Default", fields: fields{}, want: "Broker-Command"},
		{name: "Broker-Command", fields: fields{CommandName: "Broker-Command"}, want: "Broker-Command"},
		{name: "Test-Command", fields: fields{CommandName: "Test-Command"}, want: "Test-Command"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:                   tt.fields.Address,
				TimeoutInSecond:           tt.fields.TimeoutInSecond,
				MaxMessageBytes:           tt.fields.MaxMessageBytes,
				CompressionType:           tt.fields.CompressionType,
				ReconnectIntervalInSecond: tt.fields.ReconnectIntervalInSecond,
				MaxReconnectRetry:         tt.fields.MaxReconnectRetry,
				CleanupTimeInSecond:       tt.fields.CleanupTimeInSecond,
				CircuitBreaker:            tt.fields.CircuitBreaker,
				CommandName:               tt.fields.CommandName,
			}
			if got := c.commandName(); got != tt.want {
				t.Errorf("Config.commandName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionType_UnmarshalJSON(t *testing.T) {
	type fields struct {
		codec sarama.CompressionCodec
		value string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    CompressionType
	}{
		{name: "Default-Name", fields: fields{}, args: args{[]byte("")}, wantErr: false, want: None},
		{name: "Given Name - None", fields: fields{}, args: args{[]byte("\"None\"")}, wantErr: false, want: None},
		{name: "Given Name - NONE", fields: fields{}, args: args{[]byte("\"NONE\"")}, wantErr: false, want: None},
		{name: "Given Name - GZIP", fields: fields{}, args: args{[]byte("\"GZIP\"")}, wantErr: false, want: GZIP},
		{name: "Given Name - GZIP", fields: fields{}, args: args{[]byte("\"gzip\"")}, wantErr: false, want: GZIP},
		{name: "Given Name - Snappy", fields: fields{}, args: args{[]byte("\"SNAPPY\"")}, wantErr: false, want: Snappy},
		{name: "Given Name - Snappy", fields: fields{}, args: args{[]byte("\"snappy\"")}, wantErr: false, want: Snappy},
		{name: "Given Name - LZ4", fields: fields{}, args: args{[]byte("\"LZ4\"")}, wantErr: false, want: LZ4},
		{name: "Given Name - LZ4", fields: fields{}, args: args{[]byte("\"lz4\"")}, wantErr: false, want: LZ4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompressionType{
				codec: tt.fields.codec,
				value: tt.fields.value,
			}
			if err := c.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CompressionType.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if *c != tt.want {
				t.Errorf("CompressionType.UnmarshalJSON() = %v, want %v", c, tt.want)
			}
		})
	}
}

func TestCompressionType_MarshalJSON(t *testing.T) {
	type fields struct {
		codec sarama.CompressionCodec
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{name: "None", fields: fields{value: None.value}, wantErr: false, want: []byte("\"NONE\"")},
		{name: "GZIP", fields: fields{value: GZIP.value}, wantErr: false, want: []byte("\"GZIP\"")},
		{name: "SNAPPY", fields: fields{value: Snappy.value}, wantErr: false, want: []byte("\"SNAPPY\"")},
		{name: "LZ4", fields: fields{value: LZ4.value}, wantErr: false, want: []byte("\"LZ4\"")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CompressionType{
				codec: tt.fields.codec,
				value: tt.fields.value,
			}
			got, err := c.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("CompressionType.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CompressionType.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
