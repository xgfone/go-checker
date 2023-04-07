// Copyright 2023 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package checker provides a checker to check whether a condition is ok
// periodically by the checker config strategy.
package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// DefaultInterval is the default check interval.
	DefaultInterval = time.Second * 10

	// DefaultConfig is the default checker config.
	DefaultConfig = Config{Failure: 1, Timeout: time.Second, Interval: DefaultInterval}

	// DefaultCondition is the default condition.
	DefaultCondition Condition = AlwaysTrue()
)

// Config is used to configure the checker.
type Config struct {
	// More than this, the checker will change the ok status from true to false.
	Failure  uint64        // The number that the condition fails.
	Timeout  time.Duration // The timeout duration to check the condition.
	Interval time.Duration // The interval duration between two checkers.
	Delay    time.Duration // The delay duration for the first start.
}

// Condition is used to check whether a condition is ok.
type Condition interface {
	Check(context.Context) (ok bool)
}

// ConditionFunc is a condition function.
type ConditionFunc func(context.Context) bool

// Check implements the interface Condition.
func (f ConditionFunc) Check(c context.Context) bool { return f(c) }

// AlwaysTrue returns a condition that returns true always.
func AlwaysTrue() Condition {
	return ConditionFunc(func(context.Context) bool { return true })
}

// AlwaysFalse returns a condition that returns false always.
func AlwaysFalse() Condition {
	return ConditionFunc(func(context.Context) bool { return false })
}

// Checker is used to check whether a condition is ok.
type Checker struct {
	ckid string
	ckcb func(string, bool)
	conf atomic.Value // Config
	cond atomic.Value // Condition

	ctxlock sync.Mutex
	cancelf context.CancelFunc
	fail    uint64
	ok      uint32

	jitter atomic.Value // func(interval time.Duration) time.Duration
}

// NewChecker returns a new condition checker with DefaultConfig.
//
// id is mandatory, and condition and callback are optional.
// If condition is equal to nil, use DefaultCondition instead.
// When the ok status has changed, callback will be called if it is set.
func NewChecker(id string, condition Condition, callback func(id string, ok bool)) *Checker {
	if id == "" {
		panic("Checker: the checker id must not be empty")
	}
	c := &Checker{ckid: id, ckcb: callback}
	c.SetCondition(condition)
	c.SetConfig(DefaultConfig)
	c.SetJitter(nil)
	return c
}

// ID returns the checker id.
func (c *Checker) ID() string { return c.ckid }

// Condition returns the condition.
func (c *Checker) Condition() Condition { return c.cond.Load().(Condition) }

// SetCondition resets the condition.
//
// If cond is nil, use DefaultCondition.
func (c *Checker) SetCondition(cond Condition) {
	if cond == nil {
		if cond = DefaultCondition; cond == nil {
			panic("Checker: the condition must not be nil")
		}
	}
	c.cond.Store(cond)
}

// Ok reports whether the checker status is ok.
func (c *Checker) Ok() bool { return atomic.LoadUint32(&c.ok) == 1 }

// SetOk sets the status to ok.
func (c *Checker) SetOk(ok bool) { c.updateStatus(ok, 0) }

// Config returns the config of the checker.
func (c *Checker) Config() Config { return c.conf.Load().(Config) }

// SetConfig resets the config of the checker.
func (c *Checker) SetConfig(config Config) {
	if config.Interval <= 0 {
		if DefaultInterval > 0 {
			config.Interval = DefaultInterval
		} else {
			config.Interval = time.Second * 10
		}
	}
	if config.Failure < 0 {
		config.Failure = 0
	}
	c.conf.Store(config)
}

// SetJitter sets the jitter function to adjust the interval duration
// to wait for the next check.
//
// If jitter is nil, clear the jitter function.
func (c *Checker) SetJitter(jitter func(interval time.Duration) time.Duration) {
	c.jitter.Store(jitter)
}

// Stop stops the checker, which does nothing if the checker has not been started.
func (c *Checker) Stop() {
	c.ctxlock.Lock()
	if c.cancelf != nil {
		c.cancelf()
		c.cancelf = nil
	}
	c.ctxlock.Unlock()
}

// Started reports whether the checker has already been started.
func (c *Checker) Started() (yes bool) {
	c.ctxlock.Lock()
	yes = c.cancelf != nil
	c.ctxlock.Unlock()
	return
}

// Start starts the checker until the context is done or the checker is stopped。
//
// NOTICE: it will panic if the checker has been started.
// The checker can be started more times, however, only if it is not started.
func (c *Checker) Start(ctx context.Context) {
	c.ctxlock.Lock()
	if c.cancelf != nil {
		c.ctxlock.Unlock()
		panic(fmt.Errorf("Checker: %s has been started", c.ckid))
	}

	ctx, c.cancelf = context.WithCancel(ctx)
	c.ctxlock.Unlock()
	defer c.Stop()

	if c.beforeStart(ctx) {
		c.loop(ctx)
	}
}

/*
func (c *Checker) loop(ctx context.Context) {
	interval := c.Config().Interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			config := c.Config()
			if config.Interval != interval {
				interval = config.Interval
				ticker.Reset(interval)
			}
			c.checkConfig(ctx, config)
		}
	}
}
*/

func (c *Checker) loop(ctx context.Context) {
	timer := time.NewTimer(c.getInterval(c.Config().Interval))
	defer func() {
		if timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			config := c.Config()
			timer = time.NewTimer(c.getInterval(config.Interval))
			c.checkConfig(ctx, config)
		}
	}
}

func (c *Checker) getInterval(interval time.Duration) time.Duration {
	if f := c.jitter.Load().(func(time.Duration) time.Duration); f != nil {
		interval = f(interval)
	}
	return interval
}

func (c *Checker) beforeStart(ctx context.Context) (ok bool) {
	config := c.Config()
	if config.Delay > 0 {
		wait := time.NewTimer(config.Delay)
		select {
		case <-wait.C:
		case <-ctx.Done():
			wait.Stop()
			return false
		}
	}
	c.checkConfig(ctx, config)
	return true
}

func (c *Checker) checkConfig(ctx context.Context, config Config) {
	defer c.wrapPanic()
	c.updateStatus(c.checkCondtion(ctx, config), config.Failure)
}

func (c *Checker) wrapPanic() {
	if r := recover(); r != nil {
		log.Printf("panic when checker '%s' checks the condition: %v", c.ckid, r)
	}
}

func (c *Checker) checkCondtion(ctx context.Context, config Config) (ok bool) {
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	return c.Condition().Check(ctx)
}

func (c *Checker) updateStatus(success bool, failure uint64) {
	var changed bool
	if success {
		c.ctxlock.Lock()
		if c.fail > 0 {
			c.fail = 0
		}
		c.ctxlock.Unlock()
		changed = atomic.CompareAndSwapUint32(&c.ok, 0, 1)
	} else {
		switch {
		case failure == 0:
			c.ctxlock.Lock()
			if c.fail > 0 {
				c.fail = 0
			}
			c.ctxlock.Unlock()
			changed = atomic.CompareAndSwapUint32(&c.ok, 1, 0)
		case failure > 0:
			c.ctxlock.Lock()
			c.fail++
			fail := c.fail
			c.ctxlock.Unlock()

			if fail > failure {
				changed = atomic.CompareAndSwapUint32(&c.ok, 1, 0)
			}
		}
	}

	if changed && c.ckcb != nil {
		c.ckcb(c.ckid, success)
	}
	return
}
