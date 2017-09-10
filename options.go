package cache

import "time"

// Option represents an option that can be used to customize a Cache being
// created.
type Option interface {
	modify(*options)
}

// WithCleanInterval sets the interval that 'clean' operations are run.
// Default: 10 seconds.
func WithCleanInterval(dur time.Duration) Option {
	return modifyFn(func(ops *options) {
		ops.cleanInterval = dur
	})
}

// WithExpirer sets the expiry method used by the cache during 'clean'
// operations.
func WithExpirer(e Expirer) Option {
	return modifyFn(func(ops *options) {
		ops.expirer = e
	})
}

// WithStartingSize creates the cache optimized to contain 'n' values.
func WithStartingSize(n int) Option {
	return modifyFn(func(ops *options) {
		ops.startingSize = n
	})
}

var defaultOptions = options{
	cleanInterval: 10 * time.Second,
	expirer:       NewExpirePartial(1000, 0.2),
}

type options struct {
	cleanInterval time.Duration
	expirer       Expirer
	startingSize  int
}

type modifyFn func(*options)

func (fn modifyFn) modify(ops *options) { fn(ops) }
