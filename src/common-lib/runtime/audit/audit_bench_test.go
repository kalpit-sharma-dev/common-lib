package audit

import (
	"os"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

var (
	_userOne = user{
		Name:   "Jane Doe",
		Email:  "jane@test.com",
		Scores: []int{10, 60, 33, 90, 34},
	}
	_userTwo = user{
		Name:   "Jane Doe",
		Email:  "jane@Newtest.com",
		Scores: []int{10, 33, 55, 34},
	}
)

type users []*user

type user struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Scores []int  `json:"Scores"`
}

func BenchmarkAudit(b *testing.B) {
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, os.ModeAppend)

	ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
	ctx := contextutil.WithValue(httpConttext, ctxDataMap)
	config := Config{AuditName: "bench_audit_events", LoggerConfig: &logger.Config{Destination: logger.STDOUT}}
	a, _ := NewAuditLogger(config)

	b.Run("bench_create_event", func(b *testing.B) {

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				a.CreateEvent(ctx, "Record was created", "", _userOne)
			}
		})
	})

	b.Run("bench_delete_event", func(b *testing.B) {

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				a.DeleteEvent(ctx, "Record was deleted", "", _userOne)
			}
		})
	})

	b.Run("bench_update_event", func(b *testing.B) {

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				a.UpdateEvent(ctx, "Record was updated", "", _userOne, _userTwo)
			}
		})
	})

}
