package consumer

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/mock/gomock"
	mock "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/consumer/extmock"
)

var mockPartitionConsumer *mock.MockPartitionConsumer

func initialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPartitionConsumer = mock.NewMockPartitionConsumer(ctrl)
}

func Test_newMessageFrmOffsetStash(t *testing.T) {
	type args struct {
		offsetStash OffsetStash
	}
	tests := []struct {
		name string
		args args
		want Message
	}{
		{
			name: "newMessageFrmOffsetStash: Without Header", args: args{offsetStash: OffsetStash{
				Message: []byte("message"), Topic: "test", Partition: 1, Offset: 2}},
			want: Message{Message: []byte("message"), Topic: "test", Partition: 1, Offset: 2, headers: nil},
		},
		{
			name: "newMessageFrmOffsetStash: With Header", args: args{offsetStash: OffsetStash{
				Message: []byte("message"), Topic: "test", Partition: 1, Offset: 2, Header: map[string]string{"Key": "Value"}}},
			want: Message{Message: []byte("message"), Topic: "test", Partition: 1, Offset: 2,
				headers: map[string]string{"Key": "Value"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMessageFrmOffsetStash(tt.args.offsetStash)
			tt.want.PulledDateTimeUTC = got.PulledDateTimeUTC
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMessageFrmOffsetStash() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_getNewpullOrderedWithOffsetReplay(t *testing.T) {
	cfg := Config{}
	type args struct {
		conf Config
	}
	tests := []struct {
		name string
		args args
		want *pullOrderedWithOffsetReplay
	}{
		{
			name: "getNewpullOrderedWithOffsetReplay: With proper config",
			args: args{
				conf: Config{
					Address:      []string{"localhost:9024"},
					Topics:       []string{"test"},
					Group:        "test",
					CommitMode:   OnMessageCompletion,
					ConsumerMode: PullOrderedWithOffsetReplay,
				},
			},
			want: &pullOrderedWithOffsetReplay{
				cfg: Config{
					Address:      []string{"localhost:9024"},
					Topics:       []string{"test"},
					Group:        "test",
					CommitMode:   OnMessageCompletion,
					ConsumerMode: PullOrderedWithOffsetReplay,
				},
				partitionKeyStore: make(map[string]bool),
			},
		}, {
			name: "getNewpullOrderedWithOffsetReplay: With empty config",
			args: args{
				conf: cfg,
			},
			want: &pullOrderedWithOffsetReplay{
				cfg:               cfg,
				partitionKeyStore: make(map[string]bool),
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := getNewpullOrderedWithOffsetReplay(tt.args.conf)
			got.partitionKeyStoreMutex = tt.want.partitionKeyStoreMutex
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNewpullOrderedWithOffsetReplay() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_pullOrderedWithOffsetReplay_getPartitionKeyFrm(t *testing.T) {
	initialize(t)
	type fields struct {
		cfg                    Config
		partitionKeyStore      map[string]bool
		partitionKeyStoreMutex sync.Mutex
	}
	type args struct {
		pc cluster.PartitionConsumer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "getPartitionKeyFrm: with proper inputs",
			fields: fields{
				cfg:               Config{},
				partitionKeyStore: nil,
			},
			args: args{
				pc: mockPartitionConsumer,
			},
			want: "test_999",
		},
		//{},
	}

	mockPartitionConsumer.EXPECT().Topic().Return("test").AnyTimes()
	mockPartitionConsumer.EXPECT().Partition().Return(int32(999)).AnyTimes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owr := &pullOrderedWithOffsetReplay{
				cfg:                    tt.fields.cfg,
				partitionKeyStore:      tt.fields.partitionKeyStore,
				partitionKeyStoreMutex: tt.fields.partitionKeyStoreMutex,
			}
			if got := owr.getPartitionKeyFrm(tt.args.pc); got != tt.want {
				t.Errorf("pullOrderedWithOffsetReplay.getPartitionKeyFrm() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_pullOrderedWithOffsetReplay_isPartitionAlreadyConsumed(t *testing.T) {
	type fields struct {
		cfg                    Config
		partitionKeyStore      map[string]bool
		partitionKeyStoreMutex sync.Mutex
	}
	type args struct {
		pc cluster.PartitionConsumer
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     bool
		testFunc func(bool, fields)
	}{
		{
			name: "isPartitionAlreadyConsumed: with proper data",
			fields: fields{
				cfg:               Config{},
				partitionKeyStore: make(map[string]bool),
			},
			args: args{
				pc: mockPartitionConsumer,
			},
			want: true,
			testFunc: func(want bool, f fields) {
				if want {
					f.partitionKeyStore["test_999"] = true
				}
			},
		},
		{
			name: "isPartitionAlreadyConsumed: with proper new unconsumed partition key",
			fields: fields{
				cfg:               Config{},
				partitionKeyStore: make(map[string]bool),
			},
			args: args{
				pc: mockPartitionConsumer,
			},
			want: false,
			testFunc: func(want bool, f fields) {
				if !want {
					mockPartitionConsumer.EXPECT().Topic().Return("test1").AnyTimes()
					mockPartitionConsumer.EXPECT().Partition().Return(int32(998)).AnyTimes()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(tt.want, tt.fields)
			owr := &pullOrderedWithOffsetReplay{
				cfg:                    tt.fields.cfg,
				partitionKeyStore:      tt.fields.partitionKeyStore,
				partitionKeyStoreMutex: tt.fields.partitionKeyStoreMutex,
			}
			if got := owr.isPartitionAlreadyConsumed(tt.args.pc); got != tt.want {
				t.Errorf("pullOrderedWithOffsetReplay.isPartitionAlreadyConsumed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pullOrderedWithOffsetReplay_handleOffsetforPartition(t *testing.T) {
	type fields struct {
		cfg                    Config
		partitionKeyStore      map[string]bool
		partitionKeyStoreMutex sync.Mutex
	}
	type args struct {
		sc *saramaConsumer
		pc cluster.PartitionConsumer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "handleOffsetforPartition: Return no error",
			fields: fields{
				cfg:               Config{},
				partitionKeyStore: nil,
			},
			args: args{
				pc: mockPartitionConsumer,
				sc: &saramaConsumer{},
			},
			wantErr: false,
		},
		{
			name: "handleOffsetforPartition: Return with error",
			fields: fields{
				cfg: Config{
					HandleCustomOffsetStash: func(string, int32) ([]OffsetStash, error) {
						return nil, errors.New("return fake error")
					},
					RetryCount: 1,
				},
				partitionKeyStore: nil,
			},
			args: args{
				pc: mockPartitionConsumer,
				sc: &saramaConsumer{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.sc.cfg = tt.fields.cfg
			owr := &pullOrderedWithOffsetReplay{
				cfg:                    tt.fields.cfg,
				partitionKeyStore:      tt.fields.partitionKeyStore,
				partitionKeyStoreMutex: tt.fields.partitionKeyStoreMutex,
			}
			if err := owr.handleOffsetforPartition(tt.args.sc, tt.args.pc); (err != nil) != tt.wantErr {
				t.Errorf("pullOrderedWithOffsetReplay.handleOffsetforPartition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
