package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/linshule/go-distributed/registry"
)

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`     // "healthy", "unhealthy", "unknown"
	LastCheck time.Time `json:"last_check"` // 最后检查时间
	Latency   int64     `json:"latency"`    // 响应延迟（毫秒）
}

// MonitorService 监控服务
type MonitorService struct {
	checks    map[string]*ServiceStatus
	checkLock sync.RWMutex
	interval  time.Duration
}

var monitor = MonitorService{
	checks:   make(map[string]*ServiceStatus),
	interval: 10 * time.Second,
}

// StartMonitoring 启动监控
func (m *MonitorService) StartMonitoring() {
	go func() {
		for {
			m.checkAllServices()
			time.Sleep(m.interval)
		}
	}()
}

func (m *MonitorService) checkAllServices() {
	regs, err := registry.GetServices()
	if err != nil {
		log.Println("Failed to get services:", err)
		return
	}

	m.checkLock.Lock()
	defer m.checkLock.Unlock()

	// 标记所有服务为unknown，然后检查
	currentServices := make(map[string]bool)
	for _, reg := range regs {
		currentServices[string(reg.ServiceName)] = true
		if _, exists := m.checks[string(reg.ServiceName)]; !exists {
			m.checks[string(reg.ServiceName)] = &ServiceStatus{
				Name: string(reg.ServiceName),
				URL:  reg.ServiceUrl,
			}
		}
	}

	// 检查每个服务
	for _, reg := range regs {
		m.checkService(reg.ServiceName, reg.ServiceUrl)
	}

	// 清理已注销的服务
	for name := range m.checks {
		if !currentServices[name] {
			delete(m.checks, name)
		}
	}
}

func (m *MonitorService) checkService(name registry.ServiceName, url string) {
	start := time.Now()
	resp, err := http.Get(url)
	latency := time.Since(start).Milliseconds()

	status := &ServiceStatus{
		Name:      string(name),
		URL:       url,
		LastCheck: time.Now(),
		Latency:   latency,
	}

	if err != nil {
		status.Status = "unhealthy"
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "healthy"
		} else {
			status.Status = "unhealthy"
		}
	}

	m.checks[string(name)] = status
}

// GetStatus 获取所有服务状态
func (m *MonitorService) GetStatus() []ServiceStatus {
	m.checkLock.RLock()
	defer m.checkLock.RUnlock()

	result := make([]ServiceStatus, 0, len(m.checks))
	for _, status := range m.checks {
		result = append(result, *status)
	}
	return result
}

// GetServiceStatus 获取单个服务状态
func (m *MonitorService) GetServiceStatus(name string) (ServiceStatus, error) {
	m.checkLock.RLock()
	defer m.checkLock.RUnlock()

	if status, exists := m.checks[name]; exists {
		return *status, nil
	}
	return ServiceStatus{}, fmt.Errorf("service %s not found", name)
}

// MonitorHTTPService HTTP服务
type MonitorHTTPService struct{}

func (s MonitorHTTPService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := r.URL.Path
		if path == "/health" || path == "/health/" {
			// 获取所有服务健康状态
			statuses := monitor.GetStatus()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(statuses)
		} else if len(path) > 7 && path[:7] == "/health/" {
			// 获取单个服务健康状态
			name := path[7:]
			status, err := monitor.GetServiceStatus(name)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(status)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RegisterHandlers 注册HTTP处理器
func RegisterHandlers() {
	http.Handle("/monitor", &MonitorHTTPService{})
	// 启动监控
	monitor.StartMonitoring()
}
