package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Caik/go-mock-server/internal/util"
)

type HostsConfig struct {
	Hosts map[string]HostConfig `json:"hosts"`
}

type HostConfig struct {
	LatencyConfig *LatencyConfig         `json:"latency"`
	ErrorsConfig  map[string]ErrorConfig `json:"errors"`
	UrisConfig    map[string]UriConfig   `json:"uris"`
}

type UriConfig struct {
	LatencyConfig *LatencyConfig         `json:"latency"`
	ErrorsConfig  map[string]ErrorConfig `json:"errors"`
}

type LatencyConfig struct {
	Min *int `json:"min"`
	P95 *int `json:"p95"`
	P99 *int `json:"p99"`
	Max *int `json:"max"`
}

type ErrorConfig struct {
	Percentage    *int           `json:"percentage"`
	LatencyConfig *LatencyConfig `json:"latency"`
}

func (h *HostsConfig) Validate() error {
	for host, hostConfig := range h.Hosts {
		if !util.HostRegex.MatchString(host) {
			return errors.New("invalid hosts config found: it doesn't a host pattern")
		}

		if err := hostConfig.Validate(); err != nil {
			return fmt.Errorf("invalid hosts config found: %v", err)
		}
	}

	return nil
}

func (h *HostsConfig) GetHostConfig(host string) *HostConfig {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil
	}

	return &hostConfig
}

func (h *HostsConfig) SetHostConfig(host string, newConfig HostConfig) {
	h.Hosts[host] = newConfig
}

func (h *HostsConfig) DeleteHostConfig(host string) {
	delete(h.Hosts, host)
}

func (h *HostsConfig) UpdateHostLatencyConfig(host string, latencyConfig *LatencyConfig) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.LatencyConfig = latencyConfig
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) DeleteHostLatencyConfig(host string) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.LatencyConfig = nil
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) UpdateHostErrorsConfig(host string, errorsConfig map[string]ErrorConfig) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.ErrorsConfig = errorsConfig
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) DeleteHostErrorConfig(host, errorCode string) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	delete(hostConfig.ErrorsConfig, errorCode)
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) UpdateHostUrisConfig(host string, urisConfig map[string]UriConfig) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.UrisConfig = urisConfig
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) GetAppropriateErrorsConfig(host, uri string) *map[string]ErrorConfig {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil
	}

	errorsConfig := hostConfig.ErrorsConfig
	uriConfig, exists := hostConfig.UrisConfig[uri]

	if exists && len(uriConfig.ErrorsConfig) > 0 {
		errorsConfig = uriConfig.ErrorsConfig
	}

	if len(errorsConfig) > 0 {
		return &errorsConfig
	}

	return nil
}

func (h *HostsConfig) GetAppropriateLatencyConfig(host, uri string) *LatencyConfig {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil
	}

	latencyConfig := hostConfig.LatencyConfig
	uriConfig, exists := hostConfig.UrisConfig[uri]

	if exists && uriConfig.LatencyConfig != nil {
		latencyConfig = uriConfig.LatencyConfig
	}

	return latencyConfig
}

func (h *HostConfig) Validate() error {
	if h.LatencyConfig != nil {
		if err := h.LatencyConfig.validate(); err != nil {
			return fmt.Errorf("invalid host config found: %v", err)
		}
	}

	sumPercentage := 0

	for errorCode, errorConfig := range h.ErrorsConfig {
		intErrorCode, err := strconv.Atoi(errorCode)

		if err != nil {
			return fmt.Errorf("invalid host config found: invalid error code: %v", err)
		}

		if intErrorCode < 400 || intErrorCode > 599 {
			return errors.New("invalid host config found: invalid error code: error should belong to either 4xx or 5xx classes")
		}

		if err := errorConfig.validate(); err != nil {
			return fmt.Errorf("invalid host config found: %v", err)
		}

		sumPercentage += *errorConfig.Percentage
	}

	if sumPercentage > 100 {
		return errors.New("invalid host config found: the sum of all percentages should not exceed 100")
	}

	for uri, uriConfig := range h.UrisConfig {
		if !util.UriRegex.MatchString(uri) {
			return errors.New("invalid host config provided: invalid uri config found: it doesn't match a uri pattern")
		}

		if err := uriConfig.validate(); err != nil {
			return fmt.Errorf("invalid host config found: %v", err)
		}
	}

	return nil
}

func (l *LatencyConfig) validate() error {
	hasMin := l.Min != nil
	hasP95 := l.P95 != nil
	hasP99 := l.P99 != nil
	hasMax := l.Max != nil

	if !hasMin || !hasMax {
		return errors.New("invalid latency config found: you should define at least 'min' and 'max'")
	}

	if *l.Min > *l.Max {
		return errors.New("invalid latency config found: min can not be greater than max")
	}

	if hasP95 && (*l.P95 < *l.Min || *l.P95 > *l.Max) {
		return errors.New("invalid latency config found: p95 can not be lesser than min or greater than max")
	}

	if hasP99 && (*l.P99 < *l.Min || *l.P99 < *l.P95 || *l.P99 > *l.Max) {
		return errors.New("invalid latency config found: p99 can not be lesser than min/p95 or greater than max")
	}

	return nil
}

func (e *ErrorConfig) validate() error {
	if e.Percentage == nil || *e.Percentage <= 0 || *e.Percentage > 100 {
		return errors.New("invalid error config found: percentage should be greater than 0 and lesser than 100")
	}

	if e.LatencyConfig == nil {
		return nil
	}

	if err := e.LatencyConfig.validate(); err != nil {
		return fmt.Errorf("invalid error config found: %v", err)
	}

	return nil
}

func (u *UriConfig) validate() error {
	if u.ErrorsConfig == nil && u.LatencyConfig == nil {
		return errors.New("invalid uri config found: latency or errors should not be both null")
	}

	if u.LatencyConfig != nil {
		if err := u.LatencyConfig.validate(); err != nil {
			return fmt.Errorf("invalid uri config found: %v", err)
		}
	}

	sumPercentage := 0

	for statusCode, errorConfig := range u.ErrorsConfig {
		intStatusCode, err := strconv.Atoi(statusCode)

		if err != nil {
			return fmt.Errorf("invalid uri config found: invalid status code: %v", err)
		}

		if intStatusCode < 400 || intStatusCode > 599 {
			return errors.New("invalid uri config found: error status code should be between 400 and 599")
		}

		if err = errorConfig.validate(); err != nil {
			return fmt.Errorf("invalid uri config found: %v", err)
		}

		sumPercentage += *errorConfig.Percentage
	}

	if sumPercentage > 100 {
		return errors.New("invalid uri config found: the sum of all percentages should not exceed 100")
	}

	return nil
}
