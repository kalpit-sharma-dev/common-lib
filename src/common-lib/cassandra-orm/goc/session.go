// Package goc was inspired and actually uses pieces of code from package gockle "github.com/willfaught/gockle" under the
// The MIT License (MIT) Copyright (c) 2016
// Copyright (c) 2017 ContinuumLLC
// Package gockle was not used as the dependency as it doesn't provide functionality
// satisfying the needs of the current project.
// Package is the wrapper around gocql allowing mocking gocql behavior
package goc

import (
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

//go:generate mockgen -package mock -destination=mock/session_mock.go gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/goc Session,SessionStatus

// NewSimpleSession returns a new Session for hosts. It uses native protocol
// version 4.
var NewSimpleSession = newSimpleSession

// NewSimpleSessionStatus returns new Session with Status for health-checks
var NewSimpleSessionStatus = newSimpleSessionStatus

type (
	// Session is a Cassandra connection. The Query methods run CQL queries. The
	// Columns and Tables methods provide simple metadata.
	Session interface {
		// Close closes the Session.
		Close()

		// Closed checks the Session is closed.
		Closed() bool

		// Exec executes the query for statement and arguments.
		Exec(statement string, arguments ...interface{}) error

		// ExecuteBatch executes a batch operation and returns nil if successful
		// otherwise an error is returned describing the failure.
		ExecuteBatch(*gocql.Batch) error

		// NewBatch creates new Session.Batch (gocql.NewBatch is deprecated)
		NewBatch(batchType gocql.BatchType) *gocql.Batch

		// Query generates a new query object for interacting with the database.
		// Further details of the query may be tweaked using the resulting query
		// value before the query is executed. Query is automatically prepared
		// if it has not previously been executed.
		Query(stmt string, values ...interface{}) *gocql.Query

		// Select executes the select query for statement and arguments.
		Select(query string, values ...interface{}) ([]map[string]interface{}, error)

		// SetConsistency sets the default consistency level for this session. This
		// setting can also be changed on a per-query basis and the default value
		// is Quorum.
		SetConsistency(cons gocql.Consistency)
	}

	SessionStatus interface {
		Session

		// Status allows Session to implement rest.Statuser for connection monitoring
		Status(status rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus
	}

	sessionStatus struct {
		Session
		keyspace string
		hosts    []string
	}

	session struct {
		s *gocql.Session
	}
)

// NewSession returns a new Session for s.
func NewSession(s *gocql.Session) Session {
	return &session{s: s}
}

func newSimpleSessionStatus(keyspace string, hosts []string, timeout time.Duration) (SessionStatus, error) {
	sessionI, err := NewSimpleSession(keyspace, hosts, timeout)
	if err != nil {
		return nil, err
	}

	return &sessionStatus{
		Session:  sessionI,
		keyspace: keyspace,
		hosts:    hosts,
	}, nil
}

func newSimpleSession(keyspace string, hosts []string, timeout time.Duration) (Session, error) {
	var c = gocql.NewCluster(hosts...)
	c.Keyspace = keyspace
	c.ProtoVersion = 4
	c.Timeout = timeout
	//c.ConnectTimeout = timeout // this require newer version of gocql lib

	var s, err = c.CreateSession()
	if err != nil {
		return nil, err
	}
	return &session{s: s}, nil
}

func (s *sessionStatus) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.Name = "Cassandra session"
	conn.ConnectionType = "TCP"
	conn.Type = s.keyspace
	conn.ConnectionURLs = s.hosts

	// session is not initialized
	if s.Session == nil {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
		return &conn
	}

	if s.Session.Closed() {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
	} else {
		conn.ConnectionStatus = rest.ConnectionStatusActive
	}

	return &conn
}

func (s *session) Exec(statement string, arguments ...interface{}) error {
	return s.s.Query(statement, arguments...).Exec()
}

func (s *session) NewBatch(batchType gocql.BatchType) *gocql.Batch {
	return s.s.NewBatch(batchType)
}

func (s *session) ExecuteBatch(batch *gocql.Batch) error {
	return s.s.ExecuteBatch(batch)
}

func (s *session) Close() {
	s.s.Close()
}

func (s *session) Closed() bool {
	return s.s.Closed()
}

func (s *session) Query(stmt string, values ...interface{}) *gocql.Query {
	return s.s.Query(stmt, values...)
}

func (s *session) Select(query string, values ...interface{}) ([]map[string]interface{}, error) {
	return s.s.Query(query, values...).Iter().SliceMap()
}

func (s *session) SetConsistency(cons gocql.Consistency) {
	s.s.SetConsistency(cons)
}
