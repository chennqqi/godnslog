package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsURLAllowed_AllowlistEmpty(t *testing.T) {
	s := NewOutboundSecurity(nil)
	assert.False(t, s.IsURLAllowed("https://example.com/hook"))
}

func TestIsURLAllowed_MatchAllowlist(t *testing.T) {
	s := NewOutboundSecurity([]string{"example.com", "hook.io"})
	assert.True(t, s.IsURLAllowed("https://example.com/webhook"))
	assert.True(t, s.IsURLAllowed("https://hook.io/incoming"))
	assert.False(t, s.IsURLAllowed("https://evil.com/exfil"))
}

func TestIsURLAllowed_Wildcard(t *testing.T) {
	s := NewOutboundSecurity([]string{"*"})
	assert.True(t, s.IsURLAllowed("https://example.com/webhook"))
	assert.True(t, s.IsURLAllowed("http://10.0.0.1/hook"))
}

func TestIsSSRFBlocked_PrivateIP(t *testing.T) {
	s := NewOutboundSecurity([]string{"*"})
	blocked := []string{
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
		"http://127.0.0.1/",
		"http://169.254.169.254/",
		"http://localhost/",
	}
	for _, u := range blocked {
		assert.True(t, s.IsSSRF(u), "expected SSRF block for %s", u)
	}
}

func TestIsSSRFNotBlocked_PublicIP(t *testing.T) {
	s := NewOutboundSecurity([]string{"*"})
	assert.False(t, s.IsSSRF("https://8.8.8.8/dns"))
	assert.False(t, s.IsSSRF("https://example.com/hook"))
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	s := NewOutboundSecurity([]string{"*"})
	actionID := "action-1"
	for i := 0; i < 10; i++ {
		assert.True(t, s.CheckRate(actionID), "request %d should be allowed", i)
	}
	assert.False(t, s.CheckRate(actionID), "11th request should be blocked")
}

func TestRateLimiter_ResetAfterWindow(t *testing.T) {
	s := NewOutboundSecurity([]string{"*"})
	s.rateWindow = 100 * time.Millisecond
	actionID := "action-2"
	for i := 0; i < 10; i++ {
		s.CheckRate(actionID)
	}
	assert.False(t, s.CheckRate(actionID))
	time.Sleep(110 * time.Millisecond)
	assert.True(t, s.CheckRate(actionID), "after window reset, request should be allowed")
}
