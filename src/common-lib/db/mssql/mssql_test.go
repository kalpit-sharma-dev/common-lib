package mssql

import (
	"errors"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"
)

func TestGetConnectionString(t *testing.T) {
	m := mssql{}
	t.Run("Error missing config", func(t *testing.T) {
		_, err := m.GetConnectionString(db.Config{})
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		conStr, err := m.GetConnectionString(db.Config{DbName: "NOCBO",
			Server:     "10.2.27.41",
			Password:   "its",
			UserID:     "its",
			CacheLimit: 200})
		if err != nil {
			t.Errorf("Expecting nil but found err := %v", err)
		}

		if conStr == "" {
			t.Errorf("Expecting connection string found empty")
		}
	})
}

func TestValidCbError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid cb error",
			args: args{
				err: errors.New("connection was refused"),
			},
			wantErr: true,
		},
		{
			name: "Invalid cb error",
			args: args{
				err: errors.New("duplicate column"),
			},
			wantErr: false,
		},
		{
			name: "No error",
			args: args{
				err: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mssql{}
			if err := m.ValidCbError(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("mssql.ValidCbError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
