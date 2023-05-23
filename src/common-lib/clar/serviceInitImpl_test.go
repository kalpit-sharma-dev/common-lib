package clar

import (
	"strings"
	"testing"
)

func TestGetServiceInit(t *testing.T) {
	service := ServiceInitFactoryImpl{}.GetServiceInit()
	_, ok := service.(*serviceInit)
	if !ok {
		t.Error("serviceInit is not serviceInit")
	}
}

func Test_serviceInit_SetupOsArgs(t *testing.T) {
	type fields struct {
		configFilePath string
		configIndex    int
		logFilePath    string
		logIndex       int
		executablePath string
	}
	type args struct {
		defaultConfig string
		defaultLog    string
		args          []string
		configIdex    int
		logIndex      int
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantConfigPath     string
		wantLogFilePath    string
		wantExecutablePath string
	}{
		{name: "No Setup"},
		{name: "Default Config", args: args{defaultConfig: "defaultConfig.json"}, wantConfigPath: "defaultConfig.json"},
		{name: "Default Log", args: args{defaultLog: "defaultLog.log"}, wantLogFilePath: "defaultLog.log"},
		{
			name: "Setup Files", wantLogFilePath: "test.log", wantConfigPath: "test.json", wantExecutablePath: "test",
			args: args{configIdex: 1, logIndex: 2, args: []string{"test/test.json", "test.json", "test.log"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serviceInit{
				configFilePath: tt.fields.configFilePath,
				configIndex:    tt.fields.configIndex,
				logFilePath:    tt.fields.logFilePath,
				logIndex:       tt.fields.logIndex,
				executablePath: tt.fields.executablePath,
			}
			s.SetupOsArgs(tt.args.defaultConfig, tt.args.defaultLog, tt.args.args, tt.args.configIdex, tt.args.logIndex)

			if strings.Compare(s.GetConfigPath(), tt.wantConfigPath) != 0 {
				t.Errorf("serviceInit_SetupOsArgs.GetConfigPath() = got %v expected %v", s.GetConfigPath(), tt.wantConfigPath)
			}

			if strings.Compare(s.GetLogFilePath(), tt.wantLogFilePath) != 0 {
				t.Errorf("serviceInit_SetupOsArgs.GetLogFilePath() = got %v expected %v", s.GetLogFilePath(), tt.wantLogFilePath)
			}

			if strings.Compare(s.GetExecutablePath(), tt.wantExecutablePath) != 0 {
				t.Errorf("serviceInit_SetupOsArgs.GetExecutablePath() = got %v expected %v", s.GetExecutablePath(), tt.wantExecutablePath)
			}
		})
	}
}

func TestGetDirectoryPath(t *testing.T) {
	type args struct {
		inputPath         string
		directoryToSearch string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Invalid", args: args{inputPath: "", directoryToSearch: "Invalid"}, want: ""},
		{name: "Test no Dir", args: args{inputPath: "t", directoryToSearch: "t"}, want: ""},
		{name: "Test", args: args{inputPath: "Test/q", directoryToSearch: "Test"}, want: "Test"},
		{name: "Sub Dir", args: args{inputPath: "a/b/c/d", directoryToSearch: "c"}, want: "a/b/c"},
		{name: "Top Dir", args: args{inputPath: "a/b/c/d", directoryToSearch: "a"}, want: "a"},
		{name: "Not exist", args: args{inputPath: "a/b/c/d", directoryToSearch: "e"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDirectoryPath(tt.args.inputPath, tt.args.directoryToSearch); got != tt.want {
				t.Errorf("GetDirectoryPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
