package main

import (
	"fmt"
	"os"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/clar"
)

// go run example.go 1 4 'test.json' 'test.log'
func main() {
	si := clar.ServiceInitFactoryImpl{}.GetServiceInit()

	si.SetupOsArgs("defaultConfig.json", "defaultLog.log", os.Args, 3, 4)

	fmt.Println(si.GetConfigPath())
	fmt.Println(si.GetLogFilePath())
	fmt.Println(si.GetExecutablePath())
}
