# GODNSLOG 2.0 Sprint A 关闭验收结论

## 验收对象

- `Sprint A` 全部实施包与修正包
- Windsurf 最新一轮 closing patch
- 当前仓库实现状态

## 验收结论

**结论：有条件通过。**

Sprint A 的目标是建立统一领域模型与 API 契约基线。到当前状态，这一目标已经基本达成，可以进入 Sprint B，但需要挂明 1 个小型遗留项。

## 通过依据

### 1. 完整 API Key 鉴权已落地

- API Key 现在按完整 key 校验，不再只按 prefix
- `IsAgent` 已进入 server 侧桥接模型

### 2. 运行时 Swagger 已切换到 2.0 契约

[server/webserver.go](/data/dev/github.com/chennqqi/godnslog/server/webserver.go:255) 已明确将 Swagger UI 指向 `/docs/openapi.yaml`。

### 3. 旧业务错误码体系已基本清理完毕

`server/middleware.go` 中的旧 `code: 4/5` 已全部收敛。

当前扫描结果显示，仅剩 1 处残留：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2197)

### 4. 测试基线保持通过

本轮重新执行：

```bash
GOCACHE=/tmp/gocache go test ./server ./internal/auth ./...
```

结果通过。

## 遗留项

### P2：`v2GetEvidence` 仍残留旧业务码

[server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2193) 中 `v2GetEvidence` 仍返回：

- HTTP `404`
- 业务 `code: 4`

该问题仍属于契约一致性问题，但它不在 Sprint A 明确要求的核心 5 条路径：

- `/api/v2/auth/login`
- `/api/v2/cases`
- `/api/v2/payloads`
- `/api/v2/interactions`
- `/api/v2/apikeys`

因此该项不再阻断 Sprint A 关闭，但必须在 Sprint B 开始前顺手清掉。

## 验收判断

### 结论说明

Sprint A 已达到“可进入下一 Sprint”的标准，因此给予 **有条件通过**：

- 允许进入 `Sprint B`
- 必须将 `v2GetEvidence` 的旧错误码收口列为 Sprint B 的进入前修正项或首个小 patch

## 下一步

`Codex` 可以开始编排 `Sprint B：Probe 创建与 Payload 渲染`。
