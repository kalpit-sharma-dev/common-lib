package cassandraorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRowsEqualByKeys(t *testing.T) {
	tests := []struct {
		name  string
		want  bool
		data1 map[string]interface{}
		data2 map[string]interface{}
		keys  []string
	}{
		{
			name:  "Correct",
			want:  true,
			data1: map[string]interface{}{"string": "string", "int": 1, "float": 1.0},
			data2: map[string]interface{}{"string": "string", "int": 1, "float": 1.0},
			keys:  []string{"string", "int", "float"},
		},
		{
			name:  "Incorrect",
			want:  false,
			data1: map[string]interface{}{"string": "string", "int": 1, "float": 1.0},
			data2: map[string]interface{}{"errStr": "string", "int": 1, "float": 1.0},
			keys:  []string{"string", "int", "float"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := isRowsEqualByKeys(tt.data1, tt.data2, tt.keys)
			assert.Equal(t, tt.want, res)
		})
	}
}
