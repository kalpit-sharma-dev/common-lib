package checksum

import (
	"errors"
	"io"
	"strings"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

type fakeReader struct{}

// here's a fake ReadFile method that matches the signature of ioutil.ReadFile
func (f fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("Error")
}

func Test_md5Impl_Calculate(t *testing.T) {
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
			name: "calculate Success", args: args{strings.NewReader("Tests")},
			want: "90792de52961c34118f976ebe4af3a75", wantErr: false,
		},
		{name: "calculate Fail", args: args{fakeReader{}}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := md5Impl{}
			got, err := c.Calculate(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("md5Impl.Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("md5Impl.Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_md5Impl_Validate(t *testing.T) {
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
			name: "validate Success", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792de52961c34118f976ebe4af3a75"},
		},
		{
			name: "validate Success Upper case", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792DE52961C34118f976EBE4AF3A75"},
		},
		{
			name: "validate Failed", want: false, wantErr: true,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792de52961c34118f976ebe4af3a34"},
		},
		{
			name: "validate Trimming Success", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792DE52961C34118f976EBE4AF3A75\n"},
		},
		{
			name: "validate Trimming Failed", want: false, wantErr: true,
			args: args{reader: strings.NewReader("Tests"), checksum: "90792DE52961C34118f976EBE4AF3A76\n"},
		},
		{
			name: "validate Trimming", want: true, wantErr: false,
			args: args{reader: strings.NewReader("Tests"), checksum: "\t \n 90792DE52961C34118f976EBE4AF3A75\n\t \t"},
		},
		{name: "validate Fail", args: args{reader: fakeReader{}}, want: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := md5Impl{}
			got, err := c.Validate(tt.args.reader, tt.args.checksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("md5Impl.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("md5Impl.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
