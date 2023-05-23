package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_slidingWindow(t *testing.T) {
	store, finish := mockStorage(t)
	defer finish()

	currentKey, prevKey := mockKeys()

	t.Run("Default", func(t *testing.T) {
		p := mockParams()

		store.EXPECT().Get(currentKey).Return(int64(3), nil)
		store.EXPECT().Incr(currentKey).Return(int64(4), nil)
		store.EXPECT().Expire(currentKey, time.Second*time.Duration(p.Interval*expireMultiplier)).Return(true, nil)
		store.EXPECT().Get(prevKey).Return("4", nil)

		expected := int64(5)
		got, err := slidingWindow(store, p)
		require.NoError(t, err)
		require.Equal(t, expected, got)

	})

	t.Run("Limit reached", func(t *testing.T) {
		p := mockParams()
		p.Limit = 2
		expected := int64(3)

		store.EXPECT().Get(currentKey).Return(expected, nil)

		got, err := slidingWindow(store, p)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})
}

func mockParams() countParams {
	return countParams{
		Now:      1646042799,
		Group:    mockGroup,
		Key:      mockKey,
		Interval: 4,
		Limit:    10,
	}
}

func mockKeys() (current, prev string) {
	p := mockParams()
	current = storageKey(p.Group, p.Key, windowTimestamp(p.Now, p.Interval))
	prev = storageKey(p.Group, p.Key, windowTimestamp(p.Now-p.Interval, p.Interval))
	return
}
