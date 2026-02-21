package registry

type Registration struct {
	ServiceName ServiceName
	ServiceUrl  string
	// RequiredNodes int    `json:"required_nodes" validate:"required"`
}

type ServiceName string

const (
	LogService = ServiceName("LogService")
)
