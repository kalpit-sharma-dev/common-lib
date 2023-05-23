package postgresql

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"

	// Postgres driver
	_ "github.com/lib/pq"
)

const (
	//Dialect is a database name used for registration
	Dialect = "postgres"
	//defaultPort is the default port of PostgreSQL server
	defaultPort int64 = 5432
	//defaultSSLMode is the default SSL Mode to be used for connecting to PostgreSQL server
	defaultSSLMode string = "disable"
	//ServerPortKey is the key used to pass server port, in Additional Config map
	ServerPortKey = "port"
	//SSLModeKey is the key used to pass SSL mode while communicating with server, in Additional Config map
	SSLModeKey = "sslmode"
	// cbError string gets appended at the beginning of a valid circuit breaker error.
	cbError string = "Valid circuit breaker error"
)

// postgresqlErrors contains few Postgresql connection exception conditions.
//
//nolint:gofumpt
var postgresqlErrors = []string{"connection_exception", "connection_does_not_exist",
	"connection_failure", "bad connection", "connection refused",
	"sqlclient_unable_to_establish_sqlconnection",
	"sqlserver_rejected_establishment_of_sqlconnection", "broken pipe"}

func init() {
	db.RegisterDialect(Dialect, postgresql{})
}

type postgresql struct {
}

func (postgresql) GetConnectionString(config db.Config) (string, error) {
	if config.Server == "" || config.DbName == "" || config.UserID == "" || config.Password == "" {
		return "", fmt.Errorf("getDbConnInfo: One or more required db configuration  missing")
	}

	port, _ := strconv.ParseInt(config.AdditionalConfig[ServerPortKey], 0, 0)
	if port == 0 {
		port = defaultPort
	}

	sslMode, ok := config.AdditionalConfig[SSLModeKey]
	if !ok {
		sslMode = defaultSSLMode
	}
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Server, port, config.UserID, config.Password, config.DbName, sslMode)
	return connString, nil
}

func (postgresql) ValidCbError(err error) error {
	if err == nil {
		return nil
	}

	for _, v := range postgresqlErrors {
		if strings.Contains(err.Error(), v) {
			//nolint:goerr113
			//nolint:errorlint
			return fmt.Errorf("%s : %v", cbError, err)
		}
	}

	return nil
}
