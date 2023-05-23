package main

import (
	"github.com/maraino/go-mock"
	"github.com/samuel/go-zookeeper/zk"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/zookeeper"
)

// main contain examples for zookeeper package
func main() {
	zkMockObj, original := zookeeper.InitMock()
	defer zookeeper.Restore(original)

	l := &zookeeper.LockMock{}
	l.When("Lock").Return(nil)
	l.When("Unlock").Return(nil)
	zkMockObj.When("NewLock", mock.Any, mock.Any).Return(l)
	zkMockObj.When("Children", mock.Any).Return([]string{}, &zk.Stat{}, nil)
	zkMockObj.When("isCBEnabled").Return(false)

	// Create new zookeeper lock
	lock := zookeeper.NewLock("some_name_test")
	if err := lock.Lock(); err != nil {
		handleError(err)
	}

	// Unlock with defer
	defer func() {
		err := lock.Unlock()
		if err != nil {
			handleError(err)
		}
		// unlock should delete the parent node created "some_name_test"
		// wait to all DeleteLock goroutines to finish
		zookeeper.DeleteLockWG.Wait()
		// checks can be made and if the parent node still exists for any reason zookeeper.DeleteLock("some_name_test") can be called
		pnExists, _, err := zookeeper.Client.Exists("/zookeeper_root/locks/some_name_test")
		handleError(err)
		if pnExists {
			err = zookeeper.DeleteLock("some_name_test")
			handleError(err)
		}
	}()

	// Init broadcast
	broadCast, err := zookeeper.InitBroadcast("example", 0)
	if err != nil {
		handleError(err)
	}

	// Add handler
	broadCast.AddHandler("example_handler", func(*distributed.Event) {})

	// Create Event
	if err := broadCast.CreateEvent(distributed.Event{Type: "example_name", Payload: "example_payload"}); err != nil {
		handleError(err)
	}

}

func handleError(_ error) {
	// actions to handle error
}
