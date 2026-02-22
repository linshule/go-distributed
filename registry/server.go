package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServiceUrl = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.Mutex
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查是否已存在同名服务，如果是则更新
	for i, existing := range r.registrations {
		if existing.ServiceName == reg.ServiceName {
			r.registrations[i] = reg
			return nil
		}
	}
	r.registrations = append(r.registrations, reg)
	return nil
}

func (r *registry) remove(url string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i := range r.registrations {
		if r.registrations[i].ServiceUrl == url {
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Service at %s not found", url)
}

func (r *registry) getRegistrations() []Registration {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	result := make([]Registration, len(r.registrations))
	copy(result, r.registrations)
	return result
}

// findByName 根据服务名称查找服务
func (r *registry) findByName(serviceName ServiceName) []Registration {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var result []Registration
	for _, reg := range r.registrations {
		if reg.ServiceName == serviceName {
			result = append(result, reg)
		}
	}
	return result
}

// findByTag 根据标签查找服务
func (r *registry) findByTag(tag string) []Registration {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var result []Registration
	for _, reg := range r.registrations {
		for _, t := range reg.Tags {
			if t == tag {
				result = append(result, reg)
				break
			}
		}
	}
	return result
}

// healthCheck 检查服务健康状态
func (r *registry) healthCheck(serviceName ServiceName) (bool, int64) {
	regs := r.findByName(serviceName)
	if len(regs) == 0 {
		return false, 0
	}

	for _, reg := range regs {
		healthURL := reg.HealthCheckURL
		if healthURL == "" {
			healthURL = reg.ServiceUrl
		}

		start := time.Now()
		resp, err := http.Get(healthURL + "/health")
		latency := time.Since(start).Milliseconds()

		if err == nil && resp.StatusCode == http.StatusOK {
			return true, latency
		}
	}
	return false, 0
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
}

// RegistryService HTTP处理器
type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL.Path)

	// 路由处理
	path := r.URL.Path

	switch {
	// 健康检查端点: /health
	case path == "/health":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})

	// 按名称查询服务: /services/{serviceName}
	case strings.HasPrefix(path, "/services/") && r.Method == http.MethodGet:
		serviceName := strings.TrimPrefix(path, "/services/")
		if serviceName == "" {
			// 返回所有服务
			w.Header().Set("Content-Type", "application/json")
			regs := reg.getRegistrations()
			json.NewEncoder(w).Encode(regs)
			return
		}
		regs := reg.findByName(ServiceName(serviceName))
		w.Header().Set("Content-Type", "application/json")
		if len(regs) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "service not found",
			})
			return
		}
		json.NewEncoder(w).Encode(regs)

	// 按标签查询服务: /services/tag/{tag}
	case strings.HasPrefix(path, "/services/tag/") && r.Method == http.MethodGet:
		tag := strings.TrimPrefix(path, "/services/tag/")
		regs := reg.findByTag(tag)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(regs)

	// 服务健康检查: /health/{serviceName}
	case strings.HasPrefix(path, "/health/") && r.Method == http.MethodGet:
		serviceName := strings.TrimPrefix(path, "/health/")
		healthy, latency := reg.healthCheck(ServiceName(serviceName))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"serviceName": serviceName,
			"healthy":     healthy,
			"latency":     latency,
		})

	// 核心服务注册接口: /services
	case path == "/services":
		switch r.Method {
		case http.MethodPost:
			dec := json.NewDecoder(r.Body)
			var regData Registration
			err := dec.Decode(&regData)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// 设置注册时间
			regData.RegisteredAt = time.Now()
			log.Printf("Adding service: %v with URL: %s\n", regData.ServiceName, regData.ServiceUrl)
			err = reg.add(regData)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			url := string(payload)
			log.Printf("Removing service at URL: %s\n", url)
			err = reg.remove(url)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			regs := reg.getRegistrations()
			json.NewEncoder(w).Encode(regs)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
