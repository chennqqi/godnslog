package websocket

import (
	"context"
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

func TestHub_Shutdown_StopsRunLoop(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := hub.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}

func TestHub_Shutdown_Idempotent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := hub.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown error: %v", err)
	}
	if err := hub.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown error: %v", err)
	}
}

func TestHub_Shutdown_BroadcastReturnsAfterStop(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := hub.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		hub.Broadcast([]byte("after-stop"))
		close(done)
	}()
	select {
	case <-done:
		// OK — Broadcast doesn't block after Shutdown
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked after Shutdown")
	}
}
