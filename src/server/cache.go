package main

import (
	"sync"
	"time"
	"github.com/miekg/dns"
)

type Cache interface {
	Get(reqType uint16, domain string) dns.RR
	Set(reqType uint16, domain string, ip dns.RR)
}

type CacheItem struct {
	Ip dns.RR
	Die time.Time
}

type MemoryCache struct {
	cache map[uint16]map[string]*CacheItem
	locker sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		cache: make(map[uint16]map[string]*CacheItem),
	}
	go cache.cleaner()

	return cache
}

func (c *MemoryCache) Get(reqType uint16, domain string) dns.RR {
	c.locker.RLock()
	defer c.locker.RUnlock()

	if m, ok := c.cache[reqType]; ok {
		if ip, ok := m[domain]; ok {
			if ip.Die.After(time.Now()) {
				return ip.Ip
			}
		}
	}

	return nil
}

func (c *MemoryCache) Set(reqType uint16, domain string, ip dns.RR) {
	c.locker.Lock()
	defer c.locker.Unlock()

	var m map[string]*CacheItem

	m, ok := c.cache[reqType]
	if !ok {
		m = make(map[string]*CacheItem)
		c.cache[reqType] = m
	}

	m[domain] = &CacheItem{
		Ip: ip,
		Die: time.Now().Add(time.Duration(ip.Header().Ttl) * time.Second),
	}
}

func (c *MemoryCache) cleaner() {
	for c != nil {
		c.locker.Lock()
		now := time.Now()

		for _, v := range c.cache {
			for k, vv := range v {
				if vv.Die.Before(now) {
					delete(v, k)
				}
			}
		}

		c.locker.Unlock()
		time.Sleep(1 * time.Minute)
	}
}