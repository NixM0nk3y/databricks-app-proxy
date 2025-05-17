package caching

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type AppCache struct {
	_cache *cache.Cache
}

const (
	defaultExpiration = 2 * time.Minute
	purgeTime         = 5 * time.Minute
)

func NewAppCache() *AppCache {
	Cache := cache.New(defaultExpiration, purgeTime)
	return &AppCache{
		_cache: Cache,
	}
}

func (c *AppCache) Read(cachekey string) (cachevalue interface{}, ok bool) {
	cachevalue, ok = c._cache.Get(cachekey)
	if ok {
		return cachevalue, true
	}
	return cachevalue, false
}

func (c *AppCache) Update(cachekey string, value interface{}, expire time.Duration) {
	c._cache.Set(cachekey, value, expire)
}
