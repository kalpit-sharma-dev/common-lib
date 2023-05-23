package postgresql

import (
	"errors"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"
)

func TestGetConnectionString(t *testing.T) {
	p := postgresql{}
	t.Run("Error missing config", func(t *testing.T) {
		configString, err := p.GetConnectionString(db.Config{})
		if configString != "" {
			t.Errorf("Expecting blank connection string but got %s", configString)
		}
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("SuccessWithDefaultPortAndSSLMode", func(t *testing.T) {
		config := db.Config{DbName: "MockServiceDB",
			Server:     "1.2.3.4",
			Password:   "password",
			UserID:     "user",
			CacheLimit: 200}
		sentinelConfigString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Server, defaultPort, config.UserID, config.Password, config.DbName, defaultSSLMode)
		configString, err := p.GetConnectionString(config)
		if err != nil {
			t.Errorf("Expecting nil but found err : %v", err)
		}

		if configString == "" {
			t.Errorf("Expecting valid connection string but found empty")
		}

		if configString != sentinelConfigString {
			t.Errorf("Expecting config string value : %s, but found : %s", sentinelConfigString, configString)
		}
	})

	t.Run("SuccessWithDefaultPort", func(t *testing.T) {
		sslMode := "secure"
		additionalParam := make(map[string]string)
		additionalParam[SSLModeKey] = sslMode
		config := db.Config{DbName: "MockServiceDB",
			Server:           "1.2.3.5",
			Password:         "password",
			UserID:           "user",
			AdditionalConfig: additionalParam,
			CacheLimit:       200}
		sentinelConfigString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Server, defaultPort, config.UserID, config.Password, config.DbName, sslMode)
		configString, err := p.GetConnectionString(config)
		if err != nil {
			t.Errorf("Expecting nil but found err : %v", err)
		}

		if configString == "" {
			t.Errorf("Expecting valid connection string but found empty")
		}

		if configString != sentinelConfigString {
			t.Errorf("Expecting config string value : %s, but found : %s", sentinelConfigString, configString)
		}
	})

	t.Run("SuccessWithDefaultSSLMode", func(t *testing.T) {
		port := "1237"
		additionalParam := make(map[string]string)
		additionalParam[ServerPortKey] = port
		config := db.Config{DbName: "MockServiceDB",
			Server:           "1.2.3.6",
			Password:         "password",
			UserID:           "user",
			AdditionalConfig: additionalParam,
			CacheLimit:       200}
		sentinelConfigString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Server, port, config.UserID, config.Password, config.DbName, defaultSSLMode)
		configString, err := p.GetConnectionString(config)
		if err != nil {
			t.Errorf("Expecting nil but found err : %v", err)
		}

		if configString == "" {
			t.Errorf("Expecting valid connection string but found empty")
		}

		if configString != sentinelConfigString {
			t.Errorf("Expecting config string value : %s, but found : %s", sentinelConfigString, configString)
		}
	})

	t.Run("SuccessWithInvalidDatatypeForPort", func(t *testing.T) {
		port := "non-int-value"
		additionalParam := make(map[string]string)
		additionalParam[ServerPortKey] = port
		config := db.Config{DbName: "MockServiceDB",
			Server:           "1.2.3.6",
			Password:         "password",
			UserID:           "user",
			AdditionalConfig: additionalParam,
			CacheLimit:       200}
		sentinelConfigString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Server, defaultPort, config.UserID, config.Password, config.DbName, defaultSSLMode)
		configString, err := p.GetConnectionString(config)
		if err != nil {
			t.Errorf("Expecting nil but found err : %v", err)
		}

		if configString == "" {
			t.Errorf("Expecting valid connection string but found empty")
		}

		if configString != sentinelConfigString {
			t.Errorf("Expecting config string value : %s, but found : %s", sentinelConfigString, configString)
		}
	})

	t.Run("Success", func(t *testing.T) {
		port := "1237"
		sslMode := "secure"
		additionalParam := make(map[string]string)
		additionalParam[ServerPortKey] = port
		additionalParam[SSLModeKey] = sslMode
		config := db.Config{DbName: "MockServiceDB",
			Server:           "1.2.3.7",
			Password:         "password",
			UserID:           "user",
			AdditionalConfig: additionalParam,
			CacheLimit:       200}
		sentinelConfigString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Server, port, config.UserID, config.Password, config.DbName, sslMode)
		configString, err := p.GetConnectionString(config)
		if err != nil {
			t.Errorf("Expecting nil but found err : %v", err)
		}

		if configString == "" {
			t.Errorf("Expecting valid connection string but found empty")
		}

		if configString != sentinelConfigString {
			t.Errorf("Expecting config string value : %s, but found : %s", sentinelConfigString, configString)
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
				err: errors.New("connection refused"),
			},
			wantErr: true,
		},
		{
			name: "Invalid cb error",
			args: args{
				err: errors.New("No command text was set"),
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
			p := postgresql{}
			if err := p.ValidCbError(tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("postgresql.ValidCbError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
