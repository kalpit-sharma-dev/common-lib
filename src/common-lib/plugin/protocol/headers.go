package protocol

// HeaderKey is protocol header
type HeaderKey string

//ResponseStatus stataus code for responses from the server
type ResponseStatus int

//MessageDeliveryType of a message allows agent to make decisions about its delivery. Possible values- offline, batched, supercritical etc.
type MessageDeliveryType string

// Headers constant values
const (
	//HdrConstTrueValue is a constant for true
	HdrConstTrueValue string = "true"
	//HdrConstFalseValue is a constant for false
	HdrConstFalseValue string = "false"
	//HdrConstForceSend is a constant for header HdrForceSend
	HdrConstForceSend string = "true"
	//HdrConstPluginDataPersist is a constant for header HdrPluginDataPersist
	HdrConstPluginDataPersist string = "true"
	//HdrRateLimitingTypeSequential is a constant for header HdrRateLimitingType
	HdrRateLimitingTypeSequential string = "sequential"
	//HdrRateLimitingTypeSequential is a constant for header HdrRateLimitingType
	HdrRateLimitingTypeExponential string = "exponential"
	//HdrRateLimitingTypeSequential is a constant for header HdrInactivePath
	HdrInactivePathBroker string = "broker"
	//HdrRateLimitingTypeSequential is a constant for header HdrInactivePath
	HdrInactivePathHeartbeat string = "heartbeat"
	//HdrMessageTyepOffline denotes that this is an offline persisted message
	HdrMessageTyepOffline MessageDeliveryType = "offline"
	//HdrAgentErrSigningPayload denotes that the agent failed to sign payload
	HdrAgentErrSigningPayload string = "HdrAgentErrSigningPayload"
)

const (
	// HdrUserAgent describes client making protocol request
	HdrUserAgent HeaderKey = "User-Agent"

	// HdrContentType describes type of content in request or response
	HdrContentType HeaderKey = "Content-Type"

	// HdrProductType describes type of product in request or response
	HdrProductType HeaderKey = "Product-Type"

	// HdrEndpointID describes unique ID of machine
	HdrEndpointID HeaderKey = "Private-Endpoint"

	// HdrRetryAfter describes how long the user agent should wait before making a follow-up request
	HdrRetryAfter HeaderKey = "Retry-After"

	// HdrRateLimitingFactor describes the factor by which client should limit further requests. To be used i conjunction with HdrRateLimitingType
	HdrRateLimitingFactor HeaderKey = "Continuum-Rate-Limiting-Factor"

	// HdrRateLimitingType describes if rate limiting is sequential or exponential. To be used i conjunction with HdrRateLimitingFactor
	HdrRateLimitingType HeaderKey = "Continuum-Rate-Limiting-Type"

	//HdrInactivePath describes which path flow on AMS has been rate limited oe delayed. Possible values "broker" and "heartbeat"
	HdrInactivePath HeaderKey = "Continuum-Inactive-Path"

	//HdrPluginDataPersist describes whether to persist plugin data if server is offline
	HdrPluginDataPersist HeaderKey = "Continuum-Plugin-Persist-Data"

	//HdrMessageDeliveryType allows agent to make decisions about its delivery. Possible values- offline, batched, supercritical etc.
	HdrMessageDeliveryType HeaderKey = "Continuum-Message-Delivery-Type"

	// HdrForceSend describes whether to try sending data even if server is offline
	HdrForceSend HeaderKey = "Continuum-Plugin-Force-Send"

	// HdrBatchSend describes whether to send messages in a batch
	HdrBatchSend HeaderKey = "Continuum-Batch-Send"

	// HdrErrorCode is for top level error code for a failed request
	HdrErrorCode HeaderKey = "Continuum-Plugin-Error-Code"

	//HdrContentMD5 is MD5 hash key
	HdrContentMD5 HeaderKey = "Content-MD5"

	//HdrContentWebhook indicates message contains webhook
	HdrContentWebhook HeaderKey = "Content-Webhook"

	//HdrAPIVersion indicates version of the API
	HdrAPIVersion HeaderKey = "Continuum-API-Version"

	//HdrDataCompressionType indicates message contains data compression type
	HdrDataCompressionType = "Data-Compression-Type"

	//HdrAcceptEncoding indicates message contains Accept Encoding
	HdrAcceptEncoding = "Accept-Encoding"

	//Ok Response status for successful response
	Ok ResponseStatus = 200

	//StatusCreated Response status for Created response
	StatusCreated ResponseStatus = 201

	//StatusCodeInternalError status for internal exception in executing the task
	StatusCodeInternalError ResponseStatus = 500

	//StatusCodeBadRequest error status for bad request
	StatusCodeBadRequest ResponseStatus = 400

	//StatusUnauthorized error status for unauthorized request
	StatusUnauthorized ResponseStatus = 401

	//PathNotFound Error status for incorrect Plugin Path
	PathNotFound ResponseStatus = 404

	//StatusNoContent Response status for Success response with No Content
	StatusNoContent ResponseStatus = 204

	//HdrPluginPath describes Broker URL where the data would be posted
	HdrPluginPath HeaderKey = "Continuum-Plugin-Path"

	//HdrBrokerPath describes Broker URL where the data would be posted
	HdrBrokerPath HeaderKey = "Continuum-Plugin-Broker-Path"

	//HdrCommunicationPath describes Broker URL where the data would be posted
	HdrCommunicationPath HeaderKey = "Continuum-Plugin-Communication-Path"

	//HdrCommunicationURI describes Communication URL where the data would be posted
	HdrCommunicationURI HeaderKey = "Continuum-Communication-URI"

	//HdrCommunicationMethod describes Communication Method where the data would be posted
	HdrCommunicationMethod HeaderKey = "Continuum-Communication-Method"

	//HdrTaskInput describes Task Input where for execution
	HdrTaskInput HeaderKey = "Continuum-Plugin-Task-Input"

	//HdrTaskBody describes Task Body where for execution
	HdrTaskBody HeaderKey = "Continuum-Plugin-Task-Body"

	//HdrMessageType describes Message Type to process mailbox message at plugin
	HdrMessageType HeaderKey = "Continuum-Plugin-Message-Type"

	//HdrProbeEndpointId describes Probe EndpointId for get ids for vms and hyperv
	HdrProbeEndpointId HeaderKey = "Continuum-Probe-Endpointid"

	//HdrEsxiHostID describes unique identification ID to identify Esxi Host
	HdrEsxiHostID HeaderKey = "Continuum-Esxi-Host-Id"

	//HdrNetworkDeviceID describes unique identification ID to identify network device
	HdrNetworkDeviceID HeaderKey = "Continuum-Network-Device-Id"

	//HdrVcenterID describes unique identification ID to identify vCenter
	HdrVcenterID HeaderKey = "Continuum-Vcenter-Id"

	//HdrIPAddress describes unique IPAddress to identify Esxi Host
	HdrIPAddress HeaderKey = "Continuum-Ipaddress"

	//HdrRegID describes unique regID to identify Esxi Host
	HdrRegID HeaderKey = "Continuum-Regid"

	//HdrTransactionID describes RequestID/TransactionID/CorreleationID to track data across servers and processes.
	HdrTransactionID HeaderKey = "X-Request-Id"

	//HdrHTTPSecure This is temporary Key used for heartbeat, would be removed once the heartbeat changes are done in communication service
	HdrHTTPSecure HeaderKey = "Continuum-HTTP-Secure"

	//HdrAgentOS : This is a header key to pass Agent OS; as a part of any request from Agent
	HdrAgentOS string = "Continuum-Agent-OS"

	//HdrAgentVersion : This is a header key to pass Agent version; as a part of any request from Agent
	HdrAgentVersion string = "Continuum-Agent-Version"

	//HdrAgentCoreVersion : This is a header key to pass AgentCore version; as a part of any request from Agent
	HdrAgentCoreVersion string = "Continuum-AgentCore-Version"

	//HdrPluginTimeout :  This is a header key to pass timeout to the respective plugin
	HdrPluginTimeout HeaderKey = "Plugin-Execution-Timeout"

	//HdrResourcePath describes Resource URL for which data is to be fetched
	HdrResourcePath HeaderKey = "Continuum-Plugin-Resource-Path"

	//HdrEventName describes list of events captured by plugin
	HdrEventName HeaderKey = "Continuum-Plugin-Event-Id"

	//HdrDirectSend describes Resource URL for which data is to be fetched
	HdrDirectSend HeaderKey = "Continuum-Platform-Direct-Send"

	// HdrOfflineMessageHash denotes hash for offline persisted message in the form of unixnano time
	HdrOfflineMessageHash HeaderKey = "Continuum-Offline-Message-Hash"

	//HdrPlatformMessage denotes that this message is meant to be consumed by platform service (aka AMS)
	HdrPlatformMessage HeaderKey = "Continuum-Platform-Message"

	//HdrHeartbeatPersist denotes that this message is meant to be consumed by platform service (aka AMS)
	HdrHeartbeatPersist HeaderKey = "Continuum-Platform-Heartbeat-Persist"

	//HdrAgentDeleteSource denotes that the source of agent deregistration
	HdrAgentDeleteSource HeaderKey = "Continuum-Agent-Delete-Source"

	//HdrAgentDeleteTimestamp denotes that the timestamp of agent deregistration
	HdrAgentDeleteTimestamp HeaderKey = "Continuum-Agent-Delete-Timestamp"

	//HdrAgentPayloadSign denotes the payload's digest signature, signed by agent's private key
	HdrAgentPayloadSign HeaderKey = "Continuum-Agent-Payload-Sign"

	//HdrResponseErrCode error code along with HTTP(s) response
	HdrResponseErrCode HeaderKey = "Continuum-Response-Err-Code"

	//HdrHeartbeatCounter denotes the number of successfull heartbeat counts.
	HdrHeartbeatCounter HeaderKey = "Continuum-Platform-Heartbeat-Counter"

	//HdrProxySetting denotes the proxy configuration to be used while communication
	HdrProxySetting HeaderKey = "Continuum-Platform-Proxy-Setting"

	//HdrRegistrationToken denotes the token used during registration
	HdrRegistrationToken HeaderKey = "Continuum-Platform-Registration-Token"

	//HdrAgentServiceURL denotes the url used by AgentCore to communicate with DC
	HdrAgentServiceURL HeaderKey = "Continuum-Platform-Agent-Service-Url"

	//HdrWebFunctionName denotes the name of function used by AgentCore to communicate with DC
	HdrWebFunctionName HeaderKey = "Continuum-Plugin-Web-Function-Name"

	//HdrProviderName denotes the name of provider used by AgentCore to communicate with DC
	HdrProviderName HeaderKey = "Continuum-Plugin-Provider-Name"
	//HdrAgentUninstallReason denotes the reason used by AgentCore for uninstall reason to communicator plugin
	HdrAgentShutdownReason HeaderKey = "Agent-Shutdown-Reason"

	//HdrAgentAutoUpdatePackages denotes the list of packages agent is updating/updated
	HdrAgentAutoUpdatePackages HeaderKey = "Agent-Auto-Update-Packages"
)

// Headers is a map for Request Response structures
type Headers map[HeaderKey][]string

// SetKeyValue sets a key to a value (overwriting if it exists)
func (h Headers) SetKeyValue(key HeaderKey, value string) {
	h.SetKeyValues(key, []string{value})
}

// SetKeyValues sets a key to values (overwrting if it exists)
func (h Headers) SetKeyValues(key HeaderKey, values []string) {
	h[key] = values
}

// GetKeyValue returns the value for a given key
func (h Headers) GetKeyValue(key HeaderKey) (value string) {
	values := h[key]
	if len(values) > 0 {
		value = values[0]
	}
	return
}

// GetKeyValues returns values array for given key
func (h Headers) GetKeyValues(key HeaderKey) (values []string) {
	return h[key]
}

// Parameters is a map of request parameters
type Parameters map[string][]string
