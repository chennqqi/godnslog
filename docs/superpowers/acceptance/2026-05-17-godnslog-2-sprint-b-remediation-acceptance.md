# GODNSLOG 2.0 Sprint B 修正后复验

## 验收对象

- `internal/mcp/server.go`
- `internal/payload/service.go`
- `internal/payload/service_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `models/v2.go`

## 验收结论

**结论：不通过。**

本轮修正完成了两件正确的事：

1. Payload 创建主链路已接入 `RenderTemplateWithCase`
2. `v2GetPayload` / `v2PreviewPayload` 已补齐标准 404

但 Sprint B 的核心卡点仍未关闭，尤其是 MCP 仍未真正复用后端真实契约，因此不能进入 Sprint C。

## 本轮已完成的修正

### 1. `case` 变量已进入主创建链路

- `CreatePayload` 现在走 `renderPayloadWithCase`：
  [internal/payload/service.go](/data/dev/github.com/chennqqi/godnslog/internal/payload/service.go:38)

这是正确方向，解决了上轮“`case` 变量不在主路径”的问题。

### 2. Payload 读取与预览的 404 已收口

- `v2GetPayload` 不存在时返回 HTTP `404` / 业务 `404`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:990)
- `v2PreviewPayload` 不存在时返回 HTTP `404` / 业务 `404`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1093)

这两点通过。

## 仍未通过的关键问题

### 1. MCP 仍然没有复用后端真实契约

`create_oast_probe` 和 `create_payload` 仍在发送 `expires_in`：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:110)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:189)

而这次不是“后端兼容一下就算过”。Sprint B 的要求是：

- MCP 复用后端真实契约
- 不是给 MCP 再开一套兼容口

当前实现反而把 `expires_in` 加回了 API 请求模型：

- [models/v2.go](/data/dev/github.com/chennqqi/godnslog/models/v2.go:134)
- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:948)

这属于继续保留旁路语义，不符合 Sprint B 的收口目标。

### 2. `expected_protocol` 被错误地以数组形式传给 API

`create_oast_probe` 把 `expected_protocols` 归一化成切片后，直接塞进单数字段 `expected_protocol`：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:89)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:115)

后端契约定义的是单值字符串：

- [internal/models/payload.go](/data/dev/github.com/chennqqi/godnslog/internal/models/payload.go:139)

这不是小问题，而是明确的契约错配。现在之所以没暴露，是因为测试没有验证请求体字段类型。

### 3. 预览测试仍然是占位测试，没有验证 Sprint B 关键行为

`TestPayloadPreviewReturnsRenderedTemplate` 仍旧只是在未鉴权前提下检查“路由存在”：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:328)

它没有验证：

- 创建与预览使用同一渲染逻辑
- 不存在对象返回 404
- 返回值确实来自统一 `template_rendered`

这一项仍不满足 Sprint B 的测试要求。

### 4. Payload 服务测试虽增加，但关键断言仍然偏弱

新增测试覆盖面比上轮好，但 `case` 变量测试没有证明模板渲染结果中真的包含 `case`：

- [internal/payload/service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/payload/service_test.go:80)

当前断言只检查 `TemplateRendered` 非空，无法证明 `case` 变量渲染真的成立。

## 对 6 个验收问题的判断

1. Payload 契约是否真正向统一模型收敛：**部分满足，但未完成**
2. 模板渲染是否只有一套主逻辑：**基本满足**
3. 创建和预览是否复用同一渲染规则：**尚未被有效测试证明**
4. Probe 最小输出是否足够给后续 Interaction 使用：**部分满足**
5. MCP 是否复用后端真实契约：**否**
6. 是否严格没有越界：**是**

## 修正要求

Windsurf 继续停留在 Sprint B，补齐以下问题后再回传：

1. MCP 请求侧彻底收敛到后端真实契约，不再发送 `expires_in`
2. `create_oast_probe` 不得把 `[]string` 直接塞到 `expected_protocol`，必须对齐单值字段语义
3. 删除为 MCP 旁路兼容而新增到 API 请求模型中的临时字段，避免继续固化双轨契约
4. 把 `server/v2_api_test.go` 中的预览测试改成真实行为测试，而不是未鉴权占位测试
5. 把 `case` 变量渲染测试补成强断言，证明统一渲染主逻辑确实生效

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/payload ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果：测试通过，但**Sprint B 复验仍不通过**。
