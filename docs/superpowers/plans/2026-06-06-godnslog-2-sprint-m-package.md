# GODNSLOG 2.0 Sprint M Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a minimal single-Agent-Run replay/follow-up action loop that starts from the Review Packet and records the new action as Agent Operation + Audit without expanding into full Agent management.

**Architecture:** Reuse the existing Agent Run service, Agent Operation model, Review Packet API, and audit service. Add a small follow-up action service/API for one Agent Run only, with frontend controls on the Agent Run detail page and E2E proof from Review Packet to follow-up operation.

**Tech Stack:** Go, xorm, Gin server routes in `server/v2_api.go`, existing `internal/agentrun`, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint M
- **Sprint 主题**：Agent Run Replay-lite & Review Actions
- **所属阶段**：Phase 5 - Agent Governance and Replay

## Sprint 背景

Sprint J 建立了 Agent Run MVP。

Sprint K 补齐了 Agent API Key、MCP scope 和 risk gate。

Sprint L 将单个 Agent Run 推进到可复核的 Review Packet，并通过 API、MCP、UI、E2E 建立：

`Agent Run -> Review Packet -> Evidence Summary / Markdown -> Export Audit`

现在安全团队能回答“这次 Agent Run 做了什么、证据是什么”。Sprint M 的下一步不是完整 Agent 管理平台，也不是批量重放或生命周期治理，而是补齐复核后最小下一步：

- 基于一个已复核 Agent Run 创建一次 follow-up action。
- follow-up action 必须明确来自哪个 source Agent Run。
- action 必须写入 Agent Operation 和 Audit。
- UI 必须从 Agent Run detail 的 Review Packet 区块触发。
- E2E 必须证明 Review Packet -> Follow-up action -> Operation timeline / Audit payload 的闭环。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 为单个 Agent Run 增加最小 follow-up action API。
2. 支持 3 个安全、非破坏性的 follow-up action type：
   - `recheck_evidence`
   - `wait_more_interactions`
   - `create_followup_note`
3. 每次 follow-up action 写入 Agent Operation，operation result 中包含 `source_agent_run_id`、`action_type`、`reason`、`review_packet_id`。
4. 每次 follow-up action 写入 Audit，action 使用 `agent_run.followup_created`。
5. Agent Run detail 页面在 Review Packet 区块提供 Review Actions，并用 E2E 证明真实 API 请求和 Operation timeline 更新。

## 明确不做

本 Sprint 严格不做：

- 完整 Agent 管理平台。
- Agent replay 引擎或后台任务队列。
- 批量 replay / 批量操作。
- Scanner Hub 扩展。
- 生命周期治理、归档、retention 策略。
- 真实 LLM 调用。
- 删除、撤销、修改配置等高风险动作。
- 自动重新发送 payload 到目标。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-j-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-j-acceptance.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-k-package.md`
- `docs/superpowers/acceptance/2026-05-25-godnslog-2-sprint-k-acceptance.md`
- `docs/superpowers/plans/2026-05-31-godnslog-2-sprint-l-package.md`
- `docs/superpowers/acceptance/2026-05-31-godnslog-2-sprint-l-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/service.go` 已有 Agent Run create / list / get / status update / append operation。
- `internal/agentrun/review.go` 已有单 Agent Run Review Packet。
- `server/v2_api.go` 已有 `/api/v2/agent-runs/:id/review` 和 `/api/v2/agent-runs/:id/operations`。
- `internal/mcp/server.go` 已会把 MCP tool 结果写入 Agent Operation。
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已有 Agent Run detail、Review Packet、Evidence 回链。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖 Review API 请求和 Review Packet 渲染。

### 主要缺口

- Review Packet 之后没有“复核后动作”。
- UI 只能查看/导出，不能记录一次人工或 Agent follow-up 决策。
- Agent Operation timeline 不能区分原始执行与 review 后 follow-up。
- Audit 没有稳定记录“谁基于哪个 review packet 创建了后续动作”。

## 术语边界

### Follow-up Action

Follow-up Action 是针对单个 Agent Run 的复核后动作记录。它是审计记录和操作记录，不是自动 replay 引擎。

最小字段：

- `action_type`
- `reason`
- `source_agent_run_id`
- `review_packet_id`
- `case_id`
- `payload_id`
- `created_by`
- `created_at`

### Replay-lite

Replay-lite 指“基于已有 Agent Run 创建一次可审计 follow-up operation”。它不发起真实扫描，不重新投递 payload，不调用 Scanner Hub，不调 LLM。

允许 action：

- `recheck_evidence`：标记需要复核证据，并记录 review context。
- `wait_more_interactions`：标记需要继续等待回连，并记录等待原因。
- `create_followup_note`：记录一次安全分析说明。

## 数据契约

### AgentRunFollowupRequest

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunFollowupRequest struct {
	ActionType     string `json:"action_type" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
}
```

### AgentRunFollowupResponse

```go
type AgentRunFollowupResponse struct {
	AgentRunID     string         `json:"agent_run_id"`
	OperationID    string         `json:"operation_id"`
	ActionType     string         `json:"action_type"`
	Reason         string         `json:"reason"`
	ReviewPacketID string         `json:"review_packet_id,omitempty"`
	Operation      AgentOperation `json:"operation"`
	CreatedAt      time.Time      `json:"created_at"`
}
```

### Agent Operation result

Follow-up action 写入 `AgentOperation.Result` 的 JSON 必须包含：

```json
{
  "success": true,
  "source_agent_run_id": "agent-run-1",
  "action_type": "recheck_evidence",
  "reason": "Evidence looks high confidence, request second review",
  "review_packet_id": "agent-run-1",
  "case_id": "case-1",
  "payload_id": "payload-1"
}
```

### Audit details

Audit action：

```text
agent_run.followup_created
```

Audit details 必须包含：

```json
{
  "agent_run_id": "agent-run-1",
  "operation_id": "op-1",
  "action_type": "recheck_evidence",
  "reason": "Evidence looks high confidence, request second review",
  "review_packet_id": "agent-run-1",
  "case_id": "case-1",
  "payload_id": "payload-1"
}
```

不得包含完整 API Key、Authorization header、token secret 或生产敏感数据。

## API 范围

新增：

```text
POST /api/v2/agent-runs/:id/followups
```

请求：

```json
{
  "action_type": "recheck_evidence",
  "reason": "Evidence looks high confidence, request second review",
  "review_packet_id": "agent-run-1"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "agent_run_id": "agent-run-1",
    "operation_id": "op-1",
    "action_type": "recheck_evidence",
    "reason": "Evidence looks high confidence, request second review",
    "review_packet_id": "agent-run-1",
    "operation": {},
    "created_at": "2026-06-06T10:00:00Z"
  }
}
```

错误语义：

- unknown Agent Run：404。
- invalid action type：400。
- empty reason：400。
- reason 超过 500 字符：400。
- terminal / non-terminal Agent Run 都允许创建 follow-up，因为这是复核记录，不是状态迁移。

## 文件结构

### 后端

- Modify: `internal/models/agent_run.go`
  - 增加 request / response type。
  - 增加 allowed follow-up action 常量或 helper。

- Modify: `internal/agentrun/service.go`
  - 增加 `CreateFollowupAction(agentRunID string, req *models.AgentRunFollowupRequest, userID string) (*models.AgentRunFollowupResponse, error)`。
  - 复用 `AppendOperation` 写 Agent Operation。
  - 使用 `authService.CreateAuditLog` 写 audit。

- Modify: `internal/agentrun/service_test.go`
  - 增加 follow-up 成功测试。
  - 增加 invalid action / empty reason / not found 测试。
  - 增加 operation result 和 audit details 测试。

- Modify: `server/v2_api.go`
  - 注册 `POST /api/v2/agent-runs/:id/followups`。
  - 新增 handler `v2CreateAgentRunFollowup`。

- Modify: `server/v2_api_test.go`
  - 增加 API 成功与错误路径测试。
  - 验证 response 不泄露敏感字段。

### 前端

- Modify: `frontend-next/src/types/index.ts`
  - 增加 `AgentRunFollowupRequest` 和 `AgentRunFollowupResponse`。

- Modify: `frontend-next/src/lib/api-client.ts`
  - 增加 `agentRunApi.createFollowup(id, data)`。

- Modify: `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
  - 在 Review Packet 区块添加 Review Actions。
  - 提供 action type selector、reason textarea、submit button。
  - 成功后刷新 Agent Run detail，operation timeline 展示新增 follow-up operation。

- Modify: `frontend-next/e2e/agent-runs.spec.ts`
  - 增加 Review Actions E2E。
  - 必须真实等待 `POST /api/v2/agent-runs/agent-run-1/followups`。
  - 必须断言 request body。
  - 必须断言 operation timeline 中出现 follow-up action。

## Task 1: Follow-up Model and Service Tests

**Files:**
- Modify: `internal/models/agent_run.go`
- Modify: `internal/agentrun/service_test.go`

- [ ] **Step 1: Add failing service tests**

在 `internal/agentrun/service_test.go` 新增测试：

```go
func TestCreateFollowupAction(t *testing.T) {
	engine := setupTestDB(t)
	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	created, err := service.CreateAgentRun(&models.AgentRunCreateRequest{
		AgentID:    "agent-1",
		OperatorID: "operator-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://target.example",
		Title:      "Review target",
	}, "1")
	if err != nil {
		t.Fatalf("create agent run: %v", err)
	}

	resp, err := service.CreateFollowupAction(created.ID, &models.AgentRunFollowupRequest{
		ActionType:     "recheck_evidence",
		Reason:         "Evidence is high confidence and needs second review",
		ReviewPacketID: created.ID,
	}, "1")
	if err != nil {
		t.Fatalf("create followup: %v", err)
	}

	if resp.AgentRunID != created.ID {
		t.Fatalf("expected agent run id %s, got %s", created.ID, resp.AgentRunID)
	}
	if resp.OperationID == "" {
		t.Fatal("expected operation id")
	}
	if resp.ActionType != "recheck_evidence" {
		t.Fatalf("unexpected action type %s", resp.ActionType)
	}

	var op models.AgentOperation
	has, err := engine.ID(resp.OperationID).Get(&op)
	if err != nil || !has {
		t.Fatalf("expected operation row, has=%v err=%v", has, err)
	}
	if op.Action != "followup.recheck_evidence" {
		t.Fatalf("unexpected operation action %s", op.Action)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(op.Result), &result); err != nil {
		t.Fatalf("parse operation result: %v", err)
	}
	if result["source_agent_run_id"] != created.ID {
		t.Fatalf("operation result missing source_agent_run_id")
	}
	if result["action_type"] != "recheck_evidence" {
		t.Fatalf("operation result missing action_type")
	}

	var auditLog models.AuditLog
	has, err = engine.Where("action = ? AND resource_type = ?", "agent_run.followup_created", "agent_run").Get(&auditLog)
	if err != nil || !has {
		t.Fatalf("expected followup audit, has=%v err=%v", has, err)
	}
	if auditLog.Details["agent_run_id"] != created.ID {
		t.Fatalf("audit missing agent_run_id")
	}
}
```

- [ ] **Step 2: Add validation tests**

新增：

```go
func TestCreateFollowupActionValidation(t *testing.T) {
	engine := setupTestDB(t)
	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	created, err := service.CreateAgentRun(&models.AgentRunCreateRequest{
		AgentID:    "agent-1",
		OperatorID: "operator-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://target.example",
		Title:      "Review target",
	}, "1")
	if err != nil {
		t.Fatalf("create agent run: %v", err)
	}

	cases := []struct {
		name string
		id   string
		req  *models.AgentRunFollowupRequest
	}{
		{
			name: "unknown run",
			id:   "missing",
			req:  &models.AgentRunFollowupRequest{ActionType: "recheck_evidence", Reason: "valid reason"},
		},
		{
			name: "invalid action",
			id:   created.ID,
			req:  &models.AgentRunFollowupRequest{ActionType: "delete_payload", Reason: "valid reason"},
		},
		{
			name: "empty reason",
			id:   created.ID,
			req:  &models.AgentRunFollowupRequest{ActionType: "recheck_evidence", Reason: ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.CreateFollowupAction(tc.id, tc.req, "1"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
```

- [ ] **Step 3: Run tests and confirm failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun -run 'TestCreateFollowupAction' -count=1
```

Expected: FAIL because `AgentRunFollowupRequest` and `CreateFollowupAction` do not exist.

## Task 2: Follow-up Service Implementation

**Files:**
- Modify: `internal/models/agent_run.go`
- Modify: `internal/agentrun/service.go`

- [ ] **Step 1: Add model types**

Add to `internal/models/agent_run.go`:

```go
const (
	AgentRunFollowupRecheckEvidence      = "recheck_evidence"
	AgentRunFollowupWaitMoreInteractions = "wait_more_interactions"
	AgentRunFollowupCreateNote           = "create_followup_note"
)

type AgentRunFollowupRequest struct {
	ActionType     string `json:"action_type" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
}

type AgentRunFollowupResponse struct {
	AgentRunID     string         `json:"agent_run_id"`
	OperationID    string         `json:"operation_id"`
	ActionType     string         `json:"action_type"`
	Reason         string         `json:"reason"`
	ReviewPacketID string         `json:"review_packet_id,omitempty"`
	Operation      AgentOperation `json:"operation"`
	CreatedAt      time.Time      `json:"created_at"`
}

func IsAllowedAgentRunFollowupAction(action string) bool {
	switch action {
	case AgentRunFollowupRecheckEvidence,
		AgentRunFollowupWaitMoreInteractions,
		AgentRunFollowupCreateNote:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 2: Add service method**

Add to `internal/agentrun/service.go`:

```go
func (s *Service) CreateFollowupAction(agentRunID string, req *models.AgentRunFollowupRequest, userID string) (*models.AgentRunFollowupResponse, error) {
	agentRun, err := s.GetAgentRunByID(agentRunID)
	if err != nil {
		return nil, err
	}
	if agentRun == nil {
		return nil, errors.New("agent run not found")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}
	if !models.IsAllowedAgentRunFollowupAction(req.ActionType) {
		return nil, fmt.Errorf("invalid followup action type: %s", req.ActionType)
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, errors.New("reason is required")
	}
	if len(reason) > 500 {
		return nil, errors.New("reason must be 500 characters or less")
	}

	result := map[string]interface{}{
		"success":             true,
		"source_agent_run_id": agentRun.ID,
		"action_type":         req.ActionType,
		"reason":              reason,
		"review_packet_id":    req.ReviewPacketID,
		"case_id":             agentRun.CaseID,
		"payload_id":          agentRun.PayloadID,
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	opReq := &models.AgentOperationCreateRequest{
		Action:    "followup." + req.ActionType,
		RiskLevel: "low",
		Request:   fmt.Sprintf(`{"reason":%q,"review_packet_id":%q}`, reason, req.ReviewPacketID),
		Result:    string(resultJSON),
		StartedAt: &now,
		EndedAt:   &now,
	}

	operation, err := s.AppendOperation(agentRun.ID, opReq, userID)
	if err != nil {
		return nil, err
	}

	userIDPtr := &userID
	resourceIDPtr := &agentRun.ID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.followup_created",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"agent_run_id":     agentRun.ID,
			"operation_id":     operation.ID,
			"action_type":      req.ActionType,
			"reason":           reason,
			"review_packet_id": req.ReviewPacketID,
			"case_id":          agentRun.CaseID,
			"payload_id":       agentRun.PayloadID,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create followup audit log: %w", err)
	}

	return &models.AgentRunFollowupResponse{
		AgentRunID:     agentRun.ID,
		OperationID:    operation.ID,
		ActionType:     req.ActionType,
		Reason:         reason,
		ReviewPacketID: req.ReviewPacketID,
		Operation:      *operation,
		CreatedAt:      now,
	}, nil
}
```

Also add `strings` import.

- [ ] **Step 3: Run service tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun -run 'TestCreateFollowupAction' -count=1
```

Expected: PASS.

## Task 3: Follow-up API

**Files:**
- Modify: `server/v2_api.go`
- Modify: `server/v2_api_test.go`

- [ ] **Step 1: Add failing API tests**

Add to `server/v2_api_test.go`:

```go
func TestV2CreateAgentRunFollowup(t *testing.T) {
	server, r, token, userID := setupV2TestServer(t)

	agentRunID := "agent-run-followup-1"
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-1",
		OperatorID: fmt.Sprintf("%d", userID),
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://target.example",
		Title:      "Followup target",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("insert agent run: %v", err)
	}

	body := strings.NewReader(`{"action_type":"recheck_evidence","reason":"Evidence needs second review","review_packet_id":"agent-run-followup-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/followups", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AgentRunID  string `json:"agent_run_id"`
			OperationID string `json:"operation_id"`
			ActionType  string `json:"action_type"`
			Reason      string `json:"reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0")
	}
	if resp.Data.OperationID == "" {
		t.Fatalf("expected operation id")
	}
	if strings.Contains(w.Body.String(), "Authorization") || strings.Contains(w.Body.String(), "secret") {
		t.Fatalf("response leaks sensitive data")
	}
}
```

- [ ] **Step 2: Add route and handler**

In `server/v2_api.go`, under `agentRuns` routes:

```go
agentRuns.POST("/:id/followups", self.v2CreateAgentRunFollowup)
```

Add handler:

```go
func (self *WebServer) v2CreateAgentRunFollowup(c *gin.Context) {
	id := c.Param("id")
	var req v2models.AgentRunFollowupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	userID := getUserIDFromContext(c)
	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	resp, err := agentRunService.CreateFollowupAction(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid followup") ||
			strings.Contains(err.Error(), "reason") ||
			strings.Contains(err.Error(), "request is required") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2CreateAgentRunFollowup] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create followup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}
```

Use existing user ID helper style in `server/v2_api.go`; if no helper exists, follow the nearby Agent Run handlers.

- [ ] **Step 3: Run API tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run 'TestV2CreateAgentRunFollowup' -count=1
```

Expected: PASS.

## Task 4: Frontend API and Types

**Files:**
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/lib/api-client.ts`

- [ ] **Step 1: Add frontend types**

Add to `frontend-next/src/types/index.ts`:

```ts
export type AgentRunFollowupActionType =
  | 'recheck_evidence'
  | 'wait_more_interactions'
  | 'create_followup_note'

export interface AgentRunFollowupRequest {
  action_type: AgentRunFollowupActionType
  reason: string
  review_packet_id?: string
}

export interface AgentRunFollowupResponse {
  agent_run_id: string
  operation_id: string
  action_type: AgentRunFollowupActionType
  reason: string
  review_packet_id?: string
  operation: AgentOperation
  created_at: string
}
```

- [ ] **Step 2: Add API client method**

In `frontend-next/src/lib/api-client.ts`, import the new types and add:

```ts
createFollowup: (id: string, data: AgentRunFollowupRequest) =>
  api.post<{ data: AgentRunFollowupResponse }>(`/agent-runs/${id}/followups`, data),
```

- [ ] **Step 3: Run target ESLint**

Run:

```bash
cd frontend-next && npx eslint src/lib/api-client.ts src/types/index.ts
```

Expected: PASS.

## Task 5: Agent Run Detail Review Actions UI

**Files:**
- Modify: `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`

- [ ] **Step 1: Add state and handler**

In `page.tsx`, add state:

```ts
const [followupAction, setFollowupAction] = useState<'recheck_evidence' | 'wait_more_interactions' | 'create_followup_note'>('recheck_evidence')
const [followupReason, setFollowupReason] = useState('')
const [submittingFollowup, setSubmittingFollowup] = useState(false)
```

Add handler:

```ts
const handleCreateFollowup = async () => {
  if (!agentRun || !followupReason.trim()) return
  setSubmittingFollowup(true)
  setError('')
  try {
    await agentRunApi.createFollowup(agentRun.id, {
      action_type: followupAction,
      reason: followupReason.trim(),
      review_packet_id: reviewPacket?.id,
    })
    setFollowupReason('')
    const response = await agentRunApi.get(agentRun.id)
    if (response.data) {
      setAgentRun(response.data.data)
    }
  } catch (error: unknown) {
    console.error('Failed to create followup:', error)
    setError('创建Follow-up失败')
  } finally {
    setSubmittingFollowup(false)
  }
}
```

- [ ] **Step 2: Add Review Actions block**

Inside the Review Packet card, after review content:

```tsx
{reviewPacket && (
  <div className="border-t pt-4 space-y-3">
    <p className="text-sm font-medium">Review Actions</p>
    <div className="grid gap-3 md:grid-cols-[220px_1fr_auto]">
      <select
        className="border rounded-md px-3 py-2 text-sm"
        value={followupAction}
        onChange={(event) => setFollowupAction(event.target.value as typeof followupAction)}
      >
        <option value="recheck_evidence">Recheck evidence</option>
        <option value="wait_more_interactions">Wait more interactions</option>
        <option value="create_followup_note">Create follow-up note</option>
      </select>
      <textarea
        className="border rounded-md px-3 py-2 text-sm min-h-10"
        value={followupReason}
        maxLength={500}
        onChange={(event) => setFollowupReason(event.target.value)}
        placeholder="Reason for this follow-up"
      />
      <Button
        onClick={handleCreateFollowup}
        disabled={submittingFollowup || followupReason.trim().length === 0}
      >
        Create Follow-up
      </Button>
    </div>
  </div>
)}
```

Keep the UI compact. Do not create a new page, modal-heavy flow, dashboard section, or Agent management surface.

- [ ] **Step 3: Run target ESLint**

Run:

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx
```

Expected: PASS.

## Task 6: E2E Review Actions Closure

**Files:**
- Modify: `frontend-next/e2e/agent-runs.spec.ts`

- [ ] **Step 1: Extend mocks**

In `agent-runs.spec.ts`, update the detail mock so a follow-up operation can be appended after POST:

```ts
let currentAgentRun = structuredClone(mockAgentRun)

await page.route('**/api/v2/agent-runs/agent-run-1/followups', async route => {
  const request = route.request()
  if (request.method() !== 'POST') {
    await route.fallback()
    return
  }
  const body = request.postDataJSON()
  currentAgentRun = {
    ...currentAgentRun,
    operations: [
      ...currentAgentRun.operations,
      {
        id: 'op-followup-1',
        agent_run_id: 'agent-run-1',
        agent_id: 'agent-123',
        action: `followup.${body.action_type}`,
        risk_level: 'low',
        request: JSON.stringify(body),
        result: JSON.stringify({
          success: true,
          source_agent_run_id: 'agent-run-1',
          action_type: body.action_type,
          reason: body.reason,
          review_packet_id: body.review_packet_id,
          case_id: 'case-1',
          payload_id: 'payload-1',
        }),
        error: '',
        started_at: '2026-06-06T10:00:00Z',
        ended_at: '2026-06-06T10:00:00Z',
        created_at: '2026-06-06T10:00:00Z',
      },
    ],
  }
  await route.fulfill({
    json: {
      code: 0,
      message: 'success',
      data: {
        agent_run_id: 'agent-run-1',
        operation_id: 'op-followup-1',
        action_type: body.action_type,
        reason: body.reason,
        review_packet_id: body.review_packet_id,
        operation: currentAgentRun.operations.at(-1),
        created_at: '2026-06-06T10:00:00Z',
      },
    },
  })
})
```

- [ ] **Step 2: Add E2E test**

Add:

```ts
test('should create review follow-up action and refresh operation timeline', async ({ page }) => {
  await page.goto('/dashboard/agent-runs/agent-run-1')
  await page.waitForLoadState('networkidle')

  await page.getByRole('button', { name: '生成 JSON Review' }).click()
  await expect(page.getByText('Review Actions')).toBeVisible()

  await page.getByRole('combobox').selectOption('recheck_evidence')
  await page.getByPlaceholder('Reason for this follow-up').fill('Evidence needs second review')

  const followupRequest = page.waitForRequest(request =>
    request.method() === 'POST' &&
    request.url().includes('/api/v2/agent-runs/agent-run-1/followups')
  )
  await page.getByRole('button', { name: 'Create Follow-up' }).click()
  const request = await followupRequest
  const body = request.postDataJSON()

  expect(body.action_type).toBe('recheck_evidence')
  expect(body.reason).toBe('Evidence needs second review')
  expect(body.review_packet_id).toBe('agent-run-1')

  await expect(page.getByText('followup.recheck_evidence')).toBeVisible()
  await expect(page.getByText('操作历史 (3)')).toBeVisible()
})
```

If existing combobox locators conflict, scope the select with the Review Actions section using `page.getByText('Review Actions').locator('..')`.

- [ ] **Step 3: Run E2E**

Run:

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

Expected: PASS, no `show-report`, no blocking HTML report server.

## Task 7: Full Verification and Documentation

**Files:**
- Modify: `docs/verification.md` only if project convention requires adding this Sprint command set.

- [ ] **Step 1: Run backend focused tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

Expected: PASS.

- [ ] **Step 2: Run full Go tests**

```bash
GOCACHE=/tmp/gocache go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run frontend lint**

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

Expected: PASS.

- [ ] **Step 4: Run production build**

```bash
cd frontend-next && npm run build
```

Expected: PASS. If Turbopack fails only because the sandbox cannot bind a local port, rerun outside sandbox and record that detail.

- [ ] **Step 5: Run E2E**

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

Expected: PASS. This must be non-interactive and must not call `npx playwright show-report`.

## 验收清单

Codex 验收 Sprint M 时必须逐项检查：

1. `POST /api/v2/agent-runs/:id/followups` 是否真实存在。
2. Unknown Agent Run 是否返回 404。
3. Invalid action / empty reason / too long reason 是否返回 400。
4. Follow-up 是否写入 Agent Operation。
5. Operation result 是否包含 `source_agent_run_id`、`action_type`、`reason`、`review_packet_id`、`case_id`、`payload_id`。
6. Follow-up 是否写入 `agent_run.followup_created` audit。
7. Audit details 是否不泄露完整 API Key / Authorization / secret。
8. Agent Run detail UI 是否只能在 Review Packet 上下文中创建 follow-up。
9. E2E 是否真实等待 follow-up POST 请求，并断言 request body。
10. E2E 是否证明 Operation timeline 刷新出现 follow-up operation。
11. 是否没有 `test.skip`、`test.only`、`waitForTimeout`、只检查静态文字的空测。
12. 是否严格没有越界到 Scanner Hub、生命周期治理、批量操作、真实 replay engine、完整 Agent 管理平台。

## Windsurf 回传要求

Windsurf 完成后请回传：

- 修改文件列表。
- 每条验证命令和结果。
- 如果 E2E 失败，必须说明是代码失败还是环境失败，并贴出失败用例名。
- 是否改动了 `.windsurf/`、`.devin/`、Playwright config 或测试 reporter；如有，说明原因。
- 是否存在未提交或无关改动。
