package cassandra

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

const commandName = "Database-Command"

// DbConfig - configuration which is required to connect to Cassandra db.
type DbConfig struct {
	// Hosts - Addresses for the initial connections. It is recommended to use the value set in
	// the Cassandra config for broadcast_address or listen_address, an IP address not
	// a domain name. This is because events from Cassandra will use the configured IP
	// address, which is used to index connected hosts. If the domain name specified
	// resolves to more than 1 IP address then the driver may connect multiple times to
	// the same host, and will not mark the node being down or up from events.
	// This is a mandatory field
	Hosts []string

	// Keyspace - initial keyspace
	// This is a mandatory field
	Keyspace string

	// TimeoutMillisecond - connection timeout
	// (default is 1 second)
	TimeoutMillisecond time.Duration

	// initial connection timeout, used during initial dial to server
	// (default: 600ms)
	ConnectTimeout time.Duration

	// NumConn - number of connections per host
	// default - 20
	NumConn int

	// CircuitBreaker - Configuration for the Circuit breaker
	// default - circuit.New()
	CircuitBreaker *circuit.Config

	// CommandName - Name for Database command
	// defaults to - Database-Command
	CommandName string

	// ValidErrors - List of error to participates in the Circuit state calculation
	// Default values are -
	ValidErrors []string

	// Authenticator
	// authenticator (default: nil)
	Authenticator gocql.Authenticator

	// SslOptions configures TLS use
	SslOpts *gocql.SslOptions

	// Consistency sets the consistency requirements for the session. Defaults to Quorum.
	Consistency gocql.Consistency

	// If DisableInitialHostLookup then the driver will not attempt to get host info
	// from the system.peers table, this will mean that the driver will connect to
	// hosts supplied and will not attempt to lookup the hosts information, this will
	// mean that data_centre, rack and token information will not be available and as
	// such host filtering and token aware query routing will not be available.
	// default - false
	DisableInitialHostLookup bool

	// ProtoVersion sets the version of the native protocol to use, this will
	// enable features in the driver for specific protocol versions, generally this
	// should be set to a known version (2,3,4) for the cluster being connected to.
	// default - 4
	ProtoVersion int
}

// NewConfig - returns a configration object having default values
func NewConfig() *DbConfig {
	return &DbConfig{
		NumConn:            20,
		TimeoutMillisecond: time.Second,
		ConnectTimeout:     time.Millisecond * 600,
		CircuitBreaker: &circuit.Config{
			Enabled: false, TimeoutInSecond: 5, MaxConcurrentRequests: 2500,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 300, SleepWindowInSecond: 10,
		},
		Consistency: gocql.Quorum,
		CommandName: commandName,
		ValidErrors: []string{},
	}
}

func validate(conf *DbConfig) error {
	if conf == nil || len(conf.Hosts) == 0 || conf.Keyspace == "" {
		return errors.New(ErrDbHostsAndKeyspaceRequired)
	}

	if conf.Consistency == 0 {
		conf.Consistency = gocql.Quorum
	}

	if conf.NumConn == 0 {
		conf.NumConn = 20
	}

	if conf.TimeoutMillisecond == 0 {
		conf.TimeoutMillisecond = 1 * time.Second
	}

	if conf.ConnectTimeout == 0 {
		conf.ConnectTimeout = 600 * time.Millisecond
	}

	// We are adding this additional check to avoid any failure due to
	// worng timeout configuration by microservice team
	if conf.TimeoutMillisecond < time.Millisecond {
		conf.TimeoutMillisecond = conf.TimeoutMillisecond * time.Millisecond
	}

	if conf.ConnectTimeout < time.Millisecond {
		conf.ConnectTimeout = conf.ConnectTimeout * time.Millisecond
	}

	if conf.CircuitBreaker == nil {
		conf.CircuitBreaker = &circuit.Config{
			Enabled: false, TimeoutInSecond: 5, MaxConcurrentRequests: 2500,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 300, SleepWindowInSecond: 10,
		}
	}

	if conf.CommandName == "" {
		conf.CommandName = commandName
	}

	if len(conf.ValidErrors) == 0 {
		conf.ValidErrors = []string{}
	}
	circuit.Register("Circuit-Breaker", conf.CommandName, conf.CircuitBreaker, nil)
	return nil
}
