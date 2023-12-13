package content

import (
	"sync"
)

var (
	once           sync.Once
	contentService ContentService
)

func GetContentService() ContentService {
	ensureInit()

	return contentService
}

func ensureInit() {
	if contentService != nil {
		return
	}

	once.Do(func() {
		contentService = newFilesystemContentService()
	})
}
