package cache

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	gocache "github.com/patrickmn/go-cache"
)

const cachePath = "/tmp/hitpoints"

// HitGoCache implements the HitCache interface for go-cache
type HitGoCache struct {
	cache *gocache.Cache
}

// NewHitGoCache is a constructor for HitGoCache
func NewHitGoCache() *HitGoCache {
	return &HitGoCache{
		cache: loadCache(),
	}
}

// HandleHit manages error handling in AddHit and runs in the worker
func (c *HitGoCache) HandleHit(hit string) {
	err := c.AddHit(hit)
	if err != nil {
		log.Println(err)
	}
}

// AddHit updates the cache with a URL value
func (c *HitGoCache) AddHit(hit string) error {
	err := c.cache.Increment(hit, 1)
	if err != nil {
		return c.cache.Add(hit, 1, 0*time.Second)
	}
	return nil
}

// Items returns a mapping of hits to integer counts
func (c *HitGoCache) Items() map[string]int {
	hitMap := make(map[string]int)

	cacheItems := c.cache.Items()
	for k, v := range cacheItems {
		hitMap[k] = v.Object.(int)
	}
	return hitMap
}

// Clear empties the cache
func (c *HitGoCache) Clear() {
	c.cache.Flush()
}

// OnCron saves the cache to disk
func (c *HitGoCache) OnCron() {
	log.Println("Saving cache to disk...")
	err := c.saveCache()
	if err != nil {
		log.Fatal(err)
	}
}

func loadCache() *gocache.Cache {
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Println("Cache file not found, creating new cache...")
		return cache.New(cache.NoExpiration, 0*time.Second)
	}

	cacheMap := map[string]gocache.Item{}
	buf := new(bytes.Buffer)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&cacheMap)
	if err != nil {
		log.Println("Unable to load cache from file, creating a new cache...")
		return gocache.New(gocache.NoExpiration, 0*time.Second)
	}
	log.Print("Successfully loaded cache from file")
	return gocache.NewFrom(gocache.NoExpiration, 0*time.Second, cacheMap)
}

// SaveCache saves the current HitServer cache to disk
func (c *HitGoCache) saveCache() error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(c.Items())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cachePath, buf.Bytes(), 0644)
}
