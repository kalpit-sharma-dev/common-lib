package publisher

import (
	"reflect"
	"testing"

	"github.com/Shopify/sarama"
)

func TestMessage_AddHeader(t *testing.T) {
	t.Run("Add a Header Without header map", func(t *testing.T) {
		m := &Message{}
		m.AddHeader("Test-Key", "Test-Value")
		value := m.headers["Test-Key"]
		if value != "Test-Value" {
			t.Errorf("common.AddHeader() want Test-Value got %s", value)
		}
	})

	t.Run("Add a Header With header map", func(t *testing.T) {
		m := &Message{headers: map[string]string{"Init-Key": "Init-Value"}}
		m.AddHeader("Test-Key", "Test-Value")
		value := m.headers["Test-Key"]
		if value != "Test-Value" {
			t.Errorf("common.AddHeader() want Test-Value got %s", value)
		}

		value = m.headers["Init-Key"]
		if value != "Init-Value" {
			t.Errorf("common.AddHeader() want Init-Value got %s", value)
		}
	})
}

func TestMessage_toRecordHeader(t *testing.T) {
	type fields struct {
		Topic   string
		Key     Encoder
		Value   Encoder
		headers map[string]string
	}
	type args struct {
		transaction string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []sarama.RecordHeader
	}{
		{
			name: "1. No header", args: args{transaction: "transaction"}, want: []sarama.RecordHeader{},
		},
		{
			name: "2. Additional Headers", args: args{transaction: "transaction"},
			fields: fields{headers: map[string]string{"Key": "Value"}}, want: []sarama.RecordHeader{
				sarama.RecordHeader{Key: []byte("Key"), Value: []byte("Value")},
			},
		},
		{
			name: "3. Additional Headers", args: args{transaction: "transaction"},
			fields: fields{headers: map[string]string{"Key-Test": "Value-Test"}}, want: []sarama.RecordHeader{
				sarama.RecordHeader{Key: []byte("Key-Test"), Value: []byte("Value-Test")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Topic:   tt.fields.Topic,
				Key:     tt.fields.Key,
				Value:   tt.fields.Value,
				headers: tt.fields.headers,
			}
			if got := m.toRecordHeader(tt.args.transaction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Message.toRecordHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeString(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want Encoder
	}{
		{name: "1. Blank String", args: args{value: ""}, want: sarama.StringEncoder("")},
		{name: "2. test String", args: args{value: "test"}, want: sarama.StringEncoder("test")},
		{name: "3. broker String", args: args{value: "broker"}, want: sarama.StringEncoder("broker")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeString(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeBytes(t *testing.T) {
	type args struct {
		value []byte
	}
	tests := []struct {
		name string
		args args
		want Encoder
	}{
		{name: "1. Blank String", args: args{value: []byte("")}, want: sarama.ByteEncoder([]byte(""))},
		{name: "2. test String", args: args{value: []byte("test")}, want: sarama.ByteEncoder([]byte("test"))},
		{name: "3. broker String", args: args{value: []byte("broker")}, want: sarama.ByteEncoder([]byte("broker"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBytes(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeObject(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    Encoder
		wantErr bool
	}{
		{name: "1. Blank String", args: args{value: ""}, want: sarama.ByteEncoder([]byte{34, 34, 10})},
		{name: "2. test String", args: args{value: "test"}, want: sarama.ByteEncoder([]byte{34, 116, 101, 115, 116, 34, 10})},
		{name: "3. broker String", args: args{value: "broker"}, want: sarama.ByteEncoder([]byte{34, 98, 114, 111, 107, 101, 114, 34, 10})},
		{name: "3. broker String", args: args{value: struct{ name string }{name: "new object"}}, want: sarama.ByteEncoder([]byte{123, 125, 10})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeObject(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeObject() = %v, want %v", got, tt.want)
			}
		})
	}
}
