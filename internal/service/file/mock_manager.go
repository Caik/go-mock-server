package file

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Caik/go-mock-server/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	pathSeparator = string(os.PathSeparator)
)

var (
	mocksDirConfig *config.MocksDirectoryConfig
	once           sync.Once
)

func GetFinalFilePath(host, uri, method string) (string, error) {
	if err := ensureInit(); err != nil {
		return "", err
	}

	parts := strings.Split(uri, "?")
	isRootPath := strings.HasSuffix(parts[0], "/")
	uriFixed := strings.ReplaceAll(parts[0], "/", pathSeparator)

	finalPath := strings.Join([]string{
		strings.TrimSuffix(mocksDirConfig.Path, pathSeparator),
		strings.Trim(host, pathSeparator),
		strings.TrimPrefix(uriFixed, pathSeparator)}, pathSeparator)

	if len(parts) > 1 {
		finalPath += "?" + parts[1]
	}

	if isRootPath {
		finalPath += "root"
	}

	finalPath += "." + strings.ToLower(method)

	return finalPath, nil
}

func SaveUpdateFile(host, uri, method string, data *[]byte) error {
	absolutePath, err := GetFinalFilePath(host, uri, method)

	if err != nil {
		return err
	}

	// making sure all parent dirs are created
	parentDir := absolutePath[:strings.LastIndex(absolutePath, pathSeparator)+1]
	err = os.MkdirAll(parentDir, os.ModePerm)

	if err != nil {
		return fmt.Errorf("error while creating parent directories: %v", err)
	}

	// writing file to disk
	err = os.WriteFile(absolutePath, *data, 0644)

	if err != nil {
		return fmt.Errorf("error while writing file: %v", err)
	}

	return nil
}

func DeleteFile(host, uri, method string) error {
	absolutePath, err := GetFinalFilePath(host, uri, method)

	if err != nil {
		return err
	}

	//TODO implement
	if err = os.Remove(absolutePath); err != nil {
		return fmt.Errorf("error while removing file: %v", err)
	}

	return nil
}

func ensureInit() error {
	if mocksDirConfig != nil {
		return nil
	}

	once.Do(func() {
		newDirConfig, err := config.GetMockDirectoryConfig()

		if err != nil {
			log.Error(fmt.Sprintf("error while getting mocks directory config: %v", err))
			return
		}

		mocksDirConfig = newDirConfig
	})

	if mocksDirConfig == nil {
		return errors.New("error while getting mocks directory config")
	}

	return nil
}
