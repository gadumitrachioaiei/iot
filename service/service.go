package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gadumitrachioaiei/iot/idgenerator"
	"github.com/gadumitrachioaiei/iot/registrator"

	"golang.org/x/sync/singleflight"
)

type Service struct {
	ctx         context.Context // ctx received from main, so we can interrupt pending requests if necessary
	deviceCount int
	dr          *registrator.DeviceRegistrator

	group *singleflight.Group // singleflight pattern based on clientID
}

// NewService returns a handler for registering devices belonging to a client
func NewService(ctx context.Context, deviceCount int, dr *registrator.DeviceRegistrator) *Service {
	return &Service{ctx: ctx, deviceCount: deviceCount, dr: dr, group: &singleflight.Group{}}
}

// ServeHTTP handles the registration endpoint, a post with clientID that wants their devices registered
func (s *Service) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "wrong http method", http.StatusBadRequest)
		return
	}
	var body struct {
		ClientID string
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.ClientID == "" {
		http.Error(rw, "wrong body", http.StatusBadRequest)
		return
	}
	data, _, _ := s.group.Do(body.ClientID, func() (interface{}, error) {
		return register(s.ctx, s.dr, s.deviceCount), nil
	})
	if err := json.NewEncoder(rw).Encode(data); err != nil {
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
}

// RegisteredDevices represents the response for device registration from a client
type RegisteredDevices struct {
	Deveuis []string
	Errs    []string
}

// register registers devices and returns what has been registered and any errors
func register(ctx context.Context, dr *registrator.DeviceRegistrator, deviceCount int) RegisteredDevices {
	devices := idgenerator.Get(deviceCount)
	registered, errs := dr.Register(ctx, devices)
	var data RegisteredDevices
	data.Deveuis = registered
	for _, err := range errs {
		data.Errs = append(data.Errs, err.Error())
	}
	return data
}
