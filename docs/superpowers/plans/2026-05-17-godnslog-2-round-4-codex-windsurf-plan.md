# GODNSLOG 2.0 Round 4 Codex-Windsurf Collaboration Plan

**定位**：从本轮开始，职责明确分离：

- **Codex**：负责规划、设计、任务切分、阶段验收、风险控制、范围管理
- **Windsurf**：负责具体开发实施、联调、自测、提交执行结果

Codex 不直接承担业务实现，Windsurf 不自行扩展规划范围。

## 本轮目标

把现有设计成果转化为一套可执行的实施协作机制，使后续开发能够稳定推进而不重新回到“边做边改方向”的状态。

## 本轮输入

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/official-support-boundary.md`
- `docs/implementation-dependencies.md`
- 当前工程基线和 `/api/v2` 实际现状

## 本轮输出

本轮应由 Codex 产出实施治理包，由 Windsurf 按包执行。

### 输出 1：Phase-to-Sprint 切分

Codex 需要把当前 5 个大 Phase 继续下沉成可执行 Sprint 包。每个 Sprint 包必须满足：

- 范围单一
- 依赖清楚
- 可独立测试
- 可独立验收
- 不跨多个主线同时推进

建议首批切分如下：

1. Sprint A：统一领域模型与 API 口径
2. Sprint B：Probe 创建与 Payload 渲染
3. Sprint C：DNS/HTTP Interaction 捕获与归因
4. Sprint D：Evidence 汇总、评分、导出
5. Sprint E：Audit 链路与 CLI 对齐
6. Sprint F：首批控制面页面
7. Sprint G：Primary 工具集成
8. Sprint H：Agent 治理与 Agent Run

### 输出 2：标准实施包模板

Codex 负责定义 Windsurf 每次实施必须遵守的任务模板。每个实施包至少包含：

- **目标**
- **输入文档**
- **实施范围**
- **禁止越界项**
- **依赖前置条件**
- **必须修改的文件范围**
- **必须补齐的测试**
- **完成定义**
- **Codex 验收问题**

Windsurf 只能在模板规定的范围内实施，不得自行扩展到未批准的功能。

### 输出 3：交付回传模板

Windsurf 每次完成实施后，必须向 Codex 回传统一格式的结果：

- 实际修改范围
- 实际执行的验证命令
- 测试结果
- 未完成项
- 风险与阻塞项
- 偏离规划的地方

Codex 只基于这份回传和仓库现状做验收，不接受口头“已经完成”作为交付依据。

### 输出 4：阶段验收机制

Codex 对 Windsurf 的每个 Sprint 只做三类判断：

1. **通过**：范围完成，验证充分，可进入下一个 Sprint
2. **有条件通过**：主目标完成，但有明确遗留项，需在后续 Sprint 补齐
3. **不通过**：范围漂移、验证不足、关键闭环未完成，必须回退到当前 Sprint

## Codex 职责边界

Codex 负责：

- 维护统一术语和规划文档
- 把 Phase 切成 Sprint
- 定义每个 Sprint 的输入输出
- 定义每个 Sprint 的验收标准
- 审核 Windsurf 的实施结果
- 阻止范围漂移

Codex 不负责：

- 直接写业务实现代码
- 替 Windsurf 做具体联调
- 在未验收当前 Sprint 前跳到下一个 Sprint

## Windsurf 职责边界

Windsurf 负责：

- 按 Codex 下发的 Sprint 包实施
- 在限定范围内补代码、补测试、补文档
- 运行验证命令并回传结果
- 说明偏差、阻塞和技术风险

Windsurf 不负责：

- 修改产品定位
- 擅自新增能力范围
- 改写术语定义
- 绕过验收直接推进下个阶段

## 协作流程

### Step 1：Codex 下发 Sprint 包

Codex 基于现有规划文档，生成单个 Sprint 的实施说明，范围必须足够小，确保 Windsurf 能独立完成。

### Step 2：Windsurf 实施并自测

Windsurf 根据 Sprint 包完成实现，执行规定的验证命令，整理交付回传。

### Step 3：Codex 验收

Codex 根据：

- 仓库实际变化
- 验证命令结果
- 规划范围是否被遵守
- 是否满足完成定义

给出通过/有条件通过/不通过结论。

### Step 4：进入下一个 Sprint

只有当前 Sprint 被 Codex 验收通过后，才能进入下一 Sprint。

## 本轮首要任务

本轮优先不是开发，而是把 Sprint A 的实施包写清楚。Sprint A 应只覆盖：

- 统一领域模型
- `/api/v2` 统一响应格式
- 鉴权和权限口径统一
- OpenAPI/Swagger 对齐

这一步由 Codex 规划，由 Windsurf 实施。

## 本轮验收标准

只有满足以下条件，本轮才算完成：

- 已经明确写清 Codex 和 Windsurf 的职责边界
- 已经定义标准实施包模板
- 已经定义标准回传模板
- 已经定义 Sprint 级别的协作流程
- 已经明确下一步由 Windsurf 先执行哪个 Sprint

## 下一步建议

Round 4 完成后，Codex 应直接产出 `Sprint A` 的实施包；Windsurf 开始按该包执行第一轮真正开发。
