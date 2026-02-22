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
