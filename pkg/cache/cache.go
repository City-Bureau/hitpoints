package cache

// HitCache is an interface for all cache backends
type HitCache interface {
	HandleHit(string)
	AddHit(string) error
	Items() map[string]int
	Clear()
	OnCron()
}
