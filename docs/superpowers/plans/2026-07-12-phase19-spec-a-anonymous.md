# Phase 19 Spec A: 匿名模式 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development

**Goal:** 实现匿名模式开关，开启后 Interaction 不记录真实 SourceIP

**Architecture:** 配置项 `AnonymousMode` 控制 Service 层的 IP 匿名化行为

**Tech Stack:** Go

---

### Task 1: Service + Model 层匿名化

- [ ] **Step 1: 还原未完成的更改并整理**

```bash
cd /data/dev/github.com/chennqqi/godnslog
```

检查 `server/webserver.go` 中已有的 `AnonymousMode` 字段。如果格式有问题（缩进不一致），修正为：
```go
AnonymousMode bool `json:"anonymous_mode"` // when true, don't record real source IPs
```

- [ ] **Step 2: Service 结构体添加字段**

在 `internal/interaction/service.go` 的 `Service` 结构体中添加匿名模式字段：

```go
type Service struct {
	engine        *xorm.Engine
	wsHub         *websocket.Hub
	fingerprinter *fingerprint.Fingerprinter
	anonymousMode bool
}
```

修改 `NewService` 签名：

```go
func NewService(engine *xorm.Engine, wsHub *websocket.Hub, fp *fingerprint.Fingerprinter, anonymousMode bool) *Service {
	return &Service{
		engine:        engine,
		wsHub:         wsHub,
		fingerprinter: fp,
		anonymousMode: anonymousMode,
	}
}
```

在 `CreateInteraction` 方法开头（`if interaction.ID == ""` 之前）添加 IP 匿名化：

```go
// Anonymous mode: mask source IP for privacy
if s.anonymousMode {
	interaction.SourceIP = "0.0.0.0"
}
```

- [ ] **Step 3: 更新所有 NewService 调用点**

添加 `false` 作为第 4 个参数。涉及的调用点：
- `server/v2_api.go` — 多处
- `server/v2_api_test.go`
- `internal/interaction/service_test.go`
- `internal/interaction/evidence_service_test.go`
- `internal/evidencehub/service.go`
- `internal/agentrun/complete_test.go`
- `internal/agentrun/review_test.go`

- [ ] **Step 4: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./...
```

- [ ] **Step 5: 运行测试**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go test ./internal/interaction/... -v 2>&1 | grep -E "PASS|FAIL|ok"
```

- [ ] **Step 6: Commit**

```bash
git add server/webserver.go internal/interaction/service.go
git add .  # for caller updates
git commit -m "feat: add anonymous mode to mask source IPs in interactions"
```
