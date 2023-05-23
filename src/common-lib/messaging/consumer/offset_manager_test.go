package consumer

import "testing"

func Test_getCommitStrategy(t *testing.T) {
	t.Run("OnPull", func(t *testing.T) {
		got := getCommitStrategy(&Config{CommitMode: OnPull}, nil)
		_, ok := got.(*onPull)
		if !ok {
			t.Errorf("getCommitStrategy() = %v, want %v", got, &onPull{})
		}
	})

	t.Run("onMessageCompletion", func(t *testing.T) {
		got := getCommitStrategy(&Config{CommitMode: OnMessageCompletion}, nil)
		_, ok := got.(*onMessageCompletion)
		if !ok {
			t.Errorf("getCommitStrategy() = %v, want %v", got, &onMessageCompletion{})
		}
	})

	t.Run("Unknown", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		got := getCommitStrategy(&Config{CommitMode: 0}, nil)
		t.Errorf("getCommitStrategy() = %v, want %v", got, "Panic")
	})

	t.Run("Unknown", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		got := getCommitStrategy(&Config{CommitMode: -3}, nil)
		t.Errorf("getCommitStrategy() = %v, want %v", got, "Panic")
	})
}
