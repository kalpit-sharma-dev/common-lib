// +build linux darwin

package namedpipes

import (
	"testing"
	"time"
	"sync"
)

func TestGetPipeServer(t *testing.T) {
	ps := GetPipeServer()
	_, ok := ps.(pipeServer)
	if !ok {
		t.Error("Invalid cast")
	}
	pipeName := "testPipe"
	pipe, err := ps.CreatePipe(pipeName, nil)
	if err != nil {
		t.Error(err)
	}
	defer pipe.Close()
}

func TestGetPipeClient(t *testing.T) {
	var wg sync.WaitGroup
	ps := GetPipeServer()
	pc := GetPipeClient()
	сonfig := PipeConfig{
		MessageMode: false,
	}
	_, ok := pc.(pipeClient)
	if !ok {
		t.Error("Invalid cast")
	}

	pipe, err := ps.CreatePipe(`\\.\pipe\testPipe`, &сonfig)
	defer pipe.Close()
	if err != nil {
		t.Error(err)
	}

	wg.Add(1)
	go func() {
		conn, _ := pipe.Accept()
		wg.Done()
		defer conn.Close()
	}()

	timeout := time.Second
	conn, err := pc.DialPipe(`\\.\pipe\testPipe`, &timeout)
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
	conn.Close()
}
