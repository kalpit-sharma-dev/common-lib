package http

import (
	"io"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/plugin/protocol"
)

type responseSerializerImpl struct{}

func (ser *responseSerializerImpl) Serialize(res *protocol.Response, dst io.Writer) (err error) {
	return responseSerializeRaw(res, dst)
}

func (ser *responseSerializerImpl) Deserialize(src io.Reader) (res *protocol.Response, err error) {
	return responseDeserializeRaw(src)
}
