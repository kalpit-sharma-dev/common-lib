package main

import (
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"go.uber.org/zap/zapcore"
)

var destination = logger.STDOUT
var format = logger.TextFormat

func main() {
	createLogger([]string{"Logger-1", "Logger-2"})
	printMessage("Logger-1", "test")
	printMessageWithParam("Logger-2", "test")
	printMessageWithAdditionalData("Logger-3", "test")

	updateLogLevel("Logger-1", "test", logger.DEBUG)
	printMessage("Logger-1", "test")
	updateLogLevel("Logger-2", "test", logger.DEBUG)
	printMessageWithParam("Logger-2", "test")

	updateLogLevel("Logger-1", "test", logger.TRACE)
	printMessage("Logger-1", "test")
	updateLogLevel("Logger-2", "test", logger.TRACE)
	printMessageWithParam("Logger-2", "test")

	updateLogLevel("Logger-1", "test", logger.WARN)
	printMessage("Logger-2", "test")
	updateLogLevel("Logger-2", "test1", logger.WARN)
	printMessageWithParam("Logger-1", "test")

	updateLogLevel("Logger-1", "test", logger.ERROR)
	printMessage("Logger-1", "test")
	updateLogLevel("Logger-2", "test", logger.ERROR)
	printMessageWithParam("Logger-2", "test")

	updateLogLevel("Logger-1", "test", logger.FATAL)
	printMessage("Logger-1", "test")
	updateLogLevel("Logger-2", "test", logger.FATAL)
	printMessageWithParam("Logger-2", "test")

	updateLogLevel("Logger-1", "test", logger.OFF)
	printMessage("Logger-1", "test")
	updateLogLevel("Logger-2", "test", logger.OFF)
	printMessageWithParam("Logger-2", "test")

	flushLogger([]string{"Logger-1", "Logger-2"})
}

// createLogger is a function to explain how to create a logger instance
func createLogger(names []string) {
	for _, name := range names {
		_, err := logger.Create(logger.Config{Name: name, MaxSize: 1, Destination: destination, LogFormat: format})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func flushLogger(names []string) {
	for _, name := range names {
		log := logger.GetViaName(name)
		err := log.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}
}

// printMessage is a function to showcase all message type prinitng
func printMessage(loggerName string, transaction string) {
	log := logger.GetViaName(loggerName)
	log.Trace(transaction, "This is a TRACE Message")
	log.Debug(transaction, "This is a DEBUG Message")
	log.Info(transaction, "This is a INFO Message")
	log.Warn(transaction, "This is a WARN Message")
	log.Error(transaction, "ERROR-CODE", "This is a ERROR Message")
	log.Fatal(transaction, "FATAL-CODE", "This is a FATAL Message")
}

// printMessageWithParam is a function to showcase all message type prinitng
func printMessageWithParam(loggerName string, transaction string) {
	log := logger.GetViaName(loggerName)
	log.Trace(transaction, "This is a %s Message", "TRACE")
	log.Debug(transaction, "This is a %s Message", "DEBUG")
	log.Info(transaction, "This is a %s Message", "INFO")
	log.Warn(transaction, "This is a %s Message", "WARN")
	log.Error(transaction, "ERROR-CODE", "This is a %s Message", "ERROR")
	log.Fatal(transaction, "FATAL-CODE", "This is a %s Message", "FATAL")
}

// printMessageWithParam is a function to showcase all message type prinitng
func printMessageWithAdditionalData(loggerName string, transaction string) {
	log := logger.GetViaName(loggerName)

	// Using a map
	user1 := map[string]string{
		"Name": "John Smith",
		"Age":  "35",
	}
	log.With(logger.AddData("UserInfo", user1)).Info(transaction, "This is a Message user object is a map")

	// Using a struct
	user2 := Person{
		Name: "John Smith",
		Age:  35,
	}
	log.With(logger.AddData("UserInfo", user2)).Info(transaction, "This is a Message user object is a struct")

	// Using a struct when performance and type safety are critical
	// Implement zap object marshaller to the struct
	user3 := User{
		Name: "John Smith",
		Age:  35,
	}
	log.With(logger.AddData("UserInfo", user3)).Info(transaction, "This is a Message user object is a struct implements MarshalLogObject")
}

// updateLogLevel is a function to showcase how to update log level
func updateLogLevel(loggerName string, transaction string, loglevel logger.LogLevel) {
	log, _ := logger.Update(logger.Config{Name: loggerName, MaxSize: 1, LogLevel: loglevel, Destination: destination, LogFormat: format}) //nolint
	log.Info(transaction, "-----------------------------------------------------")
	log.Info(transaction, "Update Loglevel to %v", loglevel)
	log.Info(transaction, "-----------------------------------------------------")
}

// Person struct
type Person struct {
	Name string
	Age  int
}

// User struct
type User struct {
	Name string
	Age  int
}

// MarshalLogObject Marshal Resource to zap Object
func (u User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("Age", u.Age)
	if u.Name != "" {
		enc.AddString("Name", u.Name)
	}
	return nil
}
