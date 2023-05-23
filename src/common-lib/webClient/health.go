package webClient

import (
	"errors"
	"net/url"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

// Health  returns the health of all the hosts using circuit breaker state
// []CircuitBreakerConfig should be same as config passed to RegisterCircuitBreaker function
func Health(cbList []CircuitBreakerConfig) ([]rest.Statuser, error) {
	var (
		statusList []rest.Statuser
		errURLs    []string
	)

	for _, cbConfig := range cbList {

		u, err := url.Parse(cbConfig.BaseURL)
		if err != nil {
			errURLs = append(errURLs, cbConfig.BaseURL)
			continue
		}

		st := status{
			u.Host,
			cbConfig.BaseURL,
		}
		statusList = append(statusList, st)
	}

	if len(errURLs) > 0 {
		return statusList, errors.New("Failed to parse URLs: " + strings.Join(errURLs, ","))
	}

	return statusList, nil
}

type status struct {
	commandName string
	baseURL     string
}

func (k status) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = k.baseURL
	conn.ConnectionURLs = []string{k.baseURL}
	conn.ConnectionStatus = rest.ConnectionStatusActive

	state := circuit.CurrentState(k.commandName)
	if state != circuit.Close {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
	}

	return &conn
}
