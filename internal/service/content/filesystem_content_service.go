package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	pathSeparator = string(os.PathSeparator)
	rootToken     = "root"
)

type FilesystemContentService struct {
	mocksDirConfig *config.MocksDirectoryConfig
	broadcaster    *util.Broadcaster[ContentEvent]
}

func (f *FilesystemContentService) GetContent(host, uri, method, uuid string, statusCode int) (*ContentResult, error) {
	absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absolutePath)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err // real I/O error — propagate it
	}

	if err == nil {
		return &ContentResult{
			Data:   &data,
			Source: "filesystem",
			Path:   absolutePath,
		}, nil
	}

	// file not found — try _default.<method>.<statusCode> fallback
	defaultPath, defErr := f.getDefaultFilePath(host, method, statusCode)

	if defErr == nil {
		defaultData, defReadErr := os.ReadFile(defaultPath)

		if defReadErr != nil && !errors.Is(defReadErr, os.ErrNotExist) {
			return nil, defReadErr // real I/O error on default file — propagate it
		}

		if defReadErr == nil {
			return &ContentResult{
				Data:   &defaultData,
				Source: "filesystem",
				Path:   defaultPath,
			}, nil
		}
	}

	// No specific mock or default found — return empty body
	log.Info().
		Str("uuid", uuid).
		Str("path", absolutePath).
		Msg("mock not found")

	empty := []byte("")

	return &ContentResult{
		Data:   &empty,
		Source: "filesystem",
		Path:   "",
	}, nil
}

func (f *FilesystemContentService) SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error {
	absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)

	if err != nil {
		return err
	}

	// making sure all parent dirs are created
	parentDir := absolutePath[:strings.LastIndex(absolutePath, pathSeparator)+1]
	err = os.MkdirAll(parentDir, os.ModePerm)

	if err != nil {
		msg := fmt.Sprintf("error while creating parent directories: %v", err)

		log.Err(err).
			Stack().
			Str("uuid", uuid).
			Str("path", absolutePath).
			Msg("error while creating parent directories")

		return errors.New(msg)
	}

	// writing file to disk
	err = os.WriteFile(absolutePath, *data, 0644)

	if err != nil {
		msg := fmt.Sprintf("error while writing file: %v", err)

		log.Err(err).
			Stack().
			Str("uuid", uuid).
			Str("path", absolutePath).
			Msg("error while writing file")

		return errors.New(msg)
	}

	return nil
}

func (f *FilesystemContentService) DeleteContent(host, uri, method, uuid string, statusCode int) error {
	absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)

	if err != nil {
		return err
	}

	if err := os.Remove(absolutePath); err != nil {
		msg := fmt.Sprintf("error while removing file: %v", err)

		log.Err(err).
			Stack().
			Str("uuid", uuid).
			Str("path", absolutePath).
			Msg("error while removing file")

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
		if err != nil {
			return nil
		}

		contents = append(contents, *data)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error while listing contents: %v", err)
	}

	return &contents, nil
}

func (f *FilesystemContentService) ListDefaultContents(uuid string) (*[]ContentData, error) {
	contents := make([]ContentData, 0)

	if err := f.retrieveContents(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := f.defaultFilePathToContentData(path)
		if err != nil {
			return nil
		}

		contents = append(contents, *data)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error while listing default contents: %v", err)
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

func (f *FilesystemContentService) getFinalFilePath(host, uri, method string, statusCode int) (string, error) {
	// Validate inputs before using them in a path expression.
	// The regexes allow only safe characters (no ".." or path separators in host,
	// no ".." in uri), breaking the taint chain before any path is constructed.
	if !util.HostRegex.MatchString(host) {
		return "", errors.New("invalid host")
	}

	if !util.HttpMethodRegex.MatchString(strings.ToUpper(method)) {
		return "", errors.New("invalid method")
	}

	parts := strings.SplitN(uri, "?", 2)
	uriPath := parts[0]

	// Root path "/" is valid but won't match UriRegex, so handle it explicitly.
	if uriPath != "/" && !util.UriRegex.MatchString(uriPath) {
		return "", errors.New("invalid uri")
	}

	isRootPath := strings.HasSuffix(uriPath, "/")
	uriFixed := strings.ReplaceAll(uriPath, "/", pathSeparator)

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

	finalPath += "." + strings.ToLower(method) + "." + strconv.Itoa(statusCode)

	// Verify the resolved path is within the mocks directory by computing the relative path
	mocksDir := filepath.Clean(f.mocksDirConfig.Path)
	rel, err := filepath.Rel(mocksDir, filepath.Clean(finalPath))

	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("invalid path: outside mocks directory")
	}

	// Reconstruct from the clean base so the returned value is not tainted.
	return filepath.Join(mocksDir, rel), nil
}

func (f *FilesystemContentService) getDefaultFilePath(host, method string, statusCode int) (string, error) {
	return f.getFinalFilePath(host, "/_default", method, statusCode)
}

func (f *FilesystemContentService) filePathToContentData(path string) (*ContentData, error) {
	rootPath := strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator) + pathSeparator
	relativePath := strings.TrimPrefix(path, rootPath)

	firstSlashIndex := strings.Index(relativePath, pathSeparator)

	// Skip _default.* files — these are fallbacks, not real mocks
	if firstSlashIndex != -1 {
		fileName := relativePath[firstSlashIndex+1:]

		if strings.HasPrefix(fileName, "_default.") {
			return nil, fmt.Errorf("skipping fallback file: %s", path)
		}
	}

	// Expect format: host/uri.method.status — two trailing dots
	lastDotIndex := strings.LastIndex(relativePath, ".")

	if lastDotIndex == -1 {
		return nil, fmt.Errorf("incorrect file name pattern, ignoring it: %s", path)
	}

	secondLastDotIndex := strings.LastIndex(relativePath[:lastDotIndex], ".")

	if firstSlashIndex == -1 || secondLastDotIndex == -1 || firstSlashIndex >= secondLastDotIndex {
		return nil, fmt.Errorf("incorrect file name pattern, ignoring it: %s", path)
	}

	host := relativePath[:firstSlashIndex]
	uri := relativePath[firstSlashIndex:secondLastDotIndex]
	method := strings.ToUpper(relativePath[secondLastDotIndex+1 : lastDotIndex])
	statusStr := relativePath[lastDotIndex+1:]

	statusCode, err := strconv.Atoi(statusStr)

	if err != nil || statusCode < 100 || statusCode > 599 {
		return nil, fmt.Errorf("invalid status code in filename: %s", path)
	}

	// validating host
	if !util.HostRegex.MatchString(host) {
		return nil, fmt.Errorf("invalid host: %s", host)
	}

	// validating URI — skip regex for root path
	if uri != "/" && !util.UriRegex.MatchString(uri) {
		return nil, fmt.Errorf("invalid uri: %s", uri)
	}

	// validating method
	if !util.HttpMethodRegex.MatchString(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	// checking if root suffix has been added (e.g. uri ends with /root → trim to /)
	if strings.HasSuffix(uri, fmt.Sprintf("%s%s", pathSeparator, rootToken)) {
		uri = strings.TrimSuffix(uri, rootToken)
	}

	return &ContentData{
		Host:       host,
		Uri:        uri,
		Method:     method,
		StatusCode: statusCode,
	}, nil
}

// defaultFilePathToContentData parses a _default.method.status filename into ContentData.
// Expected format: <mocksDir>/<host>/_default.<method>.<status>
func (f *FilesystemContentService) defaultFilePathToContentData(path string) (*ContentData, error) {
	rootPath := strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator) + pathSeparator
	relativePath := strings.TrimPrefix(path, rootPath)

	firstSlashIndex := strings.Index(relativePath, pathSeparator)
	if firstSlashIndex == -1 {
		return nil, fmt.Errorf("not a default file: %s", path)
	}

	fileName := relativePath[firstSlashIndex+1:]
	if !strings.HasPrefix(fileName, "_default.") {
		return nil, fmt.Errorf("not a default file: %s", path)
	}

	host := relativePath[:firstSlashIndex]

	// Parse _default.<method>.<status>
	rest := strings.TrimPrefix(fileName, "_default.")
	lastDotIndex := strings.LastIndex(rest, ".")
	if lastDotIndex == -1 {
		return nil, fmt.Errorf("invalid default filename: %s", path)
	}

	method := strings.ToUpper(rest[:lastDotIndex])
	statusStr := rest[lastDotIndex+1:]

	statusCode, err := strconv.Atoi(statusStr)
	if err != nil || statusCode < 100 || statusCode > 599 {
		return nil, fmt.Errorf("invalid status code in default filename: %s", path)
	}

	if !util.HostRegex.MatchString(host) {
		return nil, fmt.Errorf("invalid host in default filename: %s", host)
	}

	if !util.HttpMethodRegex.MatchString(method) {
		return nil, fmt.Errorf("invalid method in default filename: %s", method)
	}

	return &ContentData{
		Host:       host,
		Uri:        "/_default",
		Method:     method,
		StatusCode: statusCode,
	}, nil
}

func (f *FilesystemContentService) startContentWatcher() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Err(err).
			Stack().
			Msg("error while starting new watcher")

		return
	}

	if err := filepath.Walk(f.mocksDirConfig.Path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		return watcher.Add(path)
	}); err != nil {
		log.Err(err).
			Stack().
			Msg("error while watching host directories")

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

				log.Err(err).
					Stack().
					Msg("error received while watching filesystem")
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

	log.Info().
		Str("uuid", uuid).
		Str("operation", event.Op.String()).
		Str("filepath", event.Name).
		Msg("received change from filesystem")

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
			log.Warn().
				Str("uuid", uuid).
				Str("operation", event.Op.String()).
				Str("filepath", event.Name).
				Msgf("error while reading file info: %v", err)

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
					log.Info().
						Str("uuid", uuid).
						Str("filepath", path).
						Msg("new directory found, starting to watch it for changes")

					watcher.Add(path)

					return nil
				}

				if data, err := f.filePathToContentData(path); err == nil {
					f.broadcaster.Publish(ContentEvent{Type: eventType, Data: *data}, uuid)
				}

				return nil
			}); err != nil {
				log.Warn().
					Str("uuid", uuid).
					Str("operation", event.Op.String()).
					Str("filepath", event.Name).
					Msgf("error while retrieving contents of new directory: %v", err)
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
			log.Warn().
				Str("uuid", uuid).
				Str("operation", event.Op.String()).
				Str("filepath", event.Name).
				Msgf("error while converting filepath to content data: %v", err)
		}

		return
	}

	// publishing filesystem event
	f.broadcaster.Publish(ContentEvent{Type: eventType, Data: *data}, uuid)
}

func NewFilesystemContentService(mocksDirConfig *config.MocksDirectoryConfig) *FilesystemContentService {
	service := &FilesystemContentService{
		mocksDirConfig: mocksDirConfig,
		broadcaster:    &util.Broadcaster[ContentEvent]{},
	}

	service.startContentWatcher()

	return service
}
