# Phase 2.8 Spec A: Notification Enhancement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add HTTP timeout to all 8 notification channels, escape Telegram Markdown special characters, and add WebSocket Hub graceful shutdown.

**Architecture:** Inject a shared `*http.Client` (30s timeout) into `notification.Service` via functional options, mirroring `internal/rule/action.go`'s Executor pattern. Add a Telegram Markdown escaper applied before message formatting. Add `Hub.Shutdown(ctx)` that stops the event loop and closes all clients, wired into `WebServer.Shutdown`.

**Tech Stack:** Go 1.25, xorm, gorilla/websocket, net/http, archive-free (notification only), testing with `httptest.NewServer`.

## Global Constraints

- Module path: `github.com/chennqqi/godnslog`
- Test command: `GOCACHE=/tmp/gocache go test ./...`
- HTTP timeout default: 30 seconds (matching `internal/rule/action.go` and `internal/workflow/service.go`)
- Notification tests use plain `t.Fatal`/`t.Errorf` (no testify) — see `internal/notification/service_test.go`
- WebSocket tests use plain `t.Fatal`/`t.Errorf` — see `internal/websocket/hub_test.go`
- Do not break existing `notification.NewService(engine)` callers without updating them

---

## File Structure

| File | Responsibility |
|------|----------------|
| `internal/notification/service.go` | Add `httpClient` field, options, refactor 8 send methods, add Telegram escaper |
| `internal/notification/service_test.go` | Add httptest-based timeout + telegram escape tests |
| `internal/websocket/hub.go` | Add `Shutdown(ctx)` method + stop channel |
| `internal/websocket/client.go` | Add WaitGroup tracking for write pumps |
| `internal/websocket/hub_test.go` | Add shutdown tests |
| `server/websocket.go` | Add `stopWS()` helper |
| `server/webserver.go` | Call hub shutdown in `WebServer.Shutdown` |
| `server/v2_api.go` | Update 6 `notification.NewService` call sites |

---

### Task 1: Add HTTP client with timeout to notification.Service

**Files:**
- Modify: `internal/notification/service.go:17-25`
- Test: `internal/notification/service_test.go`

**Interfaces:**
- Produces: `NewService(engine *xorm.Engine, opts ...Option) *Service`, `WithHTTPTimeout(d time.Duration) Option`, `Service.httpClient *http.Client` field

- [ ] **Step 1: Write the failing test**

Add to `internal/notification/service_test.go`:

```go
func TestService_HTTPTimeoutOption(t *testing.T) {
	engine := setupNotificationEngine(t)
	svc := NewService(engine, WithHTTPTimeout(5*time.Second))
	if svc.httpClient == nil {
		t.Fatal("httpClient should be set after WithHTTPTimeout option")
	}
	if svc.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", svc.httpClient.Timeout)
	}
}

func TestService_DefaultHTTPTimeout(t *testing.T) {
	engine := setupNotificationEngine(t)
	svc := NewService(engine)
	if svc.httpClient == nil {
		t.Fatal("httpClient should be set by default")
	}
	if svc.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", svc.httpClient.Timeout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run 'TestService_HTTPTimeoutOption|TestService_DefaultHTTPTimeout' -v`
Expected: FAIL — `svc.httpClient` undefined, `WithHTTPTimeout` undefined.

- [ ] **Step 3: Write minimal implementation**

Replace the struct + constructor in `internal/notification/service.go` (lines 17-25) with:

```go
// Option configures a notification Service.
type Option func(*Service)

// WithHTTPTimeout sets the HTTP client timeout for outbound notification calls.
func WithHTTPTimeout(d time.Duration) Option {
	return func(s *Service) {
		s.httpClient.Timeout = d
	}
}

// Service handles notification operations
type Service struct {
	engine     *xorm.Engine
	httpClient *http.Client
}

// NewService creates a new notification service.
// Default HTTP timeout is 30 seconds; override with WithHTTPTimeout.
func NewService(engine *xorm.Engine, opts ...Option) *Service {
	s := &Service{
		engine: engine,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run 'TestService_HTTPTimeoutOption|TestService_DefaultHTTPTimeout' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/notification/service.go internal/notification/service_test.go
git commit -m "feat(notification): add configurable HTTP client with default 30s timeout"
```

---

### Task 2: Refactor all 8 send methods to use httpClient

**Files:**
- Modify: `internal/notification/service.go:156-409`
- Test: `internal/notification/service_test.go`

**Interfaces:**
- Consumes: `Service.httpClient` from Task 1
- Produces: all `sendXxx` methods use `s.httpClient` instead of `http.Post`

- [ ] **Step 1: Write the failing test**

Add to `internal/notification/service_test.go` — a test that asserts a webhook send respects the configured timeout by hitting a slow server. Uses `httptest`:

```go
import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/models"
)

func TestService_SendWebhook_RespectsTimeout(t *testing.T) {
	engine := setupNotificationEngine(t)

	// Server that sleeps longer than the configured timeout.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// 200ms timeout — must fail before the 2s server responds.
	svc := NewService(engine, WithHTTPTimeout(200*time.Millisecond))

	// Insert a channel pointing at the slow server.
	channel := &models.TblNotificationChannel{
		Name: "slow-webhook", Type: "webhook",
		Config: `{"url":"` + srv.URL + `"}`,
		Enabled: true, CreatedBy: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if _, err := engine.Insert(channel); err != nil {
		t.Fatalf("insert channel: %v", err)
	}

	start := time.Now()
	err := svc.SendNotification(channel.Id, "test", "payload")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed > 1*time.Second {
		t.Errorf("expected to fail within ~200ms, took %v", elapsed)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run TestService_SendWebhook_RespectsTimeout -v`
Expected: FAIL — `err == nil` because current `http.Post` has no timeout and waits the full 2s.

- [ ] **Step 3: Refactor send methods to use httpClient**

In `internal/notification/service.go`, replace each `http.Post(url, ct, body)` call with `s.httpClient.Post(url, ct, body)` and each `http.PostForm(url, data)` with `s.httpClient.PostForm(url, data)`.

Affected methods and lines (before refactor):
- `sendWebhook` line 172: `resp, err := http.Post(webhookConfig.URL, ...)`
- `sendWechat` line 202: `resp, err := http.Post(wechatConfig.WebhookURL, ...)`
- `sendFeishu` line 232: `resp, err := http.Post(feishuConfig.WebhookURL, ...)`
- `sendDingtalk` line 263: `resp, err := http.Post(dingtalkConfig.WebhookURL, ...)`
- `sendBark` line 296: `resp, err := http.Post(cfg.URL, ...)`
- `sendServerchan` line 325: `resp, err := http.PostForm(u, formData)`
- `sendTelegram` line 358: `resp, err := http.Post(u, ...)`
- `sendSlack` line 400: `resp, err := http.Post(cfg.WebhookURL, ...)`

For each, change `http.Post` → `s.httpClient.Post` and `http.PostForm` → `s.httpClient.PostForm`. Example for `sendWebhook`:

```go
resp, err := s.httpClient.Post(webhookConfig.URL, "application/json", bytes.NewBuffer(jsonBody))
```

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run TestService_SendWebhook_RespectsTimeout -v`
Expected: PASS — request fails within ~200ms with timeout error.

- [ ] **Step 5: Run full notification test suite**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -v`
Expected: All tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/notification/service.go internal/notification/service_test.go
git commit -m "refactor(notification): use shared httpClient for all 8 send methods"
```

---

### Task 3: Telegram Markdown escaping

**Files:**
- Modify: `internal/notification/service.go:336-367`
- Test: `internal/notification/service_test.go`

**Interfaces:**
- Produces: `escapeTelegramMarkdown(s string) string` function

- [ ] **Step 1: Write the failing test**

Add to `internal/notification/service_test.go`:

```go
func TestEscapeTelegramMarkdown(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"plain", "hello world", "hello world"},
		{"underscore", "a_b", `a\_b`},
		{"asterisk", "a*b", `a\*b`},
		{"brackets", "a[b]c(d)", `a\[b\]c\(d\)`},
		{"tilde", "a~b", `a\~b`},
		{"backtick", "a`b", `a\`b`},
		{"hash", "#tag", `\#tag`},
		{"plus_minus_eq", "a+b-c=d", `a\+b\-c\=d`},
		{"pipe_braces", "a|b{c}d", `a\|b\{c\}d`},
		{"dot_bang", "a.b!c", `a\.b\!c`},
		{"gt", ">quote", `\>quote`},
		{"all_chars", "_*[]()~`>#+-=|{}.!", `\_\*\[\]\(\)\~\`\>\#\+\-\=\|\{\}\.\!`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeTelegramMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("escapeTelegramMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run TestEscapeTelegramMarkdown -v`
Expected: FAIL — `escapeTelegramMarkdown` undefined.

- [ ] **Step 3: Add the escaper and apply it in sendTelegram**

Add near the top of `internal/notification/service.go` (after imports), the escaper:

```go
// telegramMarkdownReplacer escapes characters with special meaning in
// Telegram's Markdown parse_mode so user-supplied text renders literally.
var telegramMarkdownReplacer = strings.NewReplacer(
	"_", `\_`,
	"*", `\*`,
	"[", `\[`,
	"]", `\]`,
	"(", `\(`,
	")", `\)`,
	"~", `\~`,
	"`", "\\`",
	">", `\>`,
	"#", `\#`,
	"+", `\+`,
	"-", `\-`,
	"=", `\=`,
	"|", `\|`,
	"{", `\{`,
	"}", `\}`,
	".", `\.`,
	"!", `\!`,
)

// escapeTelegramMarkdown escapes Markdown special characters per Telegram's
// parse_mode=Markdown rules.
func escapeTelegramMarkdown(s string) string {
	return telegramMarkdownReplacer.Replace(s)
}
```

Add `"strings"` to the import block if not present.

Then modify `sendTelegram` (around line 349) to escape before formatting:

```go
text := fmt.Sprintf("*GODNSLOG* %s\n\n%s",
	escapeTelegramMarkdown(message),
	escapeTelegramMarkdown(payload))
```

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... -run TestEscapeTelegramMarkdown -v`
Expected: PASS

- [ ] **Step 5: Run full notification test suite + build**

Run: `GOCACHE=/tmp/gocache go test ./internal/notification/... && go build ./...`
Expected: PASS, build OK.

- [ ] **Step 6: Commit**

```bash
git add internal/notification/service.go internal/notification/service_test.go
git commit -m "fix(notification): escape Telegram Markdown special characters in message and payload"
```

---

### Task 4: Add WebSocket Hub Shutdown method

**Files:**
- Modify: `internal/websocket/hub.go`
- Modify: `internal/websocket/client.go`
- Test: `internal/websocket/hub_test.go`

**Interfaces:**
- Produces: `Hub.Shutdown(ctx context.Context) error`, `Hub` gains `stop chan struct{}` and `wg sync.WaitGroup` fields, `Client.WritePump` registers with the WaitGroup

- [ ] **Step 1: Write the failing test**

Add to `internal/websocket/hub_test.go`:

```go
package websocket

import (
	"context"
	"testing"
	"time"
)

func TestHub_Shutdown_StopsRunLoop(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Shutdown should return without blocking.
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
	// Second shutdown should not panic or block.
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

	// Broadcast after shutdown must not block (it should drop the message).
	done := make(chan struct{})
	go func() {
		hub.Broadcast([]byte("after-stop"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked after Shutdown")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./internal/websocket/... -run 'TestHub_Shutdown' -v`
Expected: FAIL — `hub.Shutdown` undefined.

- [ ] **Step 3: Implement Shutdown on Hub**

Replace `internal/websocket/hub.go` contents with:

```go
package websocket

import (
	"context"
	"sync"
	"sync/atomic"
)

// Hub manages all active WebSocket client connections.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	stop       chan struct{}
	stopped    atomic.Bool
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	if h.stopped.Load() {
		return
	}
	h.register <- client
}

// Run starts the hub's event loop. Must be called as a goroutine.
// Returns when Shutdown is called.
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
// After Shutdown, Broadcast drops the message without blocking.
func (h *Hub) Broadcast(msg []byte) {
	if h.stopped.Load() {
		return
	}
	select {
	case h.broadcast <- msg:
	case <-h.stop:
	}
}

// Shutdown stops the hub's event loop and closes all client connections.
// Safe to call multiple times. Waits for write pumps to finish or ctx to expire.
func (h *Hub) Shutdown(ctx context.Context) error {
	if !h.stopped.CompareAndSwap(false, true) {
		return nil // already stopped
	}
	close(h.stop)
	// Wait for Run() to return and write pumps to drain.
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Len returns the number of connected clients.
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
```

Then update `internal/websocket/client.go` `WritePump` to register with the hub's WaitGroup. Add `h.wg.Add(1)` / `defer h.wg.Done()`:

```go
// WritePump writes messages to the WebSocket connection.
// Runs in its own goroutine. Handles ping frames.
func (c *Client) WritePump() {
	c.hub.wg.Add(1)
	defer c.hub.wg.Done()

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(gorillawebsocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gorillawebsocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gorillawebsocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/websocket/... -v`
Expected: All tests PASS (new shutdown tests + existing `TestNewHub`, `TestHubBroadcast`, `TestHubRegisterUnregister`).

- [ ] **Step 5: Commit**

```bash
git add internal/websocket/hub.go internal/websocket/client.go internal/websocket/hub_test.go
git commit -m "feat(websocket): add Hub.Shutdown for graceful event-loop termination"
```

---

### Task 5: Wire Hub shutdown into WebServer.Shutdown

**Files:**
- Modify: `server/websocket.go`
- Modify: `server/webserver.go:401-418`

**Interfaces:**
- Consumes: `Hub.Shutdown(ctx)` from Task 4
- Produces: `WebServer.Shutdown` gracefully stops the WS hub

- [ ] **Step 1: Inspect current WebServer.Shutdown**

Read `server/webserver.go:401-418`. Current sequence:
```go
func (self *WebServer) Shutdown(ctx context.Context) error {
	var err error
	if self.listenerMgr != nil { self.listenerMgr.Stop() }
	self.shutdownHA()
	if self.redisClient != nil { self.redisClient.Close() }
	if self.s != nil { err = self.s.Shutdown(ctx) }
	<-self.storeQuit
	self.orm.Close()
	return err
}
```

The hub shutdown must happen after `self.s.Shutdown(ctx)` (so no new WS connections arrive) and before `<-self.storeQuit`.

- [ ] **Step 2: Add stopWS helper to server/websocket.go**

Append to `server/websocket.go`:

```go
// stopWS gracefully shuts down the package-level WebSocket hub.
func stopWS(ctx context.Context) {
	if err := wsHub.Shutdown(ctx); err != nil {
		logrus.Warnf("[websocket] hub shutdown error: %v", err)
	}
}
```

- [ ] **Step 3: Call stopWS in WebServer.Shutdown**

In `server/webserver.go`, modify `Shutdown` to call `stopWS(ctx)` after `self.s.Shutdown(ctx)` and before `<-self.storeQuit`:

```go
func (self *WebServer) Shutdown(ctx context.Context) error {
	var err error
	if self.listenerMgr != nil {
		self.listenerMgr.Stop()
	}
	self.shutdownHA()
	if self.redisClient != nil {
		self.redisClient.Close()
	}
	if self.s != nil {
		err = self.s.Shutdown(ctx)
	}
	// Stop accepting new WebSocket connections and drain active clients.
	stopWS(ctx)
	//important: stop input then call shutdown

	<-self.storeQuit
	self.orm.Close()
	return err
}
```

- [ ] **Step 4: Build and run server-package tests**

Run: `GOCACHE=/tmp/gocache go build ./... && GOCACHE=/tmp/gocache go test ./server/...`
Expected: Build OK, server tests PASS.

- [ ] **Step 5: Commit**

```bash
git add server/websocket.go server/webserver.go
git commit -m "feat(server): gracefully shut down WebSocket hub in WebServer.Shutdown"
```

---

## Self-Review

**Spec coverage:**
- ✅ 通知发送 HTTP 超时配置 → Task 1 + Task 2
- ✅ Telegram Markdown 转义处理 → Task 3
- ✅ WebSocket Hub 优雅关闭 → Task 4 + Task 5

**Placeholder scan:** None — all steps have concrete code.

**Type consistency:** `NewService(engine, opts ...Option)` signature consistent across tasks; `Hub.Shutdown(ctx context.Context) error` consistent; `httpClient` field name consistent.
