package namedpipes

import (
	"net"
	"time"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . ServerPipe,ClientPipe

// AllowSystem  will set permissions for System group and account to have full access to named pipes
// https://itconnect.uw.edu/wares/msinf/other-help/understanding-sddl-syntax/
// https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-definition-language-for-conditional-aces-
const AllowSystem = "D:P(A;OICI;GA;;;SY)"

// PipeConfig contain configuration for the pipe listener. It directly maps to winio.PipeConfig
type PipeConfig struct {
	SecurityDescriptor string
	MessageMode        bool
	InputBufferSize    int32
	OutputBufferSize   int32
}

// ServerPipe is an interface for server named pipe
type ServerPipe interface {
	CreatePipe(pipeName string, config *PipeConfig) (net.Listener, error)
}

// ClientPipe is an interface for client named pipe
type ClientPipe interface {
	DialPipe(path string, timeout *time.Duration) (net.Conn, error)
}
