package model

import (
	"time"

	aModel "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/healthCheck"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/env"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/procParser"
)

// HealthCheck is a common struct to create an common API struct instance for any Service
type HealthCheck struct {
	Version    Version
	ListenPort string
}

// StrartTime : is a variable holds service start time and gets updated at bootstrap
var StrartTime time.Time

// HealthCheckServiceFactory : A factory to create an instance of HealthCheck Service
type HealthCheckServiceFactory interface {
	GetHealthCheckService(f HealthCheckDependencies) HealthCheckService
}

// HealthCheckService : A service to create API model HealthCheck object, so that service can return this
type HealthCheckService interface {
	GetHealthCheck(healthCheck HealthCheck) (aModel.HealthCheck, error)
}

// HealthCheckDalFactory : A factory to create an instance of HealthCheck Dal
type HealthCheckDalFactory interface {
	GetHealthCheckDal(f HealthCheckDependencies) HealthCheckDal
}

// HealthCheckDal : A dal to create API model HealthCheck object, so that dal can return this
type HealthCheckDal interface {
	GetHealthCheck(healthCheck HealthCheck) (aModel.HealthCheck, error)
}

// HealthCheckDependencies : A dependencies for HealthCheck service and factory
type HealthCheckDependencies interface {
	HealthCheckServiceFactory
	HealthCheckDalFactory
	VersionFactory
	env.FactoryEnv
	procParser.ParserFactory
}
