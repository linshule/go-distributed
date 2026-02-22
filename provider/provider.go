package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/linshule/go-distributed/registry"
)

// ServiceProvider 服务提供者
type ServiceProvider struct {
	registrations []registry.Registration
	notifyLock    sync.RWMutex
	notifyMap     map[string][]chan<- registry.Registration
}

var sp = ServiceProvider{
	registrations: make([]registry.Registration, 0),
	notifyMap:     make(map[string][]chan<- registry.Registration),
}

// UpdateServices 更新服务列表
func (p *ServiceProvider) UpdateServices(regs []registry.Registration) {
	p.notifyLock.Lock()
	defer p.notifyLock.Unlock()

	// 检测变化
	oldServices := make(map[string]bool)
	for _, r := range p.registrations {
		oldServices[string(r.ServiceName)] = true
	}

	p.registrations = regs

	// 通知变化
	for _, r := range regs {
		wasRegistered := oldServices[string(r.ServiceName)]
		if !wasRegistered {
			// 新服务注册
			p.notifyAll(string(r.ServiceName), r)
		}
	}

	// 检查注销的服务
	for name := range oldServices {
		stillExists := false
		for _, r := range regs {
			if string(r.ServiceName) == name {
				stillExists = true
				break
			}
		}
		if !stillExists {
			// 服务被注销
			p.notifyAll(name, registry.Registration{
				ServiceName: registry.ServiceName(name),
				ServiceUrl:  "",
			})
		}
	}
}

func (p *ServiceProvider) notifyAll(serviceName string, reg registry.Registration) {
	p.notifyLock.RLock()
	defer p.notifyLock.RUnlock()

	if channels, ok := p.notifyMap[serviceName]; ok {
		for _, ch := range channels {
			select {
			case ch <- reg:
			default:
				// 如果通道已满，跳过
			}
		}
	}
}

// Subscribe 订阅服务变化
func (p *ServiceProvider) Subscribe(serviceName string) chan registry.Registration {
	p.notifyLock.Lock()
	defer p.notifyLock.Unlock()

	ch := make(chan registry.Registration, 10)
	p.notifyMap[serviceName] = append(p.notifyMap[serviceName], ch)
	return ch
}

// Unsubscribe 取消订阅
func (p *ServiceProvider) Unsubscribe(serviceName string, ch chan registry.Registration) {
	p.notifyLock.Lock()
	defer p.notifyLock.Unlock()

	if channels, ok := p.notifyMap[serviceName]; ok {
		for i, c := range channels {
			if c == ch {
				p.notifyMap[serviceName] = append(channels[:i], channels[i+1:]...)
				break
			}
		}
	}
	close(ch)
}

// GetServices 获取所有服务
func (p *ServiceProvider) GetServices() []registry.Registration {
	regs, err := registry.GetServices()
	if err != nil {
		log.Println("Failed to get services:", err)
		return p.registrations
	}
	p.UpdateServices(regs)
	return regs
}

// FindService 查找服务
func (p *ServiceProvider) FindService(serviceName registry.ServiceName) (registry.Registration, error) {
	regs := p.GetServices()
	for _, reg := range regs {
		if reg.ServiceName == serviceName {
			return reg, nil
		}
	}
	return registry.Registration{}, fmt.Errorf("service %s not found", serviceName)
}

// ProviderService HTTP服务
type ProviderService struct{}

func (s ProviderService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		regs := sp.GetServices()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(regs)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RegisterHandlers 注册HTTP处理器
func RegisterHandlers() {
	http.Handle("/providers", &ProviderService{})
}
