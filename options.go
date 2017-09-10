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
