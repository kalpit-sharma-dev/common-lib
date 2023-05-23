package checksum

import (
	"io"
	"strings"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func Test_none_Calculate(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "calculate Success",
			args:    args{strings.NewReader("Tests")},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := none{}
			got, err := c.Calculate(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("none.Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("none.Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_none_Validate(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	type args struct {
		reader   io.Reader
		checksum string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "validate Success",
			args:    args{reader: strings.NewReader("Tests"), checksum: "39fdec1194d94212b871a28b2aa04a73cd40fce1"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "validate Success upper case",
			args:    args{reader: strings.NewReader("Tests"), checksum: "39FDEC1194D94212b871A28B2aa04a73cd40fce1"},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := none{}
			got, err := c.Validate(tt.args.reader, tt.args.checksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("none.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("none.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
