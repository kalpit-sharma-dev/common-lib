package model

import (
	"fmt"

	db "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm"
)

func init() {
	Cats().base.Register(&CatObserver{})
}

// CatObserver implementation of repo.Observer for alert mappings
type CatObserver struct{}

// OnNotify implements Observer
func (CatObserver) OnNotify(eventType db.EventType, args ...interface{}) {
	var op string
	switch eventType {
	case db.EventBeforeAdd, db.EventBeforeUpdate, db.EventBeforeDelete:
		fmt.Printf("*** notify: skip before operation: %v\n", eventType)
		return
	case db.EventAfterAdd:
		op = "added"
	case db.EventAfterUpdate:
		op = "updated"
	case db.EventAfterDelete:
		op = "deleted"
	default:
		fmt.Printf("*** notify: unsupported operation: %v\n", eventType)
		return
	}

	cat, ok := args[0].(*Cat)
	if !ok {
		fmt.Println("*** notify: add/update/delete called with wrong param?")
		return
	}
	fmt.Printf("*** notify: Cat with ID: %q and name %q was %s\n", cat.ID, cat.Name, op)
}
