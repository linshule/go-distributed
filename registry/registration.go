package registry

import "time"

// Registration 服务注册信息
type Registration struct {
	ServiceName    ServiceName            `json:"serviceName"`     // 服务名称
	ServiceUrl     string                 `json:"serviceUrl"`      // 服务URL
	ServiceVersion string                 `json:"serviceVersion"`  // 服务版本
	Metadata       map[string]string      `json:"metadata"`       // 服务元数据
	Tags           []string               `json:"tags"`            // 服务标签
	HealthCheckURL string                 `json:"healthCheckUrl"`   // 健康检查URL
	RegisteredAt   time.Time              `json:"registeredAt"`    // 注册时间
}

// ServiceName 服务名称类型
type ServiceName string

// 预定义的服务名称常量
const (
	LogService      = ServiceName("LogService")
	LibraryService = ServiceName("LibraryService")
	ProviderService = ServiceName("ProviderService")
	WebService     = ServiceName("WebService")
	MonitorService = ServiceName("MonitorService")
)
