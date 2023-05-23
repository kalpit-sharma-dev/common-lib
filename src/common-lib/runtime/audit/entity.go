package audit

import "go.uber.org/zap/zapcore"

const schemaVersion string = "1.0.0"

// AuditEvent represent resource information from the event
type AuditEvent struct {
	Type       string      `json:"Type,omitempty"`
	Subtype    string      `json:"Subtype,omitempty"`
	EntityType string      `json:"EntityType,omitempty"`
	EntityID   string      `json:"EntityId,omitempty"`
	Values     interface{} `json:"Values,omitempty"`
	Change     AuditChange `json:"Change,omitempty"`

	// message and description will not be marshalled, only used to be passed to event message and event description
	Message     string
	Description string
}

// MarshalLogObject Marshal Audit to zap Object
func (a AuditEvent) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("SchemaVersion", schemaVersion)

	if a.Type != "" {
		enc.AddString("Type", a.Type)
	}
	if a.Subtype != "" {
		enc.AddString("Subtype", a.Subtype)
	}
	if a.EntityType != "" {
		enc.AddString("EntityType", a.EntityType)
	}
	if a.EntityID != "" {
		enc.AddString("EntityId", a.EntityID)
	}
	if a.Values != nil {
		valOM, implementsOM := interface{}(a.Values).(zapcore.ObjectMarshaler)
		if implementsOM {
			enc.AddObject("Values", valOM)
		} else {
			enc.AddReflected("Values", a.Values)
		}
	}
	enc.AddObject("Change", a.Change)
	return nil
}

// AuditChange represent audit fields changes
type AuditChange struct {
	ID     int64       `json:"ID,omitempty"`
	Path   string      `json:"Path,omitempty"`
	Before interface{} `json:"Before,omitempty"`
	After  interface{} `json:"After,omitempty"`
	Type   FieldType   `json:"Type,omitempty"`
	Format FieldFormat `json:"Format,omitempty"`
}

// MarshalLogObject Marshal AuditChange to zap Object
func (ac AuditChange) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if ac.Path != "" {
		enc.AddInt64("Id", ac.ID)
		enc.AddString("Path", ac.Path)
		enc.AddReflected("Before", ac.Before)
		enc.AddReflected("After", ac.After)
		if ac.Type != "" {
			enc.AddString("Type", string(ac.Type))
		}
		if ac.Format != "" {
			enc.AddString("Format", string(ac.Format))
		}
	}

	return nil
}
