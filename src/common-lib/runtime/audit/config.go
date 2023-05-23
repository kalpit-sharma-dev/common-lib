package audit

import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

// SliceChangesFormat represent the audit event format for slices
type SliceChangesFormat string

const (
	// Full will audit the full slice before and after
	Full SliceChangesFormat = "Full"

	// Unordered will only audit changes to the slice wihtout taking order into consideration
	// Slice changed from [1 2 3] to [1 3] will audit that 2 changed to nil
	Unordered SliceChangesFormat = "Unordered"

	// Ordered will audit any changes to the slice with taking order into consideration
	// Slice changes from [1 2 3] to [1 3] will audit that 2 changed to 3 and 3 changed to nil
	Ordered SliceChangesFormat = "Ordered"
)

const DefaultIdField = "id"

var DefaultGlobalTypeConfig = map[string]FieldConfig{
	"gocql.UUID": {SliceChangesFormat: Full},
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

// clone  to clone all values of config to a new config struct
func (c *Config) Clone() Config {
	newLConf := c.LoggerConfig.Clone()
	newC := Config{
		EntitiesConfig:               c.EntitiesConfig,
		GlobalFieldsConfig:           c.GlobalFieldsConfig,
		GlobalIdField:                c.GlobalIdField,
		GlobalSliceChangesFormat:     c.GlobalSliceChangesFormat,
		GlobalAuditEventForNoChanges: c.GlobalAuditEventForNoChanges,
		GlobalTypeConfig:             c.GlobalTypeConfig,
		AuditName:                    c.AuditName,
		LoggerConfig:                 &newLConf,
	}
	return newC
}
