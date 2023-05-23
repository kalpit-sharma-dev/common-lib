package rest

import "testing"

func TestFilePath(t *testing.T) {
	type args struct {
		pathPrefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Path /v1", want: "/v1/", args: args{pathPrefix: "/v1"}},
		{name: "Path test/v1", want: "test/v1/", args: args{pathPrefix: "test/v1"}},
		{name: "Path Blank", want: "/", args: args{pathPrefix: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilePath(tt.args.pathPrefix); got != tt.want {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
