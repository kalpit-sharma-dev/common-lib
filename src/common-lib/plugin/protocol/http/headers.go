package http

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/plugin/protocol"
)

func getHTTPHeader(hdr protocol.HeaderKey) string {
	return string(hdr)
}

func getProtocolHeader(httpHeader string) protocol.HeaderKey {
	return protocol.HeaderKey(httpHeader)
}
