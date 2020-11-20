package registrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gadumitrachioaiei/iot/lorawanclient"
)

// DeviceRegistrator registers concurrently a batch of devices to lorawan service
type DeviceRegistrator struct {
	config Config
	client *lorawanclient.Client
}

// Config for DeviceRegistrator
type Config struct {
	MaxAPIReqCount int
}

// NewDeviceRegistrator returns an instance of DeviceRegistrator
func NewDeviceRegistrator(config Config, lorawanClient *lorawanclient.Client) (*DeviceRegistrator, error) {
	if config.MaxAPIReqCount < 1 {
		return nil, errors.New("bad device registrator config")
	}
	return &DeviceRegistrator{config, lorawanClient}, nil
}

// Register registers a batch of devices concurrently.
//
// You can cancel it with ctx.
//
// Returns a list of devices that were registered and any errors encountered.
//
// The list of returned devices may not be complete, in case there were errors or we had a cancellation.
func (dr *DeviceRegistrator) Register(ctx context.Context, devices []string) ([]string, []error) {
	var (
		mu         sync.Mutex
		registered []string
		errs       []error
	)
	semaphor := make(chan struct{}, dr.config.MaxAPIReqCount)
	var wg sync.WaitGroup
workLoop:
	for _, device := range devices {
		device := device
		select {
		case <-ctx.Done():
			break workLoop
		case semaphor <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-semaphor }()
				err := dr.client.Register(device)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errs = append(errs, fmt.Errorf("cannot register device: %s %v", device, err))
				} else {
					registered = append(registered, device)
				}
			}()
		}
	}
	wg.Wait()
	return registered, errs
}
