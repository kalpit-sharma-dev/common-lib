package cassandra

type fullConnection struct {
	*connection
	*batchConnection
}

// NewFullDbConnection returns the struct implementation of NewFullDbConnection with single session
func NewFullDbConnection(conf *DbConfig) (FullDbConnector, error) {
	db, err := newConnection(conf)
	if err != nil {
		return nil, err
	}

	batchDb := &batchConnection{}
	batchDb.Connection = *db
	return &fullConnection{
		connection:      db,
		batchConnection: batchDb,
	}, err
}

// Close function closes the connection and does not return error
func (f fullConnection) Close() {
	if f.session != nil {
		f.session.Close()
	}
}

// Closed function to check is session is closed or not
func (f fullConnection) Closed() bool {
	if f.session != nil {
		return f.session.Closed()
	}
	return true
}
