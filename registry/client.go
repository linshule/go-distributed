package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

// GetServices 获取所有已注册的服务
func GetServices() ([]Registration, error) {
	res, err := http.Get(ServiceUrl)
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
	return regs, nil
}

// FindService 根据服务名称查找服务
func FindService(serviceName ServiceName) (Registration, error) {
	regs, err := GetServices()
	if err != nil {
		return Registration{}, err
	}
	for _, reg := range regs {
		if reg.ServiceName == serviceName {
			return reg, nil
		}
	}
	return Registration{}, fmt.Errorf("service %s not found", serviceName)
}
