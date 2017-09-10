// MIT License
//
// Copyright (c) 2017 Ryan Fowler
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cache

import (
	"errors"
	"sync"
	"time"
)

// Cache is a in-memory cache of values keyed by strings that supports expiry.
type Cache struct {
	durClean time.Duration
	expirer  Expirer

	mu      sync.Mutex
	closed  bool
	chClean chan struct{}
	objs    map[string]value
}

type value struct {
	expireAt time.Time
	data     interface{}
}

// New returns an initialized cache using any provided option.
func New(ops ...Option) *Cache {
	op := defaultOptions
	for _, option := range ops {
		option.modify(&op)
	}

	var m map[string]value
	if op.startingSize > 0 {
		m = make(map[string]value, op.startingSize)
	} else {
		m = make(map[string]value)
	}
	return &Cache{
		durClean: op.cleanInterval,
		expirer:  op.expirer,
		objs:     m,
	}
}

// Get returns a value from the cache represented by the provided key.
func (c *Cache) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.objs[key]
	if !ok {
		return nil
	}
	if isExpired(time.Now(), v) {
		delete(c.objs, key)
		return nil
	}
	return v.data
}

// Len returns the current number of values in the cache.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.objs)
}

// SetEx sets the provided key and value, using 'exp' as the expiry duration.
func (c *Cache) SetEx(key string, val interface{}, exp time.Duration) {
	if val == nil || exp <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.objs[key] = value{expireAt: time.Now().Add(exp), data: val}
	if c.chClean == nil {
		c.chClean = make(chan struct{}, 1)
		go c.cleaner()
	}
}

// TTL returns the "time-to-live" of the value represented by 'key'. If nothing
// exists with the provided key, -1 is returned.
func (c *Cache) TTL(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.objs[key]
	if !ok {
		return -1
	}

	ttl := v.expireAt.Sub(time.Now())
	if ttl <= 0 {
		delete(c.objs, key)
		return -1
	}
	return ttl
}

func (c *Cache) cleaner() {
	t := time.NewTimer(c.durClean)
	defer t.Stop()
	for {
		select {
		case <-c.chClean:
		case <-t.C:
		}

		c.mu.Lock()

		// Check if cache is closed or no keys left to expire.
		if c.closed || len(c.objs) == 0 {
			c.chClean = nil
			c.mu.Unlock()
			return
		}

		c.expirer.lockedExpire(c)

		c.mu.Unlock()
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		t.Reset(c.durClean)
	}
}

func isExpired(now time.Time, v value) bool {
	return !v.expireAt.IsZero() && now.After(v.expireAt)
}

// ErrAlreadyClosed is the error returned from the Close method when the cache
// has already been closed.
var ErrAlreadyClosed = errors.New("cache: already closed")

// Close shuts down the cache, emptying it and preventing new values from being
// set.
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrAlreadyClosed
	}
	c.closed = true
	c.objs = nil
	if c.chClean != nil {
		select {
		case c.chClean <- struct{}{}:
		default:
		}
	}
	return nil
}
