package db

//dialect interface contains behaviors that differ across SQL database
type dialect interface {
	//GetConnectionString is used to get connection string for database
	GetConnectionString(config Config) (string, error)
	// ValidCbError is used to validate whether a db error would qualify for a circuit breaker.
	ValidCbError(err error) error
}

var dialectsMap = map[string]dialect{}

// getDialect gets the dialect for the specified dialect name
func getDialect(name string) (d dialect, ok bool) {
	d, ok = dialectsMap[name]
	return
}

// RegisterDialect - registers a new dialect and loads driver for registered dialect.
func RegisterDialect(name string, d dialect) {
	dialectsMap[name] = d
}
