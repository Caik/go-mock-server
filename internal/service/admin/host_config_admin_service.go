package admin

import (
	"fmt"
	"github.com/Caik/go-mock-server/internal/config"
)

type HostAddDeleteRequest struct {
	Host          string
	LatencyConfig *config.LatencyConfig
	ErrorConfig   map[string]config.ErrorConfig
	UriConfig     map[string]config.UriConfig
}

type HostsConfigAdminService struct {
	hostsConfig *config.HostsConfig
}

func (h *HostsConfigAdminService) GetHostsConfig() *config.HostsConfig {
	return h.hostsConfig
}

func (h *HostsConfigAdminService) GetHostConfig(host string) *config.HostConfig {
	return h.hostsConfig.GetHostConfig(host)
}

func (h *HostsConfigAdminService) AddUpdateHost(addRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	hostConfig := config.HostConfig{
		LatencyConfig: addRequest.LatencyConfig,
		ErrorsConfig:  addRequest.ErrorConfig,
		UrisConfig:    addRequest.UriConfig,
	}

	if err := hostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	h.hostsConfig.SetHostConfig(addRequest.Host, hostConfig)

	return &hostConfig, nil
}

func (h *HostsConfigAdminService) DeleteHost(host string) {
	h.hostsConfig.DeleteHostConfig(host)
}

func (h *HostsConfigAdminService) AddUpdateHostLatency(addLatencyRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		LatencyConfig: addLatencyRequest.LatencyConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	hostConfig, err := h.hostsConfig.UpdateHostLatencyConfig(addLatencyRequest.Host, addLatencyRequest.LatencyConfig)

	if err != nil {
		return nil, fmt.Errorf("error while updating host latency config: %v", err)
	}

	return hostConfig, nil
}

func (h *HostsConfigAdminService) DeleteHostLatency(host string) (*config.HostConfig, error) {
	hostConfig, err := h.hostsConfig.DeleteHostLatencyConfig(host)

	if err != nil {
		return nil, fmt.Errorf("error while deleting host latency config: %v", err)
	}

	return hostConfig, nil
}

func (h *HostsConfigAdminService) AddUpdateHostErrors(addErrorsRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		ErrorsConfig: addErrorsRequest.ErrorConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host errors config: %v", err)
	}

	hostConfig, err := h.hostsConfig.UpdateHostErrorsConfig(addErrorsRequest.Host, newHostConfig.ErrorsConfig)

	if err != nil {
		return nil, fmt.Errorf("error updating host errors config: %v", err)
	}

	return hostConfig, nil
}

func (h *HostsConfigAdminService) DeleteHostError(host, errorCode string) (*config.HostConfig, error) {
	hostConfig, err := h.hostsConfig.DeleteHostErrorConfig(host, errorCode)

	if err != nil {
		return nil, fmt.Errorf("error deleting host error config: %v", err)
	}

	return hostConfig, nil
}

func (h *HostsConfigAdminService) AddUpdateHostUris(addUrisRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		UrisConfig: addUrisRequest.UriConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	hostConfig, err := h.hostsConfig.UpdateHostUrisConfig(addUrisRequest.Host, addUrisRequest.UriConfig)

	if err != nil {
		return nil, fmt.Errorf("error while updating host uris config: %v", err)
	}

	return hostConfig, nil
}

func NewHostsConfigAdminService(hostsConfig *config.HostsConfig) *HostsConfigAdminService {
	return &HostsConfigAdminService{
		hostsConfig: hostsConfig,
	}
}
