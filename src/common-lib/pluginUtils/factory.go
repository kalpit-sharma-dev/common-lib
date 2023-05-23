package pluginUtils

import "io"

//go:generate mockgen -package mock -destination=mock/mocks.go . PluginIOReader,PluginIOWriter

//PluginIOReader interface returns a Reader interface
type PluginIOReader interface {
	GetReader() io.Reader
}

//PluginIOWriter interface returns a Writer interface
type PluginIOWriter interface {
	GetWriter() io.Writer
}
