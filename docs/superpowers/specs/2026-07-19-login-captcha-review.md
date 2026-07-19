# Design Spec Review: Login Captcha

## 准确性问题

### 1. LoginRequest 定义位置不准确
文档说 `LoginRequest` 在 `server/webui.go:321` 修改，但实际定义在 `models/api.go:29-33`，`server/models.go:28` 只是类型别名：
```go
type LoginRequest models.LoginRequest
```
修改需在 `models/api.go` 中进行，而非 `webui.go`。

### 2. 遗漏 v2 登录端点
项目有两个登录 handler：
- v1: `server/webui.go:321` — `userLogin`（`/api/v1/login`，已废弃）
- v2: `server/v2_api.go:318` — `v2Login`（`/api/v2/auth/login`，前端实际使用）

前端 API client baseURL 为 `/api/v2`（`frontend-next/src/lib/api.ts:36`），调用 `/auth/login`，实际路径是 `/api/v2/auth/login`。文档只提到修改 `webui.go` 的 handler，遗漏了 `v2_api.go` 中的 `v2Login`，这才是前端实际调用的端点。

### 3. API 路径前缀不一致
文档写 `GET /api/auth/captcha` 和 `POST /api/auth/login`，但实际 v2 路由注册在 `/api/v2` group 下（`server/v2_api.go:48-52`）。新端点应为 `GET /api/v2/auth/captcha`，或在文档中明确说明是相对于 v2 base 的路径。

### 4. 配置字段 json tag 风格
`CaptchaExpire` 字段缺少 json tag（其他 Duration 字段如 `AuthExpire` 也没有 json tag，风格一致，可以接受）。

## 遗漏项

### 5. 未提及 go.mod 依赖更新
需添加 `github.com/wenlng/go-captcha/v2` 到 `go.mod`，应明确版本约束。

### 6. 未提及 servecmd.go 的 flag 注册
`servecmd.go:45-63` 中 `servePwCmd.SetFlags` 需新增 `-captcha-enabled` 和 `-captcha-expire` flag，并在 `Execute` 中传入 `WebServerConfig`（`servecmd.go:84-101`）。文档未提及这一关键修改点。

### 7. 未提及 login-schema.ts 更新
`frontend-next/src/features/auth/schemas/login-schema.ts` 当前只验证 `username` 和 `password`。captcha 字段需在 schema 外处理（因为 captcha_value 是数字且来自组件回调），或 schema 需扩展。文档应说明如何与 react-hook-form 集成。

### 8. v1 登录端点策略未说明
v1 `/api/v1/login` 已标记为 deprecated。文档应明确 v1 是否也加 captcha 验证，还是仅 v2 加。建议：v1 保持不变（向后兼容），仅 v2 加 captcha。

### 9. 与 Demo 模式的交互未考虑
Roadmap 2.9 中规划了 Demo 模式。Demo 模式下 captcha 应如何处理？建议：Demo 模式自动禁用 captcha，或预填 captcha 值。

### 10. 现有测试将受影响
`server/v2_api_test.go` 中大量测试直接调用 `/api/v2/auth/login` 且不带 captcha 字段。启用 captcha 后这些测试将全部失败。文档应说明：
- 测试中如何绕过 captcha（如 `CaptchaEnabled: false` 配置）
- 或在测试 helper 中封装 captcha 获取流程

## 设计建议

### 11. 容差 ±3px 可能过严
go-captcha 滑块验证通常使用 ±5px 或更大容差。3px 在高 DPI 屏幕或触屏设备上可能误判。建议：
- 默认 ±5px
- 通过 `CaptchaTolerance` 配置项可调

### 12. captcha 禁用时的前端探测机制
文档说禁用时 `GET /api/auth/captcha` 返回 404，前端隐藏组件。但更好的做法是提供 `GET /api/v2/auth/captcha/status` 或在 `GET /api/v2/health` 中返回 `captcha_enabled` 字段，避免用 404 做控制流。

### 13. 缓存 key 设计
`captcha:{captcha_id}` 建议加前缀 `godnslog:captcha:{id}` 避免在共享 Redis 中与其他 key 冲突。

### 14. 缺少速率限制考虑
文档第 226 行提到 "same rate limiting applies"，但当前代码中没有登录速率限制。应明确 captcha 是 rate limiting 的补充还是替代，以及是否需要新增 login rate limiting。

## 总结

| 类别 | 数量 |
|------|------|
| 准确性问题（需修正） | 4 |
| 遗漏项（需补充） | 6 |
| 设计建议（推荐采纳） | 4 |

最关键的问题是遗漏 v2 登录端点（#2）和 `servecmd.go` flag 注册（#6），不修正这两项会导致实现直接失败。建议修正后再进入实现阶段。
