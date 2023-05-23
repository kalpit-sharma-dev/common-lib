// Package cassandra-orm is for working with a cassandra DB. It originally comes from Platform-Infrastructure-lib

package cassandraorm

import (
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/goc"
)

// absentTransactionID is using in case if transactionID is absent
const absentTransactionID = ""

var (
	// Session goc
	Session goc.Session

	keyspace       string
	defaultTimeout = 600 * time.Millisecond

	gocNewSimpleSession = goc.NewSimpleSession
)

// Load inits session
func Load(hosts []string, keyspaceName, timeoutValue string, logImpl logger.Log) (err error) {
	keyspace = keyspaceName

	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil || timeout == 0 {
		logImpl.Info(absentTransactionID, "config.Config.CassandraTimeout not valid. Default value (%s) will be used", defaultTimeout)
		timeout = defaultTimeout
	}
	s, err := gocNewSimpleSession(keyspace, hosts, timeout)
	if err == nil {
		Session = s
		Session.SetConsistency(gocql.LocalQuorum)
	}

	return err
}
