package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

var (
	once           sync.Once
	initialized    bool
	appConfig      *AppConfig
	hostsConfig    *HostsConfig
	mocksDirConfig *MocksDirectoryConfig
)

func Init() (*AppConfig, error) {
	var err error

	once.Do(func() {
		initLogger()
		initAppConfig()

		if err = initHostsConfig(appConfig.MocksConfigFile); err != nil {
			log.Error(fmt.Sprintf("error while initializing hosts config: %v", err))
			return
		}

		if err = initMocksDirectory(appConfig.MocksDirectory); err != nil {
			log.Error(fmt.Sprintf("error while initializing mocks directory: %v", err))
			return
		}

		initialized = true
	})

	if appConfig == nil {
		return nil, errors.New("invalid app config")
	}

	if hostsConfig == nil {
		return nil, errors.New("invalid hosts config")
	}

	if mocksDirConfig == nil {
		return nil, errors.New("invalid mocks directory config")
	}

	return appConfig, nil
}

func GetAppConfig() (*AppConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	return appConfig, nil
}

func GetMockDirectoryConfig() (*MocksDirectoryConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	return mocksDirConfig, nil
}

func GetHostsConfig() (*HostsConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	return hostsConfig, nil
}

func GetHostConfig(host string) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	return &hostConfig, nil
}

func SetHostConfig(host string, newConfig HostConfig) error {
	if !initialized {
		return errors.New("you need to initialize the config first")
	}

	hostsConfig.Hosts[host] = newConfig

	return nil
}

func DeleteHostConfig(host string) error {
	if !initialized {
		return errors.New("you need to initialize the config first")
	}

	delete(hostsConfig.Hosts, host)

	return nil
}

func UpdateHostLatencyConfig(host string, latencyConfig *LatencyConfig) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.LatencyConfig = latencyConfig
	hostsConfig.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func DeleteHostLatencyConfig(host string) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.LatencyConfig = nil
	hostsConfig.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func UpdateHostErrorsConfig(host string, errorsConfig map[string]ErrorConfig) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.ErrorsConfig = errorsConfig
	hostsConfig.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func DeleteHostErrorConfig(host, errorCode string) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	delete(hostConfig.ErrorsConfig, errorCode)
	hostsConfig.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func UpdateHostUrisConfig(host string, urisConfig map[string]UriConfig) (*HostConfig, error) {
	if !initialized {
		return nil, errors.New("you need to initialize the config first")
	}

	hostConfig, exists := hostsConfig.Hosts[host]

	if !exists {
		return nil, nil
	}

	hostConfig.UrisConfig = urisConfig
	hostsConfig.Hosts[host] = hostConfig

	return &hostConfig, nil
}

func initLogger() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func initAppConfig() {
	var newAppConfig AppConfig

	arg.MustParse(&newAppConfig)

	appConfig = &newAppConfig
}

func initHostsConfig(configFilePath string) error {
	// if config file is not passed, creating a new empty HostsConfig
	if len(configFilePath) == 0 {
		hostsConfig = &HostsConfig{
			Hosts: make(map[string]HostConfig),
		}

		return nil
	}

	absolutePath, err := filepath.Abs(configFilePath)

	if err != nil {
		return err
	}

	data, err := os.ReadFile(absolutePath)

	if err != nil {
		return err
	}

	var newHostsConfig HostsConfig

	err = json.Unmarshal(data, &newHostsConfig)

	if err != nil {
		return err
	}

	if err = newHostsConfig.Validate(); err != nil {
		return err
	}

	hostsConfig = &newHostsConfig

	return nil
}

func initMocksDirectory(mocksDirPath string) error {
	absolutePath, err := filepath.Abs(mocksDirPath)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(absolutePath, os.ModePerm); err != nil {
		return err
	}

	mocksDirConfig = &MocksDirectoryConfig{
		Path: absolutePath,
	}

	return nil
}
