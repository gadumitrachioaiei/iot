package lorawanservice

import (
	"encoding/json"
	"net/http"
	"sync"
)

// DeviceRegistrator represents an http handler for registering devices, simulating the LoRaWAN API
type DeviceRegistrator struct {
	mu      sync.Mutex
	devices map[string]bool
}

// NewDeviceRegistrator returns a DeviceRegistrator
func NewDeviceRegistrator() *DeviceRegistrator {
	return &DeviceRegistrator{devices: make(map[string]bool)}
}

// ServeHTTP handles the registration endpoint, a post with id of the device to be registered
func (dr *DeviceRegistrator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "wrong http method", http.StatusBadRequest)
		return
	}
	data := struct {
		Deveui string
	}{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(rw, "wrong body", http.StatusBadRequest)
		return
	}
	if dr.register(data.Deveui) {
		rw.WriteHeader(http.StatusOK)
		return
	}
	rw.WriteHeader(http.StatusUnprocessableEntity)
}

func (dr *DeviceRegistrator) register(id string) bool {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	if _, ok := dr.devices[id]; ok {
		return false
	}
	dr.devices[id] = true
	return true
}
