package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"

	"github.com/google/uuid"
)

const (
	createEvent   string = "field-created"
	deleteEvent   string = "field-deleted"
	updateEvent   string = "field-updated"
	auditLogLevel string = "AUDIT"
)

type AuditLogger struct {
	logger logger.Log
	config *Config

	values interface{}
}

type TypeFormat struct {
	fieldType   FieldType
	fieldFormat FieldFormat
}

type Option interface {
	apply(*AuditLogger)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*AuditLogger)

func (f optionFunc) apply(a *AuditLogger) {
	f(a)
}

// AddValues add values to Audit.Values
func AddValues(values interface{}) Option {
	return optionFunc(func(a *AuditLogger) {
		a.values = values
	})
}

// CallDepth sets the call depth to the current Logger RelativeCallDepth
func CallDepth(callDepth int) Option {
	return optionFunc(func(a *AuditLogger) {
		a.logger = a.logger.With(logger.CallDepth(callDepth))
	})
}

func (a *AuditLogger) clone() AuditLogger {
	newConfig := a.config.Clone()
	copy := AuditLogger{
		logger: a.logger,
		config: &newConfig,
		values: a.values,
	}
	return copy
}

func (a *AuditLogger) With(options ...Option) *AuditLogger {
	c := a.clone()
	for _, opt := range options {
		opt.apply(&c)
	}
	return &c
}

type FieldType string
type FieldFormat string

const (
	String  FieldType = "string"
	Number  FieldType = "number"
	Integer FieldType = "integer"
	Boolean FieldType = "boolean"
	Array   FieldType = "array"
	Object  FieldType = "object"
)

const (
	Date     FieldFormat = "date"
	DateTime FieldFormat = "date-time"
	Password FieldFormat = "password"
	Byte     FieldFormat = "byte"
	Binary   FieldFormat = "binary"
	Email    FieldFormat = "email"
	UUID     FieldFormat = "uuid"
	URI      FieldFormat = "uri"
	Hostname FieldFormat = "hostname"
	IPV4     FieldFormat = "ipv4"
	IPV6     FieldFormat = "ipv6"

	Float  FieldFormat = "float"
	Double FieldFormat = "double"
	Int32  FieldFormat = "int32"
	Int64  FieldFormat = "int64"
)

// audit wraps the logger, ensure the caller of the audit functions is logged as the logger source.
const loggerCallDepth int = 1

// NewAudit returns new audit object
func NewAuditLogger(config Config) (*AuditLogger, error) {
	if config.AuditName == "" {
		return nil, fmt.Errorf("AuditNameIsRequired")
	}

	if config.LoggerConfig == nil {
		config.LoggerConfig = &logger.Config{Name: config.AuditName}
	} else {
		config.LoggerConfig.Name = config.AuditName
	}

	if config.LoggerConfig.Destination == (logger.Destination{}) {
		config.LoggerConfig.Destination = logger.STDOUT
	}

	if config.LoggerConfig.ServiceName == "" {
		config.LoggerConfig.ServiceName = config.AuditName
	}

	config.LoggerConfig.LogLevel = logger.INFO
	config.LoggerConfig.LogFormat = logger.JSONFormat
	config.LoggerConfig.RelativeCallDepth += loggerCallDepth

	log, err := logger.Create(*config.LoggerConfig)
	if err != nil {
		return nil, err
	}

	if config.GlobalSliceChangesFormat == "" {
		config.GlobalSliceChangesFormat = Unordered
	}

	// populate DefaultGlobalTypeConfig to the configuation config.GlobalTypeConfig if items don't exist
	if config.GlobalTypeConfig == nil {
		config.GlobalTypeConfig = make(map[string]FieldConfig)
	}
	for k, v := range DefaultGlobalTypeConfig {
		if _, ok := config.GlobalTypeConfig[k]; !ok {
			config.GlobalTypeConfig[k] = v
		}
	}

	audit := &AuditLogger{
		logger: log,
		config: &config,
	}

	return audit, nil
}

func (a *AuditLogger) EventS(ctx context.Context, eventType string, message string, description string) (string, error) {
	if eventType == "" {
		return "", fmt.Errorf("Event type is required")
	}

	auditEvent := AuditEvent{
		Type:   eventType,
		Values: a.values,
	}

	event := logger.Event{
		Message:     message,
		Description: description,
		Audit:       auditEvent,
		LogLevel:    auditLogLevel,
		ID:          eventID(),
	}
	a.logger.LogEvent(ctx, event)

	return event.ID, nil
}

func (a *AuditLogger) Event(ctx context.Context, auditEvent AuditEvent) (string, error) {
	if auditEvent.Type == "" {
		return "", fmt.Errorf("Event type is required")
	}

	event := logger.Event{
		Message:     auditEvent.Message,
		Description: auditEvent.Description,
		Audit:       auditEvent,
		LogLevel:    auditLogLevel,
		ID:          eventID(),
	}
	a.logger.LogEvent(ctx, event)

	return event.ID, nil
}

func (a *AuditLogger) UpdateEvent(ctx context.Context, eventType string, entityType string, before interface{}, after interface{}) ([]string, error) {
	if eventType == "" {
		return nil, fmt.Errorf("Event type is required")
	}

	entityConfig := getEntityConfig(a.config, entityType, before)
	auditChanges := createAuditChanges(before, after, entityConfig.SliceChangesFormat, entityConfig.FieldsConfig, false)
	events := createEvents(auditChanges, eventType, updateEvent, entityType, entityConfig.AuditEventForNoChanges, entityConfig.IdField, before, a.values)
	var eventIDs []string
	for _, e := range events {
		eventIDs = append(eventIDs, e.ID)
		a.logger.LogEvent(ctx, e)
	}

	return eventIDs, nil
}

func (a *AuditLogger) CreateEvent(ctx context.Context, eventType string, entityType string, entity interface{}) ([]string, error) {
	if eventType == "" {
		return nil, fmt.Errorf("Event type is required")
	}

	entityConfig := getEntityConfig(a.config, entityType, entity)
	auditChanges := createAuditChanges(nil, entity, entityConfig.SliceChangesFormat, entityConfig.FieldsConfig, true)
	events := createEvents(auditChanges, eventType, createEvent, entityType, entityConfig.AuditEventForNoChanges, entityConfig.IdField, entity, a.values)
	var eventIDs []string
	for _, e := range events {
		eventIDs = append(eventIDs, e.ID)
		a.logger.LogEvent(ctx, e)
	}

	return eventIDs, nil
}

func (a *AuditLogger) DeleteEvent(ctx context.Context, eventType string, entityType string, entity interface{}) ([]string, error) {
	if eventType == "" {
		return nil, fmt.Errorf("Event type is required")
	}

	entityConfig := getEntityConfig(a.config, entityType, entity)
	auditChanges := createAuditChanges(entity, nil, entityConfig.SliceChangesFormat, entityConfig.FieldsConfig, true)
	events := createEvents(auditChanges, eventType, deleteEvent, entityType, entityConfig.AuditEventForNoChanges, entityConfig.IdField, entity, a.values)
	var eventIDs []string
	for _, e := range events {
		eventIDs = append(eventIDs, e.ID)
		a.logger.LogEvent(ctx, e)
	}

	return eventIDs, nil
}

func getEntityConfig(config *Config, entityType string, entity interface{}) EntityConfig {
	// set entity config default values with the global config values
	entityConfig := EntityConfig{
		FieldsConfig:           config.GlobalFieldsConfig,
		IdField:                config.GlobalIdField,
		SliceChangesFormat:     config.GlobalSliceChangesFormat,
		AuditEventForNoChanges: config.GlobalAuditEventForNoChanges,
	}

	// override with GlobalTypeConfig
	fieldsTypesMap := utils.GetStructFieldsTypes(entity)

	if len(config.GlobalTypeConfig) > 0 {
		for k, v := range fieldsTypesMap {
			if fieldConfig, ok := config.GlobalTypeConfig[v]; ok {
				if entityConfig.FieldsConfig == nil {
					entityConfig.FieldsConfig = make(map[string]FieldConfig)
				}
				entityConfig.FieldsConfig[k] = fieldConfig
			}
		}
	}

	// override global values with entity specific config values if exist
	if eConfig, ok := config.EntitiesConfig[entityType]; ok {
		if eConfig.FieldsConfig != nil {
			entityConfig.FieldsConfig = eConfig.FieldsConfig
		}
		if eConfig.IdField != "" {
			entityConfig.IdField = eConfig.IdField
		}
		if eConfig.SliceChangesFormat != "" {
			entityConfig.SliceChangesFormat = eConfig.SliceChangesFormat
		}
		entityConfig.AuditEventForNoChanges = eConfig.AuditEventForNoChanges
	}

	return entityConfig
}

func createEvents(auditChanges []AuditChange, eventType string, subType string, entityType string, auditEventForNoChanges bool, idField string, entity interface{}, values interface{}) []logger.Event {
	var events []logger.Event

	// get EntityID from idField
	var idFieldValue interface{}
	var idFieldStr string
	var idFieldPath []string
	if idField == "" {
		idField = DefaultIdField
	}
	if strings.HasPrefix(idField, "/") {
		idFieldPath = strings.Split(idField[1:], "/")
	} else {
		idFieldPath = strings.Split(idField, "/")
	}
	idFieldValue = getPathValue(idFieldPath, entity)
	if idFieldValue != nil {
		switch idFieldValue.(type) {
		case []byte:
			idFieldStr = fmt.Sprintf("%s", idFieldValue)
		default:
			idFieldStr = fmt.Sprintf("%v", idFieldValue)
		}

	}

	if len(auditChanges) == 0 && auditEventForNoChanges {
		auditEvent := AuditEvent{Type: eventType, Subtype: subType, EntityID: idFieldStr, EntityType: entityType, Values: values}
		event := logger.Event{
			Audit:    auditEvent,
			LogLevel: auditLogLevel,
			ID:       eventID(),
		}
		events = append(events, event)
	}

	for _, auditChange := range auditChanges {
		auditEvent := AuditEvent{Change: auditChange, Type: eventType, Subtype: subType, EntityID: idFieldStr, EntityType: entityType, Values: values}
		event := logger.Event{
			Audit:    auditEvent,
			LogLevel: auditLogLevel,
			ID:       eventID(),
		}
		events = append(events, event)
	}
	return events
}

var changeID = func() int64 {
	return time.Now().UnixNano() / 1000000
}

var eventID = func() string {
	return uuid.New().String()
}

func createAuditChanges(before interface{}, after interface{}, sliceChangesFormat SliceChangesFormat, fieldsConfig map[string]FieldConfig, nullIfEmpty bool) []AuditChange {
	// all changes events coming from the same audit call will share this eventID
	changeID := changeID()

	beforeNil := false
	afterNil := false

	if before == nil {
		t := reflect.ValueOf(after)
		typ := t.Type()
		before = (reflect.New(typ).Elem()).Interface()

		beforeNil = true
	}

	if after == nil {
		t := reflect.ValueOf(before)
		typ := t.Type()
		after = (reflect.New(typ).Elem()).Interface()

		afterNil = true
	}

	d := difference(before, after, sliceChangesFormat, fieldsConfig)

	auditChanges := []AuditChange{}
	for _, c := range d {
		foudnIgnoreField := false
		for fieldPathStr, fConfig := range fieldsConfig {
			if fConfig.Ignore {
				var ignFieldPath []string
				if strings.HasPrefix(fieldPathStr, "/") {
					ignFieldPath = strings.Split(fieldPathStr[1:], "/")
				} else {
					ignFieldPath = strings.Split(fieldPathStr, "/")
				}
				allPathItemsMatches := true
				if len(ignFieldPath) > len(c.Path) && len(ignFieldPath) > 0 {
					allPathItemsMatches = false
					break
				}

				for i, fieldN := range ignFieldPath {
					if fieldN != c.Path[i] {
						allPathItemsMatches = false
						break
					}
				}

				if allPathItemsMatches {
					foudnIgnoreField = true
				}
			}
		}

		if foudnIgnoreField {
			continue
		}

		var fieldName string = getFieldName(c.Path)
		var beforeValue interface{} = c.From
		if beforeNil {
			beforeValue = nil
		}
		var afterValue interface{} = c.To
		if afterNil {
			afterValue = nil
		}
		var fieldType FieldType
		var fieldFormat FieldFormat

		// try get field type from config
		fieldType, fieldFormat = getFieldTypeByName(fieldName, fieldsConfig)

		// if field type was not found, try get field type by value
		if fieldType == "" {
			if beforeValue != nil {
				fieldType, fieldFormat = getFieldTypeByValue(beforeValue)
			} else if afterValue != nil {
				fieldType, fieldFormat = getFieldTypeByValue(afterValue)
			}
		}

		// change values to string values for consistency
		beforeValue = stringValue(beforeValue, fieldType, nullIfEmpty)
		afterValue = stringValue(afterValue, fieldType, nullIfEmpty)

		auditChange := AuditChange{
			ID:     changeID,
			Path:   fieldName,
			Before: beforeValue,
			After:  afterValue,
			Type:   fieldType,
			Format: fieldFormat,
		}

		auditChanges = append(auditChanges, auditChange)

	}
	return auditChanges
}

func stringValue(value interface{}, valueType FieldType, nullIfEmpty bool) interface{} {
	var retValue interface{}
	if !nullIfEmpty {
		retValue = ""
	}
	valueStr := fmt.Sprintf("%v", value)
	switch valueType {
	case String:
		if value != "" && value != nil {
			retValue = valueStr
		}
	case Integer:
		if value != 0 && value != nil {
			retValue = valueStr
		}
	case Number:
		if value != 0 && value != nil {
			retValue = valueStr
		}
	case Array:
		if value != nil && valueStr != "[]" {
			switch value.(type) {
			case []byte:
				retValue = fmt.Sprintf("%s", value)
			default:
				valM, _ := json.Marshal(value)
				// []byte after being marshalled the first and last bytes are '"'
				if valM[0] == '"' && valM[len(valM)-1] == '"' {
					retValue = fmt.Sprintf("%s", valM[1:len(valM)-1])
				} else {
					retValue = fmt.Sprintf("%s", valM)
				}
			}
		}
	default:
		if value != nil && valueStr != "" {
			retValue = valueStr
		}
	}
	return retValue
}

func difference(before interface{}, after interface{}, globalSliceChangesFormat SliceChangesFormat, fieldsConfig map[string]FieldConfig) []utils.Change {
	var changes []utils.Change

	changes = utils.Difference(before, after)

	changesPerArray := make(map[string][]utils.Change)
	var arrayChanges []utils.Change
	var nonArraysChanges []utils.Change

	// populate changes for same array together
	for _, change := range changes {
		isArray := false
		// path is more than one element, arrays have at least 2 elemets on path
		if len(change.Path) > 1 {
			// if last element is an integer
			_, err := strconv.Atoi(change.Path[len(change.Path)-1])
			if err == nil {
				pathAsStringWithouLastElement := getFieldName(change.Path[:len(change.Path)-1])
				changesPerArray[pathAsStringWithouLastElement] = append(changesPerArray[pathAsStringWithouLastElement], change)
				isArray = true
			}
		}

		if !isArray {
			nonArraysChanges = append(nonArraysChanges, change)
		}
	}

	for path, sameArrayChanges := range changesPerArray {
		format := globalSliceChangesFormat
		if fConfig, ok := fieldsConfig[path]; ok {
			if fConfig.SliceChangesFormat != "" {
				format = fConfig.SliceChangesFormat
			}
		}

		if format == Full {
			newChange := changeArrayChangesToFull(sameArrayChanges[0].Path[:len(sameArrayChanges[0].Path)-1], before, after)
			arrayChanges = append(arrayChanges, newChange)
		} else if format == Unordered {
			sameArrayChanges = changeArrayChangesToUnordered(sameArrayChanges)
			arrayChanges = append(arrayChanges, sameArrayChanges...)
		} else if format == Ordered {
			arrayChanges = append(arrayChanges, sameArrayChanges...)
		}

	}

	return append(nonArraysChanges, arrayChanges...)
}

func changeArrayChangesToFull(path []string, before interface{}, after interface{}) utils.Change {
	beforeParentValue := getPathValue(path, before)
	afterParentValue := getPathValue(path, after)
	return utils.Change{
		From: beforeParentValue,
		To:   afterParentValue,
		Path: path,
	}
}

func changeArrayChangesToUnordered(sameArrayChanges []utils.Change) []utils.Change {
	arrLen := len(sameArrayChanges)
	for i := 0; i < arrLen; i++ {
		currChange := sameArrayChanges[i]
		for j := 0; j < arrLen; j++ {
			if i != j {
				otherChange := sameArrayChanges[j]
				// if current change from equals previous change to
				if reflect.DeepEqual(currChange.From, otherChange.To) && reflect.DeepEqual(currChange.To, otherChange.From) {
					// remove both elements from the array
					if j > i {
						// remove j index first
						copy(sameArrayChanges[j:], sameArrayChanges[j+1:])
						copy(sameArrayChanges[i:], sameArrayChanges[i+1:])
					} else {
						// remove i index first
						copy(sameArrayChanges[i:], sameArrayChanges[i+1:])
						copy(sameArrayChanges[j:], sameArrayChanges[j+1:])
					}
					sameArrayChanges[len(sameArrayChanges)-1] = utils.Change{}
					sameArrayChanges[len(sameArrayChanges)-2] = utils.Change{}
					sameArrayChanges = sameArrayChanges[:len(sameArrayChanges)-2]

					// start loop from begining
					i = 0
					arrLen = len(sameArrayChanges)
					// break j loop
					break
				} else if reflect.DeepEqual(currChange.From, otherChange.To) && !reflect.DeepEqual(currChange.To, otherChange.From) {
					// update prev to to curr to and remove current from array
					otherChange.To = currChange.To
					sameArrayChanges[j] = otherChange

					copy(sameArrayChanges[i:], sameArrayChanges[i+1:])
					sameArrayChanges[len(sameArrayChanges)-1] = utils.Change{}
					sameArrayChanges = sameArrayChanges[:len(sameArrayChanges)-1]

					// start loop from begining
					i = 0
					arrLen = len(sameArrayChanges)
					// break j loop
					break
				}
			}
		}
	}

	return sameArrayChanges
}

func getPathValue(path []string, s interface{}) interface{} {
	var retVal interface{}
	if s == nil {
		return retVal
	}

	val := reflect.ValueOf(s)

	if len(path) == 0 {
		return val.Interface().(interface{})
	}

	switch val.Kind() {
	case reflect.Map:
		for _, e := range val.MapKeys() {
			if e.String() == path[0] {
				retVal = getPathValue(path[1:], val.MapIndex(e).Interface())
			}
		}
	case reflect.Ptr:
		retVal = getPathValue(path, val.Elem().Interface())
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := val.Type().Field(i).Name
			fieldTag := strings.Split(val.Type().Field(i).Tag.Get("json"), ",")[0]
			fieldKind := field.Kind()

			if fieldName == path[0] || fieldTag == path[0] {
				// Check if it's a pointer to a struct.
				if fieldKind == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
					if field.CanInterface() {
						retVal = getPathValue(path[1:], field.Interface())
					}
					break
				}
				// Check if it's a struct value.
				if fieldKind == reflect.Struct {
					if field.CanAddr() && field.Addr().CanInterface() {
						// Recurse using an interface of the pointer value of the field.
						retVal = getPathValue(path[1:], field.Addr().Interface())
						break
					} else if field.CanInterface() {
						retVal = getPathValue(path[1:], field.Interface())
						break
					}
				}

				retVal = getPathValue(path[1:], field.Interface().(interface{}))
			}
		}
	}

	return retVal
}

func getFieldTypeByName(fieldName string, fieldsConfig map[string]FieldConfig) (FieldType, FieldFormat) {
	var fType FieldType
	var fFormat FieldFormat
	for fieldPathStr, fieldConf := range fieldsConfig {
		if !strings.HasPrefix(fieldPathStr, "/") {
			fieldPathStr = "/" + fieldPathStr
		}
		if fieldName == fieldPathStr {
			fType = fieldConf.TypeFormat.fieldType
			fFormat = fieldConf.TypeFormat.fieldFormat
		}
	}
	return fType, fFormat
}

var fieldTypeMap = map[reflect.Kind]FieldType{
	reflect.Bool:       Boolean,
	reflect.Int:        Integer,
	reflect.Int8:       Integer,
	reflect.Int16:      Integer,
	reflect.Int32:      Integer,
	reflect.Int64:      Integer,
	reflect.Uint:       Integer,
	reflect.Uint8:      Integer,
	reflect.Uint16:     Integer,
	reflect.Uint32:     Integer,
	reflect.Uint64:     Integer,
	reflect.Uintptr:    Integer,
	reflect.Float32:    Number,
	reflect.Float64:    Number,
	reflect.Complex64:  Number,
	reflect.Complex128: Number,
	reflect.Array:      Array,
	reflect.Chan:       Object,
	reflect.Func:       Object,
	reflect.Interface:  Object,
	reflect.Map:        Object,
	reflect.Ptr:        Object,
	reflect.Slice:      Array,
	reflect.String:     String,
	reflect.Struct:     Object,
}

var fieldFormatMap = map[reflect.Kind]FieldFormat{
	reflect.Int32:   Int32,
	reflect.Int64:   Int64,
	reflect.Uint32:  Int32,
	reflect.Uint64:  Int64,
	reflect.Float32: Float,
	reflect.Float64: Float,
}

func getFieldTypeByValue(fieldValue interface{}) (FieldType, FieldFormat) {
	// try to find field type
	var fieldType FieldType
	var fieldFormat FieldFormat
	if fieldValue != nil {
		xType := reflect.TypeOf(fieldValue).Kind()
		fieldType = fieldTypeMap[xType]
		fieldFormat = fieldFormatMap[xType]
	}
	return fieldType, fieldFormat
}

func getFieldName(path []string) string {
	var fieldName string = ""
	for _, name := range path {
		fieldName = fieldName + "/" + name
	}
	return fieldName
}
