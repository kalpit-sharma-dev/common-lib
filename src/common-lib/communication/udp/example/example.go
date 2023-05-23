package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
)

func main() {
	var err error
	if os.Args[1] == "C" {
		err = setupClient()
	} else {
		err = setupServer()
		fmt.Println("stopped UDP Server")
	}
	if err != nil {
		fmt.Println(err)
	}
}

func setupClient() error {
	cfg := udp.New()

	fmt.Println("udp.Send :: Test Message Sent Over UDP")
	err := udp.Send(cfg, []byte("Test Message Sent Over UDP"), nil)
	if err != nil {
		return err
	}

	fmt.Println("udp.Send :: Test Message Over UDP - responseHandler")
	err = udp.Send(cfg, []byte("Test Message Over UDP - responseHandler"), responseHandler)
	if err != nil {
		return err
	}
	return nil
}

var responseHandler = func(response udp.Response) error {
	fmt.Println("Address : ", response.Address)
	fmt.Println("Size : ", response.Size)
	fmt.Println("Message : ", string(response.Body))
	fmt.Println("Error : ", response.Err)
	return nil
}

func setupServer() error {
	cfg := udp.New()
	server := udp.NewServer(cfg)
	go func() {
		time.Sleep(30 * time.Second)
		fmt.Println("stopping UDP Server")
		server.Shutdown(context.Background())
	}()
	return server.Receive(requestHandler)
}

var requestHandler = func(request udp.Request) []byte {
	fmt.Println("Address : ", request.Address)
	fmt.Println("Size : ", request.Size)
	fmt.Println("Message : ", string(request.Body))
	fmt.Println("Error : ", request.Err)
	return []byte("Test Message Received Over UDP")
}
