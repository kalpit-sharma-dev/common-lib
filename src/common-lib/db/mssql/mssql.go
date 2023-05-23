package mssql

import (
	"fmt"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"

	_ "github.com/denisenkom/go-mssqldb" //To load mssql driver
)

const (

	//Dialect is a database name used for registration
	Dialect = "mssql"
	// cbError string gets appended at the beginning of a valid circuit breaker error.
	cbError string = "Valid circuit breaker error"
)

// mssqlErrors contains few mssql connection exception conditions.
//
//nolint:gofumpt
var mssqlErrors = []string{"connection_exception", "connection_does_not_exist",
	"connection_failure", "bad connection", "connection was refused",
	"error has occurred while establishing a connection"}

func init() {
	db.RegisterDialect(Dialect, mssql{})
}

type mssql struct {
}

func (m mssql) GetConnectionString(config db.Config) (string, error) {
	if config.DbName == "" || config.Password == "" || config.Server == "" || config.UserID == "" {
		return "", fmt.Errorf("getDbConnInfo: One or more required db configuration  missing")
	}
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", config.Server, config.UserID, config.Password, config.DbName)
	return connString, nil
}

func (mssql) ValidCbError(err error) error {
	if err == nil {
		return nil
	}

	for _, v := range mssqlErrors {
		if strings.Contains(err.Error(), v) {
			//nolint:goerr113
			//nolint:errorlint
			return fmt.Errorf("%s : %v", cbError, err)
		}
	}

	return nil
}
