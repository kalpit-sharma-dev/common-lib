package zookeeper

import (
	"fmt"
	"testing"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	zMock, originalClient := InitMock()
	defer Restore(originalClient)
	zMock.When("CreateRecursive", "/foo", nil, int32(0), noRestrictionsACL).Return("", zk.ErrNodeExists)
	counter, err := NewCounter("foo")
	require.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		zMock.When("Create", "/foo/counter", []uint8(nil), int32(zk.FlagSequence|zk.FlagEphemeral), noRestrictionsACL).Return("/foo/counter0000000000", nil).Times(1)
		zMock.When("Delete", "/foo/counter0000000000", int32(0)).Return(nil).Times(1)
		val, err := counter.Increment()
		assert.Equal(t, int32(0), val)
		assert.NoError(t, err)
	})
	t.Run("success_with_MaxCounterValue", func(t *testing.T) {
		maxPath := fmt.Sprintf("/foo/counter%v", MaxCounterValue)
		zMock.When("Create", "/foo/counter", nil, int32(zk.FlagSequence|zk.FlagEphemeral), noRestrictionsACL).Return(maxPath, nil).Times(1)
		zMock.When("Delete", maxPath, int32(0)).Return(zk.ErrConnectionClosed).Times(1)
		val, err := counter.Increment()
		assert.Equal(t, int32(MaxCounterValue), val)
		assert.NoError(t, err)
	})
	t.Run("fatal_when_exceeding_MaxCounterValue", func(t *testing.T) {
		maxPath := fmt.Sprintf("/foo/counter-%v", MaxCounterValue)
		zMock.When("Create", "/foo/counter", nil, int32(zk.FlagSequence|zk.FlagEphemeral), noRestrictionsACL).Return(maxPath, nil).Times(1)
		zMock.When("Delete", maxPath, int32(0)).Return(zk.ErrConnectionClosed).Times(1)
		_, err := counter.Increment()
		assert.Equal(t, ErrCounterOverflow, err)
	})
	t.Run("success_despite_error_deleting-see_comments_explaining_why", func(t *testing.T) {
		// Do not return the error and just log it instead because:
		// 1. It's fine that there was an error in deletion
		// 2. If what the app needs is to see every single value (no skipping numbers allowed) then returning the error would prevent the app from seeing all values, because it would skip the value that was just generated
		// 3. If every delete failed, that would actually cause some space to be taken up, but that's an extremely unlikely case given that the create just succeeded. Also, the znode created has zero data.
		zMock.When("Create", "/foo/counter", nil, int32(zk.FlagSequence|zk.FlagEphemeral), noRestrictionsACL).Return("/foo/counter0000000001", nil).Times(1)
		zMock.When("Delete", "/foo/counter0000000001", int32(0)).Return(zk.ErrConnectionClosed).Times(1)
		val, err := counter.Increment()
		assert.Equal(t, int32(1), val)
		assert.NoError(t, err)
	})
	t.Run("error_creating", func(t *testing.T) {
		// Do not return the error and just log it instead because:
		// 1. It's fine that there was an error in deletion
		// 2. If what the app needs is to see every single value (no skipping numbers allowed) then returning the error would prevent the app from seeing all values, because it would skip the value that was just generated
		// 3. If every delete failed, that would actually cause some space to be taken up, but that's an extremely unlikely case given that the create just succeeded. Also, the znode created has zero data.
		zMock.When("Create", "/foo/counter", nil, int32(zk.FlagSequence|zk.FlagEphemeral), noRestrictionsACL).Return("", zk.ErrConnectionClosed).Times(1)
		_, err := counter.Increment()
		assert.Error(t, err)
	})
}
