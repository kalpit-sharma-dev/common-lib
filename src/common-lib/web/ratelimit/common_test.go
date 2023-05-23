package ratelimit

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/ratelimit/mock"
)

func Test_windowTimestamp(t *testing.T) {
	expected := int64(1645615320)
	got := windowTimestamp(1645615366, 120)
	require.Equal(t, expected, got)
}

func Test_key(t *testing.T) {
	expected := "test:foo:123"
	got := storageKey("test", "foo", 123)
	require.Equal(t, expected, got)
}

func Test_increment(t *testing.T) {
	interval := int64(10)
	duration := time.Second * time.Duration(interval*expireMultiplier)
	store, finish := mockStorage(t)
	defer finish()

	t.Run("Success", func(t *testing.T) {
		key := "test:foo:111"
		expected := int64(2)
		store.EXPECT().Incr(key).Return(expected, nil)
		store.EXPECT().Expire(key, duration).Return(true, nil)

		got, err := increment(store, key, interval)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})
	t.Run("Failed", func(t *testing.T) {
		key := "test:foo:222"
		errExpected := errors.New("some error")
		store.EXPECT().Incr(key).Return(int64(0), errExpected)

		_, err := increment(store, key, interval)
		require.EqualError(t, err, errExpected.Error())
	})
}

func Test_touch(t *testing.T) {
	tt := []struct {
		name   string
		key    string
		res    bool
		resErr error
		errStr string
	}{
		{
			name:   "Success",
			key:    "test:foo:111",
			res:    true,
			resErr: nil,
			errStr: "",
		},
		{
			name:   "Failed: returned error",
			key:    "test:foo:222",
			res:    true,
			resErr: errors.New("some error"),
			errStr: "some error",
		},
		{
			name:   "Failed: result is false",
			key:    "test:foo:333",
			res:    false,
			resErr: nil,
			errStr: "failed to set TTL for key test:foo:333",
		},
	}

	interval := int64(10)
	duration := time.Second * time.Duration(interval*expireMultiplier)
	store, finish := mockStorage(t)
	defer finish()

	for _, tc := range tt {
		storeKey := tc.key
		res := tc.res
		resErr := tc.resErr
		errStr := tc.errStr
		t.Run(tc.name, func(t *testing.T) {
			store.EXPECT().Expire(storeKey, duration).Return(res, resErr)
			err := touch(store, storeKey, interval)
			if errStr != "" {
				require.EqualError(t, err, errStr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_count(t *testing.T) {
	tt := []struct {
		name     string
		key      string
		expected int64
		res      interface{}
		resErr   error
	}{
		{
			name:     "Success",
			key:      "test:foo:111",
			expected: 2,
			res:      2,
			resErr:   nil,
		},
		{
			name:     "Failed: returned error",
			key:      "test:foo:222",
			expected: 0,
			res:      0,
			resErr:   errors.New("some error"),
		},
	}

	store, finish := mockStorage(t)
	defer finish()

	for _, tc := range tt {
		storeKey := tc.key
		expected := tc.expected
		res := tc.res
		resErr := tc.resErr
		t.Run(tc.name, func(t *testing.T) {
			store.EXPECT().Get(storeKey).Return(res, resErr)
			got, err := count(store, storeKey)
			if resErr != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expected, got)
		})
	}
}

func Test_toInt64(t *testing.T) {
	tt := []struct {
		name     string
		value    interface{}
		expected int64
		expErr   bool
	}{
		{
			name:     "Success: int64",
			value:    int64(3),
			expected: 3,
			expErr:   false,
		},
		{
			name:     "Success: int",
			value:    3,
			expected: 3,
			expErr:   false,
		},
		{
			name:     "Success: string",
			value:    "3",
			expected: 3,
			expErr:   false,
		},
		{
			name:     "Success: empty string",
			value:    "",
			expected: 0,
			expErr:   false,
		},
		{
			name:     "Failed: nil",
			value:    nil,
			expected: 0,
			expErr:   true,
		},
		{
			name:     "Failed: slice",
			value:    []int{3},
			expected: 0,
			expErr:   true,
		},
	}
	for _, tc := range tt {
		val := tc.value
		expected := tc.expected
		exErr := tc.expErr
		t.Run(tc.name, func(t *testing.T) {
			got, err := toInt64(val)
			if exErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expected, got)
		})
	}
}

func Test_isNotFoundError(t *testing.T) {
	tt := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "True",
			err:      errors.New("key not found"),
			expected: true,
		},
		{
			name:     "True: redis",
			err:      errors.New("redis: nil"),
			expected: true,
		},
		{
			name:     "False",
			err:      errors.New("unexpected error"),
			expected: false,
		},
	}

	for _, tc := range tt {
		expected := tc.expected
		err := tc.err
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, expected, isNotFoundError(err))
		})
	}
}

func mockStorage(t *testing.T) (*mock.MockStorage, func()) {
	ctrl := gomock.NewController(t)
	return mock.NewMockStorage(ctrl), ctrl.Finish
}
