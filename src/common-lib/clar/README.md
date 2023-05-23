# Clar

Clar stands for command line argument reader. The intention of this package is to allow for the storage of some command line arguments in a standard struct. 

## ServiceInit Interface
```go	
	GetConfigPath() string
```
GetConfigPath returns the config path

```go    
	GetLogFilePath() string
```
GetLogFilePath returns the log file path

```go	
	SetupOsArgs(defaultConfig, defaultLog string, args []string, configIdex, logIndex int)
```
SetupOsArgs setup the config & log file path based on command line argument passed. This function is expected to be called only once in main.
- `defaultConfig` (`string`) - The default config. This is used if `args` can't be parsed.
- `defaultLog` (`string`) - The default log. This is used if `args` can't be parsed.
- `args` (`[]string`) - An array of arguments. You can pass `os.Args` directly.
- `configIndex` (`int`) - The index of `args` in which the config argument exists.
- `logIndex` (`int`) - The index of `args` in which the log argument exists.

```go
	GetExecutablePath() string
```
GetExecutablePath returns Executable Path

## Example 
See the example folder for a runnable example
```go
package main

import (
	"fmt"
	"os"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/clar"
)

func main() {
	si := clar.ServiceInitFactoryImpl{}.GetServiceInit()

	si.SetupOsArgs("defaultConfig.json", "defaultLog.log", os.Args, 3, 4)

	fmt.Println(si.GetConfigPath())
	fmt.Println(si.GetLogFilePath())
}
```
``` bash
go run example.go 'otherargument' 4 'test.json' 'test.log'
```

**Output:**
```
test.json
test.log
```