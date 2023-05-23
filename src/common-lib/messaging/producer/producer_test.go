package producer

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ error = &DeliveryError{}

func Test_ErrConsts(t *testing.T) {
	tests := []struct {
		name string
		arg  error
		want error
	}{
		{
			name: "ErrPublishMessageNotAvailable",
			arg:  ErrPublishMessageNotAvailable,
			want: errors.New("Publish:Message:Not.Available"),
		},
		{
			name: "ErrPublishSendMessageRecovered",
			arg:  ErrPublishSendMessageRecovered,
			want: errors.New("Publish:SendMessage:Recovered"),
		},
		{
			name: "ErrPublishSendMessageTimeout",
			arg:  ErrPublishSendMessageTimeout,
			want: errors.New("Publish:SendMessage:Timeout"),
		},
		{
			name: "ErrPublishSendMessageFatal",
			arg:  ErrPublishSendMessageFatal,
			want: errors.New("Publish:SendMessage:Fatal"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arg.Error() != tt.want.Error() {
				t.Errorf("%s expected %s, got %s", tt.name, tt.want, tt.arg)
			}
		})
	}
}

func Test_TypeConsts(t *testing.T) {
	tests := []struct {
		name string
		arg  Type
		want Type
	}{
		{
			name: "RegularKafkaProducer",
			arg:  RegularKafkaProducer,
			want: Type("regular"),
		},
		{
			name: "BigKafkaProducer",
			arg:  BigKafkaProducer,
			want: Type("big"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arg != tt.want {
				t.Errorf("%s expected %s, got %s", tt.name, tt.want, tt.arg)
			}
		})
	}
}

func Test_SyncProducerFunc(t *testing.T) {
	tests := []struct {
		name string
		arg  Type
	}{
		{
			name: "RegularKafkaProducer",
			arg:  RegularKafkaProducer,
		},
		{
			name: "BigKafkaProducer",
			arg:  BigKafkaProducer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			p, err := SyncProducer(tt.arg, c)
			if err != nil {
				t.Errorf("SyncProducer(\"%s\", ..) err != nil", tt.arg)
			}
			if p == nil {
				t.Errorf("SyncProducer(\"%s\", ..) p == nil", tt.arg)
			}
			p.Close()
		})
	}
}

func Test_newKafkaProducer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			p, err := newKafkaProducer(c, func(b bool) {})
			if err != nil {
				t.Errorf("newKafkaProducer(..) err != nil")
			}
			if p == nil {
				t.Errorf("newKafkaProducer(..) p == nil")
			}
			p.Close()
		})
	}
}

func TestDeliveryError_Error(t *testing.T) {
	t.Run("nil_err", func(t *testing.T) {
		var e *DeliveryError
		got := e.Error()
		require.Equal(t, "", got)
	})

	type fields struct {
		FailedReports []DeliveryReport
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "no_reports",
			fields: fields{},
			want:   "",
		},
		{
			name: "good",
			fields: fields{
				FailedReports: []DeliveryReport{
					{
						Error: errors.New("oops1"),
					},
					{
						Error: errors.New("oops2"),
					},
				},
			},
			want: "oops1; oops2",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			e := &DeliveryError{
				FailedReports: tt.fields.FailedReports,
			}

			got := e.Error()

			require.Equal(t, tt.want, got)
		})
	}
}
