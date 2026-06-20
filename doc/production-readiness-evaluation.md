# GODNSLOG 2.0 生产可用性评估

- **评估日期**：2026-06-20
- **代码基线**：`9ca01da`（`feature/phase1-backend-realization`）
- **评估结论**：**尚未达到生产可用标准；ROADMAP 2.0 目标并未全部达成。**

---

## 1. 总体结论

后端当前可编译、单元测试可全部通过，但**工程化、代码质量、测试覆盖、容器化、前端 E2E 与部署**存在明显且多方面的缺口。与 `ROADMAP_2.0.md` 和已批准的 `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` 对照，当前代码处于“Phase 1 后端夯实”起点，大量标记为“已完成”的 Phase 实际上仍有 stub、占位符或测试缺失。

---

## 2. 验证结果（实际执行）

### 2.1 后端

| 检查项 | 命令 | 结果 | 说明 |
|---|---|---|---|
| 编译 | `go build ./...` | ✅ 通过 | Go 1.25.8 可完整编译 |
| 单元测试 | `go test ./...` | ✅ 通过 | 46 个 `*_test.go` 文件，22 个包有测试 |
| 测试覆盖 | `go test -cover ./...` | ⚠️ 偏低 | 关键包如 `internal/canary` 0%、`internal/listener` 20.1%、`internal/rule` 9.9%、`server` 18.3% |
| 静态检查 | `go vet ./...` | ❌ 失败 | 3 个问题：`server/dnsserver.go:222` 不可达代码、`server/dnsserver.go:401` 未命名字段、`servecmd.go:135` 未缓冲 signal channel |
| 格式化 | `gofmt -l .` | ❌ 未通过 | 26 个文件未格式化 |

### 2.2 前端

| 检查项 | 命令 | 结果 | 说明 |
|---|---|---|---|
| 依赖安装 + 构建 | `cnpm install && npm run build` | ✅ 通过 | Next.js 16.2.6 静态产物生成成功 |
| Lint | `npm run lint` | ❌ 失败 | 48 个问题（28 errors，20 warnings） |
| E2E 测试 | `npx playwright test --reporter=line` | ❌ 失败 | 117 个用例：92 passed，24 failed，1 skipped |

### 2.3 容器化

| 检查项 | 文件 | 结果 | 说明 |
|---|---|---|---|
| Dockerfile | `Dockerfile` | ❌ 无法交付前端 | 前端构建产物默认在 `.next/`，但镜像复制 `/app/dist`，导致 `/app/dist` 不存在 |
| DockerfileCN | `DockerfileCN` | ❌ 版本不匹配 | 使用 `golang:1.22-alpine`，但 `go.mod` 要求 `go 1.25.0`；同样复制 `/app/dist` |

### 2.4 关键代码缺口（TODO / 占位 / 未实现）

- `internal/rule/action.go:200`：`sendEmailNotification` 直接返回 `email notification not implemented`。
- `internal/scannerhub/service.go:170`：`evidence_count` 硬编码为 0，未实现证据表查询。
- `server/v2_api.go:3480`：audit log RBAC 未实现，当前所有认证用户可查看全部审计日志。
- `internal/workflow/service.go`：存在多个 `TODO: Implement` action executor（HTTP、DNS、Webhook、Notify）。
- `server/router.go`：v1 的 record/data/user/setting 管理端点被注释或未实现。
- `internal/auth/middleware.go`：workspace_id 映射、int64 ID 转换仍有 TODO。

---

## 3. ROADMAP 2.0 目标达成情况

| 版本 / 目标 | 状态 | 说明 |
|---|---|---|
| **2.0 MVP 核心闭环** | 部分达成 | Case / Payload / Interaction / APIKey 的 API 与前端页面存在，但 lint 未通过、E2E 大面积失败、Workflow action 仍未实现。 |
| **2.1 扫描器协同版** | 部分达成 | Scanner Hub 有 adapter 列表和 Burp 扩展，但证据回填、Nuclei 模板示例、ZAP/Yakit 插件仍为示例或占位。 |
| **2.2 Agent 赋能版** | 部分达成 | MCP 工具与 Agent Run API 已存在，但 MCP 协议实现、E2E 验证、review 闭环尚未完全到位。 |
| **2.3 平台化版本** | 名义存在 | Workspace、Canary、Rebinding、Listener、HA、Marketplace 数据模型和 API 已落地，但测试覆盖极低、功能未充分验证。 |
| **1.0 核心功能保留** | 未达成 | v1 管理端点被注释，xip/DNS 解析在 v2 中已暴露，但前端和测试覆盖不足。 |

---

## 4. 生产可用性判定

**当前状态：不可直接用于生产。**

### 可接受的部分
- 本地开发模式下，后端可编译启动、单元测试通过。
- 前端 dev 模式可运行，核心页面可渲染。
- v2 API 路由覆盖 Case、Payload、Interaction、APIKey、User、Rule、Evidence、Canary、Rebinding、Listener、Settings、Scanner Hub、Agent Runs、HA Cluster 等。

### 不可接受的部分（阻塞项）
1. **容器镜像无法交付**：前端产物路径错误，部署后无法访问 Web UI。
2. **前端 lint 未通过**：28 个 error 直接说明代码存在类型和 React Hooks 问题。
3. **E2E 测试与 UI 不同步**：24 个失败大多因为测试仍查找中文文案，而 UI 已改为英文。
4. **Go 静态检查失败**：`go vet` 报出不可达代码和 signal channel 使用错误。
5. **代码未格式化**：`gofmt` 列出 26 个文件，不符合 Go 工程规范。
6. **测试覆盖率远低于目标**：用户目标是 90%，当前大量核心包低于 30%。
7. **核心功能仍有占位**：邮件通知、标签/报告/噪声动作、工作流执行器、证据计数等未真正落地。

---

## 5. 建议下一步

1. **按 `2026-06-20-mvp-to-production-design.md` 进入 Phase 1**：优先完成后端 stub 清理、1.0 功能 v2 暴露、APIKey 安全硬化、数据模型双写。
2. **建立 CI 门禁**：将 `go build`、`go test ./...`、`go vet ./...`、`gofmt -l .`、`npm run lint`、`npm run build` 全部加入 CI。
3. **修复容器化**：Next.js 启用 `output: 'export'` 或改为独立静态产物目录；对齐 Dockerfile 与 `go.mod` 的 Go 版本。
4. **统一前端测试语言**：将所有 E2E 断言文案改为英文，与当前 UI 保持一致。
5. **提升测试覆盖率**：为核心包（listener、rule、canary、interaction、server）补齐测试，逐步逼近 90% 目标。
6. **补充 GitHub Actions / GitLab CI**：当前 `examples/ci/` 仅提供示例，仓库本身缺少持续集成流水线。
7. **安全加固**：修复 `servecmd.go:135` 的 signal channel 问题；落实 APIKey bcrypt 存储（已在 spec 中规划）。

---

## 附录：本次评估实际执行的验证命令

```bash
# 后端
go build ./...
go test ./...
go test -cover ./...
go vet ./...
gofmt -l .

# 前端
cnpm install
npm run build
npm run lint
npx playwright install chromium
CI=1 npm run test:e2e
```

---

*本评估基于当前代码基线的实际可运行结果，不依赖历史文档中的“已完成”标记。*
