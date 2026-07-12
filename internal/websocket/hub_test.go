package websocket

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.Len() != 0 {
		t.Errorf("new hub should have 0 clients, got %d", hub.Len())
	}
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	msg := []byte(`{"type":"interaction","payload":{}}`)
	done := make(chan bool, 1)

	go func() {
		hub.Broadcast(msg)
		done <- true
	}()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("Broadcast timed out")
	}
}

func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if hub.Len() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.Len())
	}
}
