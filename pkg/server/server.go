package server

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	cache "github.com/patrickmn/go-cache"
)

// TODO: Save to file if it makes sense
const cachePath = "/tmp/hitpoints"

func loadCache() *cache.Cache {
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Println("Cache file not found, creating new cache...")
		return cache.New(cache.NoExpiration, 0*time.Second)
	}

	cacheMap := map[string]cache.Item{}
	buf := new(bytes.Buffer)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&cacheMap)
	if err != nil {
		log.Println("Unable to load cache from file, creating a new cache...")
		return cache.New(cache.NoExpiration, 0*time.Second)
	}
	log.Print("Successfully loaded cache from file")
	return cache.NewFrom(cache.NoExpiration, 0*time.Second, cacheMap)
}

func saveCache(hitCache *cache.Cache) error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(hitCache)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cachePath, buf.Bytes(), 0644)
}

// HitServer is the main struct for managing hitpoints
type HitServer struct {
	hitCache *cache.Cache
	pixel    []byte
}

// PixelGifBytes returns the content of the pixel gif in https://github.com/documentcloud/pixel-ping
func PixelGifBytes() []byte {
	return []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}
}

// NewHitServer creates a HitServer
func NewHitServer() HitServer {
	return HitServer{
		hitCache: loadCache(),
		pixel:    PixelGifBytes(),
	}
}

const defaultHit = "NA"

// Get hit value from request if there
func (s *HitServer) getRequestHit(r *http.Request) string {
	var reqURL string
	urlParams := r.URL.Query()["url"]

	if len(urlParams) > 0 {
		reqURL = urlParams[0]
	} else {
		reqURL = r.Referer()
	}

	parsedURL, err := url.Parse(reqURL)
	if reqURL == "" || err != nil {
		return defaultHit
	}

	// Remove query parameters and # fragments
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	return parsedURL.String()
}

// Add or increment cache key for URL hit
func (s *HitServer) addHit(url string) error {
	err := s.hitCache.Increment(url, 1)
	if err != nil {
		return s.hitCache.Add(url, 1, 0*time.Second)
	}
	return nil
}

// CacheItems cleans up all returned cache items
func (s *HitServer) CacheItems() map[string]int {
	hitMap := make(map[string]int)

	cacheItems := s.hitCache.Items()
	for k, v := range cacheItems {
		hitMap[k] = v.Object.(int)
	}
	return hitMap
}

// ClearCache flushes the HitServer cache
func (s *HitServer) ClearCache() {
	s.hitCache.Flush()
}

// HandlePixelRequest updates the cache and returns the pixel GIF
func (s *HitServer) HandlePixelRequest(w http.ResponseWriter, r *http.Request) {
	reqHit := s.getRequestHit(r)

	log.Println(r.URL)

	err := s.addHit(reqHit)
	if err != nil {
		http.Error(w, "There was an error processing your request", 500)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "private, no-cache, proxy-revalidate, max-age=0")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Content-Length", strconv.Itoa(len(s.pixel)))
	w.Write(s.pixel)
}
