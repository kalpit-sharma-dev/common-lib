<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Service Manager

Common lib wrapper module for service manager usage

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/servicemanager"
```

**Functions**

```go
Manager(svcName string) (*ServiceManager, error)    //Manager is a function to Open service manager
```

```go
Running(manager *ServiceManager) (bool, error)    //Running is a function to check if service is running or not
```

```go
Stopped(manager *ServiceManager) (bool, error)    //Stopped is a function to check if service is in stopped state or notfound error
```

```go
Start(manager *ServiceManager) error    //Start is a function to start service
```

```go
Stop(manager *ServiceManager) (bool, error)    //Stop is a function to stop a service
```

```go
Close(manager *ServiceManager) error    //Close relinquish access to the service
```

### Contribution

Any changes in this package should be communicated to Common Frameworks Team.
