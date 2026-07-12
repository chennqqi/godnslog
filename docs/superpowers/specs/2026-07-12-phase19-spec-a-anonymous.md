# Phase 19 Spec A: 匿名模式

## 设计

在配置中新增 `anonymous_mode` 开关，开启后：
1. Interaction 的 SourceIP 记录为 `0.0.0.0`（不记录真实来源）
2. 日志默认 24 小时自动过期（由 retention 策略执行）
3. 前端不显示用户注册信息
4. 会话使用临时 token，不持久化

## 实现

**配置**：`WebServerConfig` 新增 `AnonymousMode bool`

**数据层**：在 `internal/interaction/service.go` 的 `CreateInteraction` 中，检查 `AnonymousMode`，如果开启则覆盖 SourceIP 为 `0.0.0.0`

**前端**：在设置页面新增匿名模式开关

**交付物**：
- `server/webserver.go` — 配置项
- `internal/interaction/service.go` — 匿名化逻辑
