package cassandra

import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"

// Health - Cassandra Health Implementation
func Health(cfg *DbConfig) rest.Statuser {
	return status{
		conf: cfg,
	}
}

//Status used for getting status of Casandra connection
type status struct {
	conf *DbConfig
}

//Status used for getting status of Casandra connection
func (c status) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = "Cassandra"
	conn.ConnectionURLs = c.conf.Hosts
	conn.ConnectionStatus = rest.ConnectionStatusActive

	session, err := NewDbConnection(c.conf)

	if err != nil {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
		return &conn
	}
	defer session.Close()
	return &conn
}
