package zookeeper

import (
	"sync"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/lock"
)

// zkLockErrorToAction - a map of zk error vs the expected action to be taken by the client
var zkLockErrorToAction = map[error]lock.ErrorAction{
	zk.ErrSessionExpired: lock.CreateNewLock,
	zk.ErrDeadlock:       lock.TryUnlock,
	zk.ErrNoNode:         lock.CreateNewLock,
}

// Lock - Wrapper method to encapsulate error cases for zookeeper lock
func (lw lockWrapper) Lock() error {
	var err error
	cbErr := circuit.Do(CBCommandName, lw.cbEnabled, func() error {
		err = lw.zkLock.Lock()
		if validCBError(err) {
			return err
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	le := lock.Error{Code: err}
	switch err {
	case nil:
		return nil
	default:
		le.Action = zkLockErrorToAction[le.Code]
	}

	//if we don't have an action mapped for the error, default to TryLock action
	if le.Action == 0 {
		le.Action = lock.TryLock
	}
	return le
}

// DeleteLockWG can be called to wait for the delete lock node
var DeleteLockWG sync.WaitGroup

// Unlock - Wrapper method to encapsulate error cases for zookeeper unlock
func (lw lockWrapper) Unlock() error {
	var err error
	cbErr := circuit.Do(CBCommandName, lw.cbEnabled, func() error {
		err = lw.zkLock.Unlock()
		if validCBError(err) {
			return err
		}

		// If there is not a valid CB error we still want to try to delete the lock as either the lock was released but with an error or was not released and then try the delete
		// will fail anyway because there will be children to the LockerWraper parent node which is the lock itself
		if lw.name != "" {
			// Ignore any errors happen from deleting the parent node to the lock as we don't want to override the actual unlock error
			// checks can be made and if the parent node still exists by the client and DeleteLock can be called again
			DeleteLockWG.Add(1)
			go func() {
				defer DeleteLockWG.Done()
				delLockErr := DeleteLock(lw.name)
				if delLockErr != nil {
					Logger().Debug(defaultTransaction, "Couldn't delete lock parent node: %s", delLockErr)
				}
			}()
		}
		return nil
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	le := lock.Error{Code: err}
	switch err {
	case nil:
		return nil
	default:
		//we don't know what hit us
		le.Action = lock.TryUnlock
	}

	return le
}
