package publisher

import (
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

// Health returns a Health state for Kafka
func Health(producerType ProducerType, cfg *Config) rest.Statuser {
	return status{producerType: producerType, cfg: cfg}
}

type status struct {
	producerType ProducerType
	cfg          *Config
}

func (k status) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = fmt.Sprintf("Kafka-%s-Producer", k.producerType)
	conn.ConnectionURLs = k.cfg.Address
	conn.ConnectionStatus = rest.ConnectionStatusActive

	state := circuit.CurrentState(k.cfg.commandName())
	if state != circuit.Close && state != circuit.NA {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
	}

	return &conn
}
