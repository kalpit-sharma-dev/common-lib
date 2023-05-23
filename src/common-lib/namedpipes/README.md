<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Named Pipes
This the standard implementation of named pipes used by all the Go projects for Inter-Process Communication.

## Functions

### GetPipeServer
```go
func GetPipeServer() ServerPipe
```
GetPipeServer creates and returns a ServerPipe interface.


### GetPipeClient
```go
func GetPipeClient() ClientPipe
```
GetPipeClient creates and returns a ClientPipe interface.


## Types

### type [PipeConfig](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/-/blob/master/src/namedpipes/interfaces.go#L16)

```go
type PipeConfig struct {
	SecurityDescriptor string
	MessageMode        bool
	InputBufferSize    int32
	OutputBufferSize   int32
}
```
PipeConfig contain configuration for the pipe listener. It directly maps to winio.PipeConfig

### type [ClientPipe](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/-/blob/master/src/namedpipes/interfaces.go#L29)

```go
type ClientPipe interface {
	DialPipe(path string, timeout *time.Duration) (net.Conn, error)
}
```
ClientPipe is an interface to create a named pipe client. The argument parameter path
is the name of the named pipe. The parameter timeout is the maximum amount of time
the client should wait when connecting to the server. The function DialPipe returns a
net.Conn when it is successful.

### type [ServerPipe](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/-/blob/master/src/namedpipes/interfaces.go#L24)

```go
type ServerPipe interface {
	CreatePipe(pipeName string, config *PipeConfig) (net.Listener, error)
}
```
ServerPipe is an interface to create a named pipe server to be used for IPC. The pipeName
is the named of the named pipe. The parameter config contains configuration for the pipe
listener. The function CreatePipe returns a net.Listener when it is successful.

