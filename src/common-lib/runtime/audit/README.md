<p align="center">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Audit

This is a Standard audit implementation. Currently it is only implemented for use with [logger](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/-/tree/master/src/runtime/logger), though there could be more implementations in the future. You can view the entire SDD for discussion about the system [here](https://confluence.kksharmadevdev.com/x/kAtLBw).
### [Example](example/example.go)

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/audit"
```

**Audit Logger Instance**

```go
// Create Audit Logger instance
a, err := audit.NewAuditLogger(audit.Config{AuditName: name})
```

**Writing Logs**

```go
a.EventS(ctx,"Event Type", "Message", "Description")
a.Event(ctx, AuditEvent{
    Type:       "Event Type",
    Subtype:    "Event Subtype",
    EntityType: "Event Entity Type",
    EntityID:   "Event Entity ID",
    Message: "Message",
    Description: "Description",
})

// audit struct that was created
// audit will create  seprate audit event for each field to the user struct with event type created
a.CreateEvent(ctx, "User record was created", user)

// audit struct that was updated
// audit will create  seprate audit event for each field that was changed with event type updated
a.UpdateEvent(ctx, "User record was updated", userOld, userNew)

// audit struct that was deleted
// audit will create  seprate audit event for each field to the user struct with event type deleted
a.DeleteEvent(ctx, "User record was deleted", user)

// adding values to Create, Update and Delete Events
v := map[string]string{"ID": "123", "Name": "ABC"}
a.With(audit.AddValues(v)).CreateEvent(ctx, "User record was created", user)
a.With(audit.AddValues(v)).UpdateEvent(ctx, "User record was updated", userOld, userNew)
a.With(audit.AddValues(v)).DeleteEvent(ctx, "User record was deleted", user)

```

**Configuration**

```go
// Config is a struct to holds audit configuration
// All config properties can't be changed after the audit is intialized
type Config struct {
	// Fields Configurations Per Entity
	// Key is entity name ex: Agent, Gateway, PartnerUser, etc.
	EntitiesConfig map[string]EntityConfig

	// GlobalFieldsConfig is a map that has the configurations for specific fields where the key is field path and value is FieldConfig
	// Each map key field should be in the format of JSON pointer field path ex: /Field1/Field2
	GlobalFieldsConfig map[string]FieldConfig

	// Specify the JSON pointer field path to the EntityID in the Entity from Create, Update and Delete Events. Defaults to "id"
	GlobalIdField string

	// GlobalSliceChangesFormat indicate how to capture slices changes to Full, Unordered or Ordered globally
	// Default value Unordered
	// SliceChangesFormatMap overrides the global setting
	GlobalSliceChangesFormat SliceChangesFormat

	// AuditEventForNoChanges indicates for CreateEvent, DeletedEvent and UpdateEvent whether it should still create an Audit Event when there are no changes between the before and after objects
	// true create an Audit Event with empty Change struct
	// false no creation of Audit Event at all
	// Default is false
	GlobalAuditEventForNoChanges bool

	// GlobalTypeConfig is a map that has the configurations for specific fields types where the key is type and value is FieldConfig
	// Each map key field should be the package.Type except for the built in types like string and int should be only the type ex: string, []byte, gocql.UUID
	GlobalTypeConfig map[string]FieldConfig

	// AuditName is a name for the audit and shold be unique per audit instance
	AuditName string

	// loggerConfig holds logger configurations
	LoggerConfig *logger.Config
}

// EntityConfig is a struct holds fields configurations per Entity
type EntityConfig struct {
	// FieldsConfig is a map that has the configurations for specific fields where the key is field path and value is FieldConfig
	// Each map key field should be in the format of JSON pointer field path ex: /Field1/Field2
	// Overrides GlobalFieldConfig per Entity
	FieldsConfig map[string]FieldConfig

	// Specify the JSON pointer field path to the EntityID in the Entity from Create, Update and Delete Events. Defaults to "id"
	// Overrides GlobalIdField per Entity
	IdField string

	// SliceChangesFormat slice change format Full, Ordered and Unordered
	// Overrides the GlobalSliceChangesFormat setting
	SliceChangesFormat SliceChangesFormat

	// AuditEventForNoChanges indicates for CreateEvent, DeletedEvent and UpdateEvent whether it should still create an Audit Event when there are no changes between the before and after objects
	// true create an Audit Event with empty Change struct
	// false no creation of Audit Event at all
	// Default is false
	// Overrides GlobalAuditEventForNoChanges
	AuditEventForNoChanges bool
}

// FieldConfig is a struct holds fields configurations per Field
type FieldConfig struct {
	// Ignore field while auditing if it was changed
	Ignore bool

	// SliceChangesFormat slice change format Full, Ordered and Unordered
	// Overrides the GlobalSliceChangesFormat AND EntityConfig SliceChangesFormat
	SliceChangesFormat SliceChangesFormat

	// TypeFormat specify field type while auditing
	// If not provided the library will detect the value based on the value changed
	TypeFormat TypeFormat
}
```

## Example

```json
"Audit": {
	"EntitiesConfig":{"Person": {"IgnoreFields": ["/Field1","/Field2/Field3"] ,
		"FieldsTypes": {"/Field4":{"fieldType":"string","fieldFormat":"email"}},
		"SliceChangesFormat": "Ordered"}
	},
	"AuditName" : "<processname>",
	"LoggerConfig": {
        "Destination": "FILE",
		"MaxSize": 20,
		"MaxAge": 30,
		"MaxBackups": 5
    }
},
```

## Output

Please note changes path value will be the JSON pointer tag and not the actual field name. If the tag doesn't exist then it will be the field name.

If there is a conflict between one field JSON tag and antoher field name then only changes for the field with tag will be shown and the other field changes will be ignored. For example the below struct TestEntityWithConflictFields the JSON tag for TestWithTagName is "Name" and there is another field with the name "Name", in this case only changes for TestWithTagName will be shown with path "/Name" and changes fot the field "Name" will be ignored

```go
type TestEntityWithConflictFields struct {
	TestWithTagName string `json:"Name,omitempty"`
	Name            string
}
```

## Example of json output

```json
{
	"Caller": "service/service.go:1434",
	"Event": {
		"Date": "2021-03-05T15:31:10.8074821Z",
		"Type": "AUDIT",
		"Audit": {
			"Type": "update-ticket",
			"Subtype": "status-update",
			"EntityType": "ticket",
			"EntityId": "123",
			"Change": {
				"Id": 1,
				"Path": "/Status",
				"Before": "Ticket Created",
				"After": "Ticked Assigned",
				"Type": "string"
			}
		}
	},
	"Host": {
		"HostName": "LT5CG8411075"
	},
	"Service": {
		"Name": "audit.test.exe"
	},
	"Resource": {
		"PartnerId": "part123",
		"CorrelationId": "trans123",
		"UserId": "usr123",
		"RequestId": "f48f258e-b9b7-4b49-8f19-727f5089d798"
	}
}
```

## Benchmark results for audit events create, update and delete

goos: windows
goarch: amd64
pkg: gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/audit
BenchmarkAudit/bench_create_event-8         	   32113	     33349 ns/op	   18408 B/op	     282 allocs/op
BenchmarkAudit/bench_delete_event-8         	   35068	     31569 ns/op	   18442 B/op	     282 allocs/op
BenchmarkAudit/bench_update_event-8         	   43905	     28676 ns/op	   13187 B/op	     336 allocs/op