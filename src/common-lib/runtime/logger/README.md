<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Logger

This is a Standard logger implementation used by all the Go projects in the google. So that we can implement different serach patern in the Graylog and generate alerts on any anomoly.

### Third-Party Libraties

- [lumberjack](https://gopkg.in/natefinch/lumberjack.v2) - **License** [MIT License](https://github.com/natefinch/lumberjack/blob/v2.0/LICENSE) - **Description** - Lumberjack is intended to be one part of a logging infrastructure. It is not an all-in-one solution, but instead is a pluggable component at the bottom of the logging stack that simply controls the files to which logs are written.
  Lumberjack assumes that only one process is writing to the output files. Using the same lumberjack configuration from multiple processes on the same machine will result in improper behavior.

- [zap](https://pkg.go.dev/go.uber.org/zap) - **License** [MIT License](https://github.com/uber-go/zap/blob/master/LICENSE.txt) - **Description** - Zap is an efficeinet easy Go logger. Zap is based on three concepts that optimizes the performances 1. Avoid interface{} in favor of strongly typed design. which leads to two other 2. Reflection free. Reflection comes with a cost and could be avoided since the package is aware of the used types and 3. Free of allocation in the JSON encoding. If the standard library is well optimized, allocations here could easily be avoided, as the package holds all types of the parameters sent.

### [Example](example/example.go)

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
```

**Logger Instance**

```go
// Create Logger instance
log, err := logger.Create(logger.Config{Name: name, MaxSize: 1})

// Update logger instance
log, err := logger.Create(logger.Config{Name: name, MaxSize: 1})
```

**Writing Logs**

```go
log.Trace(transaction, "This is a TRACE Message")
log.Debug(transaction, "This is a DEBUG Message")
log.Info(transaction, "This is a INFO Message")
log.Warn(transaction, "This is a WARN Message")
log.Error(transaction, "ERROR-CODE", "This is a ERROR Message")
log.Fatal(transaction, "FATAL-CODE", "This is a FATAL Message")

// Additional JSON data to be logged

// All next three examples log record JSON will look like this:
// {"Event":{"Date":"2021-01-20T15:50:12.9874688Z","Type":"INFO","Message":"This is a INFO Message","UserDataKey":{"Name":"John Smith","Age":"35"}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"CorrelationId":"test"}}

// Using a map
user := map[string]string{
	"Name": "John Smith",
	"Age":  "35",
}
log.With(logger.AddData("UserDataKey", user)).Info(transaction, "This is a INFO Message")

// Using a struct 
// Please note struct variable have to start with capital letter otherwise they will not be logged
type User struct {
	Name string
	Age  int
}
user := User{
	Name: "John Smith",
	Age:  35,
}
log.With(logger.AddData("UserDataKey", user)).Info(transaction, "This is a INFO Message")

// Using a struct when performance and type safety are critical 
// Implement zap object marshaller to the struct
type User struct {
	Name string
	Age  int
}
func (u User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("Age", u.Age)
	if u.Name != "" {
		enc.AddString("Name", u.Name)
	}
	return nil
}
user := User{
	Name: "John Smith",
	Age:  35,
}
log.With(logger.AddData("UserDataKey", user)).Info(transaction, "This is a INFO Message")

```

**Helper functions**

```go
// Return an instance of internal io.writer used by this logger
log.GetWriter() io.WriteCloser

// Set a instance of internal io.writer used by this logger
log.SetWriter(writer io.Writer)

// Current log level of a Logger
log.LogLevel() LogLevel
```

**Configuration**

```go
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
	ServiceName string `json:"servicename"`

	// Filler is a string used to fillup required logger attribute in case these are not available
	// The default is -
	Filler string `json:"filler"`

	// LogLevel is a allowed log level to write logs
	// The default is INFO
	LogLevel LogLevel `json:"logLevel"`

	// Destination is a location to write logs
	// The default is FILE
	Destination Destination `json:"destination"`

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

	// LogFormat represent different types of log format
	LogFormat LogFormat `json:"format"`

}
```

## Example

```json
"Logging": {
	"LogLevel": "INFO",
	"Destination": "FILE",
	"MaxSize": 20,
	"MaxAge": 30,
	"MaxBackups": 5,
	"Filler": "",
	"ServiceName":"<processname>",
	"Name": "<processname>",
	"FileName": "<processname>-<Name>.log"
},
```

## Example to generate JSON formatted Log

```json
"Logging": {
	"LogLevel": "INFO",
	"Destination": "FILE",
	"MaxSize": 20,
	"MaxAge": 30,
	"MaxBackups": 5,
	"Filler": "",
	"ServiceName":"<processname>",
	"Name": "<processname>",
	"FileName": "<processname>-<Name>.log",
	"format":"JSON"
},
```

## Example of json output

```json
{
   "Event":{
      "Date":"2020-10-13T08:45:45.282162636Z",
      "Caller":"logger/logger.go:198:",
      "Type":"INFO",
      "Message":"fetch entitlement feature for partner id = 50016564"
   },
   "Host":{
      "HostName":"entitlement-service-junoms-6b6c784cd-c2sdh"
   },
   "Service":{
      "Name":"EntitlementMicro-Service"
   },
   "Resource":{
      "CorrelationId":"bc77b5ad5c46f4bbd5e6c7abc9f9e034"
   }
}
```

## Example to handle panic and send log to the system
```go
func TestPanicLog(){
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n%s", r, debug.Stack())
			logger.Fatal("correlation-id", "fatal", err.Error())
		}
	}()
	fmt.Println("business logic")
	panic("panic example")
}
```

## FAQ:

**What is Transaction ID**

- This should be **business transaction id** to track complete business flow across services.
- We should have **transaction id from each of caller in the google eco-system**, in case API is the originator it should generate the new transaction id for subsequent usage
- Whenever you generate a log file, include the Transaction ID in the log message
- Transction should be an **UUID**
  - [Helper Functions](../../utils)
- Some of the examples for request originator are:
  - Agent in case of scheduler
  - Portal in case of user request
  - LRP in case of event handling
  - Job trigger point

**Where I can find Transaction ID**
- Read Transaction ID from your incoming requests, and if one is provided, send it on outgoing requests
  - `utils.GetTransactionIDFromRequest` retrieves transaction Id, if from the http request header, if this does not present it creates a new one
- If you don’t get a Transaction ID on incoming request, then generate one, and send it on outgoing requests
  - `utils.GetTransactionID` generates new transaction id

**ERROR/FATAL Code**
- Repository owners are free to define Error/Fatal codes as these will be used only for identifying logs in the Graylag

**Zap sample application logs**
Zap uses sample logs and the reason for that is applications often experience runs of errors, either because of a bug or because of a misbehaving user. Logging errors is usually a good idea, but it can easily make this bad situation worse: not only is your application coping with a flood of errors, it's also spending extra CPU cycles and I/O logging those errors. Since writes are typically serialized, logging limits throughput when you need it most.

Sampling fixes this problem by dropping repetitive log entries. Under normal conditions, your application writes out every entry. When similar entries are logged hundreds or thousands of times each second, though, zap begins dropping duplicates to preserve throughput.

### Contribution

Any changes in this package should be communicated to Juno Team.
