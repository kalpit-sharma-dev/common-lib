package zookeeper

import (
	"errors"
	"math"
	"strconv"

	"github.com/samuel/go-zookeeper/zk"
)

const MaxCounterValue = math.MaxInt32

var counterNumLen = int(math.Ceil(math.Log10(float64(MaxCounterValue))))

var ErrCounterOverflow = errors.New("overflow from Counter.Increment")

var noRestrictionsACL = zk.WorldACL(zk.PermAll)

// Counter allows you to have a permanent, distributed incrementable integer, that counts up to MaxCounterValue
type Counter struct {
	Namespace string
}

// NewCounter constructs a Counter, and initializes it.
// You must call NewCounter on a namespace at least once before being able to call Increment on a Counter for that namespace.
// Running NewCounter on the same namespace more than once doesn't reset the counter; this operation is idempotent.
func NewCounter(namespace string) (*Counter, error) {
	if namespace[0] != '/' {
		namespace = "/" + namespace
	}
	_, err := Client.CreateRecursive(namespace, nil, 0, noRestrictionsACL)
	// If it already exists, then that's not an issue - Init should be idempotent
	if err == zk.ErrNodeExists {
		err = nil
	}
	return &Counter{namespace}, err
}

// Increment returns the next number in the count, and return it.
// It starts at 0 and goes up to and including MaxCounterValue.
// After MaxCounterValue, a fatal log will occur, and an error will be returned.
func (c *Counter) Increment() (int32, error) {
	// Create a sequence znode so that we get a number from it, then delete the znode to free up space.
	// Deleting won't affect the count upwards, even if you restart zookeeper.
	path, err := Client.Create(c.Namespace+"/counter", nil, zk.FlagSequence|zk.FlagEphemeral, noRestrictionsACL)
	if err != nil {
		return 0, err
	}
	// Now delete the node as a clean up.
	// Technically, this isn't a proper transaction, and this delete can fail while the create may succeed.
	// The test that ignores this error explains why
	err = Client.Delete(path, 0)
	if err != nil {
		Logger().Warn("", "Unable to delete a zookeeper znode (%v) - no manual intervention is necessary. Error received: %v", path, err.Error())
	}

	// If there was an overflow, like {namespace}/counter-2147483647, then give an error
	if path[len(path)-counterNumLen-1] == '-' {
		Logger().Fatal("", "CounterIncrementOverflow", "Counter.Increment overflowed beyond MaxCounterValue")
		return 0, ErrCounterOverflow
	}

	seqID, err := strconv.Atoi(path[len(path)-counterNumLen:])
	return int32(seqID), err
}
