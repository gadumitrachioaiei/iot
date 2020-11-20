package lorawanclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	Scheme       string
	Addr         string
	RegisterPath string
	Timeout      time.Duration
}

type Client struct {
	registerEndpoint string
	c                *http.Client
}

func NewClient(config Config) (*Client, error) {
	if config.Scheme == "" {
		return nil, errors.New("empty server protocol")
	}
	if config.Addr == "" {
		return nil, errors.New("empty server address")
	}
	if config.RegisterPath == "" {
		return nil, errors.New("empty path for device registration")
	}
	base, err := url.Parse(config.Scheme + config.Addr)
	if err != nil {
		return nil, fmt.Errorf("parsing server address: %v", err)
	}
	path, err := url.Parse(config.RegisterPath)
	if err != nil {
		return nil, fmt.Errorf("path for device registration: %v", err)
	}
	return &Client{
		c: &http.Client{
			Timeout: config.Timeout,
		},
		registerEndpoint: base.ResolveReference(path).String(),
	}, nil
}

func (c Client) Register(id string) error {
	body := new(bytes.Buffer)
	data := struct{ Deveui string }{id}
	if err := json.NewEncoder(body).Encode(data); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.registerEndpoint, body)
	if err != nil {
		return err
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnprocessableEntity {
		return fmt.Errorf("bad status code: %v", resp.StatusCode)
	}
	return nil
}
