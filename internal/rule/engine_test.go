package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchCIDR_Exact32(t *testing.T) {
	assert.True(t, matchCIDR("192.168.1.1/32", "192.168.1.1"))
	assert.False(t, matchCIDR("192.168.1.1/32", "192.168.1.2"))
}

func TestMatchCIDR_24(t *testing.T) {
	assert.True(t, matchCIDR("192.168.1.0/24", "192.168.1.50"))
	assert.True(t, matchCIDR("192.168.1.0/24", "192.168.1.255"))
	assert.False(t, matchCIDR("192.168.1.0/24", "192.168.2.1"))
}

func TestMatchCIDR_10(t *testing.T) {
	assert.True(t, matchCIDR("10.0.0.0/8", "10.255.255.255"))
	assert.True(t, matchCIDR("10.0.0.0/8", "10.1.2.3"))
	assert.False(t, matchCIDR("10.0.0.0/8", "11.0.0.1"))
}

func TestMatchCIDR_InvalidCIDR(t *testing.T) {
	assert.False(t, matchCIDR("invalid", "192.168.1.1"))
}

func TestMatchCIDR_InvalidIP(t *testing.T) {
	assert.False(t, matchCIDR("192.168.1.0/24", "not-an-ip"))
}
