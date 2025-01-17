package cache

import (
	"github.com/rs/zerolog/log"
)

type InMemoryCacheService struct {
	cache map[string]*[]byte
}

func (l *InMemoryCacheService) Get(cacheKey, uuid string) (*[]byte, bool) {
	data, exists := l.cache[cacheKey]

	if exists {
		log.Info().
			Str("uuid", uuid).
			Str("cache_key", cacheKey).
			Msg("data retrieved from cache")
	}

	return data, exists
}

func (l *InMemoryCacheService) Set(cacheKey string, data *[]byte, uuid string) {
	if l.cache == nil {
		l.cache = make(map[string]*[]byte)
	}

	l.cache[cacheKey] = data

	log.Info().
		Str("uuid", uuid).
		Str("cache_key", cacheKey).
		Msg("data stored in cache")
}

func NewInMemoryCacheService() *InMemoryCacheService {
	return &InMemoryCacheService{
		cache: map[string]*[]byte{},
	}
}
