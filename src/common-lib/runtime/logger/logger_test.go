package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var b = &bytes.Buffer{}
var sink = &MemorySink{b}
var httpConttext = context.Background()

// configure the zap logger for capturing log output for tests
func setupLogger(config *Config, buildWithSyncWriter bool) *loggerImpl {
	b.Reset()
	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	}) //ignore error if already registered
	zapConfig := GetZapConfig(config)
	zapConfig.OutputPaths = []string{"memory://"}

	var logger *zap.Logger
	if buildWithSyncWriter {
		w, _, _ := zap.Open(zapConfig.OutputPaths...)
		logger = buildLoggerSyncWriter(w, zapConfig)
	} else {
		logger, _ = zapConfig.Build()
	}

	l := &loggerImpl{writer: nopCloser{b}, config: config, logger: logger, hostName: "hostName"}
	return l
}

func TestCreate(t *testing.T) {
	t.Run("Create Instance", func(t *testing.T) {
		_, err := Create(Config{})
		if err != nil {
			t.Errorf("Create() error = %v, wantErr %v", err, nil)
			return
		}
		_, err = Create(Config{})
		if err == nil {
			t.Errorf("Create() error = %v, wantErr %v", nil, "LoggerAlreadyInitialized")
			return
		}
	})
}

func TestGetZapConfig(t *testing.T) {
	t.Run("Get Zap Config", func(t *testing.T) {
		conf := GetZapConfig(&Config{})
		if len(conf.OutputPaths) > 0 {
			t.Errorf("GetZapConfig() want = %v, got %v", "Empty OutputPaths", conf.OutputPaths)
			return
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("Update Existing Or Create", func(t *testing.T) {
		got, err := Update(Config{Name: "Test"})
		if err != nil {
			t.Errorf("Update()  error = %v, wantErr %v", err, nil)
			return
		}

		u, err := Update(Config{Name: "Test", Destination: STDOUT})
		if err != nil {
			t.Errorf("Update()  error = %v, wantErr %v", err, nil)
			return
		}

		if !reflect.DeepEqual(got, u) {
			t.Errorf("Update() = %v, want %v", u, got)
		}

		u, err = Update(Config{Name: "Test", Destination: STDERR})
		if err != nil {
			t.Errorf("Update()  error = %v, wantErr %v", err, nil)
			return
		}

		if !reflect.DeepEqual(got, u) {
			t.Errorf("Update() = %v, want %v", u, got)
		}

		u, err = Update(Config{Name: "Test", Destination: DISCARD})
		if err != nil {
			t.Errorf("Update()  error = %v, wantErr %v", err, nil)
			return
		}

		if !reflect.DeepEqual(got, u) {
			t.Errorf("Update() = %v, want %v", u, got)
		}
	})
}

func TestGetConfig(t *testing.T) {
	oldProcessName := processName
	defer func() {
		processName = oldProcessName
	}()
	t.Run("Get Default Config", func(t *testing.T) {
		processName = func() string {
			return "logger.test"
		}
		if got := GetConfig("NOT-AVAILABLE"); !reflect.DeepEqual(got, Config{}) {
			t.Errorf("GetConfig() = %v, want %v", got, Config{})
		}
	})

	t.Run("Get Config", func(t *testing.T) {
		Update(Config{Name: "Test", ServiceName: "Test", Filler: "-", LogFormat: "Text", Destination: DISCARD, FileName: "Test"})
		processName = func() string {
			return ""
		}
		if got := GetConfig("Test"); !reflect.DeepEqual(got, Config{Name: "Test", ServiceName: "Test", Filler: "-", LogFormat: "Text", Destination: DISCARD, FileName: "Test"}) {
			t.Errorf("GetConfig() = %v, want %v", got, Config{Name: "Test", ServiceName: "Test", Filler: "-", LogFormat: "Text", Destination: DISCARD, FileName: "Test"})
		}
	})
}

func TestGetViaName(t *testing.T) {
	l, _ := Update(Config{Name: "Test", Destination: DISCARD, FileName: "Test"})
	t.Run("Get Logger Instance", func(t *testing.T) {
		if got := GetViaName("Test"); !reflect.DeepEqual(got, l) {
			t.Errorf("GetViaName() = %v, want %v", got, l)
		}
	})

	l, _ = Update(Config{Destination: DISCARD, FileName: "Test"})
	t.Run("Get Logger Instance", func(t *testing.T) {
		if got := GetViaName(""); !reflect.DeepEqual(got, l) {
			t.Errorf("GetViaName() = %v, want %v", got, l)
		}
	})

	t.Run("Get Logger Instance", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in Get Logger Instance", r)
			}
		}()
		if got := GetViaName("Test-2"); got == nil {
			t.Errorf("GetViaName() = %v, want %v", nil, "Instance")
		}
	})

	t.Run("Get Process Name Logger Instance", func(t *testing.T) {
		delete(nameToLogger, util.ProcessName())
		name := util.ProcessName()
		if got := GetViaName(name); got == nil {
			t.Errorf("GetViaName() = %v, want %v", nil, "Instance")
		}
	})
}

func TestGet(t *testing.T) {
	l, _ := Update(Config{Destination: DISCARD})
	t.Run("Get Logger Instance", func(t *testing.T) {
		if got := Get(); !reflect.DeepEqual(got, l) {
			t.Errorf("GetViaName() = %v, want %v", got, l)
		}
	})
}

func Test_loggerImpl_formatHeader(t *testing.T) {
	oldProcessName := processName
	processName = func() string {
		return "logger.test"
	}
	defer func() {
		processName = oldProcessName
	}()
	t.Run("formatHeader", func(t *testing.T) {
		date, _ := time.Parse("2006/01/02 15:04:05.999999999", "2019/04/22 11:36:11.109121")
		want := []byte("2019/04/22 11:36:11.109121 hostName logger.test transactionID logger/logger.go")
		l := &loggerImpl{
			writer:   os.Stderr,
			config:   &Config{},
			hostName: "hostName",
		}
		if got := l.formatHeader(date, 1, "transactionID", INFO); !strings.Contains(string(got), string(want)) {
			t.Errorf("loggerImpl.formatHeader() = %v, want %v", string(got), string(want))
		}
	})
}

func TestDiscardLogger(t *testing.T) {
	tests := []struct {
		name string
		want Log
	}{
		{name: "Discard", want: GetViaName(discardLoggerName)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DiscardLogger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiscardLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loggerImpl_LogMessages(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("loggerImpl_LogLevel() = %v, want %v", got, want)
			return
		}

		success = false
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() = %v, dontWant %v", got, dontWant)
		}
	}

	wantResponse := []string{
		"logger/logger_test.go:400",
	}

	t.Run("Log level TRACE", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: TRACE}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"TRACE  Test Message with test arg", "DEBUG  Test Message with test arg",
			"INFO   Test Message with test arg", "WARN   Test Message with test arg",
			"ERROR  Error Code Test Message with test arg", "FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		output := b.String()
		validate(output, want, []string{}, t)
	})

	t.Run("Log level DEBUG", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: DEBUG}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"DEBUG  Test Message with test arg", "INFO   Test Message with test arg",
			"WARN   Test Message with test arg", "ERROR  Error Code Test Message with test arg",
			"FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			"TRACE  Test Message with test arg",
			strings.Replace(wantResponse[0], ":400",
				fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
		}
		validate(b.String(), want, dontWant, t)
	})

	t.Run("Log level INFO", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: INFO}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"INFO   Test Message with test arg",
			"WARN   Test Message with test arg", "ERROR  Error Code Test Message with test arg",
			"FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{"TRACE  Test Message with test arg",
			"DEBUG  Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
		}
		validate(b.String(), want, dontWant, t)
	})

	t.Run("Log level WARN", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: WARN}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"WARN   Test Message with test arg", "ERROR  Error Code Test Message with test arg",
			"FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{"TRACE  Test Message with test arg", "DEBUG  Test Message with test arg",
			"INFO   Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
		}
		validate(b.String(), want, dontWant, t)
	})

	t.Run("Log level Error", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: ERROR}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"ERROR  Error Code Test Message with test arg", "FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1)}
		dontWant := []string{"TRACE  Test Message with test arg", "DEBUG  Test Message with test arg",
			"INFO   Test Message with test arg", "WARN   Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		validate(b.String(), want, dontWant, t)
	})

	t.Run("Log level FATAL", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: FATAL}, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1)}
		dontWant := []string{"TRACE  Test Message with test arg", "DEBUG  Test Message with test arg",
			"INFO   Test Message with test arg", "WARN   Test Message with test arg", "ERROR  Error Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
		}
		validate(b.String(), want, dontWant, t)
	})

	t.Run("Log level INFO File", func(t *testing.T) {
		l := setupLogger(&Config{LogLevel: INFO, Destination: FILE, FileName: "testLogFile"}, true)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{"INFO   Test Message with test arg",
			"WARN   Test Message with test arg", "ERROR  Error Code Test Message with test arg",
			"FATAL  Fatal Code Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{"TRACE  Test Message with test arg",
			"DEBUG  Test Message with test arg",
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
		}
		validate(b.String(), want, dontWant, t)
	})
}

func Test_loggerImpl_SetWriter(t *testing.T) {
	type fields struct {
		writer   io.WriteCloser
		config   *Config
		hostName string
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
	}{
		{name: "default"},
		{name: "buffer writer", fields: fields{writer: nopCloser{&bytes.Buffer{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loggerImpl{
				writer:   tt.fields.writer,
				config:   tt.fields.config,
				hostName: tt.fields.hostName,
			}
			writer := &bytes.Buffer{}
			l.SetWriter(writer)
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("loggerImpl.SetWriter() = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_loggerImpl_LogLevel(t *testing.T) {
	type fields struct {
		writer   io.WriteCloser
		config   *Config
		hostName string
	}
	tests := []struct {
		name   string
		fields fields
		want   LogLevel
	}{
		{name: "Default", want: INFO, fields: fields{config: &Config{}}},
		{name: "TRACE", want: TRACE, fields: fields{config: &Config{LogLevel: TRACE}}},
		{name: "DEBUG", want: DEBUG, fields: fields{config: &Config{LogLevel: DEBUG}}},
		{name: "INFO", want: INFO, fields: fields{config: &Config{LogLevel: INFO}}},
		{name: "WARN", want: WARN, fields: fields{config: &Config{LogLevel: WARN}}},
		{name: "ERROR", want: ERROR, fields: fields{config: &Config{LogLevel: ERROR}}},
		{name: "FATAL", want: FATAL, fields: fields{config: &Config{LogLevel: FATAL}}},
		{name: "OFF", want: OFF, fields: fields{config: &Config{LogLevel: OFF}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loggerImpl{
				writer:   tt.fields.writer,
				config:   tt.fields.config,
				hostName: tt.fields.hostName,
			}
			if got := l.LogLevel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loggerImpl.LogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loggerImpl_GetWriter(t *testing.T) {
	type fields struct {
		writer   io.WriteCloser
		config   *Config
		hostName string
	}
	tests := []struct {
		name   string
		fields fields
		want   io.WriteCloser
	}{
		{name: "buffer writer", want: nopCloser{&bytes.Buffer{}}, fields: fields{writer: nopCloser{&bytes.Buffer{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loggerImpl{
				writer:   tt.fields.writer,
				config:   tt.fields.config,
				hostName: tt.fields.hostName,
			}
			if got := l.GetWriter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loggerImpl.GetWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loggerImpl_LogMessages_JSON(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_loggerImpl_LogMessages_JSON() = %v, want %v", got, want)
			return
		}

		success = false
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() = %v, dontWant %v", got, dontWant)
		}
	}
	oldCurrentTime := currentTime
	defer func() {
		currentTime = oldCurrentTime
	}()
	layout := "2020-10-12T15:10:36.0342142Z"
	str := "2020-10-12T15:10:36.0342142Z"
	tvar, _ := time.Parse(layout, str)
	currentTime = func() time.Time {
		return tvar
	}

	wantResponse := []string{
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"TRACE\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"WARN\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"ERROR\",\"Message\":\"Error Code Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"FATAL\",\"Message\":\"Fatal Code Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
	}

	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	t.Run("Log level TRACE", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		output := sink.String()
		validate(output, want, []string{}, t)
	})

	t.Run("Log level DEBUG", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400:", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
		}
		output := sink.String()
		validate(output, want, dontWant, t)
	})

	t.Run("Log level INFO", func(t *testing.T) {
		config := &Config{LogLevel: INFO, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
		}
		output := sink.String()
		validate(output, want, dontWant, t)
	})

	t.Run("Log level WARN", func(t *testing.T) {
		config := &Config{LogLevel: WARN, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
		}
		validate(sink.String(), want, dontWant, t)
	})

	t.Run("Log level Error", func(t *testing.T) {
		config := &Config{LogLevel: ERROR, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		validate(sink.String(), want, dontWant, t)
	})

	t.Run("Log level FATAL", func(t *testing.T) {
		config := &Config{LogLevel: FATAL, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
		}
		validate(sink.String(), want, dontWant, t)
	})

	t.Run("Log level INFO to File", func(t *testing.T) {
		config := &Config{LogLevel: INFO, LogFormat: JSONFormat, Destination: FILE, FileName: "testLogFile"}
		l := setupLogger(config, true)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Trace("transactionID", "Test Message with %v", "test arg")
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.Warn("transactionID", "Test Message with %v", "test arg")
		l.Error("transactionID", "Error Code", "Test Message with %v", "test arg")
		l.Fatal("transactionID", "Fatal Code", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+5), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+6), 1),
		}
		dontWant := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
		}
		output := sink.String()
		validate(output, want, dontWant, t)
		os.Remove("testLogFile")
	})
}

func Test_loggerImpl_LogMessages_JSON_AddData(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_loggerImpl_LogMessages_JSON_AddData() = %v, want %v", got, want)
			return
		}

		success = false
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() = %v, dontWant %v", got, dontWant)
		}
	}
	oldCurrentTime := currentTime
	defer func() {
		currentTime = oldCurrentTime
	}()
	layout := "2020-10-12T15:10:36.0342142Z"
	str := "2020-10-12T15:10:36.0342142Z"
	tvar, _ := time.Parse(layout, str)
	currentTime = func() time.Time {
		return tvar
	}

	wantResponse := []string{
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\",\"DataKey\":{\"Name\":\"Bill\"}},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\",\"DataKey\":{\"Name\":\"Bill\"}},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
	}

	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	t.Run("Add Data Map", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		user := map[string]string{
			"Name": "Bill",
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Debug("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Info("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
	t.Run("Add Data Struct Not Encoded", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		user := User{
			Name: "Bill",
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Debug("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Info("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
	t.Run("Add Data Map Struct Encoded", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		user := UserM{
			Name: "Bill",
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Debug("transactionID", "Test Message with %v", "test arg")
		l.With(AddData("DataKey", user)).Info("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
}

func Test_loggerImpl_LogMessages_JSON_AddContext(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_loggerImpl_LogMessages_JSON_AddContext() = %v, want %v", got, want)
			return
		}

		success = false
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() = %v, dontWant %v", got, dontWant)
		}
	}
	oldCurrentTime := currentTime
	defer func() {
		currentTime = oldCurrentTime
	}()
	layout := "2020-10-12T15:10:36.0342142Z"
	str := "2020-10-12T15:10:36.0342142Z"
	tvar, _ := time.Parse(layout, str)
	currentTime = func() time.Time {
		return tvar
	}

	wantResponse := []string{
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"CorrelationId\":\"transactionID\"}}",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans123\",\"UserId\":\"usr123\",\"RequestId\":",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans123\",\"UserId\":\"usr123\",\"RequestId\":",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"transactionID\",\"UserId\":\"usr123\",\"RequestId\":",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"transactionID\",\"UserId\":\"usr123\",\"RequestId\":",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"DEBUG\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans456\",\"UserId\":\"usr123\",\"RequestId\":",
		"{\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"INFO\",\"Message\":\"Test Message with test arg\"},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans456\",\"UserId\":\"usr123\",\"RequestId\":",
	}

	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	t.Run("Add Context Using With", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.With(AddContext(&ctx)).Debug("transactionID", "Test Message with %v", "test arg")
		l.With(AddContext(&ctx)).Info("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})

	t.Run("Add Context Using With Call Depth ", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		loggerWithCallDepth(l)
		l.Debug("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
		}
		dontWant := []string{}

		output := sink.String()

		validate(output, want, dontWant, t)
	})

	t.Run("Add Context Using Logging Methods", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		_, _, lineNumber, _ := runtime.Caller(0)
		l.Debug("transactionID", "Test Message with %v", "test arg")
		l.Info("transactionID", "Test Message with %v", "test arg")
		l.DebugC(ctx, "Test Message with %v", "test arg")
		l.InfoC(ctx, "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[1], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
			strings.Replace(wantResponse[2], ":400", fmt.Sprintf("%s%d", ":", lineNumber+3), 1),
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
	t.Run("Add Context Override Transaction Id if Empty", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctxDataMap.TransactionID = ""
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		lw := l.With(AddContext(&ctx))
		_, _, lineNumber, _ := runtime.Caller(0)
		lw.Debug("transactionID", "Test Message with %v", "test arg")
		lw.Info("transactionID", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[4], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[5], ":400", fmt.Sprintf("%s%d", ":", lineNumber+2), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
	t.Run("Add Context Update Context", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		lw := l.With(AddContext(&ctx))
		_, _, lineNumber, _ := runtime.Caller(0)
		lw.Info("", "Test Message with %v", "test arg")
		ctxDataMap.TransactionID = "trans456"
		ctx = contextutil.WithValue(httpConttext, ctxDataMap)
		lw.Info("", "Test Message with %v", "test arg")

		want := []string{
			strings.Replace(wantResponse[3], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			strings.Replace(wantResponse[7], ":400", fmt.Sprintf("%s%d", ":", lineNumber+4), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
}
func Test_loggerImpl_LogEvent(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_loggerImpl_LogEvent() = %v, want %v", got, want)
			return
		}

		success = false
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() = %v, dontWant %v", got, dontWant)
		}
	}
	oldCurrentTime := currentTime
	defer func() {
		currentTime = oldCurrentTime
	}()
	layout := "2020-10-12T15:10:36.0342142Z"
	str := "2020-10-12T15:10:36.0342142Z"
	tvar, _ := time.Parse(layout, str)
	currentTime = func() time.Time {
		return tvar
	}

	wantResponse := []string{
		"\"" + CallerKey + "\":\"logger/logger_test.go:400\",\"Event\":{\"Date\":\"0001-01-01T00:00:00Z\",\"Type\":\"AUDIT\",\"Message\":\"Test Message\",\"Audit\":{\"Name\":\"Test Name\"}},\"Host\":{},\"Service\":{},\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans123\",\"UserId\":\"usr123\",\"RequestId\":",
	}

	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	t.Run("Add Context Using With", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		event := Event{
			Message:  "Test Message",
			Audit:    UserM{Name: "Test Name"},
			LogLevel: "AUDIT",
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		l.LogEvent(ctx, event)
		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})

	t.Run("Add Context Using With", func(t *testing.T) {
		config := &Config{LogLevel: TRACE, LogFormat: JSONFormat}
		l := setupLogger(config, false)
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		event := Event{
			Message:  "Test Message",
			Audit:    UserM{Name: "Test Name"},
			LogLevel: "AUDIT",
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		l.LogEvent(ctx, event)
		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
		}
		dontWant := []string{}

		output := sink.String()
		validate(output, want, dontWant, t)
	})
}

func Test_loggerImpl_LogMessages_CallDepth(t *testing.T) {
	validate := func(got string, want []string, dontWant []string, t *testing.T) {
		success := true
		missing := make([]string, 0)
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
				missing = append(missing, w)
			}
		}

		if !success {
			t.Errorf("Test_loggerImpl_LogMessages_CallDepth() =\n%v\nwant =\n%v\nmissing =\n%v", got, want, missing)
			return
		}

		success = false
		found := make([]string, 0)
		for _, w := range dontWant {
			if strings.Contains(got, w) {
				success = true
				found = append(missing, w)
			}
		}

		if success {
			t.Errorf("loggerImpl_LogLevel() =\n%v\ndontWant =\n%v\nfound =\n%v", got, dontWant, found)
		}
	}

	emitLogs := func(l Log) int {
		_, _, lineNumber, _ := runtime.Caller(0)
		wrappedTrace(l, "Test Message")
		wrappedDebug(l, "Test Message")
		wrappedInfo(l, "Test Message")
		wrappedWarn(l, "Test Message")
		wrappedError(l, "Test Message")
		wrappedFatal(l, "Test Message")
		wrappedEvent(l, "Test Message")
		return lineNumber
	}

	wantedSource := "logger/logger_test.go"
	buildWantedJsonResponse := func(source string, startingLineNumber int) (want []string, dontWant []string) {
		return []string{
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+1),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+2),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+3),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+4),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+5),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+6),
			fmt.Sprintf("\""+CallerKey+"\":\"%s:%d\"", source, startingLineNumber+7),
		}, []string{}
	}
	buildWantedTextResponse := func(source string, startingLineNumber int) (want []string, dontWant []string) {
		return []string{
			fmt.Sprintf("%s:%d:", source, startingLineNumber+1),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+2),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+3),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+4),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+5),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+6),
			fmt.Sprintf("%s:%d:", source, startingLineNumber+7),
		}, []string{}
	}
	buildCallDepthFailure := func(source string, startingLineNumber int) (want []string, dontWant []string) {
		return []string{
				fmt.Sprintf("TRACE"),
				fmt.Sprintf("DEBUG"),
				fmt.Sprintf("INFO"),
				fmt.Sprintf("WARN"),
				fmt.Sprintf("ERROR"),
				fmt.Sprintf("FATAL"),
				fmt.Sprintf("Event"),
			}, []string{
				fmt.Sprintf("%s:%d", source, startingLineNumber+1),
				fmt.Sprintf("%s:%d", source, startingLineNumber+2),
				fmt.Sprintf("%s:%d", source, startingLineNumber+3),
				fmt.Sprintf("%s:%d", source, startingLineNumber+4),
				fmt.Sprintf("%s:%d", source, startingLineNumber+5),
				fmt.Sprintf("%s:%d", source, startingLineNumber+6),
				fmt.Sprintf("%s:%d", source, startingLineNumber+7),
			}
	}

	wrappedCallDepth := 1 //+1 for wrapped trace/debug/etc functions that are called in emitLogs
	tests := []struct {
		name string
		arg  *Config
		want func(source string, startingLineNumber int) (want []string, dontWant []string)
	}{
		{
			name: "Custom_CallDepth_JSON",
			arg: &Config{
				LogLevel: TRACE, LogFormat: JSONFormat, RelativeCallDepth: wrappedCallDepth,
			},
			want: buildWantedJsonResponse,
		},
		{
			name: "Custom_CallDepth_Text",
			arg: &Config{
				LogLevel: TRACE, LogFormat: TextFormat, RelativeCallDepth: wrappedCallDepth,
			},
			want: buildWantedTextResponse,
		},
		{
			name: "Custom_CallDepth_TooDeep_JSON",
			arg: &Config{
				LogLevel: TRACE, LogFormat: JSONFormat, RelativeCallDepth: wrappedCallDepth + 10,
			},
			want: buildCallDepthFailure,
		},
		{
			name: "Custom_CallDepth_TooDeep_Text",
			arg: &Config{
				LogLevel: TRACE, LogFormat: TextFormat, RelativeCallDepth: wrappedCallDepth + 10,
			},
			want: buildCallDepthFailure,
		},
		{
			name: "Custom_CallDepth_TooShallow_JSON",
			arg: &Config{
				LogLevel: TRACE, LogFormat: JSONFormat, RelativeCallDepth: wrappedCallDepth - 20,
			},
			want: buildCallDepthFailure,
		},
		{
			name: "Custom_CallDepth_TooShallow_Text",
			arg: &Config{
				LogLevel: TRACE, LogFormat: TextFormat, RelativeCallDepth: wrappedCallDepth - 20,
			},
			want: buildCallDepthFailure,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l *loggerImpl
			l = setupLogger(tt.arg, false)

			lineNumber := emitLogs(l)

			var want, dontWant = tt.want(wantedSource, lineNumber)
			var output string
			if tt.arg.LogFormat == JSONFormat {
				output = sink.String()
			} else {
				output = b.String()
			}

			validate(output, want, dontWant, t)
		})
	}
}

type MemorySink struct {
	*bytes.Buffer
}

func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

type User struct {
	Name string
}

type UserM struct {
	Name string
}

// MarshalLogObject Marshal Resource to zap Object
func (u UserM) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if u.Name != "" {
		enc.AddString("Name", u.Name)
	}
	return nil
}

//wrapped log functions for additional call depth testing

func wrappedTrace(l Log, message string) {
	l.Trace("transactionId", message)
}

func wrappedDebug(l Log, message string) {
	l.Debug("transactionId", message)
}

func wrappedInfo(l Log, message string) {
	l.Info("transactionId", message)
}

func wrappedWarn(l Log, message string) {
	l.Warn("transactionId", message)
}

func wrappedError(l Log, message string) {
	l.Error("transactionId", "Error Code", message)
}

func wrappedFatal(l Log, message string) {
	l.Fatal("transactionId", "Fatal Code", message)
}

func wrappedEvent(l Log, message string) {
	l.LogEvent(context.Background(), Event{Message: message})
}

func loggerWithCallDepth(l *loggerImpl) {
	l.With(CallDepth(1)).Debug("transactionID", "Test Message with %v", "test arg")
	l.With(CallDepth(1)).Info("transactionID", "Test Message with %v", "test arg")
}
