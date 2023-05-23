package with

import (
	"fmt"
	"runtime/debug"
)

//Recover is a function to call any function with recovery and gives callback to @Handler by passing Error Message and Stack Trace
func Recover(name string, transaction string, fn func(), handler func(transaction string, err error)) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%s-Recovered:%v Stack Trace :: %s", name, r, debug.Stack())
			handle(name, transaction, err, handler)
		}
	}()
	fn()
}

func handle(name string, transaction string, trace error, handler func(transaction string, err error)) {
	log := log()
	log.Info(transaction, "%s Recovered and Handler called", name)

	defer func() {
		if err := recover(); err != nil {
			log.Fatal(transaction, fmt.Sprintf("%s-handlerRecovered", name), "handleRecovered : Routine %s, error %v, trace : %s", name, err, debug.Stack())
		}
	}()

	if handler != nil {
		handler(transaction, trace)
	} else {
		log.Fatal(transaction, "%s-Recovered-No-Handler", "%v", name, trace)
	}
}
