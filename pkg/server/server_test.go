package server

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	cache "github.com/patrickmn/go-cache"
)

func TestAddHit(t *testing.T) {
	hitServer := &HitServer{
		cache.New(cache.NoExpiration, 0*time.Second),
		PixelGifBytes(),
	}

	if len(hitServer.CacheItems()) != 0 {
		t.Errorf("Initial cache server has more than one item")
	}

	hitServer.addHit("test")
	testVal, found := hitServer.hitCache.Get("test")

	if !found || testVal != 1 {
		t.Errorf("Add hit not setting value to 1")
	}

	hitServer.addHit("test")
	testVal, found = hitServer.hitCache.Get("test")

	if !found || testVal != 2 {
		t.Errorf("Add hit not incrementing value to 2")
	}
}

func TestGetRequestHit(t *testing.T) {
	hitServer := &HitServer{nil, PixelGifBytes()}
	headers := http.Header{}
	baseURL, _ := url.Parse("https://example.com/pixel.gif")
	headers.Add("Referer", "referer")
	req := &http.Request{Header: headers, URL: baseURL}

	if hitServer.getRequestHit(req) != "referer" {
		t.Errorf("Hit server not parsing the hit from the referer")
	}

	paramURL, _ := url.Parse("https://example.com/?url=https://example.com/test/&url=test")
	req = &http.Request{URL: paramURL}

	if hitServer.getRequestHit(req) != "https://example.com/test/" {
		t.Errorf("Hit server not parsing the hit from the first `url` param")
	}

	req = &http.Request{URL: baseURL}

	if hitServer.getRequestHit(req) != defaultHit {
		t.Errorf("Hit server not parsing empty hit from URL")
	}

	headers = http.Header{}
	headers.Add("Referer", fmt.Sprintf("%s?param=1#fragment", baseURL))
	req = &http.Request{URL: baseURL, Header: headers}
	if hitServer.getRequestHit(req) != baseURL.String() {
		t.Errorf("Hit server not removing query param and fragment from URL")
	}
}

func TestHandlePixelRequest(t *testing.T) {
	hitServer := &HitServer{
		cache.New(cache.NoExpiration, 0*time.Second),
		PixelGifBytes(),
	}

	paramURL, _ := url.Parse("https://example.com/?url=https://example.com/")
	r := &http.Request{URL: paramURL}
	w := httptest.NewRecorder()

	hitServer.HandlePixelRequest(w, r)

	if len(hitServer.CacheItems()) != 1 {
		t.Errorf("HandlePixelRequest didn't correctly set cache")
	}

	contentLen := w.Header().Get("Content-Length")
	if contentLen != strconv.Itoa(len(PixelGifBytes())) {
		t.Errorf("Content length header set incorrectly to %s", contentLen)
	}

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	if !bytes.Equal(body, PixelGifBytes()) {
		t.Errorf("Response does not correctly return pixel body")
	}
}
