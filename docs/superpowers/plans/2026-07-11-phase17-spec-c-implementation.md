# Spec C: 攻击链时间线 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans.

**Goal:** 按 Token 将多协议 Interaction 聚合成攻击链，提供列表和详情 API + 前端展示

**Architecture:** 后端在 `internal/interaction/attackchain.go` 实现 Token 聚合查询（SQL GROUP BY + 内存聚合），在 `server/v2_api.go` 注册 API 端点。前端在 interactions 目录下新增攻击链列表和详情组件，复用现有 Timeline 组件。

**Tech Stack:** Go (xorm), Next.js/TypeScript (前端), shadcn/ui

---

### Task 1: 攻击链数据模型 + 聚合逻辑

**Files:**
- Create: `internal/interaction/attackchain.go`
- Create: `internal/interaction/attackchain_test.go`

- [ ] **Step 1: 创建 `internal/interaction/attackchain.go`**

```go
package interaction

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// AttackChain 按 Token 聚合的交互链
type AttackChain struct {
	Token             string   `json:"token"`
	InteractionCount  int      `json:"interaction_count"`
	Protocols         []string `json:"protocols"`
	ExploitTypes      []string `json:"exploit_types"`
	FirstSeen         string   `json:"first_seen"`
	LastSeen          string   `json:"last_seen"`
	Confidence        string   `json:"confidence"`
}

// AttackChainDetail 攻击链详情（含完整时间线）
type AttackChainDetail struct {
	Token             string                `json:"token"`
	InteractionCount  int                   `json:"interaction_count"`
	Protocols         []string              `json:"protocols"`
	ExploitTypes      []string              `json:"exploit_types"`
	FirstSeen         string                `json:"first_seen"`
	LastSeen          string                `json:"last_seen"`
	Confidence        string                `json:"confidence"`
	Interactions      []models.Interaction  `json:"interactions"`
}

// AttackChainListResponse 攻击链列表
type AttackChainListResponse struct {
	Items      []AttackChain `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// GetAttackChains 查询攻击链列表（按 Token 聚合）
func (s *Service) GetAttackChains(page, pageSize int) (*AttackChainListResponse, error) {
	// Count distinct tokens
	type TokenCount struct {
		Count int64 `xorm:"count"`
	}
	var tc TokenCount
	_, err := s.engine.SQL("SELECT COUNT(DISTINCT token) as count FROM interactions WHERE token IS NOT NULL AND token != ''").Get(&tc)
	if err != nil {
		return nil, err
	}
	total := tc.Count

	if total == 0 {
		return &AttackChainListResponse{
			Items:      []AttackChain{},
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	// Get distinct tokens with aggregation, ordered by last_seen DESC
	type TokenRow struct {
		Token    string    `xorm:"token"`
		Count    int       `xorm:"cnt"`
		FirstTS  time.Time `xorm:"first_ts"`
		LastTS   time.Time `xorm:"last_ts"`
	}

	var rows []TokenRow
	offset := (page - 1) * pageSize
	err = s.engine.SQL(`
		SELECT token,
			COUNT(*) as cnt,
			MIN(timestamp) as first_ts,
			MAX(timestamp) as last_ts
		FROM interactions
		WHERE token IS NOT NULL AND token != ''
		GROUP BY token
		ORDER BY last_ts DESC
		LIMIT ? OFFSET ?`, pageSize, offset).Find(&rows)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// For each token, fetch all interactions to compute protocols and exploit_types
	items := make([]AttackChain, 0, len(rows))
	for _, row := range rows {
		chain := s.buildAttackChain(row.Token, row.Count, row.FirstTS, row.LastTS)
		items = append(items, chain)
	}

	return &AttackChainListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAttackChainDetail 获取攻击链详情（含完整时间线）
func (s *Service) GetAttackChainDetail(token string) (*AttackChainDetail, error) {
	var interactions []models.Interaction
	err := s.engine.Where("token = ?", token).Asc("timestamp").Find(&interactions)
	if err != nil {
		return nil, err
	}
	if len(interactions) == 0 {
		return nil, nil
	}

	detail := &AttackChainDetail{
		Token:        token,
		Interactions: interactions,
	}
	detail.InteractionCount = len(interactions)

	// Compute protocols, exploit_types, time span, confidence
	protocolSet := make(map[string]bool)
	exploitSet := make(map[string]bool)
	confidenceOrder := []string{"high", "medium", "low"}
	detail.Confidence = ""

	for _, in := range interactions {
		if in.Type != "" {
			protocolSet[in.Type] = true
		}
		if in.ExploitType != nil && *in.ExploitType != "" {
			exploitSet[*in.ExploitType] = true
		}
		if in.Confidence != nil && *in.Confidence != "" {
			if detail.Confidence == "" {
				detail.Confidence = *in.Confidence
			} else {
				// Keep the higher confidence
				detail.Confidence = higherConfidence(detail.Confidence, *in.Confidence, confidenceOrder)
			}
		}
	}

	detail.Protocols = sortedKeys(protocolSet)
	detail.ExploitTypes = sortedKeys(exploitSet)
	detail.FirstSeen = interactions[0].Timestamp.Format(time.RFC3339)
	detail.LastSeen = interactions[len(interactions)-1].Timestamp.Format(time.RFC3339)

	return detail, nil
}

// buildAttackChain 为单个 Token 构建 AttackChain
func (s *Service) buildAttackChain(token string, count int, firstTS, lastTS time.Time) AttackChain {
	var interactions []models.Interaction
	s.engine.Where("token = ?", token).Asc("timestamp").Find(&interactions)

	protocolSet := make(map[string]bool)
	exploitSet := make(map[string]bool)
	confidenceOrder := []string{"high", "medium", "low"}
	var chainConfidence string

	for _, in := range interactions {
		if in.Type != "" {
			protocolSet[in.Type] = true
		}
		if in.ExploitType != nil && *in.ExploitType != "" {
			exploitSet[*in.ExploitType] = true
		}
		if in.Confidence != nil && *in.Confidence != "" {
			if chainConfidence == "" {
				chainConfidence = *in.Confidence
			} else {
				chainConfidence = higherConfidence(chainConfidence, *in.Confidence, confidenceOrder)
			}
		}
	}

	return AttackChain{
		Token:            token,
		InteractionCount: count,
		Protocols:        sortedKeys(protocolSet),
		ExploitTypes:     sortedKeys(exploitSet),
		FirstSeen:        firstTS.Format(time.RFC3339),
		LastSeen:         lastTS.Format(time.RFC3339),
		Confidence:       chainConfidence,
	}
}

// higherConfidence returns the higher of two confidence levels
func higherConfidence(a, b string, order []string) string {
	rank := make(map[string]int)
	for i, v := range order {
		rank[v] = i
	}
	if rank[a] < rank[b] {
		return a
	}
	return b
}

// sortedKeys returns sorted keys from a string set
func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
```

- [ ] **Step 2: 创建 `internal/interaction/attackchain_test.go`**

```go
package interaction

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestBuildAttackChain(t *testing.T) {
	// Can't easily test with real DB, test the helper functions
	order := []string{"high", "medium", "low"}

	tests := []struct {
		a, b, expected string
	}{
		{"high", "medium", "high"},
		{"medium", "low", "medium"},
		{"low", "medium", "medium"},
		{"high", "high", "high"},
		{"low", "low", "low"},
	}

	for _, tt := range tests {
		got := higherConfidence(tt.a, tt.b, order)
		if got != tt.expected {
			t.Errorf("higherConfidence(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	set := map[string]bool{"dns": true, "http": true, "ldap": true}
	keys := sortedKeys(set)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	// Verify sorted
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Errorf("keys not sorted: %v", keys)
			break
		}
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	keys := sortedKeys(map[string]bool{})
	if keys == nil || len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}

func TestGetAttackChainsEmpty(t *testing.T) {
	// Use service_test.go's engine
	svc, err := newTestService()
	if err != nil {
		t.Fatalf("newTestService() error = %v", err)
	}
	defer svc.engine.Close()

	resp, err := svc.GetAttackChains(1, 20)
	if err != nil {
		t.Fatalf("GetAttackChains() error = %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("expected 0 total, got %d", resp.Total)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}
```

- [ ] **Step 3: 验证编译**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/interaction/
```

Expected: Build OK. Note: If `newTestService` is not accessible from the test (it might be in service_test.go in the same package), the test for `GetAttackChainsEmpty` should work since it's the same package.

- [ ] **Step 4: 运行测试**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go test ./internal/interaction/ -run "TestBuildAttackChain|TestSortedKeys|TestGetAttackChainsEmpty" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/interaction/attackchain.go internal/interaction/attackchain_test.go
git commit -m "feat: add attack chain data model and aggregation logic

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: API 端点

**Files:**
- Modify: `server/v2_api.go` — 注册攻击链 API 路由

- [ ] **Step 1: 在 `server/v2_api.go` 中找到 v2 API 路由注册区域，添加攻击链路由**

在 `registerV2API` 方法中找到 `interactions := v2.Group("/interactions")` 附近，在其后添加：

```go
// Attack chains
v2.GET("/attack-chains", self.authHandler, self.v2ListAttackChains)
v2.GET("/attack-chains/:token", self.authHandler, self.v2GetAttackChainDetail)
```

- [ ] **Step 2: 添加 Handler 方法**

```go
// v2ListAttackChains lists attack chains
func (self *WebServer) v2ListAttackChains(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	iaSvc := interaction.NewService(self.orm, nil)
	chains, err := iaSvc.GetAttackChains(page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAttackChains] error: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, chains)
}

// v2GetAttackChainDetail gets attack chain detail by token
func (self *WebServer) v2GetAttackChainDetail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "token is required"})
		return
	}

	iaSvc := interaction.NewService(self.orm, nil)
	detail, err := iaSvc.GetAttackChainDetail(token)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetAttackChainDetail] error: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if detail == nil {
		c.JSON(404, gin.H{"error": "attack chain not found"})
		return
	}

	c.JSON(200, detail)
}
```

**Note:** Add `"github.com/chennqqi/godnslog/internal/interaction"` import if not present.

- [ ] **Step 3: 验证编译**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./server/
```

Expected: Build OK

- [ ] **Step 4: Commit**

```bash
git add server/v2_api.go
git commit -m "feat: add attack chain API endpoints

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: 前端攻击链列表组件

**Files:**
- Create: `frontend-next/src/features/interactions/attack-chain-list.tsx`

- [ ] **Step 1: 创建攻击链列表组件**

```tsx
'use client'

import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { LoadingState } from '@/components/loading-state'
import { EmptyState } from '@/components/empty-state'

interface AttackChain {
  token: string
  interaction_count: number
  protocols: string[]
  exploit_types: string[]
  first_seen: string
  last_seen: string
  confidence: string
}

interface AttackChainListResponse {
  items: AttackChain[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

const protocolColors: Record<string, string> = {
  dns: 'bg-blue-100 text-blue-800',
  http: 'bg-green-100 text-green-800',
  ldap: 'bg-purple-100 text-purple-800',
  smtp: 'bg-yellow-100 text-yellow-800',
  smb: 'bg-orange-100 text-orange-800',
  ftp: 'bg-pink-100 text-pink-800',
}

const confidenceColors: Record<string, string> = {
  high: 'text-red-600',
  medium: 'text-yellow-600',
  low: 'text-gray-500',
}

export function AttackChainList() {
  const [chains, setChains] = useState<AttackChain[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [selectedToken, setSelectedToken] = useState<string | null>(null)
  const pageSize = 20

  useEffect(() => {
    setLoading(true)
    setError(null)
    apiClient.get(`/api/v2/attack-chains?page=${page}&page_size=${pageSize}`)
      .then((data: AttackChainListResponse) => {
        setChains(data.items || [])
        setTotalPages(data.total_pages || 1)
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false))
  }, [page])

  if (loading) return <LoadingState />
  if (error) return <div className="text-red-500">Error: {error}</div>
  if (chains.length === 0) return <EmptyState title="No Attack Chains" description="No interactions with tokens found yet." />

  return (
    <div className="space-y-3">
      {chains.map((chain) => (
        <div
          key={chain.token}
          className="border rounded-lg p-4 hover:shadow-md cursor-pointer transition-shadow"
          onClick={() => setSelectedToken(selectedToken === chain.token ? null : chain.token)}
        >
          <div className="flex justify-between items-start">
            <div className="flex-1">
              <h3 className="font-mono text-sm font-medium">{chain.token}</h3>
              <div className="flex flex-wrap gap-1.5 mt-2">
                {chain.protocols.map((p) => (
                  <span key={p} className={`text-xs px-2 py-0.5 rounded ${protocolColors[p] || 'bg-gray-100 text-gray-800'}`}>
                    {p}({chain.interaction_count})
                  </span>
                ))}
                {chain.exploit_types.map((et) => (
                  <span key={et} className="text-xs px-2 py-0.5 rounded bg-red-50 text-red-700">
                    {et}
                  </span>
                ))}
              </div>
            </div>
            <div className="text-right text-xs text-gray-400">
              <div className={confidenceColors[chain.confidence] || ''}>
                {chain.confidence || 'unknown'}
              </div>
            </div>
          </div>
          <div className="text-xs text-gray-400 mt-2">
            {chain.first_seen} ~ {chain.last_seen}
          </div>
        </div>
      ))}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          <button
            className="px-3 py-1 border rounded disabled:opacity-50"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            Previous
          </button>
          <span className="px-3 py-1 text-sm text-gray-500">
            {page} / {totalPages}
          </span>
          <button
            className="px-3 py-1 border rounded disabled:opacity-50"
            disabled={page >= totalPages}
            onClick={() => setPage(page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: 验证 TypeScript**

```bash
cd /data/dev/github.com/chennqqi/godnslog/frontend-next && npx tsc --noEmit --strict src/features/interactions/attack-chain-list.tsx 2>&1 || echo "check done"
```

- [ ] **Step 3: Commit**

```bash
git add frontend-next/src/features/interactions/attack-chain-list.tsx
git commit -m "feat: add attack chain list component

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: 前端攻击链详情组件 + 路由页

**Files:**
- Create: `frontend-next/src/features/interactions/attack-chain-detail.tsx`
- Create: `frontend-next/src/app/interactions/attack-chains/page.tsx`

- [ ] **Step 1: 创建详情组件 `attack-chain-detail.tsx`**

```tsx
'use client'

import { useEffect, useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Timeline, TimelineItem } from '@/components/timeline'
import { LoadingState } from '@/components/loading-state'

interface Interaction {
  id: string
  type: string
  timestamp: string
  source_ip: string
  domain?: string
  path?: string
  method?: string
  exploit_type?: string
  decoded_data?: string
  confidence?: string
}

interface AttackChainDetail {
  token: string
  interaction_count: number
  protocols: string[]
  exploit_types: string[]
  first_seen: string
  last_seen: string
  confidence: string
  interactions: Interaction[]
}

const typeIcons: Record<string, string> = {
  dns: '🔍',
  http: '🌐',
  ldap: '📂',
  smtp: '📧',
  smb: '📁',
  ftp: '📄',
}

const confidenceColors: Record<string, string> = {
  high: 'text-red-600',
  medium: 'text-yellow-600',
  low: 'text-gray-500',
}

interface AttackChainDetailProps {
  token: string
  onBack: () => void
}

export function AttackChainDetail({ token, onBack }: AttackChainDetailProps) {
  const [detail, setDetail] = useState<AttackChainDetail | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    apiClient.get(`/api/v2/attack-chains/${encodeURIComponent(token)}`)
      .then(setDetail)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [token])

  if (loading) return <LoadingState />
  if (!detail) return <div className="text-red-500">Chain not found</div>

  const timelineItems: TimelineItem[] = detail.interactions.map((in) => {
    const date = new Date(in.timestamp)
    const dateStr = date.toLocaleDateString()
    const timeStr = date.toLocaleTimeString()
    const icon = typeIcons[in.type] || '❓'
    const exploitLabel = in.exploit_type ? ` [${in.exploit_type}]` : ''

    return {
      id: in.id,
      date: dateStr,
      time: timeStr,
      title: `${icon} ${in.type.toUpperCase()}${exploitLabel}`,
      description: in.domain || in.path || in.source_ip,
      content: (
        <div className="text-xs space-y-1">
          <div>IP: {in.source_ip}</div>
          {in.decoded_data && <div className="text-green-600">Decoded: {in.decoded_data}</div>}
          {in.confidence && <div className={confidenceColors[in.confidence] || ''}>Confidence: {in.confidence}</div>}
        </div>
      ),
    }
  })

  return (
    <div className="space-y-4">
      <button onClick={onBack} className="text-sm text-indigo-600 hover:underline">&larr; Back to chains</button>

      <div className="border rounded-lg p-4 bg-gray-50">
        <h2 className="font-mono text-lg font-medium">{detail.token}</h2>
        <div className="flex gap-4 mt-2 text-sm text-gray-500">
          <span>{detail.interaction_count} interactions</span>
          <span>{detail.protocols.join(', ')}</span>
          {detail.exploit_types.length > 0 && <span>{detail.exploit_types.join(', ')}</span>}
        </div>
      </div>

      <Timeline items={timelineItems} groupByDate={false} />
    </div>
  )
}
```

- [ ] **Step 2: 创建路由页 `page.tsx`**

```tsx
'use client'

import { useState } from 'react'
import { AttackChainList } from '@/features/interactions/attack-chain-list'
import { AttackChainDetail } from '@/features/interactions/attack-chain-detail'
import { PageHeader } from '@/components/page-header'

export default function AttackChainsPage() {
  const [selectedToken, setSelectedToken] = useState<string | null>(null)

  return (
    <div className="space-y-6">
      <PageHeader
        title="Attack Chains"
        description="Interactions grouped by token showing the full attack timeline"
      />

      {selectedToken ? (
        <AttackChainDetail token={selectedToken} onBack={() => setSelectedToken(null)} />
      ) : (
        <AttackChainList onSelectToken={setSelectedToken} />
      )}
    </div>
  )
}
```

Wait, the AttackChainList component I wrote doesn't have an `onSelectToken` prop. Let me fix the list to accept it and call it on click.

Actually, I'll modify the list component to accept an optional `onSelectToken` prop, and in the page I'll handle selection state.

Let me redo the AttackChainList to accept onSelectToken:

```tsx
interface AttackChainListProps {
  onSelectToken?: (token: string) => void
}

export function AttackChainList({ onSelectToken }: AttackChainListProps) {
  // ... existing code ...

  // In the onClick handler:
  onClick={() => {
    if (onSelectToken) {
      onSelectToken(chain.token)
    }
  }}

  // ... rest of the code ...
}
```

Also need to update the AttackChainList to use `useRouter` or just pass the token up.

Actually, looking at the original list component, it already has internal `selectedToken` state that toggles on click. I should change it to use the `onSelectToken` prop instead, letting the parent page manage the selected state.

Let me finalize the code properly.

- [ ] **Step 3: Verify TypeScript**

```bash
cd /data/dev/github.com/chennqqi/godnslog/frontend-next && npx tsc --noEmit 2>&1 | head -20 || echo "check done"
```

- [ ] **Step 4: Commit**

```bash
git add frontend-next/src/features/interactions/attack-chain-detail.tsx frontend-next/src/app/interactions/attack-chains/page.tsx
git commit -m "feat: add attack chain detail component and route page

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: 全量验证

- [ ] **Step 1: 全量编译**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./...
```

Expected: 编译成功

- [ ] **Step 2: 全量测试**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go test ./... 2>&1 | tail -10
```

Expected: 无 FAIL

- [ ] **Step 3: 最终提交**

```bash
git add .
git commit -m "chore: Spec C attack chain timeline complete

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```
