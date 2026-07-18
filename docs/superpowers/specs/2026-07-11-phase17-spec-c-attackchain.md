# Phase 17 Spec C: 攻击链时间线

> 按 Token 将多协议 Interaction 聚合为攻击链，展示完整攻击过程时间线
> 对应 ROADMAP 2.4 智能增强版

## 1. 概述

将同一 Token 下的所有 Interaction（DNS/HTTP/LDAP/SMTP/SMB/FTP）聚合为一条攻击链，按时间排序展示完整攻击过程。利用 Spec A 的富化字段（exploit_type、decoded_data）丰富每条命中的上下文信息。

## 2. 架构

```
┌─────────────────┐
│  AttackChain    │ ← 后端聚合层
│                 │
│  token: abc123  │
│  protocols:     │
│    dns(3)       │
│    http(1)      │
│    ldap(1)      │
│  exploit_types: │
│    log4shell    │
│  time_span:     │
│    10:00:01~    │
│    10:00:03     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Timeline       │ ← 展开后显示完整交互时间线
│  ┌───────────┐  │
│  │ 10:00:01  │  │  DNS query jndi.xxx.dnslog.fun
│  │           │  │  └─ exploit: log4shell (high)
│  ├───────────┤  │
│  │ 10:00:02  │  │  HTTP GET /xxe
│  │           │  │  └─ exploit: xxe (high)
│  ├───────────┤  │
│  │ 10:00:03  │  │  LDAP connect
│  └───────────┘  │
└─────────────────┘
```

## 3. 数据模型

```go
// AttackChain 按 Token 聚合的交互链
type AttackChain struct {
    Token        string   `json:"token"`
    InteractionCount int  `json:"interaction_count"`
    Protocols    []string `json:"protocols"`     // 去重协议列表
    ExploitTypes []string `json:"exploit_types"` // 去重利用类型
    FirstSeen    string   `json:"first_seen"`
    LastSeen     string   `json:"last_seen"`
    Confidence   string   `json:"confidence"`    // 链中最高置信度
}

// AttackChainDetail 攻击链详情（含完整时间线）
type AttackChainDetail struct {
    AttackChain
    Interactions []models.Interaction `json:"interactions"` // 按时间排序
}

// AttackChainListResponse 攻击链列表
type AttackChainListResponse struct {
    Items      []AttackChain `json:"items"`
    Total      int64         `json:"total"`
    Page       int           `json:"page"`
    PageSize   int           `json:"page_size"`
    TotalPages int           `json:"total_pages"`
}
```

## 4. 后端实现

### 4.1 聚合逻辑

在 `internal/interaction/attackchain.go` 中：

```go
// GetAttackChains 查询攻击链列表（按 Token 聚合）
func (s *Service) GetAttackChains(page, pageSize int) (*AttackChainListResponse, error)
```

SQL 逻辑：
```sql
-- 1. 查询所有有 Token 的 Interaction，按 Token 分组聚合
SELECT token,
       COUNT(*) as interaction_count,
       COUNT(DISTINCT type) as protocol_count,
       MIN(timestamp) as first_seen,
       MAX(timestamp) as last_seen
FROM interactions
WHERE token IS NOT NULL AND token != ''
GROUP BY token
ORDER BY last_seen DESC
LIMIT ? OFFSET ?
```

```go
// 2. 对每组 Token 查询详细 Interaction 列表（仅 Detail 时）
func (s *Service) GetAttackChainDetail(token string) (*AttackChainDetail, error)
```

### 4.2 协议去重

从 Interaction 列表的 `Type` 字段聚合去重。

### 4.3 利用类型聚合

从 Interaction 的 `ExploitType` 字段聚合非空值去重，按置信度排序。

### 4.4 API 端点

在 `server/v2_api.go` 中新增：

```
GET /api/v2/attack-chains        → 攻击链列表
GET /api/v2/attack-chains/:token → 攻击链详情（含完整时间线）
```

## 5. 前端实现

### 5.1 AttackChain 列表页

在 `frontend-next/src/features/interactions/` 新增 `attack-chain-list.tsx`：

```
┌──────────────────────────────────────────────────┐
│ 🔗 abc123.dnslog.example.com                     │
│ DNS(3) HTTP(1) LDAP(1)    log4shell, xxe        │
│ 🕐 2026-07-11 10:00:01 ~ 10:00:03    high        │
├──────────────────────────────────────────────────┤
│ 🔗 def456.dnslog.example.com                     │
│ DNS(1) HTTP(1)            ssrf                   │
│ 🕐 2026-07-11 09:30:00 ~ 09:30:01  medium       │
└──────────────────────────────────────────────────┘
```

每条链展示：
- Token 名称
- 协议 badge（彩色标签，DNS=蓝/HTTP=绿/LDAP=紫/...）
- 利用类型标签（从 Spec A 富化）
- 时间跨度 + 命中计数
- 最高置信度

### 5.2 攻击链详情（时间线展开）

点击链展开显示完整时间线，复用现有 `Timeline` 组件。每条 Interaction 展示：
- 时间戳
- 协议图标 + 原始域名/路径
- 利用类型标签（如有）
- 解码数据（如有，Spec A 字段）
- 来源 IP

### 5.3 路由

在 Interaction 导航下新增 Attack Chain 标签页：
- 路由: `/interactions/attack-chains`
- 默认显示攻击链列表
- 点击链 → 展开详情或跳转

## 6. 测试

```go
func TestGetAttackChains(t *testing.T) {
    // 准备：插入多条同 Token 的 Interaction
    // 验证：聚合正确，protocols 去重， exploit_types 合并
}

func TestGetAttackChainDetail(t *testing.T) {
    // 验证：返回的 Interactions 按时间排序
}

func TestAttackChainEmpty(t *testing.T) {
    // 验证：无 Interaction 时返回空列表
}
```

## 7. 交付物

- `internal/interaction/attackchain.go` — 聚合逻辑 + API 响应类型
- `internal/interaction/attackchain_test.go` — 测试
- `server/v2_api.go` — API 端点
- `frontend-next/src/features/interactions/attack-chain-list.tsx` — 列表组件
- `frontend-next/src/features/interactions/attack-chain-detail.tsx` — 详情组件
- `frontend-next/src/app/interactions/attack-chains/page.tsx` — 路由页
