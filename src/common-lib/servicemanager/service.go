package servicemanager

import (
	"fmt"

	svc "github.com/kardianos/service"
)

const (
	//ITSPlatformManagerServiceName ITSPlatformServiceName manager service name
	ITSPlatformManagerServiceName = "ITSPlatformManager"
	//ITSPlatformServiceName main service name
	ITSPlatformServiceName = "ITSPlatform"
	//ITSPlatformServiceNameDarwin /Library/LaunchDaemons/com.ITSPlatform.platform-agent-core.plist
	ITSPlatformServiceNameDarwin = "com.ITSPlatform.platform-agent-core"
	//BrightGaugeServiceName main service name (BG agent)
	BrightGaugeServiceName = "BrightGaugeITSPlatform"
	//BrightGaugeManagerServiceName BrightGaugeITSPlatform manager service name (BG agent)
	BrightGaugeManagerServiceName = "BrightGaugeITSPlatformManager"
)

//ITSPlatformService nolint
type ITSPlatformService struct {
}

//Start nolint
func (p *ITSPlatformService) Start(s svc.Service) error {
	go p.run()
	return nil
}
func (p *ITSPlatformService) run() {
}

//Stop nolint
func (p *ITSPlatformService) Stop(s svc.Service) error {
	return nil
}

//ServiceManager is a struct to hold service state
type ServiceManager struct {
	service svc.Service
	Name    string
}

//Manager is a function to Open service manager
var Manager = func(svcName string) (*ServiceManager, error) {
	svcConfig := &svc.Config{
		Name: svcName,
	}

	itsSvc := &ITSPlatformService{}
	_svc, err := svc.New(itsSvc, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create service reference: %s error: %+v", svcName, err)
	}

	if _, err := _svc.Status(); err == nil {
		return &ServiceManager{Name: svcName, service: _svc}, nil
	}

	return nil, fmt.Errorf("Unable to find %s Service", svcName)
}

//Running is a function to check if service is running or not
var Running = func(manager *ServiceManager) (bool, error) {
	var status svc.Status
	var err error
	if status, err = manager.service.Status(); err == nil {
		return status == svc.StatusRunning, nil
	}
	return false, err
}

//Stopped is a function to check if service is in stopped state or not
var Stopped = func(manager *ServiceManager) (bool, error) {
	var status svc.Status
	var err error
	if status, err = manager.service.Status(); err == nil {
		return status == svc.StatusStopped, nil
	}
	return false, err
}

//Start is a function to start service
var Start = func(manager *ServiceManager) error {
	return manager.service.Start()
}

//Stop is a function to stop a service
var Stop = func(manager *ServiceManager) (bool, error) {
	var err error
	if err = manager.service.Stop(); err == nil {
		if status, err1 := manager.service.Status(); err1 == nil {
			return status == svc.StatusStopped, nil
		}
	}
	return false, err
}

//Close relinquish access to the service
var Close = func(manager *ServiceManager) error {
	var emptyService svc.Service
	manager.service = emptyService
	return nil
}
