# Phase 18 Spec C: 自定义 HTTP Response + RMI 监听 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans.

**Goal:** 实现 Workflow 驱动的自定义 HTTP Response 控制和 RMI 协议监听器

**Architecture:** 
- **RMI 监听**: 在 `internal/listener/` 中新增 `rmi.go`，遵循现有 LDAP/SMTP 的 ProtocolListener 接口模式。实现最小 RMI 协议解析——捕获 JRMP 握手信息用于 JNDI 注入检测。对接现有 Store/Handler 体系。
- **自定义 HTTP Response**: 已有 Payload 级别的 CustomResponse 功能（`webapi.go:278-306`），Workflow 集成扩展为规则动作，允许 Workflow 规则匹配后覆盖 HTTP 响应。

**Tech Stack:** Go 标准库 `net` (RMI TCP 监听)、`encoding/binary` (RMI 协议解析)

---

### Task 1: RMI 协议模型扩展

**Files:**
- Modify: `internal/models/listener.go` — 添加 ProtocolRMI 常量 + RMIInteraction 类型
- Modify: `internal/listener/model.go` — 导出新类型

- [ ] **Step 1: 添加 RMI 协议常量**

在 `internal/models/listener.go` 中 `ProtocolFTP` 后添加：
```go
	ProtocolRMI  Protocol = "rmi"
```

- [ ] **Step 2: 添加 RMI 交互数据结构**

在 `internal/models/listener.go` 末尾（`Close`方法后）添加：

```go
// RMIInteraction represents a captured RMI connection
// For JNDI injection detection, we capture the minimum:
// - Connection details (source IP, port)
// - RMI call context (object URN, method name if available)
type RMIInteraction struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID string    `json:"listener_id" xorm:"'listener_id' varchar(36) notnull index"`
	SourceIP   string    `json:"source_ip" xorm:"varchar(64) notnull"`
	SourcePort int       `json:"source_port" xorm:"int notnull"`
	URN        string    `json:"urn" xorm:"text"`              // The object URN being looked up
	RawData    string    `json:"raw_data" xorm:"mediumtext"`   // Raw protocol data
	Timestamp  time.Time `json:"timestamp" xorm:"datetime notnull created"`
}

func (RMIInteraction) TableName() string {
	return "rmi_interactions"
}
```

- [ ] **Step 3: 更新 listener/model.go 导出**

在 `internal/listener/model.go` 中添加：
```go
type RMIInteraction = models.RMIInteraction

const (
	ProtocolRMI = models.ProtocolRMI
)
```

- [ ] **Step 4: 验证编译**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/models/ && go build ./internal/listener/
```

- [ ] **Step 5: Commit**

```bash
git add internal/models/listener.go internal/listener/model.go
git commit -m "feat: add RMI protocol constant and interaction model"
```

---

### Task 2: RMI Listener 实现

**Files:**
- Create: `internal/listener/rmi.go`

- [ ] **Step 1: 创建 `internal/listener/rmi.go`**

这是一个最小化的 RMI 监听器，用于 JNDI 注入检测。RMI/JRMP 协议握手时客户端会发送一个对象 URN，我们只需要捕获到这个信息即可。

```go
package listener

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RMIListener implements a minimal RMI server for JNDI injection detection.
type RMIListener struct {
	listener     *Listener
	config       *ListenerConfig
	server       net.Listener
	store        Store
	logger       *logrus.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewRMIListener creates a new RMI listener.
func NewRMIListener(listener *Listener, config *ListenerConfig, store Store, logger *logrus.Logger) *RMIListener {
	if config == nil {
		config = DefaultRMIConfig()
	}
	return &RMIListener{
		listener: listener,
		config:   config,
		store:    store,
		logger:   logger,
	}
}

// DefaultRMIConfig returns default RMI listener configuration.
func DefaultRMIConfig() *ListenerConfig {
	return &ListenerConfig{
		MaxConnections: 100,
		Timeout:        30 * time.Second,
		BufferSize:     4096,
	}
}

// Start starts the RMI listener.
func (r *RMIListener) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	addr := fmt.Sprintf("%s:%d", r.listener.Host, r.listener.Port)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("rmi listen on %s: %w", addr, err)
	}
	r.server = server
	r.logger.Infof("[rmi] listening on %s for token %s", addr, r.listener.Token)

	r.wg.Add(1)
	go r.acceptLoop()
	return nil
}

// Stop stops the RMI listener.
func (r *RMIListener) Stop() error {
	r.cancel()
	if r.server != nil {
		r.server.Close()
	}
	r.wg.Wait()
	return nil
}

func (r *RMIListener) acceptLoop() {
	defer r.wg.Done()

	for {
		conn, err := r.server.Accept()
		if err != nil {
			select {
			case <-r.ctx.Done():
				return
			default:
				r.logger.Warnf("[rmi] accept error: %v", err)
				continue
			}
		}

		r.wg.Add(1)
		go r.handleConnection(conn)
	}
}

// RMI Protocol Constants
const (
	rmiProtocolStreamProtocol = 0x4a524d49 // "JRMI" magic bytes (0x4a524d49 = JRMI\x00)
	rmiStreamProtocol         = 0x4b
	rmiSingleOpProtocol       = 0x4c
)

// rmiHandshake reads the RMI JRMP handshake and extracts basic info.
// JRMP handshake structure:
//   - 4 bytes: protocol magic (JRMI)
//   - 2 bytes: protocol version
//   - 1 byte: protocol type (StreamProtocol/SingleOpProtocol)
//   - for StreamProtocol: 2 bytes: port number
func parseRMIHandshake(data []byte) (string, bool) {
	if len(data) < 7 {
		return "", false
	}

	// Check JRMI magic
	magic := binary.BigEndian.Uint32(data[0:4])
	if magic != 0x4a524d49 { // "JRMI" in ASCII
		return "", false
	}

	// Protocol type at offset 6
	protoType := data[6]
	switch protoType {
	case rmiStreamProtocol:
		if len(data) >= 9 {
			port := binary.BigEndian.Uint16(data[7:9])
			return fmt.Sprintf("JRMP StreamProtocol (callbacks port: %d)", port), true
		}
		return "JRMP StreamProtocol", true
	case rmiSingleOpProtocol:
		return "JRMP SingleOpProtocol", true
	default:
		return fmt.Sprintf("JRMP protocol type: 0x%02x", protoType), true
	}
}

func (r *RMIListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer r.wg.Done()

	conn.SetDeadline(time.Now().Add(r.config.Timeout))

	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	// Read initial handshake data
	buf := make([]byte, r.config.BufferSize)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		r.logger.Debugf("[rmi] read error from %s: %v", remoteAddr.IP, err)
		return
	}

	data := buf[:n]
	urn, recognized := parseRMIHandshake(data)

	// Save interaction
	interaction := &RMIInteraction{
		ID:         generateID(),
		ListenerID: r.listener.ID,
		SourceIP:   remoteAddr.IP.String(),
		SourcePort: remoteAddr.Port,
		URN:        urn,
		RawData:    fmt.Sprintf("%x", data), // hex dump
		Timestamp:  time.Now(),
	}

	if err := r.store.SaveRMIInteraction(interaction); err != nil {
		r.logger.Warnf("[rmi] failed to save interaction: %v", err)
	}

	r.logger.Infof("[rmi] captured connection from %s:%d - %s (recognized: %v)",
		remoteAddr.IP, remoteAddr.Port, urn, recognized)

	if !recognized {
		return
	}

	// Send minimal RMI response to keep connection healthy (no-op return)
	// For SingleOpProtocol and StreamProtocol, send an empty return data
	conn.Write(buildEmptyRMIResponse())
}

// buildEmptyRMIResponse builds a minimal valid RMI response that tells the client
// the call returned without actually doing anything.
func buildEmptyRMIResponse() []byte {
	// This is a simplified empty response - in practice, a full RMI response
	// includes: stream protocol header + return data + end marker
	// For JNDI detection purposes, this allows the connection to succeed
	// and the client to log the interaction.
	return []byte{
		0x4a, 0x52, 0x4d, 0x49, // JRMI magic
		0x00, 0x02, // version 2
		0x4b, // StreamProtocol
		0x00, 0x00, // return data (empty)
	}
}

// generateID generates a simple unique ID for interactions.
func generateID() string {
	return fmt.Sprintf("rmi-%d", time.Now().UnixNano())
}
```

**Note:** Go cannot have two functions named `generateID` in the same package. Check if `generateID` already exists in the listener package. If it does, rename to `generateRMIID`.

Let me check: `grep "func generateID" internal/listener/*.go`

- [ ] **Step 2: 更新 Store 接口**

在 `internal/listener/store.go` 中添加 RMI 交互的存储接口。先读取现有 Store 接口定义：

```bash
grep -n "type Store interface\|SaveSMTP\|SaveFTP\|SaveSMB" internal/listener/store.go
```

添加：
```go
// SaveRMIInteraction saves an RMI interaction
SaveRMIInteraction(interaction *RMIInteraction) error
```

并在 `XormStore` 实现中添加该方法：
```go
func (s *XormStore) SaveRMIInteraction(interaction *RMIInteraction) error {
	_, err := s.engine.Insert(interaction)
	return err
}
```

同时需要确保 XormStore 能正确处理新表：
- 如果已有 `Sync` 调用，确保 `new(RMIInteraction)` 被传入
- 或者在 `MigrateListener` 中添加

- [ ] **Step 3: 注册 RMI 协议到 Listener Manager**

在 `internal/listener/manager.go` 中，找到协议映射（类似 `ProtocolSMTP` → `NewSMTPListener` 的地方），添加：
```go
case ProtocolRMI:
    rmiListener := NewRMIListener(listener, config, m.store, m.logger)
    m.active[listener.ID] = &managedListener{
        listener: rmiListener,
        cancel:   cancel,
        config:   config,
    }
    return rmiListener.Start(ctx)
```

- [ ] **Step 4: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/listener/
```

- [ ] **Step 5: 测试**

```go
func TestRMIProtocolParsing(t *testing.T) {
    // JRMP StreamProtocol
    data := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x4b, 0x00, 0x50}
    urn, ok := parseRMIHandshake(data)
    if !ok {
        t.Error("expected RMI handshake to be recognized")
    }
    if urn == "" {
        t.Error("expected non-empty URN")
    }

    // Not RMI
    data2 := []byte{0x00, 0x00, 0x00, 0x00}
    _, ok2 := parseRMIHandshake(data2)
    if ok2 {
        t.Error("expected non-RMI data to not be recognized")
    }

    // Short data
    _, ok3 := parseRMIHandshake([]byte{0x4a})
    if ok3 {
        t.Error("expected short data to not be recognized")
    }
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/listener/rmi.go internal/listener/store.go internal/listener/manager.go
git commit -m "feat: add minimal RMI listener for JNDI injection detection"
```

---

### Task 3: Workflow 自定义 HTTP Response

**Files:**
- Modify: `server/webapi.go` — 使 Workflow 规则能控制 HTTP 响应

- [ ] **Step 1: 分析现有自定义响应逻辑**

当前 `webapi.go:278-306` 已有基于 Payload 的 CustomResponse。需要扩展为：Workflow 规则匹配后也能覆盖 HTTP 响应。

- [ ] **Step 2: 在 Workflow 动作中添加 HTTP Response 控制**

在 `internal/workflow/` 中找到动作执行器，添加新的动作类型（如存在 Action 类型定义）：

如果 Workflow 已经有 Action 类型系统，添加 `ActionModifyResponse`；否则直接在 rule 动作中添加：

```go
// 在 Workflow 处理链路中，在现有逻辑后添加：
// 如果 Workflow 规则匹配，允许覆盖 HTTP 响应状态码、Header、Body
// 这通过设置 gin.Context 上的键值对实现
```

**注意**：这个功能依赖现有的 Workflow 动作系统。需要先了解 Workflow/rule 的动作模型。

```bash
grep -rn "type Action\|ActionRespond\|func.*Action" internal/workflow/ internal/rule/ --include="*.go" | head -20
```

- [ ] **Step 3: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/webapi.go internal/workflow/
git commit -m "feat: add Workflow-driven custom HTTP response control"
```

---

### Task 4: 全量验证

- [ ] **Step 1: 编译 + 测试**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./... && go test ./internal/listener/ ./internal/models/ -v 2>&1 | grep -E "PASS|FAIL|ok"
```

- [ ] **Step 2: 最终提交**

```bash
git add .
git commit -m "chore: Phase 18 Spec C - custom HTTP response and RMI listener complete"
```
