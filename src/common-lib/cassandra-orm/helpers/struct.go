package helpers

import (
	"reflect"

	"github.com/pkg/errors"

	ref "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/reflect"
)

var (
	// ErrParamIsNotStruct parameter is not struct
	ErrParamIsNotStruct = errors.New("a and b must be structs")
	// ErrSerialization serialization's error
	ErrSerialization = errors.New("unable to serialize")
)

// SerializeDefault converts struct to map for gocql
func SerializeDefault(item interface{}) (map[string]interface{}, error) {
	row, ok := ref.StructToMap(item)
	if !ok {
		return nil, ErrSerialization
	}

	return row, nil
}

// Equal checks if two structs are equal. List of fields to compare could be sent in fields,
// otherwise it will check all struct fields
func Equal(a, b interface{}, fields ...string) (bool, error) {
	aRef := getElem(a)
	bRef := getElem(b)

	if aRef.Kind() != reflect.Struct || bRef.Kind() != reflect.Struct {
		return false, ErrParamIsNotStruct
	}
	if len(fields) == 0 {
		for i := 0; i < aRef.NumField(); i++ {
			fields = append(fields, aRef.Type().Field(i).Name)
		}
	}

	return compareFields(aRef, bRef, fields...), nil
}

func getElem(iface interface{}) reflect.Value {
	value := reflect.ValueOf(iface)
	if value.Kind() != reflect.Ptr {
		return value
	}
	return value.Elem()
}

func compareFields(aRef, bRef reflect.Value, fields ...string) bool {
	for _, f := range fields {
		aField := aRef.FieldByName(f)
		bField := bRef.FieldByName(f)

		if aField.Kind() == reflect.Slice {
			if aField.Len() == 0 && bField.Len() == 0 {
				continue
			}

			if aField.Len() == 0 || bField.Len() == 0 {
				return false
			}
		}

		if !reflect.DeepEqual(aField.Interface(), bField.Interface()) {
			return false
		}
	}

	return true
}
