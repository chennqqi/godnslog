# 仓库协作指南

## 项目结构与模块组织

GODNSLOG 是一个 Go 后端 + 前端管理界面的安全工具，用于 DNS/HTTP OAST 回连验证、SSRF/XXE/RFI/RCE 等漏洞验证，以及 2.0 规划中的证据链、扫描器协同和 AI Agent 赋能。

根目录 Go 入口包括 `main.go`、`servecmd.go`、`resetpass.go`。旧版后端服务代码在 `server/`，数据库模型在 `models/`，缓存逻辑在 `cache/`，SDK 与示例在 `client/` 和 `examples/`。2.0 新模块主要放在 `internal/`、`cmd/`、`cli/`、`migration/`、`templates/`、`docs/` 等目录。新版前端在 `frontend-next/`，旧 Vue 前端如存在则仅作为历史兼容参考。文档、需求、设计稿和截图主要放在 `doc/`、`docs/`、`res/`。

## 构建、测试与开发命令

- `go build`：从仓库根目录构建后端。
- `go test ./...`：运行全部 Go 测试。
- `go run . serve -domain example.com -4 127.0.0.1`：本地启动服务。
- `go run ./cmd/cli --help`：查看 2.0 CLI 帮助。
- `go run ./cmd/mcp-server`：启动 MCP Server。
- `cd frontend-next && npm install`：安装新版前端依赖。
- `cd frontend-next && npm run dev`：启动新版前端开发服务。
- `cd frontend-next && npm run build`：构建新版前端。
- `cd frontend-next && npm run lint`：运行新版前端 lint。
- `docker build -t user/godnslog .`：构建容器镜像；国内环境可使用 `-f DockerfileCN`。

## 编码风格与命名约定

Go 代码必须使用 `gofmt` 格式化。包名保持短小、小写，并与目录职责一致。只有跨包 API 才使用导出标识符。新增 2.0 后端能力优先放在 `internal/<domain>/`，保持模块边界清晰，例如 `internal/payload/`、`internal/interaction/`、`internal/mcp/`。

新版前端使用 Next.js、TypeScript、Tailwind/shadcn/ui。组件和业务代码应按既有 `frontend-next/src/` 结构组织，页面保持功能完整，避免只做静态占位。界面修改需要注意响应式、可读性和真实数据接入。

## 测试规范

Go 测试与被测代码放在同目录，文件名使用 `*_test.go`，测试函数使用 `TestXxx`。后端改动至少运行相关包测试，影响共享模型、API、权限、Interaction、Payload、MCP 或 CLI 时应扩大到 `go test ./...`。

前端行为测试优先使用项目已有的 Playwright/Jest 配置。涉及可见 UI 改动时，应补充或更新相关 E2E/组件测试，并在最终说明中写明已运行的验证命令。

完整验证命令维护在 `docs/verification.md`，完成任何功能前必须记录实际运行过的命令和结果。

## 2.0 产品与实现重点

2.0 的核心定位是“面向安全团队、扫描器和 AI Agent 的自托管 OAST 证据中枢”。规划和实现应优先服务以下闭环：

- 创建 Case 和可追踪 Payload。
- 捕获 DNS/HTTP 等 Interaction。
- 自动归因到 Payload、Case、目标和测试人。
- 生成可解释证据链和报告。
- 通过 CLI、OpenAPI、MCP、插件和 Webhook 服务外部工具。

Scanner Hub 不应只集成 Nuclei，还应规划和支持 Burp Suite、Yakit/Yak、ZAP、xray/rad、Postman/Apifox、CI/CD 等主流工具。AI Agent 能力应优先提供高层、安全、可审计工具，例如 `create_oast_probe`、`wait_for_interaction`、结构化证据摘要和最小权限 APIKey。

## 提交与 PR 规范

提交标题保持简短、聚焦单一变化。PR 应包含变更摘要、测试命令与结果、关联 issue，以及前端可见变更的截图或录屏。不要把无关格式化、生成文件或临时调试改动混入功能提交。

## 安全与配置注意事项

不要提交真实域名、API Token、生产数据库、生成凭据或敏感回连数据。默认管理员密码会在首次运行时输出，可通过 `resetpw` 修改。DNS、Callback、Rebinding、Workflow 出站请求、MCP APIKey 和 Agent 操作都应视为敏感能力，默认采用最小权限、过期时间、审计日志和高风险动作禁用策略。

## codex约定
- sub agent总是使用GPT-5.4 Mini模型

## windsurf约定
- sub agent总是使用GPT-5.4 Mini模型
- 执行前端 E2E 时，禁止使用会触发 `Serving HTML report at http://localhost:9323. Press Ctrl+C to quit.` 的 Playwright 打开式报告流程。
- Windsurf 运行 E2E 只允许使用一次性、非交互式命令并直接退出，例如 `npx playwright test --reporter=line`、`npx playwright test --reporter=list`，不得执行 `npx playwright show-report`，不得让测试进程因本地报告服务常驻阻塞。
