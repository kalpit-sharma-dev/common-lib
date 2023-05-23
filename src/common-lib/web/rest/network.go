package rest

import (
	"errors"
	"net"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
)

// GetNetworkInterfaces returns list of all tcp interfaces of application
func GetNetworkInterfaces(url string) ([]string, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return nil, err
	}

	// If host specified directly, no reason to continue
	if host != "" {
		return []string{host + ":" + port}, nil
	}

	ipSlice := util.LocalIPAddress()
	if len(ipSlice) == 0 {
		return nil, errors.New("UnableToFindIPAddress")
	}

	networkInterface := make([]string, 0, len(ipSlice))
	for _, ip := range ipSlice {
		networkInterface = append(networkInterface, ip+":"+port)
	}

	return networkInterface, nil
}
