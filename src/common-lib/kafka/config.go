// Package kafka implements kafka client configuration details
//
// Deprecated: kafka is old implementation of kafka connectivity and should not be used
// except for compatibility with legacy systems.
//
// Use https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/tree/master/messaging for all kafka connectivity
// This package is frozen and no new functionality will be added.
package kafka

import "errors"

//ClientConfig base properties for configs
type ClientConfig struct {
	BrokerAddress []string
}

//ProducerConfig contains configuration information for kafka consumer
type ProducerConfig struct {
	ClientConfig
	//RequiredAcks  int16
}

//ConsumerConfig contaijns configuration information for kafka consumer
type ConsumerConfig struct {
	ClientConfig
	GroupID    string
	Topics     []string
	AutoCommit bool
}

//Validates Config for both producer and consumer
func validateConfig(config *ClientConfig) error {
	if config.BrokerAddress == nil {
		err := errors.New(ErrorBrokerAddressNotProvided)
		return err
	} else if len(config.BrokerAddress) == 0 {
		err := errors.New(ErrorBrokerAddressNotProvided)
		return err
	}
	return nil
}
