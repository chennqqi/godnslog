# Phase 5a/5b Code Review

**Date:** 2026-06-20  
**Reviewer:** Cascade (AI)  
**Scope:** Phase 5a (Multi-Protocol Listeners & Security) + Phase 5b (HA, Deployment & Marketplace)

---

## 发现的问题及修复

### 1. 速率/连接限制未实际执行（严重）

**文件：** `internal/listener/manager.go`, `internal/listener/smtp.go`, `internal/listener/ldap.go`, `internal/listener/smb.go`, `internal/listener/ftp.go`

**问题描述：**

`RateLimiter` 和 `ConnLimiter` 在 `Manager.startListener()` 中被创建并包装到 `rateLimitedStore` 中，但各协议监听器的 `acceptConnections()` 循环从未调用 `rateLim.Allow()` 或 `connLim.Acquire()`。限制器虽然存在但完全未被使用，等于没有防护。

**根因：**

设计时将限制器放在 store wrapper 中，但监听器的 accept 循环只使用 store 做数据持久化，不会通过 store 检查连接限制。限制检查应该在 accept 循环中、连接被接受后立即执行。

**修复方案：**

新增 `SecurityContext`（`internal/listener/security_context.go`），通过 `sync.Map` 按 listener ID 注册。Manager 在 `startListener()` 时创建 `SecurityContext` 并注册，在 `StopListener()` 和 `Stop()` 时清除。

每个协议的 `acceptConnections()` 循环修改为：
1. 调用 `sc.CheckConnection(conn.RemoteAddr())` 检查速率和并发限制
2. 如果返回 `false`，记录日志并关闭连接
3. 如果返回 `true`，启动 goroutine 处理连接，`defer sc.ReleaseConnection()` 释放并发槽位

**新增文件：** `internal/listener/security_context.go`, `internal/listener/security_context_test.go`

---

### 2. `Start()` 中的 data race（中等）

**文件：** `internal/listener/manager.go:74`

**问题描述：**

```go
m.logger.Infof("[listener/manager] started %d listeners", len(m.active))
```

`len(m.active)` 读取 map 时未持有 `m.mu` 锁，而 `startListener` 在其他 goroutine（如通过 API 调用 `StartListener`）中可能并发写入 `m.active`，构成数据竞争。

**修复方案：**

在读取 `len(m.active)` 前加锁：

```go
m.mu.Lock()
count := len(m.active)
m.mu.Unlock()
m.logger.Infof("[listener/manager] started %d listeners", count)
```

---

### 3. `v2UpdateClusterConfig` upsert 逻辑错误（中等）

**文件：** `server/v2_ha_handlers.go:177-206`

**问题描述：**

原逻辑：
1. 调用 `svc.GetConfig()` — 当 DB 无配置时返回内存中的默认配置（非 DB 记录），`err == nil`
2. 调用 `svc.UpdateConfig()` — 对不存在的行执行 UPDATE，xorm 不报错但也不写入
3. 因为 `UpdateConfig` 没返回错误，`CreateConfig` 分支不会执行
4. 结果：首次配置无法持久化

**修复方案：**

改为先调用 `store.ListConfigs()` 直接检查 DB 中是否已有配置记录：
- 有记录 → `UpdateConfig`
- 无记录 → `CreateConfig`

---

## 未修改的既有 lint 警告（非本次引入）

以下 lint 警告存在于本次变更之前的代码中，不在本次 review 修复范围内：

- `server/dnsserver.go:222` — unreachable code
- `server/dnsserver.go:401` — unkeyed fields in `dns.A` struct literal
- `server/dnsserver.go:414` — redundant return statement
- `server/webui.go:1449` — redundant return statement
- `server/webserver.go:376` — impossible condition: non-nil == nil

---

## 测试验证

```
go build ./...                                          # PASS
go test ./internal/listener/ ./server/ ./internal/ha/ ./internal/marketplace/ -timeout 60s  # ALL PASS
go test ./internal/listener/ -run TestSecurityContext -v                               # 4/4 PASS
go test ./internal/listener/ -run TestRateLimiter -v                                   # 2/2 PASS
go test ./internal/listener/ -run TestConnLimiter -v                                   # 2/2 PASS
go test ./internal/listener/ -run TestManager -v                                       # 3/3 PASS
go test ./server/ -run TestV2Routes -v                                                 # 1/1 PASS
```
