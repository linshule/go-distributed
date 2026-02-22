package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 服务发现客户端结构
type DiscoveryClient struct {
	cache          []Registration
	cacheMutex     sync.RWMutex
	cacheExpiry    time.Duration
	lastUpdate     time.Time
	serviceUrl     string
}

// 全局服务发现客户端
var (
	defaultClient = &DiscoveryClient{
		serviceUrl:  ServiceUrl,
		cacheExpiry: 30 * time.Second,
	}
)

// RegistrationService 向注册中心注册服务
func RegistrationService(r Registration) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		return err
	}
	res, err := http.Post(ServiceUrl, "application/json", buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service:%v", res.Status)
	}
	return nil
}

// ShutdownService 通知注册中心服务关闭
func ShutdownService(url string) error {
	req, err := http.NewRequest(http.MethodDelete, ServiceUrl, bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to shutdown service:%v", res.Status)
	}
	return nil
}

// GetServices 获取所有已注册的服务（带缓存）
func GetServices() ([]Registration, error) {
	return defaultClient.GetServices()
}

// GetServicesFresh 强制刷新获取所有已注册的服务
func GetServicesFresh() ([]Registration, error) {
	return defaultClient.GetServicesFresh()
}

// FindService 根据服务名称查找服务（带缓存）
func FindService(serviceName ServiceName) (Registration, error) {
	return defaultClient.FindService(serviceName)
}

// FindServiceFresh 强制刷新查找服务
func FindServiceFresh(serviceName ServiceName) (Registration, error) {
	return defaultClient.FindServiceFresh(serviceName)
}

// FindServicesByTag 根据标签查找服务
func FindServicesByTag(tag string) ([]Registration, error) {
	url := fmt.Sprintf("%s/services/tag/%s", ServiceUrl, tag)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to find services by tag:%v", res.Status)
	}
	defer res.Body.Close()
	var regs []Registration
	err = json.NewDecoder(res.Body).Decode(&regs)
	return regs, err
}

// HealthCheck 检查服务健康状态
func HealthCheck(serviceName ServiceName) (bool, int64, error) {
	url := fmt.Sprintf("%s/health/%s", ServiceUrl, serviceName)
	res, err := http.Get(url)
	if err != nil {
		return false, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, 0, nil
	}

	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return false, 0, err
	}

	healthy, _ := result["healthy"].(bool)
	latency, _ := result["latency"].(float64)
	return healthy, int64(latency), nil
}

// GetServices 获取所有服务（带缓存）
func (c *DiscoveryClient) GetServices() ([]Registration, error) {
	c.cacheMutex.RLock()
	if time.Since(c.lastUpdate) < c.cacheExpiry && len(c.cache) > 0 {
		defer c.cacheMutex.RUnlock()
		return c.cache, nil
	}
	c.cacheMutex.RUnlock()
	return c.GetServicesFresh()
}

// GetServicesFresh 强制刷新获取所有服务
func (c *DiscoveryClient) GetServicesFresh() ([]Registration, error) {
	res, err := http.Get(c.serviceUrl)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get services:%v", res.Status)
	}
	defer res.Body.Close()
	var regs []Registration
	err = json.NewDecoder(res.Body).Decode(&regs)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	c.cacheMutex.Lock()
	c.cache = regs
	c.lastUpdate = time.Now()
	c.cacheMutex.Unlock()

	return regs, nil
}

// FindService 查找服务（带缓存）
func (c *DiscoveryClient) FindService(serviceName ServiceName) (Registration, error) {
	// 先尝试从缓存获取
	c.cacheMutex.RLock()
	if time.Since(c.lastUpdate) < c.cacheExpiry && len(c.cache) > 0 {
		for _, reg := range c.cache {
			if reg.ServiceName == serviceName {
				c.cacheMutex.RUnlock()
				return reg, nil
			}
		}
	}
	c.cacheMutex.RUnlock()

	// 缓存未命中，刷新并重试
	return c.FindServiceFresh(serviceName)
}

// FindServiceFresh 强制刷新查找服务
func (c *DiscoveryClient) FindServiceFresh(serviceName ServiceName) (Registration, error) {
	url := fmt.Sprintf("%s/services/%s", c.serviceUrl, serviceName)
	res, err := http.Get(url)
	if err != nil {
		return Registration{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return Registration{}, fmt.Errorf("service %s not found", serviceName)
	}
	if res.StatusCode != http.StatusOK {
		return Registration{}, fmt.Errorf("failed to find service:%v", res.Status)
	}

	var regs []Registration
	err = json.NewDecoder(res.Body).Decode(&regs)
	if err != nil {
		return Registration{}, err
	}

	if len(regs) == 0 {
		return Registration{}, fmt.Errorf("service %s not found", serviceName)
	}

	// 更新缓存
	c.cacheMutex.Lock()
	c.cache = regs
	c.lastUpdate = time.Now()
	c.cacheMutex.Unlock()

	return regs[0], nil
}

// SetCacheExpiry 设置缓存过期时间
func SetCacheExpiry(expiry time.Duration) {
	defaultClient.cacheExpiry = expiry
}

// ClearCache 清除缓存
func ClearCache() {
	defaultClient.ClearCache()
}

// ClearCache 清除缓存
func (c *DiscoveryClient) ClearCache() {
	c.cacheMutex.Lock()
	c.cache = nil
	c.lastUpdate = time.Time{}
	c.cacheMutex.Unlock()
}
