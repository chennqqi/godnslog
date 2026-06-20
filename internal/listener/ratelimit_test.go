package listener

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1*time.Minute, 3)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("1.2.3.4") {
		t.Fatal("4th request should be denied")
	}

	// Different IP should be allowed
	if !rl.Allow("5.6.7.8") {
		t.Fatal("different IP should be allowed")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 2)

	rl.Allow("1.2.3.4")
	rl.Allow("1.2.3.4")

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	rl.Cleanup()

	// After cleanup, should be allowed again
	if !rl.Allow("1.2.3.4") {
		t.Fatal("should be allowed after cleanup")
	}
}

func TestConnLimiter_AcquireRelease(t *testing.T) {
	cl := NewConnLimiter(2)

	if !cl.Acquire() {
		t.Fatal("first acquire should succeed")
	}
	if !cl.Acquire() {
		t.Fatal("second acquire should succeed")
	}
	if cl.Acquire() {
		t.Fatal("third acquire should fail (at capacity)")
	}

	if cl.Current() != 2 {
		t.Fatalf("expected current=2, got %d", cl.Current())
	}

	cl.Release()

	if cl.Current() != 1 {
		t.Fatalf("expected current=1 after release, got %d", cl.Current())
	}

	if !cl.Acquire() {
		t.Fatal("acquire after release should succeed")
	}
}

func TestConnLimiter_ReleaseAtZero(t *testing.T) {
	cl := NewConnLimiter(1)
	cl.Release() // should not panic
	if cl.Current() != 0 {
		t.Fatalf("expected current=0, got %d", cl.Current())
	}
}
