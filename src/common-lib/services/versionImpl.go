package services

import (
	"time"

	aModel "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/version"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/services/model"
)

// VersionFactoryImpl returns the concrete implementation of Factory
type VersionFactoryImpl struct {
}

// GetVersionService : A factory function to create an instance of Version Service
func (VersionFactoryImpl) GetVersionService() model.VersionService {
	return versionServiceImpl{}
}

// versionServiceImpl returns the concrete implementation of VersionService
type versionServiceImpl struct{}

func (v versionServiceImpl) GetVersion(ver model.Version) aModel.Version {
	return aModel.Version{
		Name:            ver.SolutionName,
		Type:            "Version",
		TimeStampUTC:    time.Now().UTC().Round(time.Second),
		ServiceName:     ver.ServiceName,
		ServiceProvider: ver.ServiceProvider,
		ServiceVersion:  ver.Major + "." + ver.Minor + "." + ver.Patch,
		BuildNumber:     model.BuildVersion,
	}
}
