package cassandraorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBase(t *testing.T) {
	tests := []struct {
		item       Model
		tableName  string
		tableKeys  []string
		viewTables map[string][]string
		wantErr    bool
	}{
		{
			item: nil, tableName: "", tableKeys: nil, viewTables: nil, wantErr: true,
		},
		{
			item: nil, tableName: "test", tableKeys: nil, viewTables: nil, wantErr: true,
		},
		{
			item: nil, tableName: "test", tableKeys: []string{"key1", "key2"}, viewTables: nil, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run("Base", func(t *testing.T) {
			got := NewBase(tt.item, tt.tableName, tt.tableKeys, tt.viewTables)
			assert.Equal(t, tt.tableName, got.Table())
			assert.ElementsMatch(t, tt.tableKeys, got.Keys())
			if got.GetColumns() == nil && tt.wantErr != true {
				t.Errorf("expected: %v, got: %v", tt.tableKeys, got.GetColumns())
			}
			assert.Equal(t, `"test"`, got.Quote("test"))
		})
	}
}
