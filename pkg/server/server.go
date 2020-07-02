package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	rice "github.com/GeertJohan/go.rice"
)

const workerCount = 8

const chanBuffer = 8

const defaultHit = "NA"

// HitServer is the main struct for managing hitpoints
type HitServer struct {
	pixel []byte
	hits  chan string
}

// PixelGifBytes returns the content of the pixel gif in https://github.com/documentcloud/pixel-ping
func PixelGifBytes() []byte {
	return []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}
}

// NewHitServer creates a HitServer
func NewHitServer() HitServer {
	return HitServer{
		pixel: PixelGifBytes(),
		hits:  make(chan string, chanBuffer),
	}
}

// StartWorker begins accepting hits on the hits channel and caching them
func (s *HitServer) StartWorker(hitFunc func(string)) {
	log.Println("Starting worker pool...")
	for i := 0; i < workerCount; i++ {
		go func() {
			for hit := range s.hits {
				hitFunc(hit)
			}
		}()
	}
}

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

// HandlePixelRequest sends a message to the hits worker and returns the pixel GIF
func (s *HitServer) HandlePixelRequest(w http.ResponseWriter, r *http.Request) {
	s.hits <- s.getRequestHit(r)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "private, no-cache, proxy-revalidate, max-age=0")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Content-Length", strconv.Itoa(len(s.pixel)))
	_, _ = w.Write(s.pixel)
}

// HandleJS creates a handler function for the JS file used to count hits
func (s *HitServer) HandleJS(domain string, ssl bool) func(http.ResponseWriter, *http.Request) {
	jsBox, err := rice.FindBox("../../static")
	if err != nil {
		log.Fatal(err)
	}

	jsTmpl, err := jsBox.String("hitpoints.js")
	if err != nil {
		log.Fatal(err)
	}

	var scheme string
	if ssl {
		scheme = "https"
	} else {
		scheme = "http"
	}

	jsStr := fmt.Sprintf(jsTmpl, scheme, domain)
	jsBytes := []byte(jsStr)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(jsBytes)))
		_, _ = w.Write(jsBytes)
	}
}
