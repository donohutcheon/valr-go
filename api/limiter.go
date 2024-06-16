package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	// defaultRate specifies the rate at which requests are allowed.
	defaultRate = time.Minute
	// defaultMaxPerInterval specifies the number of requests allowed per interval.
	defaultMaxPerInterval = 1000
)

type Limiter interface {
	Wait(context.Context) error
}

type RateLimiter struct {
	cond           *sync.Cond
	requestCount   int
	rate           time.Duration
	maxPerInterval int
}

type RateLimiterOption func(limiter *RateLimiter)

func WithRate(rate time.Duration) RateLimiterOption {
	return func(limiter *RateLimiter) {
		limiter.rate = rate
	}
}

func WithMaxPerInterval(maxPerInterval int) RateLimiterOption {
	return func(limiter *RateLimiter) {
		limiter.maxPerInterval = maxPerInterval
	}
}

func NewRateLimiter(opts ...RateLimiterOption) *RateLimiter {
	mu := new(sync.Mutex)
	cond := sync.NewCond(mu)
	rl := &RateLimiter{
		cond:           cond,
		requestCount:   0,
		rate:           defaultRate,
		maxPerInterval: defaultMaxPerInterval,
	}

	for _, opt := range opts {
		opt(rl)
	}

	go func() {
		for {
			rl.resetCount()
		}
	}()

	return rl
}

func (l *RateLimiter) Wait(ctx context.Context) error {
	l.cond.L.Lock()
	defer l.cond.L.Unlock()

	if l.requestCount < l.maxPerInterval {
		l.requestCount++
		return nil
	}

	fmt.Printf("Rate limit exceeded. Waiting for reset\n")
	l.cond.Wait()

	return nil
}

func (l *RateLimiter) resetCount() {
	until := time.Until(nextReset(l.rate))
	time.Sleep(until)
	l.cond.L.Lock()
	defer l.cond.L.Unlock()
	l.requestCount = 0
	l.cond.Broadcast()
}

func nextReset(rate time.Duration) time.Time {
	now := time.Now()
	return now.Truncate(rate).Add(rate)
}
