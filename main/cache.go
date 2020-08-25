package main

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	NoExpiration = cache.NoExpiration
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration = cache.DefaultExpiration
)

type Cache struct {
	*cache.Cache

	rcdCh chan interface{}
}

func NewCache(def, interval time.Duration) *Cache {
	var c Cache
	{
		c.Cache = cache.New(AuthExpire, time.Minute)
		c.rcdCh = make(chan interface{}, 128)
	}
	return &c
}

func (self *Cache) Close() {
	close(self.rcdCh)
}

func (self *Cache) Input() chan<- interface{} {
	return self.rcdCh
}

func (self *Cache) Output() <-chan interface{} {
	return self.rcdCh
}
