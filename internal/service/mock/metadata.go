package mock

// Metadata keys injected into MockResponse.Metadata throughout the service chain.
// Values are human-readable so they can be displayed in the UI directly.
const (
	MetadataMatched          = "Matched"
	MetadataSource           = "Source"
	MetadataPath             = "Path"
	MetadataSimulatedStatus  = "Simulated Status"
	MetadataStatusRuleScope  = "Status Rule Scope"
	MetadataSimulatedLatency = "Simulated Latency"
	MetadataLatencyRuleScope = "Latency Rule Scope"
	MetadataLatencyRange     = "Latency Range (ms)"
)
