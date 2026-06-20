package listener

import (
	"net"
	"testing"
)

func TestSecurityContext_CheckConnection(t *testing.T) {
	// Use a reasonable window for rate test: 1 minute, max 2 per IP
	rateLim := NewRateLimiter(1000000000*60, 2)
	connLim := NewConnLimiter(2)
	sc := NewSecurityContext(rateLim, connLim)

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
	// But rate limit is per-IP, so different IP is fine for rate
	// However conn limiter is global — we released 1, so 1 slot available
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

func TestSecurityContext_GetSetClear(t *testing.T) {
	// Get for nonexistent ID should return noop
	sc := GetSecurityContext("nonexistent-listener")
	if !sc.CheckConnection(&net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}) {
		t.Fatal("nonexistent listener should have noop security context")
	}

	// Set and get
	rateLim := NewRateLimiter(1000000000*60, 10)
	connLim := NewConnLimiter(10)
	customSc := NewSecurityContext(rateLim, connLim)
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
