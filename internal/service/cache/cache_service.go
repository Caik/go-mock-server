package cache

type CacheService interface {
	Get(cacheKey, uuid string) (*[]byte, bool)
	Set(cacheKey string, data *[]byte, uuid string)
}
