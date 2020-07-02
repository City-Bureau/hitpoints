package cache

import (
	"testing"
)

func TestAddHitGoCache(t *testing.T) {
	hitCache := NewHitGoCache()

	_ = hitCache.AddHit("test")
	testVal, found := hitCache.cache.Get("test")

	if !found || testVal != 1 {
		t.Errorf("Add hit not setting value to 1")
	}

	_ = hitCache.AddHit("test")
	testVal, found = hitCache.cache.Get("test")

	if !found || testVal != 2 {
		t.Errorf("Add hit not incrementing value to 2")
	}
}
