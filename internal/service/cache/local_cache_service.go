package cache

import (
	log "github.com/sirupsen/logrus"
)

type localCacheService struct {
	cache map[string]*[]byte
}

func (l localCacheService) Get(cacheKey, uuid string) (*[]byte, bool) {
	data, exists := l.cache[cacheKey]

	if exists {
		log.WithField("uuid", uuid).
			WithField("cache_key", cacheKey).
			Info("data retrieved from cache")
	}

	return data, exists
}

func (l *localCacheService) Set(cacheKey string, data *[]byte, uuid string) {
	if l.cache == nil {
		l.cache = make(map[string]*[]byte)
	}

	l.cache[cacheKey] = data

	log.WithField("uuid", uuid).
		WithField("cache_key", cacheKey).
		Info("data stored in cache")
}

func newLocalCacheService() *localCacheService {
	return &localCacheService{
		cache: map[string]*[]byte{},
	}
}
