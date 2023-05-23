<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Exec

Interface wrapper around some os/exec functionality

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exec"
```

Command is an interface that has a run function to run a command with an optional list of arguments

```go
type Command interface {
	Run(string, ...string) error
}
```

Usage - This is an example using the built in implementation of the Command interface, CommandImpl

```go
func main() {
    runner := exec.CommandImpl{}
    runStartupCommand(runner)
}

func runStartupCommand(commandRunner exec.Command) error {
    return commandRunner.Run("startup")
}
```

Testing - One benefit is this is very easy to test with the built in mocks

```go
import "github.com/golang/mock/gomock"
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exec/mock"

//Very simplified test showing an easy way to test functions that are using an exec.Command object
func Test_RunStartupCommand(t *testing) {
    ctrl := gomock.NewController(t)
    mockRunner := mock.NewMockCommand(ctrl)
    mockCommand.EXPECT.Run("startup").Returns(nil).Times(1)
    runStartupCommand(mockRunner)
}
```

### Contribution

Any changes in this package should be communicated to Common Frameworks.
