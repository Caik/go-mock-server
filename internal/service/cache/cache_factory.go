package cache

import "sync"

var (
	once         sync.Once
	cacheService CacheService
)

func GetCacheService() CacheService {
	ensureInit()

	return cacheService
}

func ensureInit() {
	once.Do(func() {
		cacheService = &localCacheService{}
	})
}
