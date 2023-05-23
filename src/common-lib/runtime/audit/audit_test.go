package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/contextutil"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var entity1 = TestEntity{
	Field1: "Test",
	Field2: 10,
	Field3: TestSubEntity{
		Field4: "Test",
		Field5: []int{1, 2, 3},
	},
}
var entity2 = TestEntity{
	Field1: "Test",
	Field2: 20,
	Field3: TestSubEntity{
		Field4: "Test",
		Field5: []int{1, 2, 3},
	},
}
var entity3 = TestEntity{
	Field1: "Test",
	Field2: 30,
	Field3: TestSubEntity{
		Field4: "New",
		Field5: []int{1, 3},
	},
}

var entity5 = TestEntity{
	Field1: "Test",
	Field2: 30,
	Field3: TestSubEntity{
		Field4: "New",
		Field5: []int{1, 4, 3},
	},
}

var entity4 = TestEntity2{
	Field1: "Test",
	Field6: "Test",
	Field7: 10,
	Field2: 40,
}

var b = &bytes.Buffer{}
var sink = &MemorySink{b}
var httpConttext = context.Background()

type MemorySink struct {
	*bytes.Buffer
}

func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

func TestCreate(t *testing.T) {
	t.Run("Create New Audit Instance", func(t *testing.T) {
		_, err := NewAuditLogger(Config{})
		if err == nil {
			t.Errorf("NewAudit() error = %v, wantErr %v", err, "AuditNameIsRequired")
			return
		}
		_, err = NewAuditLogger(Config{AuditName: "TestNewAuditWithName"})
		if err != nil {
			t.Errorf("NewAudit() error = %v, wantErr %v", err, nil)
			return
		}
		_, err = NewAuditLogger(Config{AuditName: "TestNewAuditWithName"})
		if err == nil {
			t.Errorf("NewAudit() error = %v, wantErr %v", nil, "LoggerAlreadyInitialized")
			return
		}
	})
}

func Test_getFieldTypeByValue(t *testing.T) {
	tests := []struct {
		name  string
		field interface{}
		wantT FieldType
		wantF FieldFormat
	}{
		{name: "test type bool", field: bool(true), wantT: Boolean},
		{name: "test type int", field: int(1), wantT: Integer},
		{name: "test type int8", field: int8(1), wantT: Integer},
		{name: "test type int16", field: int16(1), wantT: Integer},
		{name: "test type int32", field: int32(1), wantT: Integer, wantF: Int32},
		{name: "test type int64", field: int64(1), wantT: Integer, wantF: Int64},
		{name: "test type uint", field: uint(1), wantT: Integer},
		{name: "test type uint8", field: uint8(1), wantT: Integer},
		{name: "test type uint16", field: uint16(1), wantT: Integer},
		{name: "test type uint32", field: uint32(1), wantT: Integer, wantF: Int32},
		{name: "test type uint64", field: uint64(1), wantT: Integer, wantF: Int64},
		{name: "test type uintptr", field: uintptr(1), wantT: Integer},
		{name: "test type float32", field: float32(1), wantT: Number, wantF: Float},
		{name: "test type float64", field: float64(1), wantT: Number, wantF: Float},
		{name: "test type complex64", field: complex64(1), wantT: Number},
		{name: "test type complex128", field: complex128(1), wantT: Number},
		{name: "test type slice", field: []int{1}, wantT: Array},
		{name: "test type array", field: [...]int{1}, wantT: Array},
		{name: "test type channel", field: make(chan int), wantT: Object},
		{name: "test type function", field: func() {}, wantT: Object},
		{name: "test type pointer", field: &TestEntity{}, wantT: Object},
		{name: "test type struct", field: TestEntity{}, wantT: Object},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotT, gotF := getFieldTypeByValue(tt.field); gotT != tt.wantT || gotF != tt.wantF {
				t.Errorf("getFieldTypeByValue = %v %v, want %v %v", gotT, gotF, tt.wantT, tt.wantF)
			}
		})
	}
}

func Test_difference(t *testing.T) {
	validate := func(got []utils.Change, want []utils.Change, t *testing.T) {
		if len(got) != len(want) {
			t.Errorf("Test_difference() Lengths don't match = %v, want %v", len(got), len(want))
			t.Errorf("Test_difference() = %v, want %v", got, want)
			return
		}
		success := true
		for _, w := range want {
			found := false
			for _, g := range got {
				if reflect.DeepEqual(g.Path, w.Path) && reflect.DeepEqual(g.From, w.From) && reflect.DeepEqual(g.To, w.To) {
					found = true
					break
				}
			}

			if !found {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_difference() = %v, want %v", got, want)
			return
		}
	}

	var entity1J, entity2J, entity3J, entity4J interface{}
	entity1M, _ := json.Marshal(entity1)
	json.Unmarshal(entity1M, &entity1J)
	entity2M, _ := json.Marshal(entity2)
	json.Unmarshal(entity2M, &entity2J)
	entity3M, _ := json.Marshal(entity3)
	json.Unmarshal(entity3M, &entity3J)
	entity4M, _ := json.Marshal(entity4)
	json.Unmarshal(entity4M, &entity4J)

	t.Run("diff Unordered", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   20,
			},
		}
		got := difference(entity1, entity2, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered Same Elements Different Positions", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   20,
			},
		}
		entity6 := entity2
		entity6.Field3.Field5 = []int{2, 3, 1}
		got := difference(entity1, entity6, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered Same Elements Different Positions 2", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   20,
			},
		}
		entity6 := entity1
		entity6.Field3.Field5 = []int{1, 2, 3, 4}

		entity7 := entity2
		entity7.Field3.Field5 = []int{3, 2, 1, 4}
		got := difference(entity6, entity7, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Ordered", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   20,
			},
		}
		got := difference(entity1, entity2, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   30,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []int{1, 2, 3},
				To:   []int{1, 3},
			},
		}
		got := difference(entity1, entity3, Full, map[string]FieldConfig{})
		validate(got, want, t)
	})
	t.Run("diff Full 2", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   30,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []int{1, 2, 3},
				To:   []int{1, 4, 3},
			},
		}
		got := difference(entity1, entity5, Full, map[string]FieldConfig{})
		validate(got, want, t)
	})
	t.Run("diff Unordered Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   30,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: 2,
				To:   nil,
			},
		}
		got := difference(entity1, entity3, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Ordered Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   30,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: 2,
				To:   3,
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: 3,
				To:   nil,
			},
		}
		got := difference(entity1, entity3, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   30,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []int{1, 2, 3},
				To:   []int{1, 3},
			},
		}
		got := difference(entity1, entity3, Full, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered JSON", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   float64(20),
			},
		}
		got := difference(entity1J, entity2J, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Ordered JSON", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   float64(20),
			},
		}
		got := difference(entity1J, entity2J, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered JSON Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   float64(30),
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: float64(2),
				To:   nil,
			},
		}
		got := difference(entity1J, entity3J, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Ordered JSON Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   float64(30),
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: float64(2),
				To:   float64(3),
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: float64(3),
				To:   nil,
			},
		}
		got := difference(entity1J, entity3J, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full JSON Multiple Changes", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   float64(30),
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   "New",
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []interface{}{float64(1), float64(2), float64(3)},
				To:   []interface{}{float64(1), float64(3)},
			},
		}
		got := difference(entity1J, entity3J, Full, map[string]FieldConfig{})
		validate(got, want, t)
	})
	t.Run("diff Unordered different structs", func(t *testing.T) {
		// missing changes from the created object, if objects are JSON it shows all changes
		want := []utils.Change{
			{
				Path: []string{"Field2"},
				From: nil,
				To:   40,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "0"},
				From: 1,
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: 2,
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: 3,
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   10,
			},
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   nil,
			},
		}
		got := difference(entity1, entity4, Unordered, map[string]FieldConfig{})
		validate(got, want, t)
	})
	t.Run("diff Ordered different structs", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2"},
				From: nil,
				To:   40,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "0"},
				From: 1,
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: 2,
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: 3,
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   10,
			},
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   nil,
			},
		}
		got := difference(entity1, entity4, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full different structs", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2"},
				From: nil,
				To:   40,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []int{1, 2, 3},
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   10,
			},
			{
				Path: []string{"Field2J"},
				From: 10,
				To:   nil,
			},
		}
		got := difference(entity1, entity4, Full, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered different structs JSON", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "0"},
				From: float64(1),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: float64(2),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: float64(3),
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   float64(10),
			},
			{
				Path: []string{"Field2"},
				From: nil,
				To:   float64(40),
			},
		}
		got := difference(entity1J, entity4J, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Ordered different structs JSON", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "0"},
				From: float64(1),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "1"},
				From: float64(2),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5", "2"},
				From: float64(3),
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   float64(10),
			},
			{
				Path: []string{"Field2"},
				From: nil,
				To:   float64(40),
			},
		}
		got := difference(entity1J, entity4J, Ordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full different structs JSON", func(t *testing.T) {
		want := []utils.Change{
			{
				Path: []string{"Field2J"},
				From: float64(10),
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field4"},
				From: "Test",
				To:   nil,
			},
			{
				Path: []string{"Field3", "Field5"},
				From: []interface{}{float64(1), float64(2), float64(3)},
				To:   nil,
			},
			{
				Path: []string{"Field6"},
				From: nil,
				To:   "Test",
			},
			{
				Path: []string{"Field7"},
				From: nil,
				To:   float64(10),
			},
			{
				Path: []string{"Field2"},
				From: nil,
				To:   float64(40),
			},
		}
		got := difference(entity1J, entity4J, Full, map[string]FieldConfig{})

		validate(got, want, t)
	})

	t.Run("diff Unordered Struct With Multiple Arrays Multiple Changes", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"4", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1", "0"},
				From: "1",
				To:   "4",
			},
			{
				Path: []string{"Field1", "1"},
				From: "2",
				To:   nil,
			},
			{
				Path: []string{"Field2", "2"},
				From: 3,
				To:   nil,
			},
		}
		got := difference(e1, e2, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Unordered Struct With Multiple Arrays", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"1", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1", "1"},
				From: "2",
				To:   nil,
			},
			{
				Path: []string{"Field2", "2"},
				From: 3,
				To:   nil,
			},
		}
		got := difference(e1, e2, Unordered, map[string]FieldConfig{})

		validate(got, want, t)
	})
	t.Run("diff Full Struct With Multiple Arrays", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"1", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1"},
				From: []string{"1", "2", "3"},
				To:   []string{"1", "3"},
			},
			{
				Path: []string{"Field2"},
				From: []int{1, 2, 3},
				To:   []int{1, 2},
			},
		}
		got := difference(e1, e2, Full, map[string]FieldConfig{})

		validate(got, want, t)
	})

	t.Run("diff Full Struct With Multiple Arrays Override Path to Unordered", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"1", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1", "1"},
				From: "2",
				To:   nil,
			},
			{
				Path: []string{"Field2"},
				From: []int{1, 2, 3},
				To:   []int{1, 2},
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field1"] = FieldConfig{SliceChangesFormat: Unordered}

		got := difference(e1, e2, Full, fieldsConfig)

		validate(got, want, t)
	})
	t.Run("diff Full Struct With Multiple Arrays Override Path to Ordered", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"1", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1", "1"},
				From: "2",
				To:   "3",
			},
			{
				Path: []string{"Field1", "2"},
				From: "3",
				To:   nil,
			},
			{
				Path: []string{"Field2"},
				From: []int{1, 2, 3},
				To:   []int{1, 2},
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field1"] = FieldConfig{SliceChangesFormat: Ordered}

		got := difference(e1, e2, Full, fieldsConfig)

		validate(got, want, t)
	})
	t.Run("diff Unordered Struct With Multiple Arrays Override Full", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"1", "2", "3"},
			Field2: []int{1, 2, 3},
		}
		e2 := TestEntity3{
			Field1: []string{"1", "3"},
			Field2: []int{1, 2},
		}
		want := []utils.Change{
			{
				Path: []string{"Field1"},
				From: []string{"1", "2", "3"},
				To:   []string{"1", "3"},
			},
			{
				Path: []string{"Field2", "2"},
				From: 3,
				To:   nil,
			},
		}
		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field1"] = FieldConfig{SliceChangesFormat: Full}

		got := difference(e1, e2, Unordered, fieldsConfig)

		validate(got, want, t)
	})
}

func Test_createAuditChange(t *testing.T) {
	validate := func(got []AuditChange, want []AuditChange, t *testing.T) {
		if len(got) != len(want) {
			t.Errorf("Test_createAuditChanges() Lengths don't match = %v, want %v", len(got), len(want))
			return
		}
		for _, w := range want {
			found := false
			for _, g := range got {
				if w.Path == g.Path && reflect.DeepEqual(w.Before, g.Before) && reflect.DeepEqual(w.After, g.After) && w.Type == g.Type {
					found = true
				}
			}
			if !found {
				t.Errorf("Test_createAuditChanges() don't match = %v, want %v", got, w)
				return
			}
		}

	}

	var entity1J, entity2J, entity3J, entity4J interface{}
	entity1M, _ := json.Marshal(entity1)
	json.Unmarshal(entity1M, &entity1J)
	entity2M, _ := json.Marshal(entity2)
	json.Unmarshal(entity2M, &entity2J)
	entity3M, _ := json.Marshal(entity3)
	json.Unmarshal(entity3M, &entity3J)
	entity4M, _ := json.Marshal(entity4)
	json.Unmarshal(entity4M, &entity4J)

	t.Run("createAuditChanges Unordered", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "20",
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity2, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "20",
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity2, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Unordered Multiple Changes", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity3, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Unordered Multiple Changes Null If Empty False", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "",
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity3, Unordered, map[string]FieldConfig{}, false)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered Multiple Changes", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity3, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})

	t.Run("createAuditChanges Ordered Multiple Changes Null If Empty False", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  "",
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity3, Ordered, map[string]FieldConfig{}, false)

		validate(got, want, t)
	})

	t.Run("createAuditChanges Unordered Multiple Changes Ignore Field 1", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "integer",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3/Field4"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1, entity3, Unordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered Multiple Changes Ignore Field 1", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "integer",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3/Field4"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1, entity3, Ordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Unordered Multiple Changes Ignore Field Parent With Multiple Childs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1, entity3, Unordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered Multiple Changes Ignore Field Parent With Multiple Childs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "integer",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["Field3"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1, entity3, Ordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Unordered", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "20",
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity2J, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Ordered", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "20",
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity2J, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Unordered Multiple Changes", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity3J, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Ordered Multiple Changes", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "New",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity3J, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Unordered Multiple Changes Ignore Field 1", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "number",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3/Field4"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1J, entity3J, Unordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Ordered Multiple Changes Ignore Field 1", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
				Type:   "number",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "number",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3/Field4"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1J, entity3J, Ordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Unordered Multiple Changes Ignore Field Parent With Multiple Childs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1J, entity3J, Unordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges JSON Ordered Multiple Changes Ignore Field Parent With Multiple Childs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "30",
				Type:   "number",
			},
		}

		fieldsConfig := make(map[string]FieldConfig)
		fieldsConfig["/Field3"] = FieldConfig{Ignore: true}

		got := createAuditChanges(entity1J, entity3J, Ordered, fieldsConfig, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Unordered Different Structs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2",
				Before: nil,
				After:  "40",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  nil,
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/0",
				Before: "1",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field6",
				Before: nil,
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: nil,
				After:  "10",
				Type:   "integer",
			},
			{
				Path:   "/Field2J",
				Before: "10",
				After:  nil,
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity4, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})

	t.Run("createAuditChanges Unordered Different Structs Null If Empty False", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2",
				Before: "",
				After:  "40",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  "",
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/0",
				Before: "1",
				After:  "",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  "",
				Type:   "integer",
			},
			{
				Path:   "/Field6",
				Before: "",
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: "",
				After:  "10",
				Type:   "integer",
			},
			{
				Path:   "/Field2J",
				Before: "10",
				After:  "",
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity4, Unordered, map[string]FieldConfig{}, false)

		validate(got, want, t)
	})

	t.Run("createAuditChanges Full Different Structs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2",
				Before: nil,
				After:  "40",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  nil,
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5",
				Before: "[1,2,3]",
				After:  nil,
				Type:   "array",
			},
			{
				Path:   "/Field6",
				Before: nil,
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: nil,
				After:  "10",
				Type:   "integer",
			},
			{
				Path:   "/Field2J",
				Before: "10",
				After:  nil,
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity4, Full, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Full Different Structs String Array", func(t *testing.T) {
		e1 := TestEntity3{
			Field1: []string{"val{ 1", "val} 2", "val 3"},
		}
		e2 := TestEntity3{
			Field1: []string{"val 4", "val 3"},
		}
		want := []AuditChange{
			{
				Path:   "/Field1",
				Before: "[\"val{ 1\",\"val} 2\",\"val 3\"]",
				After:  "[\"val 4\",\"val 3\"]",
				Type:   "array",
			},
		}
		got := createAuditChanges(e1, e2, Full, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered Different Structs", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2",
				Before: nil,
				After:  "40",
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  nil,
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/0",
				Before: "1",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "integer",
			},
			{
				Path:   "/Field6",
				Before: nil,
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: nil,
				After:  "10",
				Type:   "integer",
			},
			{
				Path:   "/Field2J",
				Before: "10",
				After:  nil,
				Type:   "integer",
			},
		}
		got := createAuditChanges(entity1, entity4, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Unordered Different Structs JSON", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  nil,
				Type:   "number",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  nil,
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/0",
				Before: "1",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field6",
				Before: nil,
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: nil,
				After:  "10",
				Type:   "number",
			},
			{
				Path:   "/Field2",
				Before: nil,
				After:  "40",
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity4J, Unordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
	t.Run("createAuditChanges Ordered Different Structs JSON", func(t *testing.T) {
		want := []AuditChange{
			{
				Path:   "/Field2J",
				Before: "10",
				After:  nil,
				Type:   "number",
			},
			{
				Path:   "/Field3/Field4",
				Before: "Test",
				After:  nil,
				Type:   "string",
			},
			{
				Path:   "/Field3/Field5/0",
				Before: "1",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  nil,
				Type:   "number",
				Format: "float",
			},
			{
				Path:   "/Field6",
				Before: nil,
				After:  "Test",
				Type:   "string",
			},
			{
				Path:   "/Field7",
				Before: nil,
				After:  "10",
				Type:   "number",
			},
			{
				Path:   "/Field2",
				Before: nil,
				After:  "40",
				Type:   "number",
			},
		}
		got := createAuditChanges(entity1J, entity4J, Ordered, map[string]FieldConfig{}, true)

		validate(got, want, t)
	})
}

func Test_createEvents(t *testing.T) {
	validate := func(got []logger.Event, want []AuditChange, requiredEventsNumber int, t *testing.T) {
		if len(got) != requiredEventsNumber {
			t.Errorf("Test_createEvents() = %v, want %v", len(got), requiredEventsNumber)
			return
		}

		for _, w := range want {
			found := false
			for _, e := range got {
				c := e.Audit.(AuditEvent).Change
				if c.Path == w.Path && c.Before == w.Before && c.After == w.After && c.Type == w.Type {
					found = true
				}
			}

			if !found {
				t.Errorf("Test_createEvents() Change Not Found = %v, want %v", got, want)
				return
			}
		}
	}

	t.Run("createEvents 1", func(t *testing.T) {
		auditChanges := []AuditChange{{
			Path:   "Field2",
			Before: "10",
			After:  "30",
		}}
		changeType := "Record was updated"
		subtype := "field-updated"

		want := auditChanges

		got := createEvents(auditChanges, changeType, subtype, "", false, "", nil, nil)

		validate(got, want, 1, t)
	})
	t.Run("createEvents 2", func(t *testing.T) {
		auditChanges := []AuditChange{
			{
				Path:   "/Field2",
				Before: "10",
				After:  "30",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  "<no value>",
			},
		}
		changeType := "Record was updated"
		subtype := "field-updated"

		want := auditChanges

		got := createEvents(auditChanges, changeType, subtype, "", false, "", nil, nil)

		validate(got, want, 3, t)
	})
	t.Run("createEvents 3", func(t *testing.T) {
		auditChanges := []AuditChange{
			{
				Path:   "/Field2",
				Before: "10",
				After:  "30",
			},
			{
				Path:   "/Field3/Field5/1",
				Before: "2",
				After:  "3",
			},
			{
				Path:   "/Field3/Field5/2",
				Before: "3",
				After:  "<no value>",
			},
		}
		changeType := "Record was updated"
		subtype := "field-updated"

		want := auditChanges

		got := createEvents(auditChanges, changeType, subtype, "", false, "", nil, nil)

		validate(got, want, 3, t)
	})
	t.Run("createEvents 4", func(t *testing.T) {
		auditChanges := []AuditChange{}
		changeType := "Record was updated"
		subtype := "field-updated"

		want := auditChanges

		got := createEvents(auditChanges, changeType, subtype, "", true, "", nil, nil)

		validate(got, want, 1, t)
	})
}

func Test_AuditEvents(t *testing.T) {
	validate := func(got string, want []string, validateString string, validateStringCount int, t *testing.T) {
		success := true
		for _, w := range want {
			if !strings.Contains(got, w) {
				success = false
			}
		}

		if !success {
			t.Errorf("Test_AuditEvents() = %v, want %v", got, want)
			return
		}

		foundStringCount := strings.Count(got, validateString)
		if foundStringCount != validateStringCount {
			t.Errorf("Test_AuditEvents() found string %v count %v, want %v", validateString, foundStringCount, validateStringCount)
			return
		}
	}

	wantResponse := []string{
		"\"" + logger.CallerKey + "\":\"audit/audit_test.go:400\",",
		"\"Resource\":{\"PartnerId\":\"part123\",\"CorrelationId\":\"trans123\",\"UserId\":\"usr123\",\"RequestId\":",
	}

	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	oldChangeID := changeID
	defer func() {
		changeID = oldChangeID
	}()
	changeID = func() int64 {
		return 101
	}

	oldEventID := eventID
	defer func() {
		eventID = oldEventID
	}()
	eventID = func() string {
		return "a1s2d3f4"
	}

	var entity1J, entity2J, entity3J interface{}
	entity1M, _ := json.Marshal(entity1)
	json.Unmarshal(entity1M, &entity1J)
	entity2M, _ := json.Marshal(entity2)
	json.Unmarshal(entity2M, &entity2J)
	entity3M, _ := json.Marshal(entity3)
	json.Unmarshal(entity3M, &entity3J)

	t.Run("Audit EventS", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_EventS", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.EventS(ctx, "EventType", "Test Audit Message", "Desc")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Message\":\"Test Audit Message\",\"Id\":\"" + eventID() + "\",\"Description\":\"Desc\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"EventType\",\"Change\":{}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{}}", 1, t)
	})

	t.Run("Audit EventS Empty Event Type", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_EventS_Empty_Event_Type", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, err := a.EventS(ctx, "", "Test Audit Message", "Desc")

		if err == nil {
			t.Errorf("Error = %v, wantErr %v", err, "Event type is required")
			return
		}
	})

	t.Run("Audit EventS Using With Values", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_EventS_Using_With_Values", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		v := map[string]string{"ID": "123", "Name": "ABC"}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.With(AddValues(v)).EventS(ctx, "EventType", "Test Audit Message", "Desc")

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Message\":\"Test Audit Message\",\"Id\":\"" + eventID() + "\",\"Description\":\"Desc\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"EventType\",\"Values\":{\"ID\":\"123\",\"Name\":\"ABC\"},\"Change\":{}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{}}", 1, t)
	})

	t.Run("Audit EventS Using With Call Depth", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_EventS_Using_With_Call_Depth", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		auditEventS(a, ctx, "EventType", "Test Audit Message", "Desc", 2)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Message\":\"Test Audit Message\",\"Id\":\"" + eventID() + "\",\"Description\":\"Desc\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"EventType\",\"Change\":{}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{}}", 1, t)
	})

	t.Run("Audit Event", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		event := AuditEvent{
			Type:       "Event Type",
			Subtype:    "Event Subtype",
			EntityType: "Event Entity Type",
			EntityID:   "Event Entity ID",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.Event(ctx, event)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Event Type\",\"Subtype\":\"Event Subtype\",\"EntityType\":\"Event Entity Type\",\"EntityId\":\"Event Entity ID\",\"Change\":{\"Id\":1,\"Path\":\"Field 1 Name\",\"Before\":\"Before 1 Value\",\"After\":\"After 1 Value\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit Event Empty Event Type", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event_Empty_Event_Type", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		event := AuditEvent{
			Subtype:    "Event Subtype",
			EntityType: "Event Entity Type",
			EntityID:   "Event Entity ID",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
		}

		_, err := a.Event(ctx, event)

		if err == nil {
			t.Errorf("Error = %v, wantErr %v", err, "Event type is required")
			return
		}
	})

	t.Run("Audit Event With Message and Description", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event_With_Message_And_Description", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		event := AuditEvent{
			Type:        "Event Type",
			Subtype:     "Event Subtype",
			EntityType:  "Event Entity Type",
			EntityID:    "Event Entity ID",
			Message:     "Event Message",
			Description: "Event Description",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.Event(ctx, event)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Message\":\"Event Message\",\"Id\":\"" + eventID() + "\",\"Description\":\"Event Description\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Event Type\",\"Subtype\":\"Event Subtype\",\"EntityType\":\"Event Entity Type\",\"EntityId\":\"Event Entity ID\",\"Change\":{\"Id\":1,\"Path\":\"Field 1 Name\",\"Before\":\"Before 1 Value\",\"After\":\"After 1 Value\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit Event With Values", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event_With_Values", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		newValues := make(map[string]string)
		newValues["f1"] = "1"
		newValues["f2"] = "2"
		event := AuditEvent{
			Type:       "Event Type",
			Subtype:    "Event Subtype",
			EntityType: "Event Entity Type",
			EntityID:   "Event Entity ID",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
			Values: newValues,
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.Event(ctx, event)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\"",
			"\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Event Type\",\"Subtype\":\"Event Subtype\",\"EntityType\":\"Event Entity Type\",\"EntityId\":\"Event Entity ID\"",
			"\"Values\":{\"f1\":\"1\",\"f2\":\"2\"}",
			"\"Change\":{",
			"{\"Id\":1,\"Path\":\"Field 1 Name\",\"Before\":\"Before 1 Value\",\"After\":\"After 1 Value\",\"Type\":\"string\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit Event With Struct Values", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event_With_Struct_Values", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		newValues := TestEntity2{
			Field1: "1",
			Field2: 2,
		}
		event := AuditEvent{
			Type:       "Event Type",
			Subtype:    "Event Subtype",
			EntityType: "Event Entity Type",
			EntityID:   "Event Entity ID",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
			Values: newValues,
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.Event(ctx, event)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\"",
			"\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Event Type\",\"Subtype\":\"Event Subtype\",\"EntityType\":\"Event Entity Type\",\"EntityId\":\"Event Entity ID\"",
			"\"Values\":{\"Field6\":\"\",\"Field7\":0,\"Field2\":2,\"Field1\":\"1\"}",
			"\"Change\":{",
			"{\"Id\":1,\"Path\":\"Field 1 Name\",\"Before\":\"Before 1 Value\",\"After\":\"After 1 Value\",\"Type\":\"string\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit Event With Struct Values With OM", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_Event_With_Struct_Values_With_OM", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)
		newValues := TestEntityOM{
			F1: "1",
			F2: 2,
		}
		event := AuditEvent{
			Type:       "Event Type",
			Subtype:    "Event Subtype",
			EntityType: "Event Entity Type",
			EntityID:   "Event Entity ID",
			Change: AuditChange{
				ID:     1,
				Path:   "Field 1 Name",
				Before: "Before 1 Value",
				After:  "After 1 Value",
				Type:   "string",
			},
			Values: newValues,
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.Event(ctx, event)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\"",
			"\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Event Type\",\"Subtype\":\"Event Subtype\",\"EntityType\":\"Event Entity Type\",\"EntityId\":\"Event Entity ID\"",
			"\"Values\":{\"f1\":\"1\",\"f2\":2}",
			"\"Change\":{",
			"{\"Id\":1,\"Path\":\"Field 1 Name\",\"Before\":\"Before 1 Value\",\"After\":\"After 1 Value\",\"Type\":\"string\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct Multiple Unordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_Struct_Multiple_Unordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 3, t)
	})

	t.Run("Audit UpdateEvent Empty Event Type", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_Empty_Event_Type", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, err := a.UpdateEvent(ctx, "", "", entity1, entity3)
		if err == nil {
			t.Errorf("Error = %v, wantErr %v", err, "Event type is required")
			return
		}
	})

	t.Run("Audit UpdateEvent Struct No Changes Empty Event", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalAuditEventForNoChanges: true, AuditName: "Test_AuditEvents_UpdateEvent_Struct_No_Changes_Empty_Event", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		entity6 := entity1

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity6)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct JSON Raw Message", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_Struct_JSON_Raw_Message", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		e1 := struct {
			Value interface{}
		}{
			&struct {
				SValue json.RawMessage
			}{
				json.RawMessage(`1`),
			},
		}
		e2 := struct {
			Value interface{}
		}{
			&struct {
				SValue json.RawMessage
			}{
				json.RawMessage(`2`),
			},
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", &e1, &e2)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"//Value//SValue/0\",\"Before\":\"49\",\"After\":\"50\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Slice", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_Slice", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		e1 := []TestEntity{
			{
				Field1: "1",
			},
		}
		e2 := []TestEntity{
			{
				Field1: "2",
			},
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", e1, e2)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/0/Field1\",\"Before\":\"1\",\"After\":\"2\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct No Changes No Event", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalAuditEventForNoChanges: false, AuditName: "Test_AuditEvents_UpdateEvent_Struct_No_Changes_No_Event", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		entity6 := entity1

		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity6)

		want := []string{}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 0, t)
	})

	t.Run("Audit UpdateEvent Struct With Date", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_Struct_With_Date", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		entity6 := TestEntity4{
			Field1: "Name",
			Field2: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		entity7 := TestEntity4{
			Field1: "Name",
			Field2: time.Date(1982, 1, 1, 12, 0, 0, 0, time.UTC),
		}
		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity6, entity7)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":\"1980-01-01 12:00:00 +0000 UTC\",\"After\":\"1982-01-01 12:00:00 +0000 UTC\",\"Type\":\"object\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct Multiple Ordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Ordered, AuditName: "Test_AuditEvents_UpdateEvent_Struct_Multiple_Ordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"3\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 4, t)
	})

	t.Run("Audit UpdateEvent JSON Multiple Unordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_UpdateEvent_JSON_Multiple_Unordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1J, entity3J)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"number\",\"Format\":\"float\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"\",\"Type\":\"number\",\"Format\":\"float\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 3, t)
	})
	t.Run("Audit UpdateEvent JSON Multiple Ordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Ordered, AuditName: "Test_AuditEvents_UpdateEvent_JSON_Multiple_Ordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1J, entity3J)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"number\",\"Format\":\"float\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"3\",\"Type\":\"number\",\"Format\":\"float\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":\"\",\"Type\":\"number\",\"Format\":\"float\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 4, t)
	})

	t.Run("Audit CreateEvent Struct Full", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Struct_Full", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field1\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":null,\"After\":\"10\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5\",\"Before\":null,\"After\":\"[1,2,3]\",\"Type\":\"array\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 4, t)
	})

	t.Run("Audit CreateEvent Empty Event Type", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Empty_Event_Type", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, err := a.CreateEvent(ctx, "", "", entity1)
		if err == nil {
			t.Errorf("Error = %v, wantErr %v", err, "Event type is required")
			return
		}
	})

	t.Run("Audit CreateEvent Struct Full Using With Values", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Struct_Full_Using_With_Values", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		v := map[string]string{"ID": "123", "Name": "ABC"}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.With(AddValues(v)).CreateEvent(ctx, "Record was created", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Values\":{\"ID\":\"123\",\"Name\":\"ABC\"},\"Change\":{\"Id\":101,\"Path\":\"/Field1\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Values\":{\"ID\":\"123\",\"Name\":\"ABC\"},\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":null,\"After\":\"10\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Values\":{\"ID\":\"123\",\"Name\":\"ABC\"},\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Values\":{\"ID\":\"123\",\"Name\":\"ABC\"},\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5\",\"Before\":null,\"After\":\"[1,2,3]\",\"Type\":\"array\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 4, t)
	})

	t.Run("Audit CreateEvent Struct With UUID", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Struct_With_UUID", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		u1uuid, _ := gocql.RandomUUID()

		testEntityUUID := TestEntityUUID{
			Id:   u1uuid,
			Name: "Test",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", testEntityUUID)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityId\":\"" + u1uuid.String() + "\",\"Change\":{\"Id\":101,\"Path\":\"/Name\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityId\":\"" + u1uuid.String() + "\",\"Change\":{\"Id\":101,\"Path\":\"/id\",\"Before\":null,\"After\":\"" + u1uuid.String() + "\",\"Type\":\"array\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})

	t.Run("Audit CreateEvent Struct With Bytes", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Struct_With_Bytes", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		u1uuid := []byte("\"Walter \\\"Heisenberg\\\" White\"")

		testEntityUUID := TestEntityUUID2{
			Id: u1uuid,
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", testEntityUUID)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityId\":\"\\\"Walter \\\\\\\"Heisenberg\\\\\\\" White\\\"\",\"Change\":{\"Id\":101,\"Path\":\"/id\",\"Before\":null,\"After\":\"\\\"Walter \\\\\\\"Heisenberg\\\\\\\" White\\\"\",\"Type\":\"array\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit CreateEvent Struct Unordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_CreateEvent_Struct_Unordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field1\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":null,\"After\":\"10\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/0\",\"Before\":null,\"After\":\"1\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":null,\"After\":\"2\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":null,\"After\":\"3\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 6, t)
	})
	t.Run("Audit CreateEvent Struct Ordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Ordered, AuditName: "Test_AuditEvents_CreateEvent_Struct_Ordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field1\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":null,\"After\":\"10\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":null,\"After\":\"Test\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/0\",\"Before\":null,\"After\":\"1\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":null,\"After\":\"2\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":null,\"After\":\"3\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 6, t)
	})

	t.Run("Audit DeleteEvent Struct Full", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_DeleteEvent_Struct_Full", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.DeleteEvent(ctx, "Record was deleted", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"Change\":{",
			"{\"Id\":101,\"Path\":\"/Field1\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5\",\"Before\":\"[1,2,3]\",\"After\":null,\"Type\":\"array\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 4, t)
	})

	t.Run("Audit DeleteEvent Empty Event Type", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_DeleteEvent_Empty_Event_Type", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, err := a.DeleteEvent(ctx, "", "", entity1)
		if err == nil {
			t.Errorf("Error = %v, wantErr %v", err, "Event type is required")
			return
		}
	})

	t.Run("Audit DeleteEvent Struct Unordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Unordered, AuditName: "Test_AuditEvents_DeleteEvent_Struct_Unordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.DeleteEvent(ctx, "Record was deleted", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"Change\":{",
			"{\"Id\":101,\"Path\":\"/Field1\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/0\",\"Before\":\"1\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":null,\"Type\":\"integer\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 6, t)
	})
	t.Run("Audit DeleteEvent Struct Ordered", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Ordered, AuditName: "Test_AuditEvents_DeleteEvent_Struct_Ordered", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.DeleteEvent(ctx, "Record was deleted", "", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"Change\":{",
			"{\"Id\":101,\"Path\":\"/Field1\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/0\",\"Before\":\"1\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":null,\"Type\":\"integer\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 6, t)
	})

	t.Run("Audit DeleteEvent Struct Ordered Provide Field Types and Formats", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		fieldsConifg := make(map[string]FieldConfig)
		fieldsConifg["/Field1"] = FieldConfig{TypeFormat: TypeFormat{fieldType: String, fieldFormat: Email}}
		fieldsConifg["Field2J"] = FieldConfig{TypeFormat: TypeFormat{fieldType: Integer, fieldFormat: Int32}}
		fieldsConifg["/Field3/Field4"] = FieldConfig{TypeFormat: TypeFormat{fieldType: Object}}

		entitiesConfig := make(map[string]EntityConfig)
		entitiesConfig["TestEntity"] = EntityConfig{FieldsConfig: fieldsConifg}

		config := Config{GlobalSliceChangesFormat: Ordered, EntitiesConfig: entitiesConfig, AuditName: "Test_AuditEvents_DeleteEvent_Struct_Ordered_Types_Formats", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.DeleteEvent(ctx, "Record was deleted", "TestEntity", entity1)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"EntityType\":\"TestEntity\",\"Change\":{",
			"{\"Id\":101,\"Path\":\"/Field1\",\"Before\":\"Test\",\"After\":null,\"Type\":\"string\",\"Format\":\"email\"}",
			"{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":null,\"Type\":\"integer\",\"Format\":\"int32\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":null,\"Type\":\"object\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/0\",\"Before\":\"1\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":null,\"Type\":\"integer\"}",
			"{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":null,\"Type\":\"integer\"}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 6, t)
	})

	t.Run("Audit CreateEvent Struct Entity ID with default value", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Entity_ID_Default_Value", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		var testEntity = TestEntity6{
			Id:     123,
			Field2: "test",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", testEntity)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/id\",\"Before\":null,\"After\":\"123\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":null,\"After\":\"test\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})

	t.Run("Audit CreateEvent Struct Entity ID with default value Not Matched", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_CreateEvent_Entity_ID_Default_Value_Not_Matched", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		var testEntity = TestEntity5{
			Id:     123,
			Field2: "test",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "", testEntity)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Id\",\"Before\":null,\"After\":\"123\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":null,\"After\":\"test\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})

	t.Run("Audit CreateEvent Struct Entity ID with default value Not Matched Override With Conf", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		entitiesConfig := make(map[string]EntityConfig)
		entitiesConfig["TestEntity"] = EntityConfig{IdField: "Id"}

		config := Config{GlobalSliceChangesFormat: Full, EntitiesConfig: entitiesConfig, AuditName: "Test_AuditEvents_CreateEvent_Entity_ID_Default_Value_Not_Matched_Override_Conf", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		var testEntity = TestEntity5{
			Id:     123,
			Field2: "test",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.CreateEvent(ctx, "Record was created", "TestEntity", testEntity)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityType\":\"TestEntity\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/Id\",\"Before\":null,\"After\":\"123\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was created\",\"Subtype\":\"field-created\",\"EntityType\":\"TestEntity\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":null,\"After\":\"test\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})

	t.Run("Audit UpdateEvent Struct Entity ID with default value", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_UpdateEvent_Entity_ID_Default_Value", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		var testEntity = TestEntity6{
			Id:     123,
			Field2: "test",
		}

		var testEntityUpdated = TestEntity6{
			Id:     123,
			Field2: "test2",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", testEntity, testEntityUpdated)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":\"test\",\"After\":\"test2\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit DeleteEvent Struct Entity ID with default value", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{GlobalSliceChangesFormat: Full, AuditName: "Test_AuditEvents_DeleteEvent_Entity_ID_Default_Value", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		var testEntity = TestEntity6{
			Id:     123,
			Field2: "test",
		}

		_, _, lineNumber, _ := runtime.Caller(0)
		a.DeleteEvent(ctx, "Record was deleted", "", testEntity)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/id\",\"Before\":\"123\",\"After\":null,\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was deleted\",\"Subtype\":\"field-deleted\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/Field2\",\"Before\":\"test\",\"After\":null,\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})

	t.Run("Audit UpdateEvent Struct Nested JSON Tag", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)
		config := Config{AuditName: "Test_AuditEvents_UpdateEvent_Nested_JSON_Tag", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		person := Person{
			Id: "123",
			Contact: PersonContact{
				PhoneNumber: "123-456-789",
			},
		}

		personUpdated := person
		personUpdated.Contact.PhoneNumber = "123-654-789"

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "Person", person, personUpdated)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityType\":\"Person\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/contact/phoneNumber\",\"Before\":\"123-456-789\",\"After\":\"123-654-789\",\"Type\":\"string\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct Nested JSON Tag With Config", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		fieldsConifg := make(map[string]FieldConfig)
		fieldsConifg["/contact/phoneNumber"] = FieldConfig{TypeFormat: TypeFormat{fieldType: Number, fieldFormat: Double}}

		entitiesConfig := make(map[string]EntityConfig)
		entitiesConfig["Person"] = EntityConfig{FieldsConfig: fieldsConifg}

		config := Config{EntitiesConfig: entitiesConfig, AuditName: "Test_AuditEvents_UpdateEvent_Nested_JSON_Tag_With_Config", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		person := Person{
			Id: "123",
			Contact: PersonContact{
				PhoneNumber: "123-456-789",
			},
		}

		personUpdated := person
		personUpdated.Contact.PhoneNumber = "123-654-789"

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "Person", person, personUpdated)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityType\":\"Person\",\"EntityId\":\"123\",\"Change\":{\"Id\":101,\"Path\":\"/contact/phoneNumber\",\"Before\":\"123-456-789\",\"After\":\"123-654-789\",\"Type\":\"number\",\"Format\":\"double\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct Entity Config Overrides Global Config", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		fieldsConifg := make(map[string]FieldConfig)
		fieldsConifg["Field2J"] = FieldConfig{TypeFormat: TypeFormat{fieldType: Integer, fieldFormat: Int32}}

		entitiesConfig := make(map[string]EntityConfig)
		entitiesConfig["TestEntity"] = EntityConfig{FieldsConfig: fieldsConifg, IdField: "/Field2", SliceChangesFormat: Unordered}

		globalFieldsConifg := make(map[string]FieldConfig)
		globalFieldsConifg["Field2J"] = FieldConfig{TypeFormat: TypeFormat{fieldType: Integer, fieldFormat: Int64}}

		config := Config{GlobalSliceChangesFormat: Ordered, GlobalFieldsConfig: globalFieldsConifg, GlobalIdField: "Field1", EntitiesConfig: entitiesConfig, AuditName: "Test_AuditEvents_UpdateEvent_Struct_Entity_Config_Overrides_Global_Config", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "TestEntity", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityType\":\"TestEntity\",\"EntityId\":\"10\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\",\"Format\":\"int32\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityType\":\"TestEntity\",\"EntityId\":\"10\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"EntityType\":\"TestEntity\",\"EntityId\":\"10\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 3, t)
	})

	t.Run("Audit UpdateEvent Struct Global Type Config Ignore Field Type Struct", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		// this will drop the commented lines from the want array
		typesConfig := make(map[string]FieldConfig)
		typesConfig["audit.TestSubEntity"] = FieldConfig{Ignore: true}

		config := Config{GlobalSliceChangesFormat: Ordered, GlobalTypeConfig: typesConfig, AuditName: "Test_AuditEvents_UpdateEvent_Global_Type_Config_Ignore_Field_Type_Struct", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"3\",\"Type\":\"integer\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 1, t)
	})

	t.Run("Audit UpdateEvent Struct Global Type Config Ignore Field Type Array", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		// this will drop the commented lines from the want array
		typesConfig := make(map[string]FieldConfig)
		typesConfig["[]int"] = FieldConfig{Ignore: true}

		config := Config{GlobalSliceChangesFormat: Ordered, GlobalTypeConfig: typesConfig, AuditName: "Test_AuditEvents_UpdateEvent_Global_Type_Config_Ignore_Field_Type_Array", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"3\",\"Type\":\"integer\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 2, t)
	})
	t.Run("Audit UpdateEvent Struct Global Type Config Ignore Field Type Array Slice Format", func(t *testing.T) {
		b.Reset()
		ctxDataMap, _ := contextutil.NewContextData("trans123", "part123", "usr123")
		ctx := contextutil.WithValue(httpConttext, ctxDataMap)

		// this will change the commented lines from the want array to one line which is the full array
		typesConfig := make(map[string]FieldConfig)
		typesConfig["[]int"] = FieldConfig{SliceChangesFormat: Full}

		config := Config{GlobalSliceChangesFormat: Ordered, GlobalTypeConfig: typesConfig, AuditName: "Test_AuditEvents_UpdateEvent_Global_Type_Config_Ignore_Field_Type_Array_Slice_Format", LoggerConfig: &logger.Config{Destination: logger.MEMORY}}
		a, _ := NewAuditLogger(config)

		_, _, lineNumber, _ := runtime.Caller(0)
		a.UpdateEvent(ctx, "Record was updated", "", entity1, entity3)

		want := []string{
			strings.Replace(wantResponse[0], ":400", fmt.Sprintf("%s%d", ":", lineNumber+1), 1),
			wantResponse[1],
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field2J\",\"Before\":\"10\",\"After\":\"30\",\"Type\":\"integer\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field4\",\"Before\":\"Test\",\"After\":\"New\",\"Type\":\"string\"}}",
			"\"Type\":\"AUDIT\",\"Id\":\"" + eventID() + "\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5\",\"Before\":\"[1,2,3]\",\"After\":\"[1,3]\",\"Type\":\"array\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/1\",\"Before\":\"2\",\"After\":\"3\",\"Type\":\"integer\"}}",
			// "\"Type\":\"AUDIT\",\"Id\":\""+eventID()+"\",\"Audit\":{\"SchemaVersion\":\"" + schemaVersion + "\",\"Type\":\"Record was updated\",\"Subtype\":\"field-updated\",\"Change\":{\"Id\":101,\"Path\":\"/Field3/Field5/2\",\"Before\":\"3\",\"After\":\"\",\"Type\":\"integer\"}}",
		}
		output := sink.String()
		validate(output, want, ",\"Change\":{", 3, t)
	})
}

func mapToString(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

type TestEntity struct {
	Field1 string
	Field2 int `json:"Field2J,omitempty"`
	Field3 TestSubEntity
}

type TestEntity2 struct {
	Field6 string
	Field7 int
	Field2 int
	Field1 string
}

type TestEntity3 struct {
	Field1 []string
	Field2 []int
}

type TestEntity4 struct {
	Field1 string
	Field2 time.Time
}

type TestSubEntity struct {
	Field4 string
	Field5 []int
}

type TestEntity5 struct {
	Id     int
	Field2 string
}

type TestEntity6 struct {
	Id     int `json:"id,omitempty"`
	Field2 string
}

type TestEntityUUID struct {
	Id   gocql.UUID `json:"id,omitempty"`
	Name string
}

type TestEntityUUID2 struct {
	Id   []byte `json:"id,omitempty"`
	Name string
}

type Person struct {
	Id      string        `json:"id,omitempty"`
	Contact PersonContact `json:"contact,omitempty"`
}

type PersonContact struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type TestEntityOM struct {
	F1 string
	F2 int `json:"Field2J,omitempty"`
}

// MarshalLogObject Marshal Resource to zap Object
func (e TestEntityOM) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e.F1 != "" {
		enc.AddString("f1", e.F1)
	}
	enc.AddInt("f2", e.F2)
	return nil
}

func replaceSpace(s string) string {
	var result []rune
	const badSpace = '\u0020'
	for _, r := range s {
		if r == badSpace {
			result = append(result, '\u00A0')
			continue
		}
		result = append(result, r)
	}
	return string(result)
}
func auditEventS(audit *AuditLogger, ctx context.Context, eventType string, message string, description string, calldepth int) (string, error) {
	return audit.With(CallDepth(calldepth)).EventS(ctx, eventType, message, description)
}
