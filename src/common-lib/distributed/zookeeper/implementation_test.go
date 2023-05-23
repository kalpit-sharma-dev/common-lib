package zookeeper

import (
	"errors"
	"fmt"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

var testHosts = []string{"test"}

func TestInit(t *testing.T) {
	_, originalClient := InitMock()
	defer Restore(originalClient)

	restoreFn := MockConnect(nil)
	defer restoreFn()

	t.Run("Success", func(t *testing.T) {
		err := Init(testHosts, "test")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Incorrect base path", func(t *testing.T) {
		basePath := ""
		expectedErr := fmt.Errorf("incorrect base path: %s", basePath)

		err := Init(testHosts, basePath)
		if err.Error() != expectedErr.Error() {
			t.Fatalf("expected err: %s, got: %s", expectedErr, err)
		}
	})

	t.Run("Connect error", func(t *testing.T) {
		expectedErr := errors.New("error")
		MockConnect(expectedErr)

		err := Init(testHosts, "test")
		if err.Error() != expectedErr.Error() {
			t.Fatalf("expected err: %s, got: %s", expectedErr, err)
		}
	})
}

func TestInitWithLogger(t *testing.T) {
	_, originalClient := InitMock()
	defer Restore(originalClient)

	restoreFn := MockConnect(nil)
	defer restoreFn()

	logImpl, err := logger.Create(logger.Config{Name: "test", Destination: logger.DISCARD})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		err = InitWithLogger(testHosts, "test", logImpl)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Incorrect base path", func(t *testing.T) {
		basePath := ""
		expectedErr := fmt.Errorf("incorrect base path: %s", basePath)

		err = InitWithLogger(testHosts, basePath, logImpl)
		if err.Error() != expectedErr.Error() {
			t.Fatalf("expected err: %s, got: %s", expectedErr, err)
		}
	})

	t.Run("Nil logger", func(t *testing.T) {
		err = InitWithLogger(testHosts, "test", nil)
		if err != nil {
			t.Fatalf("expected err: nil, got: %s", err)
		}
	})

	t.Run("Connect error", func(t *testing.T) {
		expectedErr := errors.New("error")
		MockConnect(expectedErr)

		err = InitWithLogger(testHosts, "test", logImpl)
		if err.Error() != expectedErr.Error() {
			t.Fatalf("expected err: %s, got: %s", expectedErr, err)
		}
	})
}

func TestStatus(t *testing.T) {
	t.Run("Status_Available", func(t *testing.T) {
		cs := ConnectionStatus{
			Path:  "/mock-service-path",
			Hosts: []string{"localhost:2181"},
		}

		zkMockObj, originalClient := InitMock()
		defer Restore(originalClient)

		zkMockObj.When("State").Return(stateConnected)

		conn := rest.OutboundConnectionStatus{}
		con := cs.Status(conn)

		if con.ConnectionStatus != rest.ConnectionStatusActive {
			t.Fatalf("expected status to be %s, but got %s", rest.ConnectionStatusActive, con.ConnectionStatus)
		}
	})

	t.Run("Status_Not_Available", func(t *testing.T) {
		cs := ConnectionStatus{
			Path:  "/mock-service-path",
			Hosts: []string{"localhost:2181"},
		}

		zkMockObj, originalClient := InitMock()
		defer Restore(originalClient)

		zkMockObj.When("State").Return("")

		conn := rest.OutboundConnectionStatus{}
		con := cs.Status(conn)

		if con.ConnectionStatus != rest.ConnectionStatusUnavailable {
			t.Fatalf("expected status to be %s, but got %s", rest.ConnectionStatusUnavailable, con.ConnectionStatus)
		}
	})
}
