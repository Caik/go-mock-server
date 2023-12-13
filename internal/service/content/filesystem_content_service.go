package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	pathSeparator = string(os.PathSeparator)
	rootToken     = "root"
)

type filesystemContentService struct {
	mocksDirConfig *config.MocksDirectoryConfig
	once           sync.Once
	broadcaster    util.Broadcaster[ContentEvent]
}

func (f *filesystemContentService) GetContent(host, uri, method, uuid string) (*[]byte, error) {
	absolutePath, err := f.getFinalFilePath(host, uri, method)

	if err != nil {
		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error("error while generating final file path")

		return nil, err
	}

	data, err := os.ReadFile(absolutePath)

	if err != nil {
		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Info("mock not found")

		return nil, errors.New("mock not found")
	}

	return &data, err
}

func (f *filesystemContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	absolutePath, err := f.getFinalFilePath(host, uri, method)

	if err != nil {
		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error("error while generating final file path")

		return err
	}

	// making sure all parent dirs are created
	parentDir := absolutePath[:strings.LastIndex(absolutePath, pathSeparator)+1]
	err = os.MkdirAll(parentDir, os.ModePerm)

	if err != nil {
		msg := fmt.Sprintf("error while creating parent directories: %v", err)

		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error(msg)

		return errors.New(msg)
	}

	// writing file to disk
	err = os.WriteFile(absolutePath, *data, 0644)

	if err != nil {
		msg := fmt.Sprintf("error while writing file: %v", err)

		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error(msg)

		return errors.New(msg)
	}

	return nil
}

func (f *filesystemContentService) DeleteContent(host, uri, method, uuid string) error {
	absolutePath, err := f.getFinalFilePath(host, uri, method)

	if err != nil {
		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error("error while generating final file path")

		return err
	}

	if err = os.Remove(absolutePath); err != nil {
		msg := fmt.Sprintf("error while removing file: %v", err)

		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error(msg)

		return errors.New(msg)
	}

	return nil
}

func (f *filesystemContentService) ListContents(uuid string) (*[]ContentData, error) {
	if err := f.ensureInit(); err != nil {
		return nil, err
	}

	contents := make([]ContentData, 0)

	if err := filepath.Walk(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := f.filePathToContentData(path)

		if err == nil {
			contents = append(contents, *data)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error while listing contents: %v", err)
	}

	return &contents, nil
}

func (f *filesystemContentService) Subscribe(subscriberId string, eventTypes ...ContentEventType) <-chan ContentEvent {
	return f.broadcaster.Subscribe(subscriberId, func(event ContentEvent) bool {
		// if there's no filter being passed, allows all event types
		if len(eventTypes) == 0 {
			return true
		}

		for _, eventType := range eventTypes {
			if eventType == event.Type {
				return true
			}
		}

		return false
	})
}

func (f *filesystemContentService) Unsubscribe(subscriberId string) {
	f.broadcaster.Unsubscribe(subscriberId)
}

func (f *filesystemContentService) getFinalFilePath(host, uri, method string) (string, error) {
	if err := f.ensureInit(); err != nil {
		return "", err
	}

	parts := strings.Split(uri, "?")
	isRootPath := strings.HasSuffix(parts[0], "/")
	uriFixed := strings.ReplaceAll(parts[0], "/", pathSeparator)

	finalPath := strings.Join([]string{
		strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator),
		strings.Trim(host, pathSeparator),
		strings.TrimPrefix(uriFixed, pathSeparator)}, pathSeparator)

	if len(parts) > 1 {
		finalPath += "?" + parts[1]
	}

	if isRootPath && len(parts) == 1 {
		finalPath += rootToken
	}

	finalPath += "." + strings.ToLower(method)

	return finalPath, nil
}

func (f *filesystemContentService) filePathToContentData(path string) (*ContentData, error) {
	rootPath := strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator) + pathSeparator

	relativePath := strings.TrimPrefix(path, rootPath)

	firstSlashIndex := strings.Index(relativePath, pathSeparator)
	lastDotIndex := strings.LastIndex(relativePath, ".")

	if firstSlashIndex == -1 || lastDotIndex == -1 || firstSlashIndex >= lastDotIndex {
		return nil, fmt.Errorf("incorrect file name pattern, ignoring it: %s", path)
	}

	host := relativePath[:firstSlashIndex]
	uri := relativePath[firstSlashIndex:lastDotIndex]
	method := strings.ToUpper(relativePath[lastDotIndex+1:])

	// validating host
	if !util.HostRegex.MatchString(host) {
		return nil, fmt.Errorf("invalid host: %s", host)
	}

	// validating URI
	if !util.UriRegex.MatchString(uri) {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	// validating method
	if !util.HttpMethodRegex.MatchString(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	// checking if root suffix has been added
	if strings.HasSuffix(uri, fmt.Sprintf("%s%s", pathSeparator, rootToken)) {
		uri = strings.TrimSuffix(uri, rootToken)
	}

	data := ContentData{
		Host:   host,
		Uri:    uri,
		Method: method,
	}

	return &data, nil
}

func (f *filesystemContentService) ensureInit() error {
	if f.mocksDirConfig != nil {
		return nil
	}

	f.once.Do(func() {
		newDirConfig, err := config.GetMockDirectoryConfig()

		if err != nil {
			log.Error(fmt.Sprintf("error while getting mocks directory config: %v", err))
			return
		}

		f.mocksDirConfig = newDirConfig
	})

	if f.mocksDirConfig == nil {
		return errors.New("error while getting mocks directory config")
	}

	f.startContentWatcher()

	return nil
}

func (f *filesystemContentService) startContentWatcher() {
	// Watching host directories
	hostDirWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("error while starting new watcher: %v", err)
		return
	}

	if err := filepath.Walk(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		if strings.Compare(strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator), strings.TrimSuffix(path, pathSeparator)) == 0 {
			return nil
		}

		return hostDirWatcher.Add(path)
	}); err != nil {
		log.Errorf("error while watching host directories: %v", err)
		return
	}

	// Watching main mock directory to find new hosts created
	mockDirWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("error while starting new watcher: %v", err)

		hostDirWatcher.Close()

		return
	}

	if err = mockDirWatcher.Add(f.mocksDirConfig.Path); err != nil {
		log.Errorf("error while adding new path to be watched: %v", err)

		mockDirWatcher.Close()

		return
	}

	// starting to receive events from the host dir watcher
	go func() {
		for {
			select {
			case event, ok := <-hostDirWatcher.Events:
				if !ok {
					continue
				}

				// ignoring CHMOD changes
				if event.Has(fsnotify.Chmod) {
					continue
				}

				uuid := uuid.NewString()

				if err != nil {
					log.Error("error while generating new uuid")
					continue
				}

				log.WithField("uuid", uuid).
					WithField("operation", event.Op).
					WithField("filepath", event.Name).
					Info("received change from filesystem")

				var eventType ContentEventType

				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					eventType = Removed
				} else if event.Has(fsnotify.Create) {
					eventType = Created
				} else {
					eventType = Updated
				}

				data, err := f.filePathToContentData(event.Name)

				if err != nil {
					log.WithField("uuid", uuid).
						WithField("operation", event.Op).
						WithField("filepath", event.Name).
						Warn(fmt.Sprintf("error while converting filepath to content data: %v", err))

					continue
				}

				f.broadcaster.Publish(ContentEvent{Type: eventType, Data: *data}, uuid)

			case err, ok := <-hostDirWatcher.Errors:
				if !ok {
					continue
				}

				log.Errorf("error received while watching filesystem: %v", err)
			}
		}
	}()

	// starting to receive events from the mock dir watcher
	go func() {
		for {
			select {
			case event, ok := <-mockDirWatcher.Events:
				if !ok {
					continue
				}

				// ignoring CHMOD and REMOVE changes
				if event.Has(fsnotify.Chmod) || event.Has(fsnotify.Remove) {
					continue
				}

				uuid := uuid.NewString()

				if err != nil {
					log.Error("error while generating new uuid")
					continue
				}

				log.WithField("uuid", uuid).
					WithField("filepath", event.Name).
					Info("new host directory found, starting to watch it for changes")

				// adding new path to the host dir watcher
				hostDirWatcher.Add(event.Name)
			case err, ok := <-mockDirWatcher.Errors:
				if !ok {
					continue
				}

				log.Errorf("error received while watching filesystem: %v", err)
			}
		}
	}()
}

func newFilesystemContentService() *filesystemContentService {
	service := filesystemContentService{}
	service.ensureInit()

	return &service
}
