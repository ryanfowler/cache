package cache

import (
	"runtime"
	"time"
)

// Expirer represents an expiry technique used by a Cache.
type Expirer interface {
	lockedExpire(*Cache)
}

// NewExpireAll returns an Expirer that will iterate through all entries in the
// Cache, removing any that are expired.
func NewExpireAll() Expirer {
	return expireAll{}
}

type expireAll struct{}

func (e expireAll) lockedExpire(c *Cache) {
	lockedExpireAll(c.objs)
}

type expirePartial struct {
	batchSize     int
	continueRatio float64
}

// NewExpirePartial returns an Expirer that will iterate through a maximum of
// 'batchSize' entries, stopping only if less than 'continueRatio' entries were
// expired.
// The advantage to using this Expirer is that entries may be get/set in between
// batches. This makes this expiry method more performant for larger caches.
func NewExpirePartial(batchSize int, continueRatio float64) Expirer {
	if batchSize <= 0 {
		batchSize = 1
	}
	if continueRatio <= 0.0 {
		continueRatio = 0.01
	} else if continueRatio > 1.0 {
		continueRatio = 1.0
	}
	return expirePartial{
		batchSize:     batchSize,
		continueRatio: continueRatio,
	}
}

func (e expirePartial) lockedExpire(c *Cache) {
	if e.batchSize >= len(c.objs) {
		lockedExpireAll(c.objs)
		return
	}
	for {
		now := time.Now()
		if lockedExpireSome(now, e.batchSize, c.objs) < e.continueRatio {
			return
		}
		c.mu.Unlock()
		runtime.Gosched()
		c.mu.Lock()
		if c.closed {
			return
		}
	}
}

func lockedExpireAll(m map[string]value) {
	now := time.Now()
	for k, v := range m {
		if isExpired(now, v) {
			delete(m, k)
		}
	}
}

func lockedExpireSome(now time.Time, size int, m map[string]value) float64 {
	var count int
	var expired int
	for k, v := range m {
		if isExpired(now, v) {
			expired++
			delete(m, k)
		}
		count++
		if count >= size {
			break
		}
	}
	if count == 0 {
		return 0.0
	}
	return float64(expired) / float64(count)
}
