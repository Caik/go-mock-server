package content

import (
	"errors"
	"fmt"
	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	pathSeparator = string(os.PathSeparator)
	rootToken     = "root"
)

type FilesystemContentService struct {
	mocksDirConfig *config.MocksDirectoryConfig
	broadcaster    util.Broadcaster[ContentEvent]
}

func (f *FilesystemContentService) GetContent(host, uri, method, uuid string) (*[]byte, error) {
	absolutePath := f.getFinalFilePath(host, uri, method)
	data, err := os.ReadFile(absolutePath)

	if err != nil {
		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Info("mock not found")

		return nil, errors.New("mock not found")
	}

	return &data, err
}

func (f *FilesystemContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	absolutePath := f.getFinalFilePath(host, uri, method)

	// making sure all parent dirs are created
	parentDir := absolutePath[:strings.LastIndex(absolutePath, pathSeparator)+1]
	err := os.MkdirAll(parentDir, os.ModePerm)

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

func (f *FilesystemContentService) DeleteContent(host, uri, method, uuid string) error {
	absolutePath := f.getFinalFilePath(host, uri, method)

	if err := os.Remove(absolutePath); err != nil {
		msg := fmt.Sprintf("error while removing file: %v", err)

		log.WithField("uuid", uuid).
			WithField("path", absolutePath).
			Error(msg)

		return errors.New(msg)
	}

	return nil
}

func (f *FilesystemContentService) ListContents(uuid string) (*[]ContentData, error) {
	contents := make([]ContentData, 0)

	if err := f.retrieveContents(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
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

func (f *FilesystemContentService) Subscribe(subscriberId string, eventTypes ...ContentEventType) <-chan ContentEvent {
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

func (f *FilesystemContentService) Unsubscribe(subscriberId string) {
	f.broadcaster.Unsubscribe(subscriberId)
}

func (f *FilesystemContentService) getFinalFilePath(host, uri, method string) string {
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

	return finalPath
}

func (f *FilesystemContentService) filePathToContentData(path string) (*ContentData, error) {
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

func (f *FilesystemContentService) startContentWatcher() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("error while starting new watcher: %v", err)
		return
	}

	if err := filepath.Walk(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		return watcher.Add(path)
	}); err != nil {
		log.Errorf("error while watching host directories: %v", err)
		return
	}

	// starting to receive events from the watcher
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				f.handleFilesystemEvent(event, watcher)

			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}

				log.Errorf("error received while watching filesystem: %v", err)
			}
		}
	}()
}

func (f *FilesystemContentService) retrieveContents(path string, fn func(path string, info fs.FileInfo, err error) error) error {
	if err := filepath.Walk(path, fn); err != nil {
		return fmt.Errorf("error while listing contents: %v", err)
	}

	return nil
}

func (f *FilesystemContentService) handleFilesystemEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	// ignoring CHMOD changes
	if event.Has(fsnotify.Chmod) {
		return
	}

	uuid := uuid.NewString()

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

	if eventType != Removed {
		// it's only possible to get stats from a file/dir in case it still exists, which wouldn't be the case when
		// a Removed event is sent
		info, err := os.Stat(event.Name)

		if err != nil {
			log.WithField("uuid", uuid).
				WithField("operation", event.Op).
				WithField("filepath", event.Name).
				Warn(fmt.Sprintf("error while reading file info: %v", err))

			return
		}

		// in case a dir is added/changed, we want to make sure that all its subdirectories are also watched and
		// any files found along the way are broadcast
		if info.IsDir() {
			// recursively adding new dirs to be watched
			if err = f.retrieveContents(event.Name, func(path string, fileInfo fs.FileInfo, err2 error) error {
				if err2 != nil {
					return err2
				}

				if fileInfo.IsDir() {
					log.WithField("uuid", uuid).
						WithField("filepath", path).
						Info("new directory found, starting to watch it for changes")

					watcher.Add(path)

					return nil
				}

				if data, err := f.filePathToContentData(path); err == nil {
					f.broadcaster.Publish(ContentEvent{Type: eventType, Data: *data}, uuid)
				}

				return nil
			}); err != nil {
				log.WithField("uuid", uuid).
					WithField("operation", event.Op).
					WithField("filepath", event.Name).
					Warn(fmt.Sprintf("error while retrieving contents of new directory: %v", err))
			}

			// given we know at this point that the event was a dir that was created/updated,
			// nothing else needs to be done apart from adding the new dirs (recursively) to the watch list
			// and publishing the files that were found
			return
		}

	}

	data, err := f.filePathToContentData(event.Name)

	if err != nil {
		// for remove events it might happen that the object deleted was a dir, and in that case an error would be thrown
		// It is not possible to know if the deleted object is a dir or a file in advance
		if eventType != Removed {
			log.WithField("uuid", uuid).
				WithField("operation", event.Op).
				WithField("filepath", event.Name).
				Warn(fmt.Sprintf("error while converting filepath to content data: %v", err))
		}

		return
	}

	// publishing filesystem event
	f.broadcaster.Publish(ContentEvent{Type: eventType, Data: *data}, uuid)
}

func NewFilesystemContentService(mocksDirConfig *config.MocksDirectoryConfig) *FilesystemContentService {
	var broadcaster util.Broadcaster[ContentEvent]

	service := FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
		broadcaster:    broadcaster,
	}

	service.startContentWatcher()

	return &service
}
