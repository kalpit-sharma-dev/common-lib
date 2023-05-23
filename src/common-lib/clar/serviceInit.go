// Package clar is - command line argument reader.
package clar

//go:generate mockgen -package mock -destination=mock/mocks.go . ServiceInit,ServiceInitFactory

//ServiceInit interface is used for service initilization
type ServiceInit interface {
	//GetConfigPath returns the config path
	GetConfigPath() string
	//GetLogFilePath returns the log file path
	GetLogFilePath() string
	//SetupOsArgs setup the config & log file path based on command line argument passed. This function is expected to be called only once in main.
	SetupOsArgs(defaultConfig, defaultLog string, args []string, configIdex, logIndex int)
	//GetExecutablePath returns Executable Path
	GetExecutablePath() string
}

//ServiceInitFactory interface gives the instance of the ServiceInit
type ServiceInitFactory interface {
	//GetServiceInit returns the implementation of ServiceInit interface
	GetServiceInit() ServiceInit
}
