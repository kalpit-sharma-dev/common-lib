package zookeeper

import (
	"testing"

	"github.com/samuel/go-zookeeper/zk"
)

func TestLock(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "ErrConnectionClosed",
			expectedErr: zk.ErrConnectionClosed,
		},
		{
			name:        "ErrSessionExpired",
			expectedErr: zk.ErrSessionExpired,
		},
		{
			name:        "ErrDeadlock",
			expectedErr: zk.ErrDeadlock,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := &LockMock{}
			lock.When("Lock").Return(test.expectedErr)
			lWrapper := lockWrapper{
				zkLock: lock,
			}
			err := lWrapper.Lock()
			if err.Error() != test.expectedErr.Error() {
				t.Fatalf("expected err: %s, got: %s", test.expectedErr, err)
			}
		})
	}
}

func TestUnlock(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "ErrConnectionClosed",
			expectedErr: zk.ErrConnectionClosed,
		},
		{
			name:        "ErrNoNode",
			expectedErr: zk.ErrNoNode,
		},
		{
			name:        "ErrSessionExpired",
			expectedErr: zk.ErrSessionExpired,
		},
		{
			name:        "ErrNotEmpty",
			expectedErr: zk.ErrNotEmpty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := &LockMock{}
			lock.When("Unlock").Return(test.expectedErr)
			lWrapper := lockWrapper{
				zkLock: lock,
			}
			if err := lWrapper.Unlock(); err.Error() != test.expectedErr.Error() {
				t.Fatalf("expected err: %s, got: %s", test.expectedErr, err)
			}
		})
	}
}

func TestUnlockWithLWName(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "ErrConnectionClosed",
			expectedErr: zk.ErrConnectionClosed,
		},
		{
			name:        "ErrNoNode",
			expectedErr: zk.ErrNoNode,
		},
		{
			name:        "ErrSessionExpired",
			expectedErr: zk.ErrSessionExpired,
		},
		{
			name:        "ErrNotEmpty",
			expectedErr: zk.ErrNotEmpty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Init([]string{"1.1.1.1:2181"}, "/test")
			if err != nil {
				t.Errorf("failed to init zk err: %v", err)
				return
			}
			lock := &LockMock{}
			lock.When("Unlock").Return(test.expectedErr)
			lWrapper := lockWrapper{
				zkLock: lock,
				name:   "lw-name",
			}

			if err := lWrapper.Unlock(); err.Error() != test.expectedErr.Error() {
				t.Fatalf("expected err: %s, got: %s", test.expectedErr, err)
			}
			// wait for the DeleteLock method to finish otherwise we get a panic as el ClientMock will finish and disposed before the Exist or Delete
			// Panic Errors:
			// Mock call missing for Delete(
			// 	"/test/locks/lw-name",
			// 	int32(-1),
			// )
			// or
			// Mock call missing for Exists(
			// 	"/test/locks/lw-name",
			// )
			DeleteLockWG.Wait()
		})
	}
}
