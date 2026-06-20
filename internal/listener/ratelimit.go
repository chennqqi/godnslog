package listener

import (
	"sync"
	"time"
)

// RateLimiter implements a sliding-window per-IP connection rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxCount int
	hits     map[string][]time.Time
}

// NewRateLimiter creates a new RateLimiter with the given window and max connections per IP.
func NewRateLimiter(window time.Duration, maxCount int) *RateLimiter {
	return &RateLimiter{
		window:   window,
		maxCount: maxCount,
		hits:     make(map[string][]time.Time),
	}
}

// Allow returns true if the IP is within the rate limit, false otherwise.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter out expired entries
	var fresh []time.Time
	for _, t := range rl.hits[ip] {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}

	if len(fresh) >= rl.maxCount {
		rl.hits[ip] = fresh
		return false
	}

	fresh = append(fresh, now)
	rl.hits[ip] = fresh
	return true
}

// Cleanup removes expired entries for all IPs. Should be called periodically.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for ip, timestamps := range rl.hits {
		var fresh []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				fresh = append(fresh, t)
			}
		}
		if len(fresh) == 0 {
			delete(rl.hits, ip)
		} else {
			rl.hits[ip] = fresh
		}
	}
}

// ConnLimiter enforces a max concurrent connections counter.
type ConnLimiter struct {
	mu      sync.Mutex
	current int
	max     int
}

// NewConnLimiter creates a new ConnLimiter with the given max concurrent connections.
func NewConnLimiter(max int) *ConnLimiter {
	return &ConnLimiter{max: max}
}

// Acquire returns true if a connection slot is available, false if at capacity.
func (cl *ConnLimiter) Acquire() bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.current >= cl.max {
		return false
	}
	cl.current++
	return true
}

// Release frees a connection slot.
func (cl *ConnLimiter) Release() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.current > 0 {
		cl.current--
	}
}

// Current returns the current number of active connections.
func (cl *ConnLimiter) Current() int {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.current
}
