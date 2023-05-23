package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exec/with"
)

//go run example.go Recover
//go run example.go Context
func main() {
	runType := os.Args[1]
	if runType == "Recover" {
		with.Recover("List Various 2 Powers", "transaction-id", listTwoPowers, recoverFromTwoPowersError)
		return
	} else if runType == "Context" {
		ctx := context.Background()
		with.Context(ctx, "List Various 2 Powers", "transaction-id", listTwoPowersAlternate)
		fmt.Println("Program errored but recovered and the error was logged")
		return
	}

	fmt.Println("Select a valid runType (Recover or Context)")
}

//This is the function that is being called that could potentially panic
func listTwoPowers() {
	currentNumber, power := 2, 1
	for ; ; power++ {
		time.Sleep(time.Second / 10)
		fmt.Println("2 to the", power, "power is", currentNumber)
		currentNumber *= 2
		//Note this is bad design to showcase how this work, do not intentionally panic for something like this
		if currentNumber <= 0 {
			panic("currentNumber went out of bounds")
		}
	}
}

//This is the recovery function that gets called if the above function panics
func recoverFromTwoPowersError(transaction string, err error) {
	fmt.Printf("Error occured during getting powers of 2 in transaction id: %s \nError: %v\n\nGracefully handled the error", transaction, err)
}

//This is the function that is being called by Context that could potentially panic or return an error
func listTwoPowersAlternate() error {
	currentNumber, power := 2, 1
	for ; ; power++ {
		time.Sleep(time.Second / 10)
		fmt.Println("2 to the", power, "power is", currentNumber)
		currentNumber *= 2
		//Note this is bad design to showcase how this work, do not intentionally errored for something like this
		if currentNumber <= 0 {
			return errors.New("currentNumber went out of bounds")
		}
	}
}
