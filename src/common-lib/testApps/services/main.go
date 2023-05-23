package main

import (
	"fmt"

	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/env"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/procParser"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
)

type healthCheckDependencyImpl struct {
	services.HealthCheckServiceFactoryImpl
	services.HealthCheckDalFactoryImpl
	services.VersionFactoryImpl
	env.FactoryEnvImpl
	procParser.ParserFactoryImpl
}

func main() {
	h := services.HealthCheckServiceFactoryImpl{}
	s := h.GetHealthCheckService(healthCheckDependencyImpl{})
	model.StrartTime = time.Now()
	health, _ := s.GetHealthCheck(model.HealthCheck{
		Version: model.Version{
			SolutionName:    "SolutionName",
			ServiceName:     "ServiceName",
			ServiceProvider: "ContinuumLLC",
			Major:           "1",
			Minor:           "1",
			Patch:           "11",
		},
		ListenPort: ":8081",
	})

	fmt.Println(health)
}
