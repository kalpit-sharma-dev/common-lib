package db

import (
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestInitializeCache(t *testing.T) {

	t.Run("Cache limit correct", func(t *testing.T) {
		expectedCacheLimit := 200
		data = nil
		cacheLimit = 0
		initializeCache(Config{CacheLimit: 200})

		if cacheLimit != expectedCacheLimit {
			t.Errorf("execpting cache limit to be %v but got %v", expectedCacheLimit, cacheLimit)
		}
	})
}

func TestGetStatement(t *testing.T) {
	initializeCache(Config{CacheLimit: 200})
	type args struct {
		scenario       string
		input          string
		expectedOutput *sqlx.Stmt
		cacheKey       string
		cacheValue     interface{}
	}

	sentinelStmt := &sqlx.Stmt{}

	tests := []args{
		{
			scenario: "Cache Empty",
			input:    "select * from abc",
		},
		{
			scenario:   "Key Not In Cache",
			input:      "select * from abc",
			cacheKey:   "abcd",
			cacheValue: nil,
		},
		{
			scenario:   "Invalid Value In Cache",
			input:      "select * from abc",
			cacheKey:   "select * from abc",
			cacheValue: 42,
		},
		{
			scenario:       "Valid Value In Cache",
			input:          "select * from abc",
			cacheKey:       "select * from abc",
			cacheValue:     sentinelStmt,
			expectedOutput: sentinelStmt,
		},
	}

	for _, test := range tests {
		if test.cacheKey != "" {
			data.Set(test.cacheKey, test.cacheValue, -1)
		}
		stmt := getStatement(test.input)
		if stmt != test.expectedOutput {
			t.Errorf("%s : Failed as expected return value : %v but got : %v", test.scenario, test.expectedOutput, stmt)
		}
	}
}

func TestAddStatement(t *testing.T) {
	initializeCache(Config{CacheLimit: 200})
	type args struct {
		scenario   string
		inputKey   string
		inputValue *sqlx.Stmt
		cacheFull  bool
	}

	sentinelStmt := &sqlx.Stmt{}

	tests := []args{
		{
			scenario:   "Cache Not Full",
			inputKey:   "select * from mocktable",
			inputValue: sentinelStmt,
			cacheFull:  false,
		},
		{
			scenario:   "Cache Full",
			inputKey:   "select * from mocktable",
			inputValue: sentinelStmt,
			cacheFull:  true,
		},
	}

	for _, test := range tests {
		if test.cacheFull {
			for i := 0; data.ItemCount() < cacheLimit; i++ {
				data.Set(strconv.Itoa(i), sentinelStmt, -1)
			}
		}

		currentSize := data.ItemCount()
		addStatement("", test.inputKey, test.inputValue)

		if test.cacheFull && data.ItemCount() != 1 {
			t.Errorf("Expected cache to have only one element, found %d", data.ItemCount())
		}
		if !test.cacheFull && data.ItemCount() != currentSize+1 {
			t.Errorf("Expected cache to have %d element(s), found %d", currentSize+1, data.ItemCount())
		}
	}
}

func TestDelete(t *testing.T) {
	initializeCache(Config{CacheLimit: 200})
	sentinelStmt := &sqlx.Stmt{}
	key := "abcd"
	data.Set(key, sentinelStmt, -1)

	stmt := getStatement(key)
	if stmt == nil {
		t.Errorf("Expected value to be present for key %s", key)
	}
	deleteKey(key)
	stmt = getStatement(key)
	if stmt != nil {
		t.Errorf("Expected value to be not present for key %s", key)
	}
}
