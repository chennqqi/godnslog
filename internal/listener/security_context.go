package listener

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SecurityContext provides rate limiting, connection limiting, and IP whitelisting
// for protocol listeners. It also records rejected connections as interactions so
// that analysts can see connection attempts that were blocked by limits.
type SecurityContext struct {
	rateLim    *RateLimiter
	connLim    *ConnLimiter
	whitelist  []*net.IPNet
	store      Store
	listenerID string
	protocol   Protocol
	logger     *logrus.Logger
}

// NewSecurityContext creates a new SecurityContext with the given limiters, whitelist,
// and store for recording rejected connections.
func NewSecurityContext(
	rateLim *RateLimiter,
	connLim *ConnLimiter,
	whitelistCIDRs []string,
	store Store,
	listenerID string,
	protocol Protocol,
	logger *logrus.Logger,
) *SecurityContext {
	sc := &SecurityContext{
		rateLim:    rateLim,
		connLim:    connLim,
		store:      store,
		listenerID: listenerID,
		protocol:   protocol,
		logger:     logger,
	}

	// Parse whitelist CIDRs
	for _, cidr := range whitelistCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			if logger != nil {
				logger.Warnf("[security_context] invalid whitelist CIDR %q: %v", cidr, err)
			}
			continue
		}
		sc.whitelist = append(sc.whitelist, ipNet)
	}

	return sc
}

// isWhitelisted returns true if the IP is in any whitelist CIDR.
func (sc *SecurityContext) isWhitelisted(ip string) bool {
	if len(sc.whitelist) == 0 {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, ipNet := range sc.whitelist {
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// CheckConnection returns true if the connection from the given address should be
// accepted (within rate limit and under max concurrent connections).
// It acquires a connection slot; the caller MUST call ReleaseConnection when done.
// If the connection is rejected, it records a "rejected" interaction so analysts
// can see the attempt.
func (sc *SecurityContext) CheckConnection(remoteAddr net.Addr) bool {
	if sc == nil {
		return true
	}

	ip := parseIPFromAddr(remoteAddr.String())

	// Whitelisted IPs bypass all limits
	if sc.isWhitelisted(ip) {
		// Still acquire a conn slot for accounting, but don't reject if full
		if sc.connLim != nil {
			sc.connLim.Acquire()
		}
		return true
	}

	// Check rate limit first
	if sc.rateLim != nil && !sc.rateLim.Allow(ip) {
		sc.recordRejectedConnection(ip, remoteAddr.String(), "rate_limited")
		return false
	}

	// Check concurrent connection limit
	if sc.connLim != nil && !sc.connLim.Acquire() {
		sc.recordRejectedConnection(ip, remoteAddr.String(), "conn_limit_reached")
		return false
	}

	return true
}

// recordRejectedConnection saves a ListenerInteraction record for a rejected connection
// so analysts can see that a callback attempt was made but blocked by limits.
func (sc *SecurityContext) recordRejectedConnection(ip, fullAddr, reason string) {
	if sc.store == nil {
		return
	}

	port := 0
	fmt.Sscanf(fullAddr, "%*s:%d", &port)

	interaction := &ListenerInteraction{
		ID:         generateInteractionID(),
		ListenerID: sc.listenerID,
		Protocol:   sc.protocol,
		SourceIP:   ip,
		SourcePort: port,
		Data:       fmt.Sprintf("connection rejected: %s", reason),
		Metadata:   Metadata{"reject_reason": reason},
		Timestamp:  time.Now(),
	}

	if err := sc.store.CreateListenerInteraction(nil, interaction); err != nil {
		if sc.logger != nil {
			sc.logger.Errorf("[security_context] failed to record rejected connection: %v", err)
		}
	}

	if sc.logger != nil {
		sc.logger.Infof("[security_context] connection from %s rejected (%s), recorded as interaction", ip, reason)
	}
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

// generateInteractionID generates a unique interaction ID for rejected connection records.
func generateInteractionID() string {
	return fmt.Sprintf("inter-%d", time.Now().UnixNano())
}
