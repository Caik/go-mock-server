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

func GetHostsConfig() (*config.HostsConfig, error) {
	hostsConfig, err := config.GetHostsConfig()

	if err != nil {
		return nil, fmt.Errorf("error while getting hosts config: %v", err)
	}

	return hostsConfig, nil
}

func GetHostConfig(host string) (*config.HostConfig, error) {
	hostConfig, err := config.GetHostConfig(host)

	if err != nil {
		return nil, fmt.Errorf("error while getting host config: %v", err)
	}

	return hostConfig, nil
}

func AddUpdateHost(addRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	hostConfig := config.HostConfig{
		LatencyConfig: addRequest.LatencyConfig,
		ErrorsConfig:  addRequest.ErrorConfig,
		UrisConfig:    addRequest.UriConfig,
	}

	if err := hostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	if err := config.SetHostConfig(addRequest.Host, hostConfig); err != nil {
		return nil, fmt.Errorf("error while setting host config: %v", err)
	}

	return &hostConfig, nil
}

func DeleteHost(host string) error {
	err := config.DeleteHostConfig(host)

	if err != nil {
		return fmt.Errorf("error while deleting host config: %v", err)
	}

	return nil
}

func AddUpdateHostLatency(addLatencyRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		LatencyConfig: addLatencyRequest.LatencyConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	hostConfig, err := config.UpdateHostLatencyConfig(addLatencyRequest.Host, addLatencyRequest.LatencyConfig)

	if err != nil {
		return nil, fmt.Errorf("error while updating host latency config: %v", err)
	}

	return hostConfig, nil
}

func DeleteHostLatency(host string) (*config.HostConfig, error) {
	hostConfig, err := config.DeleteHostLatencyConfig(host)

	if err != nil {
		return nil, fmt.Errorf("error while deleting host latency config: %v", err)
	}

	return hostConfig, nil
}

func AddUpdateHostErrors(addErrorsRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		ErrorsConfig: addErrorsRequest.ErrorConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host errors config: %v", err)
	}

	hostConfig, err := config.UpdateHostErrorsConfig(addErrorsRequest.Host, newHostConfig.ErrorsConfig)

	if err != nil {
		return nil, fmt.Errorf("error updating host errors config: %v", err)
	}

	return hostConfig, nil
}

func DeleteHostError(host, errorCode string) (*config.HostConfig, error) {
	hostConfig, err := config.DeleteHostErrorConfig(host, errorCode)

	if err != nil {
		return nil, fmt.Errorf("error deleting host error config: %v", err)
	}

	return hostConfig, nil
}

func AddUpdateHostUris(addUrisRequest HostAddDeleteRequest) (*config.HostConfig, error) {
	newHostConfig := config.HostConfig{
		UrisConfig: addUrisRequest.UriConfig,
	}

	if err := newHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("error while validating host config: %v", err)
	}

	hostConfig, err := config.UpdateHostUrisConfig(addUrisRequest.Host, addUrisRequest.UriConfig)

	if err != nil {
		return nil, fmt.Errorf("error while updating host uris config: %v", err)
	}

	return hostConfig, nil
}
