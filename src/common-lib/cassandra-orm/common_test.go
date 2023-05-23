package cassandraorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetQueryKeys(t *testing.T) {
	type testStruct struct {
		ValString string  `json:"valString"`
		ValInt    int64   `json:"valInt"`
		ValFloat  float64 `json:"valFloat"`
		ValBool   bool    `json:"valBool"`
	}

	var keyCols = []string{"ValString", "ValInt", "ValFloat", "ValBool"}

	keys, err := GetQueryKeys(nil, keyCols)
	assert.Error(t, err)

	keys, err = GetQueryKeys(testStruct{}, keyCols)
	assert.NoError(t, err)

	errKeyCols := append(keyCols, "errCol")
	keys, err = GetQueryKeys(testStruct{}, errKeyCols)
	assert.Len(t, keys, 0)

	keys, err = GetQueryKeys(testStruct{
		ValString: "string", ValInt: 7, ValFloat: 3.142, ValBool: true,
	}, keyCols)
	assert.Len(t, keys, 4)
	assert.Equal(t, "string", keys[0].(string))
	assert.Exactly(t, int64(7), keys[1].(int64))
	assert.Equal(t, 3.142, keys[2].(float64))
	assert.Equal(t, true, keys[3].(bool))

	keys, err = GetQueryKeys(testStruct{
		ValString: "string", ValInt: 7, ValFloat: 0.0, ValBool: false,
	}, keyCols)
	assert.Len(t, keys, 2)
	assert.Equal(t, "string", keys[0].(string))
	assert.Exactly(t, int64(7), keys[1].(int64))

	keys, err = GetQueryKeys(testStruct{
		ValString: "string", ValInt: 7, ValFloat: 0.0, ValBool: true,
	}, keyCols)
	assert.Len(t, keys, 4)
	assert.Equal(t, "string", keys[0].(string))
	assert.Exactly(t, int64(7), keys[1].(int64))
	assert.Empty(t, keys[2].(float64))
	assert.Equal(t, true, keys[3].(bool))
}
