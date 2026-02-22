package registry

type Registration struct {
	ServiceName ServiceName `json:"serviceName"`
	ServiceUrl  string      `json:"serviceUrl"`
}

type ServiceName string

const (
	LogService = ServiceName("LogService")
)
