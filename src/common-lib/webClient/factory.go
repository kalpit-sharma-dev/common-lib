package webClient

import (
	"fmt"
	"strings"
)

// ClientFactoryImpl implementation of ClientFactory
type ClientFactoryImpl struct {
}

// GetClientServiceByType method to return Client configuration by its Type
func (ClientFactoryImpl) GetClientServiceByType(clientType ClientType, config ClientConfig) HTTPClientService {
	switch clientType {
	case TLSClient:
		return &tlsClientService{
			config: config,
		}
	default:
		return &httpClientServiceImpl{
			config: config,
		}
	}
}

// GetClientService return a concreate implementation of Client
func (ClientFactoryImpl) GetClientService(fact HTTPClientFactory, config ClientConfig) ClientService {
	service := clientServiceImpl{
		config: config,
	}
	service.httpClientFact = fact
	return service
}

// HTTPClientFactoryImpl implements HTTPCommandFactory
type HTTPClientFactoryImpl struct {
}

// GetHTTPClient returns HTTPCommandService
func (HTTPClientFactoryImpl) GetHTTPClient(config ClientConfig) HTTPClientService {
	return &httpClientServiceImpl{
		config: config,
	}
}

const (
	cOfflineErrorKeywords string = "timeout, connection refused, connection was aborted, no such host, connection was forcibly closed"
	// ErrorClientOffline is returned when the client is offline
	ErrorClientOffline string = "ErrClientOffline"
)

func checkOffline(err error) error {
	errStrings := strings.Split(cOfflineErrorKeywords, ",")
	for _, v := range errStrings {
		if strings.Contains(err.Error(), v) {
			return fmt.Errorf(fmt.Sprintf("%s : %v", ErrorClientOffline, err))
		}
	}
	return err
}
