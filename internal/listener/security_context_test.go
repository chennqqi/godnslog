package listener

import (
	"net"
	"testing"
	"time"
)

func TestSecurityContext_CheckConnection(t *testing.T) {
	// Use a reasonable window for rate test: 1 minute, max 2 per IP
	rateLim := NewRateLimiter(time.Minute, 2)
	connLim := NewConnLimiter(2)
	sc := NewSecurityContext(rateLim, connLim, nil, nil, "test-listener", ProtocolSMTP, nil)

	addr1 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}
	addr2 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1235}
	addr3 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1236}
	addr4 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1237}

	// First two connections from same IP should be allowed (rate limit = 2)
	if !sc.CheckConnection(addr1) {
		t.Fatal("first connection should be allowed")
	}
	if !sc.CheckConnection(addr2) {
		t.Fatal("second connection should be allowed")
	}

	// Third connection from same IP should be denied by rate limit
	if sc.CheckConnection(addr3) {
		t.Fatal("third connection should be denied by rate limit")
	}

	// Release one slot
	sc.ReleaseConnection()

	// Different IP should still be allowed
	addrOther := &net.TCPAddr{IP: net.ParseIP("5.6.7.8"), Port: 1234}
	// Rate limit is per-IP, so different IP is fine for rate.
	// Conn limiter is global — we released 1, so 1 slot available.
	if !sc.CheckConnection(addrOther) {
		t.Fatal("connection from different IP should be allowed after release")
	}

	// Now conn limiter should be full again (2 active)
	if sc.CheckConnection(addr4) {
		t.Fatal("fourth connection should be denied by conn limit")
	}
}

func TestSecurityContext_Nil(t *testing.T) {
	var sc *SecurityContext

	// Nil security context should allow all
	if !sc.CheckConnection(&net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}) {
		t.Fatal("nil security context should allow all connections")
	}

	// Release should be no-op
	sc.ReleaseConnection()
}

func TestSecurityContext_NoopContext(t *testing.T) {
	sc := NoopSecurityContext()
	if !sc.CheckConnection(&net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}) {
		t.Fatal("noop security context should allow all connections")
	}
}

func TestSecurityContext_Whitelist(t *testing.T) {
	rateLim := NewRateLimiter(time.Minute, 1) // very strict: 1 per IP per minute
	connLim := NewConnLimiter(1)              // very strict: 1 concurrent
	sc := NewSecurityContext(
		rateLim, connLim,
		[]string{"10.0.0.0/8", "192.168.1.0/24"},
		nil, "test-wl", ProtocolSMTP, nil,
	)

	// Whitelisted IP should bypass rate limit
	wlAddr1 := &net.TCPAddr{IP: net.ParseIP("10.1.2.3"), Port: 1234}
	wlAddr2 := &net.TCPAddr{IP: net.ParseIP("10.1.2.3"), Port: 1235}
	if !sc.CheckConnection(wlAddr1) {
		t.Fatal("whitelisted IP first connection should be allowed")
	}
	sc.ReleaseConnection()
	if !sc.CheckConnection(wlAddr2) {
		t.Fatal("whitelisted IP second connection should bypass rate limit")
	}
	sc.ReleaseConnection()

	// Another whitelisted IP in different CIDR
	wlAddr3 := &net.TCPAddr{IP: net.ParseIP("192.168.1.50"), Port: 1234}
	if !sc.CheckConnection(wlAddr3) {
		t.Fatal("whitelisted IP in 192.168.1.0/24 should be allowed")
	}
	sc.ReleaseConnection()

	// Non-whitelisted IP should be rate limited
	nonWl1 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}
	nonWl2 := &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1235}
	if !sc.CheckConnection(nonWl1) {
		t.Fatal("non-whitelisted IP first connection should be allowed")
	}
	if sc.CheckConnection(nonWl2) {
		t.Fatal("non-whitelisted IP second connection should be rate limited")
	}
	// Release the first non-whitelisted connection
	sc.ReleaseConnection()
}

func TestSecurityContext_InvalidWhitelistCIDR(t *testing.T) {
	// Invalid CIDR should be silently skipped, not crash
	sc := NewSecurityContext(
		NewRateLimiter(time.Minute, 10),
		NewConnLimiter(10),
		[]string{"not-a-cidr", "10.0.0.0/8"},
		nil, "test-invalid-cidr", ProtocolSMTP, nil,
	)

	// Valid CIDR should still work
	if !sc.CheckConnection(&net.TCPAddr{IP: net.ParseIP("10.1.1.1"), Port: 1234}) {
		t.Fatal("valid whitelist CIDR should work")
	}
	sc.ReleaseConnection()
}

func TestSecurityContext_GetSetClear(t *testing.T) {
	// Get for nonexistent ID should return noop
	sc := GetSecurityContext("nonexistent-listener")
	if !sc.CheckConnection(&net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}) {
		t.Fatal("nonexistent listener should have noop security context")
	}

	// Set and get
	rateLim := NewRateLimiter(time.Minute, 10)
	connLim := NewConnLimiter(10)
	customSc := NewSecurityContext(rateLim, connLim, nil, nil, "test-listener-sc", ProtocolSMTP, nil)
	SetSecurityContext("test-listener-sc", customSc)

	retrieved := GetSecurityContext("test-listener-sc")
	if retrieved != customSc {
		t.Fatal("GetSecurityContext should return the same context that was set")
	}

	// Clear
	ClearSecurityContext("test-listener-sc")
	noop := GetSecurityContext("test-listener-sc")
	if !noop.CheckConnection(&net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}) {
		t.Fatal("after clear, should return noop context")
	}
}
