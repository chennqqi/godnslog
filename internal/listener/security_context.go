package listener

import (
	"net"
	"sync"
)

// SecurityContext provides rate limiting and connection limiting for listeners.
// It is passed to each protocol listener and should be checked in the accept loop.
type SecurityContext struct {
	rateLim *RateLimiter
	connLim *ConnLimiter
}

// NewSecurityContext creates a new SecurityContext with the given limiters.
func NewSecurityContext(rateLim *RateLimiter, connLim *ConnLimiter) *SecurityContext {
	return &SecurityContext{
		rateLim: rateLim,
		connLim: connLim,
	}
}

// CheckConnection returns true if the connection from the given address should be
// accepted (within rate limit and under max concurrent connections).
// It acquires a connection slot; the caller MUST call ReleaseConnection when done.
func (sc *SecurityContext) CheckConnection(remoteAddr net.Addr) bool {
	if sc == nil {
		return true
	}

	ip := parseIPFromAddr(remoteAddr.String())

	// Check rate limit first
	if sc.rateLim != nil && !sc.rateLim.Allow(ip) {
		return false
	}

	// Check concurrent connection limit
	if sc.connLim != nil && !sc.connLim.Acquire() {
		return false
	}

	return true
}

// ReleaseConnection frees a connection slot. Must be called when a connection
// is closed if CheckConnection returned true.
func (sc *SecurityContext) ReleaseConnection() {
	if sc == nil || sc.connLim == nil {
		return
	}
	sc.connLim.Release()
}

// RateLimiter returns the underlying rate limiter (may be nil).
func (sc *SecurityContext) RateLimiter() *RateLimiter {
	if sc == nil {
		return nil
	}
	return sc.rateLim
}

// ConnLimiter returns the underlying connection limiter (may be nil).
func (sc *SecurityContext) ConnLimiter() *ConnLimiter {
	if sc == nil {
		return nil
	}
	return sc.connLim
}

// noopSecurityContext is a no-op security context for listeners without limits.
var noopSecurityContext = &SecurityContext{}

// NoopSecurityContext returns a SecurityContext that allows all connections.
func NoopSecurityContext() *SecurityContext {
	return noopSecurityContext
}

// securityContextPool provides per-listener security contexts.
// Using a pool avoids re-creating limiters on restart.
var securityContextPool sync.Map

// GetSecurityContext returns the SecurityContext for a listener ID,
// or a no-op context if none exists.
func GetSecurityContext(listenerID string) *SecurityContext {
	if v, ok := securityContextPool.Load(listenerID); ok {
		return v.(*SecurityContext)
	}
	return NoopSecurityContext()
}

// SetSecurityContext stores a SecurityContext for a listener ID.
func SetSecurityContext(listenerID string, sc *SecurityContext) {
	securityContextPool.Store(listenerID, sc)
}

// ClearSecurityContext removes the SecurityContext for a listener ID.
func ClearSecurityContext(listenerID string) {
	securityContextPool.Delete(listenerID)
}
