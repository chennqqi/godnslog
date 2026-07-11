# ROADMAP 2.0 剩余工作收尾设计规格

- **创建日期**：2026-07-11
- **状态**：已批准
- **对应路线图**：`ROADMAP_2.0.md`

---

## 概述

基于 ROADMAP 2.0 完成度分析（`docs/superpowers/acceptance/roadmap2.0-production-readiness-report.md`），当前代码基线已实现 2.0 全部功能目标，但存在 5 项 P0/P1 剩余工作需在 GA 前闭合。本文档定义这三阶段的实施设计。

---

## Phase 1：P0 安全门禁

### 1.1 协议监听器安全审计

**文件范围**：
- `internal/listener/smtp.go`
- `internal/listener/ldap.go`
- `internal/listener/smb.go`
- `internal/listener/ftp.go`
- `internal/listener/ratelimit.go`
- `internal/listener/security_context.go`
- `internal/listener/manager.go`
- `config/config.go`（若无则确认配置入口）

**实施项**：

1. **代码审查**：逐协议检查监听器启动逻辑，确认以下风险点：
   - 所有监听器默认关闭，需显式配置协议 + 端口才能启用
   - 监听器绑定的 IP/端口可配置，避免意外暴露到公网
   - 请求处理路径无命令注入、路径穿越、内存泄漏

2. **配置加固**：
   - 在配置结构中新增 `ListenerSecurity` 字段：
     - `AllowPrivateIPs`（bool，默认 false）— 禁止监听器回连到 RFC 1918 地址
     - `MaxConnectionsPerIP`（int，默认 10）— 单 IP 最大并发连接
     - `RateLimit`（int，默认 100）— 每秒最大请求数
   - 验证 `internal/listener/ratelimit.go` 已接入这些配置项，补充测试

3. **安全文档**：新增 `docs/listener-security-guide.md`，包含：
   - 各协议的风险等级说明
   - 推荐的网络隔离方案（独立 subnet、iptables 规则示例）
   - 生产环境最小开放端口列表

**验收标准**：
- [ ] 所有协议监听器默认关闭
- [ ] `ratelimit_test.go` 覆盖速率限制逻辑
- [ ] 安全配置文档提交

### 1.2 HA 集群验证

**文件范围**：
- `internal/ha/service.go`
- `internal/ha/store.go`
- `internal/ha/model.go`
- `server/v2_ha_handlers.go`
- `server/webui.go`（HA 集成点）
- `docker-compose.ha.yml`

**实施项**：

1. **集成验证**：在 `server/webui.go` 的启动流程中确认 HA 模块已初始化，补充以下集成点：
   - 启动时检测 `cfg.HA.Enabled` 标志
   - 启用 HA 时初始化 Redis 连接和节点注册
   - 提供健康检查端点 `GET /api/v2/health/leader` 返回当前 leader 信息

2. **Docker Compose 环境**：修复/确认 `docker-compose.ha.yml` 可用，包含：
   - 2 个 godnslog 节点（端口映射 8080:8080 和 8081:8080）
   - 1 个 Redis 实例
   - 共享卷或外部数据库（SQLite 仅用于单机 HA 测试数据只读场景，真实 HA 需 MySQL）

3. **集成测试**：新增 `internal/ha/ha_integration_test.go`，覆盖：
   - 单节点启动 → 自动成为 leader
   - 第二节点加入 → 成为 follower，可查询集群状态
   - leader 心跳超时 → follower 提升为 leader
   - 集群状态 API 返回正确节点列表

4. **验证 Runbook**：输出 `docs/ha-verification-runbook.md`，包含：
   - 启动 2 节点集群的步骤
   - 验证 leader 选举的命令序列
   - 故障切换测试的步骤
   - 恢复后的集群状态确认

**验收标准**：
- [ ] `go test ./internal/ha/...` 通过
- [ ] `docker compose -f docker-compose.ha.yml up` 可启动 2 节点集群
- [ ] leader 故障后 follower 在 30s 内完成提升
- [ ] HA 验证 Runbook 提交

---

## Phase 2：Redis 公共组件 + 功能收尾

### 2.1 Redis 公共组件

**新增文件**：`internal/redislib/redis.go`、`internal/redislib/redis_test.go`

**设计**：
- 封装 `go-redis` 客户端，提供：
  - `NewClient(cfg *RedisConfig) (*RedisClient, error)`
  - `(*RedisClient).Ping() error`
  - `(*RedisClient).Close() error`
  - Key 前缀管理：所有 key 自动添加 `godnslog:` 前缀
- 配置结构体：
  ```go
  type RedisConfig struct {
      Addr     string // host:port
      Password string
      DB       int
      UseTLS   bool
  }
  ```
- 支持从全局配置读取 Redis 地址，与现有 `redis/go-redis/v9` 依赖保持一致

**迁移**：
- `internal/mcp/redis_session.go` 中的 Redis 连接逻辑迁移到 `redislib`
- 保持 `RedisSessionStore` 的接口不变

### 2.2 MCP Redis 集成

**文件范围**：
- `internal/mcp/transport.go`（`v2MCPHandler` 或等价入口）
- `internal/mcp/server.go`
- `server/v2_api.go`（MCP 路由注册）

**实施项**：
- 在 MCP handler 初始化时，检查全局配置 `cfg.Redis.Addr` 是否为空
- 若配置了 Redis → 使用 `internal/redislib` 创建 `RedisSessionStore`
- 若未配置 Redis → 使用现有内存 SessionStore（向后兼容）
- 已有 `internal/mcp/redis_session.go` 无需修改

**验收标准**：
- [ ] 无 Redis 配置时 MCP 使用内存存储（已有行为不变）
- [ ] 有 Redis 配置时 MCP session 跨实例共享
- [ ] `go test ./internal/mcp/...` 通过

### 2.3 Workflow 持久化收尾

**文件范围**：
- `internal/workflow/persistent_log.go`
- `internal/workflow/queue.go`
- `internal/workflow/service.go`
- `server/webui.go`（启动时恢复）

**实施项**：

1. **模型确认**：`PersistentActionLog` 已有定义，包含 `status` 字段，确认字段值覆盖 `pending` / `running` / `completed` / `failed` / `dead`

2. **启动恢复**：在 server 启动流程中（`server/webui.go` 的 `initDatabase` 或后续初始化步骤）：
   - 查询 `PersistentActionLog` 中 `status = pending` 的所有记录
   - 将这些任务重新入队到 `internal/workflow/queue.go` 的 Workflow 队列
   - 日志记录恢复的任务数量

3. **重试上限**：在 queue 处理逻辑中，当 `PersistentActionLog` 执行失败时：
   - 递增 `retry_count` 字段
   - 若 `retry_count >= DefaultMaxCallbackErrorCount`（当前为 5），标记 `status = dead`
   - 否则保持 `pending`，下次启动时重试

4. **执行后更新**：Action 执行完成后（无论成功/失败），更新 `PersistentActionLog` 对应的 `status` 和 `executed_at`

**验收标准**：
- [ ] `go test ./internal/workflow/...` 通过
- [ ] 存在 pending 记录时，server 启动后自动恢复执行
- [ ] 超过最大重试次数后自动标记为 dead

---

## Phase 3：容器 HEALTHCHECK

**文件范围**：
- `Dockerfile`
- `docker-compose.yml`
- `docker-compose.ha.yml`

**实施项**：
- 在 `Dockerfile` 末尾添加：
  ```dockerfile
  HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v2/health || exit 1
  ```
- 确认基础镜像已包含 `curl`（alpine 默认包含，若使用 scratch 需调整）
- `docker-compose.yml` 和 `docker-compose.ha.yml` 中已有的 healthcheck 配置保持不变（编排层覆盖镜像层）

**验收标准**：
- [ ] `docker build` 后 `docker inspect` 输出包含 `Healthcheck` 字段
- [ ] 容器启动后 10s 内 health 状态变为 healthy

---

## 依赖关系

```
Phase 1
  │
  ├── 1.1 安全审计 ──→ 无下游依赖，独立
  │
  └── 1.2 HA 验证 ──→ 依赖 Docker Compose 环境，独立验证

Phase 2
  │
  ├── 2.1 Redis 组件 ──→ 被 2.2 依赖
  │         │
  │         └── 2.2 MCP Redis ──→ 依赖 2.1
  │
  └── 2.3 Workflow 持久化 ──→ 独立（仅依赖数据库表，不依赖 Redis）

Phase 3
  └── 3.1 HEALTHCHECK ──→ 无依赖，可随时合并
```

## 测试策略

| 模块 | 测试类型 | 命令 |
|------|---------|------|
| 安全审计 | 代码审查 + 手动 | `go vet ./internal/listener/...` |
| HA 验证 | 单元 + 集成 | `go test ./internal/ha/...` |
| Redis 组件 | 单元 | `go test ./internal/redislib/...` |
| MCP Redis | 单元 | `go test ./internal/mcp/...` |
| Workflow 持久化 | 单元 | `go test ./internal/workflow/...` |
| HEALTHCHECK | 容器验证 | `docker inspect` |

## 交付物清单

1. 代码修改：
   - 监听器安全加固代码
   - `internal/redislib/` 新包
   - MCP Redis 集成连通
   - Workflow 持久化恢复逻辑
   - Dockerfile HEALTHCHECK

2. 文档：
   - `docs/listener-security-guide.md`
   - `docs/ha-verification-runbook.md`

3. 测试：
   - `internal/ha/ha_integration_test.go`
   - `internal/redislib/redis_test.go`
   - 现有单元测试全部通过
