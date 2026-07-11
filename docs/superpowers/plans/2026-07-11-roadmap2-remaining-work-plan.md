# ROADMAP 2.0 剩余工作实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 ROADMAP 2.0 GA 前剩余的 P0/P1 工作。

**架构:** 基于现有代码基线，计划包含 5 项实际需改动的工作。HEALTHCHECK（Dockerfile 已存在 line 56-57）和 MCP Redis 集成（`v2MCPHandler` line 4980 已传入 `self.redisClient`）已实现无需改动。剩余工作分两阶段：Phase 1 安全门禁（监听器审计 + HA leader election），Phase 2 功能收尾（Redis 公共组件 + Workflow retry DB 同步）。

**Tech Stack:** Go 1.25, XORM, go-redis/v9, Docker Compose, Alpine Linux

---

## 文件改动清单

### Phase 1.1：监听器安全审计

**审计范围：**
- `internal/listener/smtp.go` — SMTP 协议监听器实现
- `internal/listener/ldap.go` — LDAP 协议监听器实现
- `internal/listener/smb.go` — SMB 协议监听器实现
- `internal/listener/ftp.go` — FTP 协议监听器实现
- `internal/listener/manager.go` — 监听器生命周期管理

**文档：**
- Create: `docs/listener-security-guide.md`

### Phase 1.2：HA leader election

**新增：**
- `internal/ha/election.go` — 基于数据库乐观锁的 leader election

**修改：**
- `internal/ha/model.go` — 追加 LeaderElection 模型
- `internal/ha/store.go` — 追加 TryAcquireLock/ReleaseLock/GetLeader 方法
- `internal/ha/service.go` — 追加 ElectLeader/IsLeader/GetLeaderID/Resign 方法
- `server/webserver.go` — 在现有 initHA/heartbeat 中集成 leader election
- `server/v2_ha_handlers.go` — 添加 GET /ha/leader 端点

**文档：**
- Create: `docs/ha-verification-runbook.md`

### Phase 2.1：Redis 公共组件

**Create:** `internal/redislib/redis.go`、`internal/redislib/redis_test.go`

### Phase 2.3：Workflow 持久化 retry DB 同步

**修改：**
- `internal/workflow/queue.go` — processJob 中写回 retry_count + dead 状态

### 无需改动（已实现）

| 项目 | 已有位置 |
|------|---------|
| MCP Redis 集成 | `server/v2_api.go:4980` 已传 `self.redisClient` |
| Docker HEALTHCHECK | `Dockerfile:56-57` 已存在 |
| HA 初始化流程 | `server/webserver.go:330` 已调用 `initHA()` |
| WebServer HA 字段 | `webserver.go:76-78` 已有 `haSvc`/`haNodeID`/`haCancel` |
| HA 表同步 | `server/webui.go:83-86` 已包含 ClusterNode/ClusterConfig/HealthCheck |

---

### Task 1: 协议监听器安全审计 + 文档

**Files:**
- Audit: `internal/listener/smtp.go`, `ldap.go`, `smb.go`, `ftp.go`
- Review: `internal/listener/manager.go:151-224`
- Create: `docs/listener-security-guide.md`

- [ ] **Step 1: 代码审计 SMTP 监听器**

  审查 `internal/listener/smtp.go`：
  1. 启动逻辑是否绑定到可配置的主机/端口
  2. 请求处理路径是否有命令注入/路径穿越风险
  3. 连接处理是否接了 `SecurityContext.CheckConnection()`
  4. 默认是否关闭（需 `IsEnabled` 为 true 才启动）
  
  在审查结果中记录每个风险点。

- [ ] **Step 2: 代码审计 LDAP 监听器**

  `internal/listener/ldap.go` — 同上检查项，重点关注 LDAP 协议解析器的内存安全性。

- [ ] **Step 3: 代码审计 SMB 监听器**

  `internal/listener/smb.go` — 检查 SMB 协议实现的完整性，确认不接受未授权的文件操作。

- [ ] **Step 4: 代码审计 FTP 监听器**

  `internal/listener/ftp.go` — 检查 FTP 命令处理是否限制路径穿越风险。

- [ ] **Step 5: 确认 Manager 默认关闭策略**

  在 `manager.go` 的 `startListener()` 方法中确认：Manager 只启动 `IsEnabled == true` 的监听器。默认情况下无数据库记录或 `IsEnabled` 为 false，监听器不会启动。

  `Listener` 结构体 `IsEnabled` 的默认值为 false（Go 零值规则），无需代码改动。

- [ ] **Step 6: 编写安全文档**

  `docs/listener-security-guide.md`：
  ```markdown
  # 协议监听器安全指南

  ## 概述
  GODNSLOG 支持 SMTP、LDAP、SMB、FTP 四种协议监听器，用于捕获 OAST 回连。
  本文档说明各协议的风险点和安全配置要求。

  ## 风险等级

  | 协议 | 风险等级 | 说明 |
  |------|---------|------|
  | SMTP | 中 | 可能被用于邮件中继 |
  | LDAP | 中 | 可能泄露目录结构 |
  | SMB | 高 | 可能泄露 NTLM 哈希 |
  | FTP | 中 | 匿名访问风险 |

  ## 安全配置

  ### 默认关闭
  所有协议监听器默认关闭，需通过 API `POST /api/v2/listeners` 显式创建并设置 `is_enabled: true`。

  ### 速率限制
  - 默认单 IP 每秒 100 请求
  - 默认单 IP 最大并发连接 10
  - 通过 `ListenerConfig.RateLimitMax` 和 `MaxConcurrentConnections` 配置

  ### 网络隔离建议
  iptables 规则示例：
  ```bash
  # 仅允许内网访问监听器端口
  iptables -A INPUT -p tcp --dport 25 -s 10.0.0.0/8 -j ACCEPT
  iptables -A INPUT -p tcp --dport 25 -j DROP
  ```

  ### 生产环境最小端口列表
  | 端口 | 协议 | 必需 |
  |------|------|------|
  | 53/udp | DNS | 是 |
  | 80/tcp | HTTP | 是 |
  | 8080/tcp | API | 是 |
  | 25/tcp | SMTP | 按需 |
  | 389/tcp | LDAP | 按需 |
  | 445/tcp | SMB | 按需 |
  | 21/tcp | FTP | 按需 |
  ```

- [ ] **Step 7: 运行现有测试确保不破坏**

  ```bash
  go test ./internal/listener/... -v
  ```
  Expected: 全部 PASS

- [ ] **Step 8: Commit**

  ```bash
  git add internal/listener/ docs/listener-security-guide.md
  git commit -m "docs: add listener security guide and complete security audit"
  ```

---

### Task 2: HA leader election 实现

**Files:**
- Create: `internal/ha/election.go`
- Modify: `internal/ha/model.go`（追加 LeaderElection 模型）
- Modify: `internal/ha/store.go`（追加 TryAcquireLock/ReleaseLock/GetLeader）
- Modify: `internal/ha/service.go`（Store 接口追加三个方法）
- Modify: `server/webserver.go:410-456`（在 initHA 的 heartbeat 中集成 leader election）
- Modify: `server/v2_ha_handlers.go`（添加 leader 状态 API）

- [ ] **Step 1: 添加 LeaderElection 模型**

  在 `internal/ha/model.go` 追加：

  ```go
  // LeaderElection represents a leader election record for HA failover.
  type LeaderElection struct {
      ID        string    `json:"id" xorm:"'id' varchar(64) pk notnull"`
      LeaderID  string    `json:"leader_id" xorm:"varchar(64) notnull"`
      Term      int64     `json:"term" xorm:"bigint notnull default 0"`
      LeaseEnd  time.Time `json:"lease_end" xorm:"datetime notnull"`
      CreatedAt time.Time `json:"created_at" xorm:"datetime created"`
      UpdatedAt time.Time `json:"updated_at" xorm:"datetime updated"`
  }

  // TableName returns the table name for LeaderElection
  func (LeaderElection) TableName() string {
      return "ha_leader_election"
  }
  ```

- [ ] **Step 2: 添加数据库操作方法**

  在 `internal/ha/store.go` 追加：

  ```go
  // TryAcquireLock attempts to acquire the leader lock using a database row.
  // Returns true if the lock was acquired, false if held by another node.
  // Uses optimistic locking via term version check.
  func (s *XormStore) TryAcquireLock(ctx context.Context, leaderID string, leaseDuration time.Duration) (bool, error) {
      session := s.engine.Context(ctx).NewSession()
      defer session.Close()

      var current LeaderElection
      has, err := session.ID("leader").Get(&current)
      if err != nil {
          return false, err
      }

      now := time.Now()
      if has && current.LeaseEnd.After(now) && current.LeaderID != leaderID {
          return false, nil
      }

      if has {
          affected, err := session.Where("id = ? AND term = ?", "leader", current.Term).Update(&LeaderElection{
              LeaderID:  leaderID,
              Term:      current.Term + 1,
              LeaseEnd:  now.Add(leaseDuration),
              UpdatedAt: now,
          })
          if err != nil {
              return false, err
          }
          if affected == 0 {
              return false, nil
          }
      } else {
          _, err := session.Insert(&LeaderElection{
              ID:        "leader",
              LeaderID:  leaderID,
              Term:      1,
              LeaseEnd:  now.Add(leaseDuration),
              CreatedAt: now,
              UpdatedAt: now,
          })
          if err != nil {
              return false, err
          }
      }

      return true, nil
  }

  // ReleaseLock releases the leader lock if held by the given leaderID.
  func (s *XormStore) ReleaseLock(ctx context.Context, leaderID string) error {
      _, err := s.engine.Context(ctx).Where("id = ? AND leader_id = ?", "leader", leaderID).Delete(&LeaderElection{})
      return err
  }

  // GetLeader returns the current leader information.
  func (s *XormStore) GetLeader(ctx context.Context) (*LeaderElection, error) {
      var le LeaderElection
      has, err := s.engine.Context(ctx).ID("leader").Get(&le)
      if err != nil {
          return nil, err
      }
      if !has {
          return nil, nil
      }
      return &le, nil
  }
  ```

- [ ] **Step 3: 更新 Store 接口**

  在 `internal/ha/service.go` 的 Store 接口追加三个方法：

  ```go
  // TryAcquireLock attempts to acquire the leader lock
  TryAcquireLock(ctx context.Context, leaderID string, leaseDuration time.Duration) (bool, error)
  // ReleaseLock releases the leader lock
  ReleaseLock(ctx context.Context, leaderID string) error
  // GetLeader returns current leader info
  GetLeader(ctx context.Context) (*LeaderElection, error)
  ```

- [ ] **Step 4: 创建 election.go**

  `internal/ha/election.go`：

  ```go
  package ha

  import (
      "context"
      "time"
  )

  // ElectLeader attempts to become the leader. Returns true if this node is the leader.
  func (s *Service) ElectLeader(ctx context.Context, nodeID string) (bool, error) {
      return s.store.TryAcquireLock(ctx, nodeID, 30*time.Second)
  }

  // IsLeader checks if the given nodeID is the current leader.
  func (s *Service) IsLeader(ctx context.Context, nodeID string) (bool, error) {
      leader, err := s.store.GetLeader(ctx)
      if err != nil {
          return false, err
      }
      if leader == nil {
          return false, nil
      }
      return leader.LeaderID == nodeID && leader.LeaseEnd.After(time.Now()), nil
  }

  // GetLeaderID returns the current leader's node ID, or empty string if none.
  func (s *Service) GetLeaderID(ctx context.Context) (string, error) {
      leader, err := s.store.GetLeader(ctx)
      if err != nil {
          return "", err
      }
      if leader == nil || leader.LeaseEnd.Before(time.Now()) {
          return "", nil
      }
      return leader.LeaderID, nil
  }

  // Resign releases the leader lock if held by this node.
  func (s *Service) Resign(ctx context.Context, nodeID string) error {
      return s.store.ReleaseLock(ctx, nodeID)
  }
  ```

- [ ] **Step 5: 在 initHA 的 heartbeat 中集成 leader election**

  修改 `server/webserver.go` 的 `haHeartbeat()` 方法（~line 458-480），在心跳中增加 leader election：

  ```go
  // haHeartbeat periodically updates the node's last ping time and attempts leader election.
  func (self *WebServer) haHeartbeat(ctx context.Context) {
      ticker := time.NewTicker(10 * time.Second)
      defer ticker.Stop()

      var isLeader bool

      for {
          select {
          case <-ctx.Done():
              if isLeader {
                  if err := self.haSvc.Resign(ctx, self.haNodeID); err != nil {
                      logrus.Errorf("[webserver.go::haHeartbeat] failed to resign: %v", err)
                  }
              }
              return
          case <-ticker.C:
              // Update node heartbeat
              node, err := self.haSvc.GetNode(ctx, self.haNodeID)
              if err != nil {
                  logrus.Errorf("[webserver.go::haHeartbeat] failed to get node: %v", err)
                  continue
              }
              node.LastPing = time.Now()
              node.Status = "online"
              if err := self.haSvc.UpdateNode(ctx, node); err != nil {
                  logrus.Errorf("[webserver.go::haHeartbeat] failed to update node: %v", err)
              }

              // Attempt leader election if Redis is available
              if self.redisClient != nil {
                  becameLeader, err := self.haSvc.ElectLeader(ctx, self.haNodeID)
                  if err != nil {
                      logrus.Errorf("[webserver.go::haHeartbeat] leader election failed: %v", err)
                      continue
                  }
                  if becameLeader && !isLeader {
                      logrus.Infof("[webserver.go::haHeartbeat] node %s became leader", self.haNodeID)
                  } else if !becameLeader && isLeader {
                      logrus.Warnf("[webserver.go::haHeartbeat] node %s lost leader status", self.haNodeID)
                  }
                  isLeader = becameLeader
              }
          }
      }
  }
  ```

- [ ] **Step 6: 添加 leader 状态 API**

  在 `server/v2_ha_handlers.go` 中添加：

  ```go
  // v2GetLeader returns the current cluster leader information.
  func (self *WebServer) v2GetLeader(c *gin.Context) {
      leaderID, err := self.haSvc.GetLeaderID(c)
      if err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
          return
      }
      c.JSON(http.StatusOK, gin.H{"leader_id": leaderID})
  }
  ```

  在 `server/v2_ha_handlers.go` 的 HA 路由组中注册：

  ```go
  haGroup.GET("/leader", self.v2GetLeader)
  ```

- [ ] **Step 7: 将 LeaderElection 表加入数据库同步**

  编辑 `server/webui.go`，在 `initDatabase()` 的 `Sync2` 调用中追加：

  ```go
  &ha.LeaderElection{},
  ```

- [ ] **Step 8: 测试编译和测试**

  ```bash
  go build ./...
  go test ./internal/ha/... -v
  ```
  Expected: 编译成功，测试全部 PASS

- [ ] **Step 9: Commit**

  ```bash
  git add internal/ha/ server/webserver.go server/v2_ha_handlers.go server/webui.go
  git commit -m "feat: add HA leader election with heartbeat integration"
  ```

---

### Task 3: HA 集成验证 + Runbook

**Files:**
- Create: `docs/ha-verification-runbook.md`
- Modify: `docker-compose.ha.yml`（如有必要）

- [ ] **Step 1: 构建并启动 2 节点 HA 集群**

  ```bash
  docker compose -f docker-compose.ha.yml up -d --build
  ```
  Expected: 3 个容器（godnslog-1, godnslog-2, redis）均启动成功

- [ ] **Step 2: 验证节点注册和 leader 选举**

  ```bash
  docker compose -f docker-compose.ha.yml logs godnslog-1 | grep "haHeartbeat"
  docker compose -f docker-compose.ha.yml logs godnslog-2 | grep "haHeartbeat"
  ```
  Expected: 一个节点显示 "became leader"，另一个不显示。

- [ ] **Step 3: 验证 leader API**

  ```bash
  curl -s http://localhost:8081/api/v2/ha/leader
  curl -s http://localhost:8082/api/v2/ha/leader
  ```
  Expected: 两个端点的 `leader_id` 相同，指向成为 leader 的节点。

- [ ] **Step 4: 验证集群状态 API**

  ```bash
  curl -s http://localhost:8080/api/v2/ha/nodes
  ```
  Expected: 包含 2 个节点，状态均为 online。

- [ ] **Step 5: 验证故障切换**

  ```bash
  docker compose -f docker-compose.ha.yml stop godnslog-1
  sleep 35  # 等 leader lease 过期（30s）+ 1 个心跳周期
  curl -s http://localhost:8082/api/v2/ha/leader
  ```
  Expected: godnslog-2 成为新的 leader（leader_id 变为 godnslog-2 的 ID）。

- [ ] **Step 6: 恢复原节点**

  ```bash
  docker compose -f docker-compose.ha.yml start godnslog-1
  sleep 10
  curl -s http://localhost:8080/api/v2/ha/nodes
  ```
  Expected: 2 个节点都在线。

- [ ] **Step 7: 编写 HA 验证 Runbook**

  `docs/ha-verification-runbook.md`，包含上述 Step 1-6 的完整命令和预期输出。

- [ ] **Step 8: Commit**

  ```bash
  git add docs/ha-verification-runbook.md docker-compose.ha.yml
  git commit -m "docs: add HA verification runbook with leader election validation"
  ```

---

### Task 4: Redis 公共组件

**Files:**
- Create: `internal/redislib/redis.go`
- Create: `internal/redislib/redis_test.go`

- [ ] **Step 1: 创建 redislib 包**

  `internal/redislib/redis.go`：

  ```go
  // Package redislib provides a common Redis client wrapper for shared use
  // across MCP session store, workflow queue, and HA modules.
  package redislib

  import (
      "context"
      "fmt"
      "time"

      "github.com/redis/go-redis/v9"
  )

  // Config holds Redis connection configuration.
  type Config struct {
      Addr     string
      Password string
      DB       int
      UseTLS   bool
  }

  // Client wraps a go-redis client with key prefix management.
  type Client struct {
      raw    *redis.Client
      prefix string
  }

  // NewClient creates a new Redis client. Returns nil if addr is empty.
  func NewClient(cfg Config) *Client {
      if cfg.Addr == "" {
          return nil
      }
      opts := &redis.Options{
          Addr:         cfg.Addr,
          Password:     cfg.Password,
          DB:           cfg.DB,
          DialTimeout:  5 * time.Second,
          ReadTimeout:  3 * time.Second,
          WriteTimeout: 3 * time.Second,
      }
      return &Client{
          raw:    redis.NewClient(opts),
          prefix: "godnslog:",
      }
  }

  // Ping checks the Redis connection health.
  func (c *Client) Ping(ctx context.Context) error {
      if c == nil || c.raw == nil {
          return fmt.Errorf("redis client is nil")
      }
      return c.raw.Ping(ctx).Err()
  }

  // Close closes the Redis connection.
  func (c *Client) Close() error {
      if c == nil || c.raw == nil {
          return nil
      }
      return c.raw.Close()
  }

  // Raw returns the underlying go-redis client for direct use.
  func (c *Client) Raw() *redis.Client {
      if c == nil {
          return nil
      }
      return c.raw
  }

  // Key returns a prefixed Redis key.
  func (c *Client) Key(suffix string) string {
      return c.prefix + suffix
  }
  ```

- [ ] **Step 2: 创建测试文件**

  `internal/redislib/redis_test.go`：

  ```go
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
  ```

- [ ] **Step 3: 运行测试**

  ```bash
  go test ./internal/redislib/... -v
  ```
  Expected: 全部 PASS

- [ ] **Step 4: Commit**

  ```bash
  git add internal/redislib/
  git commit -m "feat: add common Redis client wrapper package (redislib)"
  ```

---

### Task 5: Workflow 持久化 retry DB 同步

**Files:**
- Modify: `internal/workflow/queue.go:180-257`

- [ ] **Step 1: 在 processJob 中添加 retry_count DB 持久化和 dead 状态**

  修改 `internal/workflow/queue.go` 的 `processJob` 方法。在重试逻辑处（约 line 244-256），将 in-memory 的 retry 改为 DB 持久化：

  ```go
  if err != nil {
      if job.Attempt < q.maxRetries {
          job.Attempt++

          // Update retry_count in DB if engine is available
          if q.engine != nil && plogID != "" {
              q.engine.ID(plogID).Update(&PersistentActionLog{
                  Status:     "pending",
                  Attempt:    job.Attempt,
                  Error:      err.Error(),
                  FinishedAt: time.Now(),
              })
          }

          backoff := time.Duration(job.Attempt*job.Attempt) * time.Second
          log.Printf("[workflow-queue] Job failed (attempt %d), retrying in %v: %v", job.Attempt, backoff, err)
          time.Sleep(backoff)
          if enqueueErr := q.Enqueue(job); enqueueErr != nil {
              log.Printf("[workflow-queue] Failed to re-enqueue job: %v", enqueueErr)
          }
      } else {
          // Mark as dead after max retries
          log.Printf("[workflow-queue] Job failed after %d attempts, giving up: %v", job.Attempt+1, err)
          if q.engine != nil && plogID != "" {
              q.engine.ID(plogID).Update(&PersistentActionLog{
                  Status:     "dead",
                  Error:      err.Error(),
                  FinishedAt: time.Now(),
              })
          }
      }
  }
  ```

  注意：`PersistentActionLog.Status` 是 `varchar(16)`（无数据库枚举约束），`"dead"` 值可直接使用无需改 schema。

- [ ] **Step 2: 运行 Workflow 测试**

  ```bash
  go test ./internal/workflow/... -v
  ```
  Expected: 全部 PASS

- [ ] **Step 3: 运行全局测试确保不破坏**

  ```bash
  go test ./... 2>&1 | tail -20
  ```
  Expected: 全部 PASS

- [ ] **Step 4: Commit**

  ```bash
  git add internal/workflow/queue.go
  git commit -m "fix: persist workflow retry count to DB and add dead status"
  ```

---

## 最终验证

- [ ] **完整编译和测试**

  ```bash
  go build ./...
  go test ./...
  go vet ./...
  ```
  Expected: 全部通过

- [ ] **前端构建**

  ```bash
  npm --prefix ./frontend-next run build
  ```
  Expected: 构建成功（25 个页面）

- [ ] **容器构建**

  ```bash
  docker build -t godnslog .
  ```
  Expected: 构建成功，HEALTHCHECK 已内置
