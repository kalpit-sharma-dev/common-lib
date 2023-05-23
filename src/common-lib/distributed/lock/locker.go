package lock

const (
	// Ignore - error can be ignored and proceed further
	Ignore ErrorAction = iota + 1
	// TryLock - try lock as unlock failed
	TryLock
	// TryUnlock - try unlock as locking failed
	TryUnlock
	// CreateNewLock - create a new lock, forget the last lock
	CreateNewLock
)

type (
	// Locker presents distributed lock/unlock
	Locker interface {
		Lock() error
		Unlock() error
	}

	// ErrorAction - action to be taken on getting a lock/unlock error
	ErrorAction int

	// Error - an error wrapper
	Error struct {
		Code   error
		Action ErrorAction
	}
)

func (err Error) Error() string {
	return err.Code.Error()
}
