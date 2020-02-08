package main

import (
	"time"

	"github.com/ReneKroon/ttlcache"
)

type Cache interface {
	SetWithTTL(key, val string, duration time.Duration)
	Get(key string) (interface{}, bool)
	Close()
}

type inMemCache struct {
	cache *ttlcache.Cache
}

func (i inMemCache) SetWithTTL(key, val string, dur time.Duration) {
	i.cache.SetWithTTL(key, val, dur)
}

func (i inMemCache) Get(key string) (interface{}, bool) {
	return i.cache.Get(key)
}

func (i inMemCache) Close() {
	i.cache.Close()
}

func NewImMemCache() Cache {
	return &inMemCache{cache: ttlcache.NewCache()}
}
