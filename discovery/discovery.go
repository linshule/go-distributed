package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/linshule/go-distributed/registry"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	Name      string            `json:"name"`       // 服务名称
	URL       string            `json:"url"`        // 服务URL
	Version   string            `json:"version"`    // 服务版本
	Metadata  map[string]string `json:"metadata"`   // 元数据
	Tags      []string          `json:"tags"`       // 标签
	Healthy   bool              `json:"healthy"`    // 健康状态
	Latency   int64             `json:"latency"`    // 响应延迟
	LastCheck time.Time         `json:"lastCheck"`  // 最后检查时间
}

// ServiceWatcher 服务变化观察者
type ServiceWatcher struct {
	serviceName string
	callback    func(*ServiceInstance)
}

// Discovery 服务发现
type Discovery struct {
	instances  map[string][]*ServiceInstance // 服务实例列表
	watchers   map[string][]*ServiceWatcher // 观察者列表
	mutex      sync.RWMutex
	httpClient *http.Client
	registryURL string
}

// 全局服务发现实例
var d = &Discovery{
	instances:  make(map[string][]*ServiceInstance),
	watchers:   make(map[string][]*ServiceWatcher),
	httpClient: &http.Client{Timeout: 5 * time.Second},
	registryURL: "http://localhost:3000/services",
}

// New 创建新的服务发现实例
func New(registryURL string) *Discovery {
	return &Discovery{
		instances:   make(map[string][]*ServiceInstance),
		watchers:    make(map[string][]*ServiceWatcher),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		registryURL: registryURL,
	}
}

// Refresh 刷新服务列表
func (d *Discovery) Refresh() error {
	regs, err := registry.GetServices()
	if err != nil {
		return err
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// 创建新的实例映射
	newInstances := make(map[string][]*ServiceInstance)

	for _, reg := range regs {
		instance := &ServiceInstance{
			Name:     string(reg.ServiceName),
			URL:      reg.ServiceUrl,
			Version:  reg.ServiceVersion,
			Metadata: reg.Metadata,
			Tags:     reg.Tags,
		}

		// 检查健康状态
		if reg.HealthCheckURL != "" {
			instance.Healthy, instance.Latency = d.checkHealth(reg.HealthCheckURL)
		} else {
			instance.Healthy, instance.Latency = d.checkHealth(reg.ServiceUrl)
		}
		instance.LastCheck = time.Now()

		newInstances[instance.Name] = append(newInstances[instance.Name], instance)
	}

	d.instances = newInstances

	// 通知观察者
	d.notifyWatchers()

	return nil
}

// checkHealth 检查服务健康状态
func (d *Discovery) checkHealth(url string) (bool, int64) {
	start := time.Now()
	resp, err := d.httpClient.Get(url + "/health")
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return false, latency
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, latency
}

// notifyWatchers 通知所有观察者
func (d *Discovery) notifyWatchers() {
	for name, instances := range d.instances {
		if watchers, ok := d.watchers[name]; ok {
			for _, w := range watchers {
				if len(instances) > 0 {
					w.callback(instances[0])
				}
			}
		}
	}
}

// GetInstances 获取所有服务实例
func (d *Discovery) GetInstances(serviceName string) []*ServiceInstance {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.instances[serviceName]
}

// GetHealthyInstance 获取一个健康的服务实例（负载均衡）
func (d *Discovery) GetHealthyInstance(serviceName string) (*ServiceInstance, error) {
	d.mutex.RLock()
	instances := d.instances[serviceName]
	d.mutex.RUnlock()

	var healthyInstances []*ServiceInstance
	for _, inst := range instances {
		if inst.Healthy {
			healthyInstances = append(healthyInstances, inst)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instance found for %s", serviceName)
	}

	// 简单轮询负载均衡
	index := time.Now().UnixNano() % int64(len(healthyInstances))
	return healthyInstances[index], nil
}

// Watch 观察服务变化
func (d *Discovery) Watch(serviceName string, callback func(*ServiceInstance)) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	watcher := &ServiceWatcher{
		serviceName: serviceName,
		callback:    callback,
	}
	d.watchers[serviceName] = append(d.watchers[serviceName], watcher)
}

// StartPolling 启动定时刷新
func (d *Discovery) StartPolling(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := d.Refresh(); err != nil {
				log.Printf("Discovery refresh error: %v", err)
			}
		}
	}()
}

// GetAllServices 获取所有服务信息
func (d *Discovery) GetAllServices() map[string][]*ServiceInstance {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	result := make(map[string][]*ServiceInstance)
	for k, v := range d.instances {
		result[k] = v
	}
	return result
}

// 全局函数

// Refresh 刷新全局服务发现实例
func Refresh() error {
	return d.Refresh()
}

// GetInstances 获取服务实例
func GetInstances(serviceName string) []*ServiceInstance {
	return d.GetInstances(serviceName)
}

// GetHealthyInstance 获取健康服务实例
func GetHealthyInstance(serviceName string) (*ServiceInstance, error) {
	return d.GetHealthyInstance(serviceName)
}

// Watch 观察服务变化
func Watch(serviceName string, callback func(*ServiceInstance)) {
	d.Watch(serviceName, callback)
}

// StartPolling 启动定时刷新
func StartPolling(interval time.Duration) {
	d.StartPolling(interval)
}

// GetAllServices 获取所有服务
func GetAllServices() map[string][]*ServiceInstance {
	return d.GetAllServices()
}

// ServiceHandler HTTP处理器，用于服务发现API
type ServiceHandler struct{}

// ServeHTTP 处理服务发现请求
func (h *ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// 刷新服务列表
		if err := d.Refresh(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		// 返回所有服务
		services := d.GetAllServices()
		json.NewEncoder(w).Encode(services)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RegisterHandlers 注册HTTP处理器
func RegisterHandlers() {
	http.Handle("/discovery", &ServiceHandler{})
}
