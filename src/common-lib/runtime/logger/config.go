package logger

import (
	"path/filepath"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
)

// LogFormat represent log structure
type LogFormat string

// represent different types of log format
const (
	TextFormat LogFormat = "Text"
	JSONFormat LogFormat = "JSON"
)

// Config is a struct to hold logger configuration
type Config struct {
	// Name is a user defined name of logger, this is used to handle multiple log files.
	// The default is <processname>, if empty
	Name string `json:"name"`

	// FileName is the file to write logs to and backup log files will be retained in the same directory.
	// It uses <processname>-Name.log in the same directory where process binary available, if empty.
	FileName string `json:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	//It defaults to 20 megabytes.
	MaxSize int `json:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	// The default is 30 Days to remove old log files based on age.
	MaxAge int `json:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.
	// The default is to retain 5 old log files (though MaxAge may still cause them to get deleted)
	MaxBackups int `json:"maxbackups"`

	// ServiceName is a user defined name of Service.
	// The default is <processname>, if empty
	ServiceName        string `json:"servicename"`
	ServiceVersion     string `json:"serviceversion"`
	ServiceOwner       string `json:"serviceowner"`
	CommitSHA          string `json:"commitid"`
	BuildNumber        string `json:"buildnumber"`
	BuildVersion       string `json:"buildversion"`
	CurrentEnvironment string `json:"environment"`

	// Filler is a string used to fillup required logger attribute in case these are not available
	// The default is -
	Filler string `json:"filler"`

	// LogLevel is a allowed log level to write logs
	// The default is INFO
	LogLevel LogLevel `json:"logLevel"`

	// Destination is a location to write logs
	// The default is FILE
	Destination Destination `json:"destination"`

	// LogFormat represent different types of log format
	LogFormat LogFormat `json:"format"`

	// RelativeCallDepth is the depth needed for the runtime to find a caller file name and line number.
	// This depth is relative to the logger package; the default is 0 and for most cases can be left as-is.
	// If you wrap this logger with your own logging code, you will need an additional +1
	// for each function that wraps this logger to get the actual source value.
	//
	// Example:
	//
	// 	l.Debug(ctx, "This message requires a depth of 0 to report this line in Source, no RelativeCallDepth customization needed")
	//  func myLogger1(ctx context.Context, l Log, m string) { l.Debug(ctx, "My custom wrapped message: " + m) }
	//  myLogger1(ctx, l, "This message requires a depth of 1 to report this line in Source")
	//  func myLogger2(ctx context.Context, l Log, m string) { myLogger1(ctx, "My even more custom wrapped message: " + m) }
	//  myLogger2(ctx, l, "This message requires a depth of 2 to report this line in Source")
	RelativeCallDepth int
}

// The default absolute depths for the logger implementation to report calling code's line number and file as the Source.
const (
	MinCallDepth     = 5
	MinCallDepthJSON = 3
)

var processName = func() string {
	return util.ProcessName()
}

// setLogFormat set log format structure
func (c *Config) setLogFormat(format LogFormat) {
	c.LogFormat = format
}

// name generates the name of the logger from the process Name.
func (c *Config) name() string {
	if c.Name == "" {
		c.Name = processName()
	}
	return c.Name
}

// filename generates the name of the logfile from the process Name.
func (c *Config) fileName() string {
	if c.FileName == "" {
		loggerName := c.name()
		process := processName()
		if process != loggerName {
			loggerName = strings.Replace(loggerName, " ", "", -1)
			process = process + "-" + loggerName
		}
		name := process + ".log"
		c.FileName = filepath.Join(util.InvocationPath(), name)
	}
	return c.FileName
}

// maxSize generates value for max size of a log file
func (c *Config) maxSize() int {
	if c.MaxSize <= 0 {
		c.MaxSize = 20
	}
	return c.MaxSize
}

// maxAge generates value for max Age
func (c *Config) maxAge() int {
	if c.MaxAge <= 0 {
		c.MaxAge = 30
	}
	return c.MaxAge
}

// maxBackups generates value for max Backups.
func (c *Config) maxBackups() int {
	if c.MaxBackups <= 0 {
		c.MaxBackups = 5
	}
	return c.MaxBackups
}

// serviceName generates the name of the service from the process Name.
func (c *Config) serviceName() string {
	if c.ServiceName == "" {
		c.ServiceName = processName()
	}
	return c.ServiceName
}

// filler generates the name of the filler.
func (c *Config) filler() string {
	if c.Filler == "" {
		c.Filler = "-"
	}
	return c.Filler
}

// logLevel generates the default value for the log-level
func (c *Config) logLevel() LogLevel {
	if c.LogLevel.name != "" {
		return c.LogLevel
	}
	return INFO
}

// destination generates the default value for the destination
func (c *Config) destination() Destination {
	if c.Destination.order <= 0 || c.Destination.order > 4 {
		c.Destination = FILE
	}
	return c.Destination
}

// calldepth generates the default value for the calldepth.
func (c *Config) calldepth() int {
	depth := MinCallDepth

	if c.LogFormat == JSONFormat {
		depth = MinCallDepthJSON
	}

	if c.RelativeCallDepth > 0 {
		depth += c.RelativeCallDepth
	}

	return depth
}

// LogLevel is a struct to enable or disable log level
type LogLevel struct {
	order int
	name  string
	value string
}

// UnmarshalJSON is a function to unmarshal Loglevel
// The default value is INFO
func (l *LogLevel) UnmarshalJSON(data []byte) error {
	level := strings.ToUpper(strings.TrimSpace(string(data)))
	lv := INFO
	switch level {
	case OFF.value:
		lv = OFF
	case FATAL.value:
		lv = FATAL
	case ERROR.value:
		lv = ERROR
	case WARN.value:
		lv = WARN
	case INFO.value:
		lv = INFO
	case DEBUG.value:
		lv = DEBUG
	case TRACE.value:
		lv = TRACE
	}

	l.order = lv.order
	l.name = lv.name
	l.value = lv.value

	return nil
}

// MarshalJSON is a function to marshal Loglevel
func (l LogLevel) MarshalJSON() ([]byte, error) {
	return []byte(l.value), nil
}

var (
	// OFF is used for disabling all the logs from a process
	OFF = LogLevel{0, "OFF   ", "\"OFF\""}
	// FATAL is used for writing only fatal logs from the process
	FATAL = LogLevel{10, "FATAL ", "\"FATAL\""}
	// ERROR is used for writing only error and fatal logs from the process
	ERROR = LogLevel{20, "ERROR ", "\"ERROR\""}
	// WARN is used for writing only warn, error and fatal logs from the process
	WARN = LogLevel{30, "WARN  ", "\"WARN\""}
	// INFO is used for writing only info, warn, error and fatal logs from the process
	INFO = LogLevel{40, "INFO  ", "\"INFO\""}
	// DEBUG is used for writing only debug, info, warn, error and fatal logs from the process
	DEBUG = LogLevel{50, "DEBUG ", "\"DEBUG\""}
	// TRACE is used for writing only all the logs from the process
	TRACE = LogLevel{60, "TRACE ", "\"TRACE\""}
)

// Destination is a struct to hold logger destination
type Destination struct {
	order int
	value string
}

var (
	// FILE is a log destination; which indicates that we need to write logs in the File
	FILE = Destination{order: 1, value: "\"FILE\""}
	// STDOUT is a log destination; which indicates that we need to write logs in the Standard Output Stream
	STDOUT = Destination{order: 2, value: "\"STDOUT\""}
	// STDERR is a log destination; which indicates that we need to write logs in the Standard Error Stream
	STDERR = Destination{order: 3, value: "\"STDERR\""}
	// DISCARD is a log destination; which indicates that we need to disscard logs written on this
	DISCARD = Destination{order: 4, value: "\"DISCARD\""}
	// MEMORY is a log destination; which indicates that we need to write logs in the Memory
	// It's useful for applications that want to unit test their log output.
	MEMORY = Destination{order: 2, value: "\"MEMORY\""}
)

// UnmarshalJSON is a function to unmarshal Destination
// The default value is FILE
func (d *Destination) UnmarshalJSON(data []byte) error {
	dest := strings.ToUpper(strings.TrimSpace(string(data)))
	dst := FILE
	switch dest {
	case FILE.value:
		dst = FILE
	case STDOUT.value:
		dst = STDOUT
	case STDERR.value:
		dst = STDERR
	case DISCARD.value:
		dst = DISCARD
	}

	d.value = dst.value
	d.order = dst.order
	return nil
}

// MarshalJSON is a function to marshal Destination
func (d Destination) MarshalJSON() ([]byte, error) {
	return []byte(d.value), nil
}

// clone  to clone all values of config to a new config struct
func (c *Config) Clone() Config {
	newC := Config{
		Name:               c.Name,
		FileName:           c.FileName,
		MaxSize:            c.MaxSize,
		MaxAge:             c.MaxAge,
		MaxBackups:         c.MaxBackups,
		ServiceName:        c.ServiceName,
		ServiceVersion:     c.ServiceVersion,
		ServiceOwner:       c.ServiceOwner,
		CommitSHA:          c.CommitSHA,
		BuildNumber:        c.BuildNumber,
		BuildVersion:       c.BuildVersion,
		CurrentEnvironment: c.CurrentEnvironment,
		Filler:             c.Filler,
		LogLevel:           c.LogLevel,
		Destination:        c.Destination,
		LogFormat:          c.LogFormat,
		RelativeCallDepth:  c.RelativeCallDepth,
	}
	return newC
}
