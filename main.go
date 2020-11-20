package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gadumitrachioaiei/iot/idgenerator"
	"github.com/gadumitrachioaiei/iot/lorawanclient"
	"github.com/gadumitrachioaiei/iot/lorawanservice"
	"github.com/gadumitrachioaiei/iot/registrator"
	"github.com/gadumitrachioaiei/iot/service"
)

const (
	maxAPIReqCount          = 10
	deviceCount             = 100
	lorawanScheme           = "http://"
	lorawanAddress          = "127.0.0.1:9090"
	lorawanRegistrationPath = "/sensor-onboarding-sample"
	lorawanClientTimeout    = 2 * time.Second
	serviceAddress          = "127.0.0.1:8080"
	serviceRegistrationPath = "/sensor-onboarding-sample"
)

func main() {
	go func() {
		if err := startLorawanService(); err != nil {
			log.Fatalf("Cannot connect to LoRaWANAPI: %v", err)
		}
	}()
	//  a little sleep to make sure lorawan service is up and running
	time.Sleep(10 * time.Millisecond)
	devices := idgenerator.Get(deviceCount)
	lorawanClient, err := lorawanclient.NewClient(lorawanclient.Config{
		Scheme:       lorawanScheme,
		Addr:         lorawanAddress,
		RegisterPath: lorawanRegistrationPath,
		Timeout:      lorawanClientTimeout,
	})
	if err != nil {
		log.Fatalf("Cannot create lorawanservice client: %v", err)
	}
	shutdown := handleInterrupts()
	ctx, cancel := context.WithCancel(context.Background())
	registrator, err := registrator.NewDeviceRegistrator(registrator.Config{MaxAPIReqCount: maxAPIReqCount}, lorawanClient)
	if err != nil {
		log.Fatalf("Cannot create device registrator: %v", err)
	}
	apiServer := createService(ctx, deviceCount, registrator)
	idleConnsClosed := make(chan struct{}) // make sure main still runs until apiServer.Shutdown returns
	go func() {
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Device registration service: %v", err)
		}
	}()
	go func() {
		<-shutdown
		cancel()
		apiServer.Shutdown(context.Background())
		close(idleConnsClosed)
	}()
	registeredDevices, errs := registrator.Register(ctx, devices)
	log.Printf("registered devices:\n%v\n", registeredDevices)
	if errs != nil {
		log.Printf("\nerrors:\n%v\n", errs)
	}
	<-idleConnsClosed
}

// catch SIGINT so we can shutdown gracefully
func handleInterrupts() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	return c
}

// startLorawanService starts the lorawan service simulator
func startLorawanService() error {
	mux := http.NewServeMux()
	mux.Handle(lorawanRegistrationPath, lorawanservice.NewDeviceRegistrator())
	s := http.Server{
		Addr:    lorawanAddress,
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// createService creates our service service for device registration
func createService(ctx context.Context, deviceCount int, dr *registrator.DeviceRegistrator) http.Server {
	mux := http.NewServeMux()
	mux.Handle(serviceRegistrationPath, service.NewService(ctx, deviceCount, dr))
	s := http.Server{
		Addr:    serviceAddress,
		Handler: mux,
	}
	return s
}
