package logger_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

var (
	errExample = errors.New("fail")

	_messages   = fakeMessages(1000)
	_tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	_tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	_tenTimes   = []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
		time.Unix(3, 0),
		time.Unix(4, 0),
		time.Unix(5, 0),
		time.Unix(6, 0),
		time.Unix(7, 0),
		time.Unix(8, 0),
		time.Unix(9, 0),
	}
	_oneUser = &user{
		Name:      "Jane Doe",
		Email:     "jane@test.com",
		CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	_tenUsers = users{
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
	}
)

func fakeMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("Test logging, but use a somewhat realistic message length. (#%v)", i)
	}
	return messages
}

func getMessage(iter int) string {
	return _messages[iter%1000]
}

type users []*user

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func fakeFmtArgs() []interface{} {
	// Need to keep this a function instead of a package-global var so that we
	// pay the cast-to-interface{} penalty on each call.
	return []interface{}{
		_tenInts[0],
		_tenInts,
		_tenStrings[0],
		_tenStrings,
		_tenTimes[0],
		_tenTimes,
		_oneUser,
		_oneUser,
		_tenUsers,
		errExample,
	}
}

func BenchmarkLogger(b *testing.B) {
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, os.ModeAppend)

	b.Run("big_text_format", func(b *testing.B) {
		l, _ := logger.Update(logger.Config{Name: "Logger-1", MaxSize: 1, Destination: logger.STDOUT, LogFormat: logger.JSONFormat})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("transaction_id", "%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})

	b.Run("samll_text_format", func(b *testing.B) {
		l, _ := logger.Update(logger.Config{Name: "Logger-1", MaxSize: 1, Destination: logger.STDOUT, LogFormat: logger.JSONFormat})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("transaction_id", getMessage(0))
			}
		})
	})

	b.Run("big_JSON_format", func(b *testing.B) {
		l, _ := logger.Update(logger.Config{Name: "Logger-1", MaxSize: 1, Destination: logger.STDOUT, LogFormat: logger.JSONFormat})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("transaction_id", "%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})

	b.Run("samll_JSON_format", func(b *testing.B) {
		l, _ := logger.Update(logger.Config{Name: "Logger-1", MaxSize: 1, Destination: logger.STDOUT, LogFormat: logger.JSONFormat})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("transaction_id", getMessage(0))
			}
		})
	})
}
