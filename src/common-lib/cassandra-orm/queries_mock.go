package cassandraorm

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/qb"
)

// QueryMockFunc describe type for select/get mocks
type QueryMockFunc func(*AccessorMock, interface{}, qb.Builder, map[string]interface{}) error

// DefaultSelectMock default mock for select function
func DefaultSelectMock(mock *AccessorMock, values interface{}, queryBuilder qb.Builder, params map[string]interface{}) error {
	ok, err := selectInMock(mock, values, queryBuilder, params)
	if ok {
		return err
	}
	res, err := find(mock, params)
	if err != nil {
		return err
	}
	SetSliceTo(values, res)
	return nil
}

// DefaultGetMock default mock for get function
func DefaultGetMock(mock *AccessorMock, value interface{}, _ qb.Builder, params map[string]interface{}) error {
	res, err := find(mock, params)
	if err != nil {
		return err
	}
	SetValueTo(value, res[0])
	return nil
}

// selectInMock - mock for 'SELECT ... WHERE <IDField> IN (<ids> ...)' query
func selectInMock(mock *AccessorMock, values interface{}, queryBuilder qb.Builder, params map[string]interface{}) (bool, error) {
	if len(params) != 1 {
		return false, nil
	}
	keyName := mock.Quote(mock.Keys()[0])
	param, found := params[keyName]
	if !found || reflect.TypeOf(param).Kind() != reflect.Slice {
		return false, nil
	}
	slice := reflect.ValueOf(param)
	for i := 0; i < slice.Len(); i++ {
		id := fmt.Sprintf("%v", slice.Index(i))
		newParams := map[string]interface{}{keyName: id}
		err := DefaultSelectMock(mock, values, queryBuilder, newParams)
		if err != nil && err != gocql.ErrNotFound {
			return true, err
		}
	}
	if values == nil {
		return true, gocql.ErrNotFound
	}
	return true, nil
}

func find(mock *AccessorMock, params map[string]interface{}) ([]interface{}, error) {
	res := mock.FindBy(unquote(params))
	if len(res) == 0 {
		return nil, gocql.ErrNotFound
	}
	return res, nil
}

func unquote(params map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(params))
	for key, value := range params {
		unquoted, err := strconv.Unquote(key)
		if err == nil {
			key = unquoted
		}

		m[key] = value
	}
	return m
}
