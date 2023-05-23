package main

import (
	"context"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/audit"
)

func main() {
	var httpConttext = context.Background()
	ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
	ctx := contextutil.WithValue(httpConttext, ctxDataMap)
	config := audit.Config{AuditName: "Test_Audit"}
	a, _ := audit.NewAuditLogger(config)

	a.EventS(ctx, "EventType", "Test Audit Message", "Desc")
	// output
	// {"Caller":"example/Main.go:17","Event":{"Date":"2021-04-07T00:02:13.3264031Z","Type":"AUDIT","Message":"Test Audit Message","Description":"Desc","Audit":{"SchemaVersion":"1.0.0","Type":"EventType","Change":{}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}

	event := audit.AuditEvent{
		Type:        "update-ticket",
		Subtype:     "status-update",
		EntityType:  "ticket",
		EntityID:    "123",
		Message:     "Event Message",
		Description: "Event Description",
		Change: audit.AuditChange{
			ID:     1,
			Path:   "/Status",
			Before: "Ticket Created",
			After:  "Ticked Assigned",
			Type:   "string",
		},
	}
	a.Event(ctx, event)
	// output
	// {"Caller":"example/Main.go:36","Event":{"Date":"2021-04-07T00:02:13.327402Z","Type":"AUDIT","Message":"Event Message","Description":"Event Description","Audit":{"SchemaVersion":"1.0.0","Type":"update-ticket","Subtype":"status-update","EntityType":"ticket","EntityId":"123","Change":{"Id":1,"Path":"/Status","Before":"Ticket Created","After":"Ticked Assigned","Type":"string"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}

	person := Person{
		Id:   "123",
		Name: "John Smith",
		Age:  45,
	}
	a.CreateEvent(ctx, "User specified event type", "Person", person)
	// output
	// {"Caller":"example/Main.go:45","Event":{"Date":"2021-04-07T00:02:13.3284038Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-created","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733328,"Path":"/id","Before":"","After":"123","Type":"string"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}
	// {"Caller":"example/Main.go:45","Event":{"Date":"2021-04-07T00:02:13.3294032Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-created","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733328,"Path":"/Name","Before":"","After":"John Smith","Type":"string"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}
	// {"Caller":"example/Main.go:45","Event":{"Date":"2021-04-07T00:02:13.3304052Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-created","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733328,"Path":"/Age","Before":0,"After":45,"Type":"integer"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}

	personUpdated := person
	personUpdated.Age = 47
	a.UpdateEvent(ctx, "User specified event type", "Person", person, personUpdated)
	//output
	// {"Caller":"example/Main.go:53","Event":{"Date":"2021-04-07T00:02:13.331404Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-updated","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733331,"Path":"/Age","Before":45,"After":47,"Type":"integer"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}

	a.DeleteEvent(ctx, "User specified event type", "Person", person)
	// output
	// {"Caller":"example/Main.go:57","Event":{"Date":"2021-04-07T00:02:13.3324043Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-deleted","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733332,"Path":"/id","Before":"123","After":"","Type":"string"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}
	// {"Caller":"example/Main.go:57","Event":{"Date":"2021-04-07T00:02:13.3334028Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-deleted","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733332,"Path":"/Name","Before":"John Smith","After":"","Type":"string"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}
	// {"Caller":"example/Main.go:57","Event":{"Date":"2021-04-07T00:02:13.334403Z","Type":"AUDIT","Audit":{"SchemaVersion":"1.0.0","Type":"User specified event type","Subtype":"field-deleted","EntityType":"Person","EntityId":"123","Change":{"Id":1617753733332,"Path":"/Age","Before":45,"After":0,"Type":"integer"}}},"Host":{"HostName":"LT5CG8411075"},"Service":{"Name":"Main.exe"},"Resource":{"PartnerId":"part123","CorrelationId":"trans123","UserId":"usr123","RequestId":"e0db535f-e118-444d-ad91-1a55801c6e29"}}
}

// Person struct
type Person struct {
	Id   string `json:"id,omitempty"`
	Name string
	Age  int
}
