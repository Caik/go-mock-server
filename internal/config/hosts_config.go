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
	LatencyConfig  *LatencyConfig          `json:"latency"`
	StatusesConfig map[string]StatusConfig `json:"statuses"`
	UrisConfig     map[string]UriConfig    `json:"uris"`
}

type UriConfig struct {
	LatencyConfig  *LatencyConfig          `json:"latency"`
	StatusesConfig map[string]StatusConfig `json:"statuses"`
}

type LatencyConfig struct {
	Min *int `json:"min"`
	P95 *int `json:"p95"`
	P99 *int `json:"p99"`
	Max *int `json:"max"`
}

type StatusConfig struct {
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

func (h *HostsConfig) UpdateHostStatusesConfig(host string, statusesConfig map[string]StatusConfig) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.StatusesConfig = statusesConfig
	h.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func (h *HostsConfig) DeleteHostStatusConfig(host, statusCode string) (*HostConfig, error) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, nil
	}

	delete(hostConfig.StatusesConfig, statusCode)
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

func (h *HostsConfig) GetAppropriateStatusesConfig(host, uri string) (*map[string]StatusConfig, string) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, ""
	}

	statusesConfig := hostConfig.StatusesConfig
	scope := "Host Default"
	uriConfig, exists := hostConfig.UrisConfig[uri]

	if exists && len(uriConfig.StatusesConfig) > 0 {
		statusesConfig = uriConfig.StatusesConfig
		scope = "URI Override"
	}

	if len(statusesConfig) > 0 {
		return &statusesConfig, scope
	}

	return nil, ""
}

func (h *HostsConfig) GetAppropriateLatencyConfig(host, uri string) (*LatencyConfig, string) {
	hostConfig, exists := h.Hosts[host]

	if !exists {
		return nil, ""
	}

	latencyConfig := hostConfig.LatencyConfig
	scope := "Host Default"
	uriConfig, exists := hostConfig.UrisConfig[uri]

	if exists && uriConfig.LatencyConfig != nil {
		latencyConfig = uriConfig.LatencyConfig
		scope = "URI Override"
	}

	return latencyConfig, scope
}

func (h *HostConfig) Validate() error {
	if h.LatencyConfig != nil {
		if err := h.LatencyConfig.validate(); err != nil {
			return fmt.Errorf("invalid host config found: %v", err)
		}
	}

	sumPercentage := 0

	for statusCode, statusConfig := range h.StatusesConfig {
		intErrorCode, err := strconv.Atoi(statusCode)

		if err != nil {
			return fmt.Errorf("invalid host config found: invalid status code: %v", err)
		}

		if intErrorCode < 100 || intErrorCode > 599 {
			return errors.New("invalid host config found: invalid error code: status code must be between 100 and 599")
		}

		if err := statusConfig.validate(); err != nil {
			return fmt.Errorf("invalid host config found: %v", err)
		}

		sumPercentage += *statusConfig.Percentage
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

func (s *StatusConfig) validate() error {
	if s.Percentage == nil || *s.Percentage <= 0 || *s.Percentage > 100 {
		return errors.New("invalid status config found: percentage should be greater than 0 and lesser than 100")
	}

	if s.LatencyConfig == nil {
		return nil
	}

	if err := s.LatencyConfig.validate(); err != nil {
		return fmt.Errorf("invalid status config found: %v", err)
	}

	return nil
}

func (u *UriConfig) validate() error {
	if u.StatusesConfig == nil && u.LatencyConfig == nil {
		return errors.New("invalid uri config found: latency or statuses should not be both null")
	}

	if u.LatencyConfig != nil {
		if err := u.LatencyConfig.validate(); err != nil {
			return fmt.Errorf("invalid uri config found: %v", err)
		}
	}

	sumPercentage := 0

	for statusCode, statusConfig := range u.StatusesConfig {
		intStatusCode, err := strconv.Atoi(statusCode)

		if err != nil {
			return fmt.Errorf("invalid uri config found: invalid status code: %v", err)
		}

		if intStatusCode < 100 || intStatusCode > 599 {
			return errors.New("invalid uri config found: status code must be between 100 and 599")
		}

		if err = statusConfig.validate(); err != nil {
			return fmt.Errorf("invalid uri config found: %v", err)
		}

		sumPercentage += *statusConfig.Percentage
	}

	if sumPercentage > 100 {
		return errors.New("invalid uri config found: the sum of all percentages should not exceed 100")
	}

	return nil
}
