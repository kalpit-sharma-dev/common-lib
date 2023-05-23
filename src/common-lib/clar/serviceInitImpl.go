package clar

import (
	"os"
	"path/filepath"
	"strings"
)

var serviceInitSingleton ServiceInit

// ServiceInitFactoryImpl is the serviceInit Factory
type ServiceInitFactoryImpl struct{}

// GetServiceInit returns an instance of ServiceInit
// This code is not thread safe. the creation of ServiceInit is unlikely to
// happen across multiple goroutines, hence mutex is not used.
func (ServiceInitFactoryImpl) GetServiceInit() ServiceInit {
	if serviceInitSingleton == nil {
		serviceInitSingleton = newServiceInit()
	}
	return serviceInitSingleton
}

func newServiceInit() ServiceInit {
	return &serviceInit{}
}

type serviceInit struct {
	configFilePath string
	configIndex    int
	logFilePath    string
	logIndex       int
	executablePath string
}

func (s *serviceInit) GetConfigPath() string {
	return s.configFilePath
}

func (s *serviceInit) GetLogFilePath() string {
	return s.logFilePath
}

//GetExecutablePath returns Executable Path
func (s *serviceInit) GetExecutablePath() string {
	return s.executablePath
}

func (s *serviceInit) SetupOsArgs(defaultConfig, defaultLog string, args []string, configIdex, logIndex int) {
	s.configFilePath = defaultConfig
	s.logFilePath = defaultLog
	s.configIndex = configIdex
	s.logIndex = logIndex
	s.setupConfigFile(args)
	s.setupLogFile(args)
	s.setupExecutablePath(args)
}

//Private function to set the executable path
func (s *serviceInit) setupExecutablePath(args []string) {
	if len(args) > 0 && len(args[0]) > 0 {
		s.executablePath = filepath.Dir(args[0])
	}
}

func (s *serviceInit) setupConfigFile(args []string) {
	if len(args) > s.configIndex {
		value := args[s.configIndex]
		if value != "" {
			s.configFilePath = value
		}
	}
}

// irrespective of whether log file parameter is present or not,
// the log writer should be setup final value of the logFilePath
func (s *serviceInit) setupLogFile(args []string) {
	if len(args) > s.logIndex {
		value := args[s.logIndex]
		if value != "" {
			s.logFilePath = value
		}
	}
}

// GetDirectoryPath this will extract directory path from input path
func GetDirectoryPath(inputPath, directoryToSearch string) string {
	if len(inputPath) > 0 {
		currDir := filepath.Dir(inputPath)
		folderName := strings.Split(currDir, string(os.PathSeparator))
		for i := len(folderName) - 1; i >= 0; i-- {
			if folderName[i] == directoryToSearch {
				return currDir
			}
			currDir = filepath.Join(currDir, "..")
		}
	}
	return ""
}
