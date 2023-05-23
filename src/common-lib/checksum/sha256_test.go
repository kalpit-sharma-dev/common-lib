package checksum

import (
	"io"
	"strings"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func Test_sha256Impl_Calculate(t *testing.T) {
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
		{name: "calculate Fail", args: args{reader: fakeReader{}}, want: "", wantErr: true},
		{
			name: "calculate Success", args: args{strings.NewReader("Tests")},
			want: "e5c9d7030bada2fe792ad7e4a26ca44379a73f3fe16805ac1359e46759a1571a", wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sha256Impl{}
			got, err := c.Calculate(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("sha256Impl.Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sha256Impl.Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sha256Impl_Validate(t *testing.T) {
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
		{name: "validate Fail", args: args{reader: fakeReader{}}, want: false, wantErr: true},
		{
			name: "validate Success", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "e5c9d7030bada2fe792ad7e4a26ca44379a73f3fe16805ac1359e46759a1571a"},
		},
		{
			name: "validate Success upper case", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "e5c9d7030bada2fe792ad7e4a26ca44379a73f3fe16805ac1359e46759a1571a"},
		},
		{
			name: "validate Failed", want: false, wantErr: true,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792de52961c34118f976ebe4af3a34"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sha256Impl{}
			got, err := c.Validate(tt.args.reader, tt.args.checksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("sha256Impl.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sha256Impl.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
