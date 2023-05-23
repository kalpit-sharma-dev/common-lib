package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

//go:generate mockgen -package mocks -destination=mocks/log_mock.go . Log

var nameToLogger = make(map[string]*loggerImpl)

const discardLoggerName = "Discard-Logger"

const CallerKey = "Caller"

// Log is a interface to hold instace of loggerImpl
type Log interface {
	Trace(transactionID string, message string, v ...interface{})
	Debug(transactionID string, message string, v ...interface{})
	Info(transactionID string, message string, v ...interface{})
	Warn(transactionID string, message string, v ...interface{})
	Error(transactionID string, errorCode string, message string, v ...interface{})
	Fatal(transactionID string, fatalCode string, message string, v ...interface{})

	TraceC(ctx context.Context, message string, v ...interface{})
	DebugC(ctx context.Context, message string, v ...interface{})
	InfoC(ctx context.Context, message string, v ...interface{})
	WarnC(ctx context.Context, message string, v ...interface{})
	ErrorC(ctx context.Context, errorCode string, message string, v ...interface{})
	FatalC(ctx context.Context, fatalCode string, message string, v ...interface{})

	LogEvent(ctx context.Context, event Event)

	With(options ...Option) Log
	Sync() error

	GetWriter() io.WriteCloser
	SetWriter(writer io.Writer)
	LogLevel() LogLevel

	// TO BE REMOVED AFTER MIGRATION TO NEW LOGGER COMPLETES

	// TraceWithLevel log trace message with calldepth
	//
	// Deprecated: TraceWithLevel should not be used except for compatibility with legacy systems.
	// Instead use trace
	TraceWithLevel(transactionID string, calldepth int, message string, v ...interface{})
	// DebugWithLevel log debug message with calldepth
	//
	// Deprecated: DebugWithLevel should not be used except for compatibility with legacy systems.
	// Instead use debug
	DebugWithLevel(transactionID string, calldepth int, message string, v ...interface{})
	// InfoWithLevel log info message with calldepth
	//
	// Deprecated: InfoWithLevel should not be used except for compatibility with legacy systems.
	// Instead use info
	InfoWithLevel(transactionID string, calldepth int, message string, v ...interface{})
	// WarnWithLevel log warn message with calldepth
	//
	// Deprecated: WarnWithLevel should not be used except for compatibility with legacy systems.
	// Instead use warn
	WarnWithLevel(transactionID string, calldepth int, message string, v ...interface{})
	// ErrorWithLevel log trace message with calldepth
	//
	// Deprecated: ErrorWithLevel should not be used except for compatibility with legacy systems.
	// Instead use error
	ErrorWithLevel(transactionID string, calldepth int, errorCode string, message string, v ...interface{})
	// FatalWithLevel log trace message with calldepth
	//
	// Deprecated: FatalWithLevel should not be used except for compatibility with legacy systems.
	// Instead use fetal
	FatalWithLevel(transactionID string, calldepth int, fatalCode string, message string, v ...interface{})
}

type loggerImpl struct {
	writer   io.WriteCloser
	config   *Config
	hostName string

	host    Host
	service Service

	mutex sync.Mutex // ensures atomic writes; protects the following fields

	logger *zap.Logger

	data eventData
	ctx  *context.Context
}

func (l *loggerImpl) clone() loggerImpl {
	newConfig := l.config.Clone()
	copy := loggerImpl{
		writer:   l.writer,
		config:   &newConfig,
		service:  l.service,
		hostName: l.hostName,
		host:     l.host,
		logger:   l.logger,
		data:     l.data,
		ctx:      l.ctx,
	}
	return copy
}

func (l *loggerImpl) Sync() error {
	return l.logger.Sync()
}

type eventData struct {
	data interface{}
	key  string
}

type Option interface {
	apply(*loggerImpl)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*loggerImpl)

func (f optionFunc) apply(l *loggerImpl) {
	f(l)
}

// AddData add object to Event with key to logger
func AddData(key string, data interface{}) Option {
	return optionFunc(func(l *loggerImpl) {
		l.data = eventData{key: key, data: data}
	})
}

// AddContext add context data to logger
func AddContext(ctx *context.Context) Option {
	return optionFunc(func(l *loggerImpl) {
		l.ctx = ctx
	})
}

// CallDepth sets RelativeCallDepth
func CallDepth(callDepth int) Option {
	return optionFunc(func(l *loggerImpl) {
		l.config.RelativeCallDepth = callDepth
	})
}

func init() {
	Create(Config{Name: discardLoggerName, Destination: DISCARD}) // nolint
}

// Create is a function to create an instance of a loggerImpl
func Create(config Config) (Log, error) {
	if config.LogFormat == "" {
		config.setLogFormat(TextFormat)
	}
	instance, ok := nameToLogger[config.name()]

	if ok {
		return instance, fmt.Errorf("LoggerAlreadyInitialized for name :: %s and File :: %s", config.name(), config.fileName())
	}

	l := &loggerImpl{
		config: &config,
	}

	host := Host{}
	host.Name = util.Hostname(config.filler())

	service := Service{
		Name:         config.serviceName(),
		Version:      config.ServiceVersion,
		Owner:        config.ServiceOwner,
		BuildNumber:  config.BuildNumber,
		CommitSHA:    config.CommitSHA,
		BuildVersion: config.BuildVersion,
		Environment:  config.CurrentEnvironment,
	}

	l.host = host
	l.service = service
	l.hostName = host.Name

	zapConfig := GetZapConfig(&config)

	l.setOutput()

	if config.destination() == FILE {
		l.logger = buildLoggerSyncWriter(l.writer, zapConfig)
	} else {
		l.logger, _ = zapConfig.Build()
	}

	nameToLogger[config.name()] = l
	return l, nil
}

// Update is a function to Update an instance of a loggerImpl
// This creates a new logger instance if logger for given name does not exist
func Update(config Config) (Log, error) {
	instance, ok := nameToLogger[config.name()]

	if !ok {
		return Create(config)
	}

	instance.config = &config

	zapConfig := GetZapConfig(&config)

	instance.setOutput()

	service := Service{
		Name:         config.serviceName(),
		Version:      config.ServiceVersion,
		Owner:        config.ServiceOwner,
		BuildNumber:  config.BuildNumber,
		CommitSHA:    config.CommitSHA,
		BuildVersion: config.BuildVersion,
		Environment:  config.CurrentEnvironment,
	}

	instance.service = service

	if config.destination() == FILE {
		instance.logger = buildLoggerSyncWriter(instance.writer, zapConfig)
	} else {
		instance.logger, _ = zapConfig.Build()
	}

	return instance, nil
}

func buildLoggerSyncWriter(writer io.Writer, config zap.Config) *zap.Logger {
	w := zapcore.AddSync(writer)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		w,
		config.Level,
	)
	return zap.New(core, zap.AddCaller())
}

// GetZapConfig get zap default logging config
func GetZapConfig(config *Config) zap.Config {
	zapConfig := zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		ErrorOutputPaths: []string{"stderr"},
	}
	if config.LogFormat == JSONFormat {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig = zapcore.EncoderConfig{
			CallerKey:    CallerKey,
			EncodeCaller: zapcore.ShortCallerEncoder,
			EncodeTime:   zapcore.RFC3339NanoTimeEncoder,
		}
	}
	switch config.LogLevel {
	case TRACE:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case DEBUG:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case INFO:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WARN:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ERROR:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case FATAL:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DPanicLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	switch config.destination() {
	case STDOUT:
		zapConfig.OutputPaths = []string{"stdout"}
	case STDERR:
		zapConfig.OutputPaths = []string{"stderr"}
	case MEMORY:
		zapConfig.OutputPaths = []string{"memory://"}
	default:
		zapConfig.OutputPaths = []string{}
	}

	return zapConfig
}

// GetConfig is a function to return an instance of a configuration used for soecified loggerImpl
func GetConfig(name string) Config {
	if name == "" {
		name = processName()
	}

	instance, ok := nameToLogger[name]
	if ok {
		return *instance.config
	}
	return Config{}
}

// GetViaName is a function to return a logger instance for given Name
// This return a default logger instance having FILE as a writer for Process name
func GetViaName(name string) Log {
	if name == "" {
		name = processName()
	}

	instance, ok := nameToLogger[name]

	if ok {
		return instance
	}

	if name == processName() {
		instance, _ := Create(Config{}) //nolint
		return instance
	}

	panic(fmt.Errorf("logger with Name: %s does not exist", name))
}

// Get is a function to return a logger instance
// Default name will be used as a process name
func Get() Log {
	return GetViaName(processName())
}

// DiscardLogger is a function to return a logger instance having ioutil.Discard as a writer
func DiscardLogger() Log {
	return GetViaName(discardLoggerName)
}

func (l *loggerImpl) Trace(transactionID string, message string, v ...interface{}) {
	if l.config.logLevel().order >= TRACE.order {
		l.output(l.config.calldepth(), transactionID, TRACE, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) Debug(transactionID string, message string, v ...interface{}) {
	if l.config.logLevel().order >= DEBUG.order {
		l.output(l.config.calldepth(), transactionID, DEBUG, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) Info(transactionID string, message string, v ...interface{}) {
	if l.config.logLevel().order >= INFO.order {
		l.output(l.config.calldepth(), transactionID, INFO, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) Warn(transactionID string, message string, v ...interface{}) {
	if l.config.logLevel().order >= WARN.order {
		l.output(l.config.calldepth(), transactionID, WARN, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) Error(transactionID string, errorCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= ERROR.order {
		l.output(l.config.calldepth(), transactionID, ERROR, fmt.Sprintf(errorCode+" "+message, v...))
	}
}

func (l *loggerImpl) Fatal(transactionID string, fatalCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= FATAL.order {
		l.output(l.config.calldepth(), transactionID, FATAL, fmt.Sprintf(fatalCode+" "+message, v...))
	}
}

func (l *loggerImpl) TraceC(ctx context.Context, message string, v ...interface{}) {
	if l.config.logLevel().order >= TRACE.order {
		l.outputC(ctx, l.config.calldepth(), TRACE, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) DebugC(ctx context.Context, message string, v ...interface{}) {
	if l.config.logLevel().order >= DEBUG.order {
		l.outputC(ctx, l.config.calldepth(), DEBUG, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) InfoC(ctx context.Context, message string, v ...interface{}) {
	if l.config.logLevel().order >= INFO.order {
		l.outputC(ctx, l.config.calldepth(), INFO, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) WarnC(ctx context.Context, message string, v ...interface{}) {
	if l.config.logLevel().order >= WARN.order {
		l.outputC(ctx, l.config.calldepth(), WARN, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) ErrorC(ctx context.Context, errorCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= ERROR.order {
		l.outputC(ctx, l.config.calldepth(), ERROR, fmt.Sprintf(errorCode+" "+message, v...))
	}
}

func (l *loggerImpl) FatalC(ctx context.Context, fatalCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= FATAL.order {
		l.outputC(ctx, l.config.calldepth(), FATAL, fmt.Sprintf(fatalCode+" "+message, v...))
	}
}

func (l *loggerImpl) LogEvent(ctx context.Context, event Event) {
	resource := getResourceFromContext(&ctx, "")

	content := LogContent{
		Event:    event,
		Resource: resource,
	}
	calldepth := l.config.calldepth() - 1 // less call stacks for this JSON output than the usual Info/Warn/Etc functions
	level := LogLevel{order: 0, name: event.LogLevel}

	if l.config.LogFormat == JSONFormat {
		l.jsonFormat(calldepth, level, content)
	} else {
		rawMessage, err := json.Marshal(content)
		var message string

		if err != nil {
			message = fmt.Sprintf("%+v", content)
		} else {
			message = string(rawMessage)
		}
		l.textFormat(calldepth, "", level, message)
	}
}

func (l *loggerImpl) With(options ...Option) Log {
	c := l.clone()
	for _, opt := range options {
		opt.apply(&c)
	}
	return &c
}

// GetWriter is a function to return an instance of internal io.writer used by this logger
func (l *loggerImpl) GetWriter() io.WriteCloser {
	return l.writer
}

// SetWriter is a function to set a instance of internal io.writer used by this logger
func (l *loggerImpl) SetWriter(writer io.Writer) {
	//Adding locks as we only update output while creating or updating logger instance
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.writer != nil {
		l.writer.Close() //nolint
	}

	l.writer = &nopCloser{writer}
}

// LogLevel - Current log level of a Logger
func (l *loggerImpl) LogLevel() LogLevel {
	return l.config.logLevel()
}

// Helper functions used for logging
func (l *loggerImpl) setOutput() {
	//Adding locks as we only update output while creating or updating logger instance
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.writer != nil {
		l.writer.Close() //nolint
	}

	switch l.config.destination() {
	case STDOUT:
		l.writer = nopCloser{os.Stdout}
	case STDERR:
		l.writer = nopCloser{os.Stderr}
	case FILE:
		l.writer = &lumberjack.Logger{
			Filename:   l.config.fileName(),
			MaxSize:    l.config.maxSize(), // megabytes
			MaxBackups: l.config.maxBackups(),
			MaxAge:     l.config.maxAge(), //days
			Compress:   true,              // disabled by default
			LocalTime:  false,             // UTC by default
		}
	case MEMORY:
		l.writer = nopCloser{ioutil.Discard}
	default:
		l.writer = nopCloser{ioutil.Discard}
	}
}

func (l *loggerImpl) output(calldepth int, transactionID string, level LogLevel, message string) {
	var data map[string]interface{}
	if l.data.data != nil {
		data = make(map[string]interface{})
		data[l.data.key] = &l.data.data
	}

	if l.config.LogFormat == JSONFormat {
		content := structureLogData(message, data, l.ctx, transactionID)
		l.jsonFormat(calldepth, level, content)
	} else {
		l.textFormat(calldepth, transactionID, level, message)
	}
}

func (l *loggerImpl) outputC(ctx context.Context, calldepth int, level LogLevel, message string) {
	var data map[string]interface{}
	if l.data.data != nil {
		data = make(map[string]interface{})
		data[l.data.key] = &l.data.data
	}

	if l.config.LogFormat == JSONFormat {
		content := structureLogData(message, data, &ctx, "")
		l.jsonFormat(calldepth, level, content)
	} else {
		l.textFormat(calldepth, "", level, message)
	}
}

func structureLogData(message string, data map[string]interface{}, ctx *context.Context, transactionID string) LogContent {
	event := Event{
		Message: message,
	}
	if data != nil {
		event.Data = data
	}

	resource := getResourceFromContext(ctx, transactionID)

	return LogContent{
		Event:    event,
		Resource: resource,
	}
}

func getResourceFromContext(ctx *context.Context, transactionID string) Resource {
	resource := Resource{}
	if ctx != nil {
		contextData := contextutil.GetData(*ctx)

		if contextData.TransactionID == "" {
			contextData.TransactionID = transactionID
		}

		resource.PartnerID = contextData.PartnerID
		resource.UserID = contextData.UserID
		resource.CompanyID = contextData.CompanyID
		resource.SiteID = contextData.SiteID
		resource.AgentID = contextData.AgentID
		resource.EndpointID = contextData.EndpointID
		resource.ClientID = contextData.ClientID
		resource.CorrelationID = contextData.TransactionID
		resource.RequestID = contextData.RequestID
	} else {
		resource.CorrelationID = transactionID
	}

	return resource
}

func (l *loggerImpl) jsonFormat(calldepth int, level LogLevel, content LogContent) {
	content.Event.Date = currentTime()
	content.Event.LogLevel = strings.Trim(level.name, " ")

	callerSkip := zap.AddCallerSkip(calldepth)
	event := zap.Object("Event", content.Event)
	host := zap.Object("Host", l.host)
	service := zap.Object("Service", l.service)
	resource := zap.Object("Resource", content.Resource)

	switch level {
	case TRACE:
		l.logger.WithOptions(callerSkip).Debug("", event, host, service, resource)
	case DEBUG:
		l.logger.WithOptions(callerSkip).Debug("", event, host, service, resource)
	case INFO:
		l.logger.WithOptions(callerSkip).Info("", event, host, service, resource)
	case WARN:
		l.logger.WithOptions(callerSkip).Warn("", event, host, service, resource)
	case ERROR:
		l.logger.WithOptions(callerSkip).Error("", event, host, service, resource)
	case FATAL:
		// using Fatal causes process termination
		l.logger.WithOptions(callerSkip).DPanic("", event, host, service, resource)
	default:
		l.logger.WithOptions(callerSkip).Info("", event, host, service, resource)
	}
}

var sourceName = func(l *loggerImpl, calldepth int) string {
	return string(l.formatFileName(calldepth))
}

var currentTime = func() time.Time {
	return time.Now().UTC()
}

func (l *loggerImpl) textFormat(calldepth int, transactionID string, level LogLevel, message string) {
	now := currentTime() // get this early and always use UTC for logging
	buf := l.formatHeader(now, calldepth, transactionID, level)
	buf = append(buf, message...)

	if len(message) == 0 || message[len(message)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.writer.Write(buf) //nolint Ignoring this error as we dont want to handle
}

func (l *loggerImpl) formatHeader(t time.Time, calldepth int, transactionID string, level LogLevel) []byte {
	buf := make([]byte, 0)
	buf = append(buf, l.formatTime(t)...)
	buf = append(buf, ' ')
	buf = append(buf, l.hostName...)
	buf = append(buf, ' ')
	buf = append(buf, l.config.serviceName()...)
	buf = append(buf, ' ')
	buf = append(buf, transactionID...)
	buf = append(buf, ' ')
	buf = append(buf, l.formatFileName(calldepth)...)
	buf = append(buf, ' ')
	buf = append(buf, level.name...)
	buf = append(buf, ' ')
	return buf
}

func (l *loggerImpl) formatTime(t time.Time) []byte {

	buf := make([]byte, 0)

	// Add year, month, day
	year, month, day := t.Date()
	l.itoa(&buf, year, 4)
	buf = append(buf, '/')
	l.itoa(&buf, int(month), 2)
	buf = append(buf, '/')
	l.itoa(&buf, day, 2)
	buf = append(buf, ' ')

	// Add hour, min, sec
	hour, min, sec := t.Clock()
	l.itoa(&buf, hour, 2)
	buf = append(buf, ':')
	l.itoa(&buf, min, 2)
	buf = append(buf, ':')
	l.itoa(&buf, sec, 2)

	// Add microseconds
	buf = append(buf, '.')
	l.itoa(&buf, t.Nanosecond()/1e3, 6)

	return buf
}

func (l *loggerImpl) formatFileName(calldepth int) []byte {
	buf := make([]byte, 0)
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "????"
		line = 0
	}

	short := file
	pathSections := 2 //Include the file name and its package folder
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			pathSections--
			if pathSections == 0 {
				short = file[i+1:]
				break
			}
		}
	}

	buf = append(buf, short...)
	buf = append(buf, ':')
	l.itoa(&buf, line, -1)
	buf = append(buf, ':')
	return buf
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func (l *loggerImpl) itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// TO BE REMOVED AFTER MIGRATION TO NEW LOGGER COMPLETES
func (l *loggerImpl) TraceWithLevel(transactionID string, calldepth int, message string, v ...interface{}) {
	if l.config.logLevel().order >= TRACE.order {
		l.output(l.config.calldepth()+calldepth, transactionID, TRACE, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) DebugWithLevel(transactionID string, calldepth int, message string, v ...interface{}) {
	if l.config.logLevel().order >= DEBUG.order {
		l.output(l.config.calldepth()+calldepth, transactionID, DEBUG, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) InfoWithLevel(transactionID string, calldepth int, message string, v ...interface{}) {
	if l.config.logLevel().order >= INFO.order {
		l.output(l.config.calldepth()+calldepth, transactionID, INFO, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) WarnWithLevel(transactionID string, calldepth int, message string, v ...interface{}) {
	if l.config.logLevel().order >= WARN.order {
		l.output(l.config.calldepth()+calldepth, transactionID, WARN, fmt.Sprintf(message, v...))
	}
}

func (l *loggerImpl) ErrorWithLevel(transactionID string, calldepth int, errorCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= ERROR.order {
		l.output(l.config.calldepth()+calldepth, transactionID, ERROR, fmt.Sprintf(errorCode+" "+message, v...))
	}
}

func (l *loggerImpl) FatalWithLevel(transactionID string, calldepth int, fatalCode string, message string, v ...interface{}) {
	if l.config.logLevel().order >= FATAL.order {
		l.output(l.config.calldepth()+calldepth, transactionID, FATAL, fmt.Sprintf(fatalCode+" "+message, v...))
	}
}
