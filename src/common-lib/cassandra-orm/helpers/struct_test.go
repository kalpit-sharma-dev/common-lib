package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	customType = customTypeStruct{
		Int:    1,
		String: "aaa",
		Slice:  []int{1, 2, 3, 4, 5},
	}

	testStructFull = testStruct{
		Int:        1,
		String:     "aaa",
		Slice:      []int{1, 2, 3, 4, 5},
		CustomType: customType,
		CustomTypeSlice: []customTypeStruct{
			customType,
			customType,
		},
	}

	testStructNoCustomTypeSlice = testStruct{
		Int:        1,
		String:     "aaa",
		Slice:      []int{1, 2, 3, 4, 5},
		CustomType: customType,
	}
)

type (
	testStruct struct {
		Int             int
		String          string
		Slice           []int
		CustomType      customTypeStruct
		CustomTypeSlice []customTypeStruct
	}

	customTypeStruct struct {
		Int    int
		String string
		Slice  []int
	}
)

func TestEqual(t *testing.T) {
	testCases := []struct {
		a      testStruct
		b      testStruct
		fields []string
		result bool
		err    error
	}{
		{
			a:      testStructFull,
			b:      testStructFull,
			fields: []string{},
			result: true,
			err:    nil,
		},
		{
			a:      testStructFull,
			b:      testStructNoCustomTypeSlice,
			fields: []string{"String", "Slice", "CustomType"},
			result: true,
			err:    nil,
		},
		{
			a:      testStructFull,
			b:      testStructNoCustomTypeSlice,
			fields: []string{"String", "Slice", "CustomTypeSlice"},
			result: false,
			err:    nil,
		},
		{
			a:      testStructFull,
			b:      testStructNoCustomTypeSlice,
			fields: []string{},
			result: false,
			err:    nil,
		},
	}
	for _, c := range testCases {
		result, err := Equal(c.a, c.b, c.fields...)
		if result != c.result || err != c.err {
			t.Errorf("Case not valid: %+v", c)
		}
	}
	_, err := Equal("2", "3")
	assert.EqualError(t, err, ErrParamIsNotStruct.Error())
}

func TestSerializeDefaultPass(t *testing.T) {
	m, err := SerializeDefault(testStructFull)
	require.NoError(t, err)

	require.Len(t, m, 5, "Expected %d items; got: %d", 5, len(m))
	checkValue(t, m, "Int", 1)
	checkValue(t, m, "String", "aaa")
	checkValue(t, m, "Slice", []int{1, 2, 3, 4, 5})
	checkValue(t, m, "CustomType", customType)
	checkValue(t, m, "CustomTypeSlice", []customTypeStruct{customType, customType})
}

func TestSerializeDefaultFail(t *testing.T) {
	_, err := SerializeDefault(nil)
	assert.EqualError(t, err, ErrSerialization.Error())
}

func checkValue(t *testing.T, m map[string]interface{}, key string, expected interface{}) {
	v, ok := m[key]
	require.True(t, ok, "Value with key %q missed", key)
	assert.Equal(t, expected, v, "Expected %v; got: %v", expected, v)
}
