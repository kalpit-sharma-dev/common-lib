package zookeeper

import (
	"fmt"
	"strings"
	"testing"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

const (
	cbClosedExists          string = "cbClosedExists"
	cbClosedGet             string = "cbClosedGet"
	cbClosedChildren        string = "cbClosedChildren"
	cbClosedCreateRecursive string = "cbClosedCreateRecursive"
	cbClosedSet             string = "cbClosedSet"
	cbClosedDelete          string = "cbClosedDelete"
	cbClosedLock            string = "cbClosedLock"
	cbClosedUnlock          string = "cbClosedUnlock"

	cbOpenExists          string = "cbOpenExists"
	cbOpenGet             string = "cbOpenGet"
	cbOpenChildren        string = "cbOpenChildren"
	cbOpenCreateRecursive string = "cbOpenCreateRecursive"
	cbOpenSet             string = "cbOpenSet"
	cbOpenDelete          string = "cbOpenDelete"
	cbOpenLock            string = "cbOpenLock"
	cbOpenUnlock          string = "cbOpenUnlock"

	cbClosedExistsInvalidCbError          string = "cbClosedExistsInvalidCbError"
	cbClosedGetInvalidCbError             string = "cbClosedGetInvalidCbError"
	cbClosedChildrenInvalidCbError        string = "cbClosedChildrenInvalidCbError"
	cbClosedCreateRecursiveInvalidCbError string = "cbClosedCreateRecursiveInvalidCbError"
	cbClosedSetInvalidCbError             string = "cbClosedSetInvalidCbError"
	cbClosedDeleteInvalidCbError          string = "cbClosedDeleteInvalidCbError"
	cbClosedLockInvalidCbError            string = "cbClosedLockInvalidCbError"
	cbClosedUnlockInvalidCbError          string = "cbClosedUnlockInvalidCbError"

	testPath        = "/testpath"
	testInvalidPath = "testInvalidPath"
	testLock        = "testLock"
	testInvalidLock = "/testInvalidLock"
)

func setupCB(reqVolThreshold int) error {
	cfg := circuit.New()
	cfg.ErrorPercentThreshold = 100
	cfg.MaxConcurrentRequests = 1000
	cfg.RequestVolumeThreshold = reqVolThreshold
	cfg.SleepWindowInSecond = 1
	cfg.TimeoutInSecond = 100
	cfg.Enabled = true

	err := Init([]string{"1.1.1.1:2181"}, "/test")
	if err != nil {
		return fmt.Errorf("failed to init zk err: %v", err)
	}

	err = RegisterCircuitBreaker(cfg)
	if err != nil {
		return fmt.Errorf("failed to register CB with err: %v", err)
	}
	return nil
}

func Test_CB_Closed(t *testing.T) {
	err := setupCB(100)
	if err != nil {
		t.Errorf("Failed to setup CB with error: %v", err)
		return
	}

	tests := []struct {
		name     string
		cbState  string
		expected string
	}{

		{
			name:     cbClosedExists,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedExistsInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedGet,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedGetInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedChildren,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedChildrenInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedCreateRecursive,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedCreateRecursiveInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedSet,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedSetInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedDelete,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedDeleteInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedLock,
			cbState:  circuit.Close,
			expected: zk.ErrNoServer.Error(),
		},
		{
			name:     cbClosedLockInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrInvalidPath.Error(),
		},
		{
			name:     cbClosedUnlock,
			cbState:  circuit.Close,
			expected: zk.ErrNotLocked.Error(),
		},
		{
			name:     cbClosedUnlockInvalidCbError,
			cbState:  circuit.Close,
			expected: zk.ErrNotLocked.Error(),
		},
	}
	for _, test := range tests {
		switch test.name {
		case cbClosedExists:
			_, _, err = Client.Exists(testPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedExistsInvalidCbError:
			_, _, err = Client.Exists(testInvalidPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedGet:
			_, _, err = Client.Get(testPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedGetInvalidCbError:
			_, _, err = Client.Get(testInvalidPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedChildren:
			_, _, err = Client.Children(testPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedChildrenInvalidCbError:
			_, _, err = Client.Children(testInvalidPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedCreateRecursive:
			_, err := Client.CreateRecursive(testPath, nil, 0, nil)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedCreateRecursiveInvalidCbError:
			_, err := Client.CreateRecursive(testInvalidPath, nil, 0, nil)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedSet:
			_, err := Client.Set(testPath, nil, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedSetInvalidCbError:
			_, err := Client.Set(testInvalidPath, nil, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedDelete:
			err := Client.Delete(testPath, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedDeleteInvalidCbError:
			err := Client.Delete(testInvalidPath, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedLock:
			lw := NewLock(testLock)
			err = lw.Lock()
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedLockInvalidCbError:
			lw := NewLock(testInvalidLock)
			err = lw.Lock()
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedUnlock:
			lw := NewLock(testLock)
			err = lw.Unlock()
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbClosedUnlockInvalidCbError:
			lw := NewLock(testInvalidLock)
			err = lw.Unlock()
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		}
	}

}

func Test_CB_Open(t *testing.T) {
	err := setupCB(1)
	if err != nil {
		t.Errorf("Failed to setup CB with error: %v", err)
		return
	}

	tests := []struct {
		name     string
		cbState  string
		expected string
	}{
		{
			name:     cbOpenExists,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},

		{
			name:     cbOpenGet,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenChildren,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenCreateRecursive,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenSet,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenDelete,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenLock,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
		{
			name:     cbOpenUnlock,
			cbState:  circuit.Open,
			expected: circuit.ErrCircuitOpenMessage,
		},
	}

	for _, test := range tests {
		switch test.name {
		case cbOpenExists:
			for i := 0; i < 10; i++ {
				_, _, err = Client.Exists(testPath)
			}
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenGet:
			_, _, err = Client.Get(testPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenChildren:
			_, _, err = Client.Children(testPath)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenCreateRecursive:
			_, err = Client.CreateRecursive(testPath, nil, 0, nil)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenSet:
			_, err = Client.Set(testPath, nil, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenDelete:
			err = Client.Delete(testPath, 0)
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Unexpected error: %v", err)
			}
		case cbOpenLock:
			lw := NewLock(testLock)
			for i := 0; i < 5; i++ {
				err = lw.Lock()
			}
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		case cbOpenUnlock:
			lw := NewLock(testLock)
			err = lw.Unlock()
			if !strings.Contains(err.Error(), test.expected) {
				t.Errorf("Expected: %v Got: %v,", test.expected, err.Error())
			}
		}
	}

}
