package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

func InitLogger() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func ParseAppArguments() *AppArguments {
	var arguments AppArguments

	arg.MustParse(&arguments)

	return &arguments
}

func NewHostsConfig(appArguments *AppArguments) (*HostsConfig, error) {
	configFilePath := appArguments.MocksConfigFile

	// if config file is not passed, creating a new empty HostsConfig
	if len(configFilePath) == 0 {
		return &HostsConfig{
			Hosts: make(map[string]HostConfig),
		}, nil
	}

	absolutePath, err := filepath.Abs(configFilePath)

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absolutePath)

	if err != nil {
		return nil, err
	}

	var newHostsConfig HostsConfig

	err = json.Unmarshal(data, &newHostsConfig)

	if err != nil {
		return nil, err
	}

	if err = newHostsConfig.Validate(); err != nil {
		return nil, err
	}

	return &newHostsConfig, nil
}

func NewMocksDirectoryConfig(appArguments *AppArguments) (*MocksDirectoryConfig, error) {
	absolutePath, err := filepath.Abs(appArguments.MocksDirectory)

	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(absolutePath, os.ModePerm); err != nil {
		return nil, err
	}

	return &MocksDirectoryConfig{
		Path: absolutePath,
	}, nil
}
