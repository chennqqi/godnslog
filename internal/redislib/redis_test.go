package redislib

import (
	"testing"
)

func TestNewClientNil(t *testing.T) {
	c := NewClient(Config{Addr: ""})
	if c != nil {
		t.Error("expected nil client for empty addr")
	}
}

func TestKeyPrefix(t *testing.T) {
	c := NewClient(Config{Addr: "localhost:6379"})
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	key := c.Key("session:abc")
	expected := "godnslog:session:abc"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestRawNil(t *testing.T) {
	var c *Client
	if c.Raw() != nil {
		t.Error("expected nil raw client")
	}
}

func TestCloseNil(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestPingNil(t *testing.T) {
	var c *Client
	if err := c.Ping(nil); err == nil {
		t.Error("expected error for nil client")
	}
}
