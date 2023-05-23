package util

import (
	"os"
	"testing"
)

func TestProcessName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Test", want: "util.test"},
	}
	oldFileBasePath := fileBasePath
	fileBasePath = func() string {
		return "util.test"
	}
	defer func() {
		fileBasePath = oldFileBasePath
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessName(); got != tt.want {
				t.Errorf("ProcessName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvocationPath(t *testing.T) {
	tests := []struct {
		name           string
		want           string
		invocationPath string
	}{
		{name: "Test", want: "test", invocationPath: "test"},
		{name: "Test", want: "test", invocationPath: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invocationPath = tt.invocationPath
			os.Args = []string{"test/test"}
			if got := InvocationPath(); got != tt.want {
				t.Errorf("InvocationPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotifyStopSignal(t *testing.T) {
	stopch := make(chan bool, 1)
	cb := func() {}
	type args struct {
		stop     chan bool
		callback []func()
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "1 Stop channel nil", wantErr: true, setup: func() {}, args: args{stop: nil, callback: []func(){}}},
		{name: "2 Stop signal", wantErr: false, setup: func() { stopch <- true }, args: args{stop: stopch, callback: []func(){}}},
		{name: "3 Stop signal with multiple callbacks", wantErr: false, setup: func() { stopch <- true }, args: args{stop: stopch, callback: []func(){cb, cb}}},
		{name: "4 Stop signal with nil callbacks", wantErr: true, setup: func() { stopch <- true }, args: args{stop: stopch, callback: nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := NotifyStopSignal(tt.args.stop, tt.args.callback...); (err != nil) != tt.wantErr {
				t.Errorf("NotifyStopSignal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalIPAddress(t *testing.T) {
	t.Run("Local Ip Address", func(t *testing.T) {
		got := LocalIPAddress()
		if len(got) == 0 {
			t.Errorf("LocalIPAddress() = Expected IP Address got nil")
		}
	})
}

func TestHostname(t *testing.T) {
	t.Run("Blank Host name", func(t *testing.T) {
		want, _ := os.Hostname()
		if got := Hostname("defaultValue"); got != want {
			t.Errorf("Hostname() = %v, want %v", got, want)
		}
	})

	t.Run("Existing Host name", func(t *testing.T) {
		want := "Hostname"
		hostName = "Hostname"
		if got := Hostname("defaultValue"); got != want {
			t.Errorf("Hostname() = %v, want %v", got, want)
		}
	})
}
