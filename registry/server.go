package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServiceUrl = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.Mutex
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
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

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %s\n", r.ServiceName, r.ServiceUrl)
		err = reg.add(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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
		return
	}
}
