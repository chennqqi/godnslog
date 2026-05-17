# GODNSLOG 2.0 Productization Phase Plan

**定位**：本文件是下一阶段规划，不包含具体实施步骤。Codex 在该阶段负责规划、设计、验收标准和阶段治理，不负责代替执行实现。

## 阶段目标

把当前已经具备的 2.0 零散能力，收敛成一个真正有竞争力的产品面，使 GODNSLOG 在三个维度形成清晰优势：

1. **能力优势**：不是单一 DNSLOG，而是面向安全验证的 OAST Evidence Hub。
2. **AI Agent 友好**：不是“可以被 Agent 调用”，而是“天然适合 Agent 安全地调用和审计”。
3. **多工具集成**：不是只接 Nuclei，而是具备可持续扩展的 Scanner Hub 适配器体系。

## 当前阶段输入

下一阶段以以下成果为输入，不重复做基线修复：

- 后端测试已可全量验证
- `/api/v2` 已成为统一演进入口
- `docs/scanner-hub.md` 已定义基础集成合同
- `docs/CLI_USAGE.md` 与 `docs/MCP_SERVER_USAGE.md` 已提供 CLI 与 Agent 入口

## 下一阶段产出物

本阶段必须交付四类文档与设计结论。

### 1. 产品定位与差异化说明

输出一份统一产品说明，回答四个问题：

- GODNSLOG 2.0 不是谁
- GODNSLOG 2.0 主要解决什么问题
- 为什么它比传统 DNSLOG/回连平台更强
- 为什么安全团队和 AI Agent 都愿意优先使用它

建议形成一页式能力地图：

- Core: Case / Payload / Interaction / Evidence
- Access: Web / API / CLI / MCP
- Integrations: Scanner Hub / Workflow / Notification
- Governance: API Key / Audit / Retention / Workspace

### 2. Scanner Hub 适配器体系设计

目标不是继续罗列工具，而是定义“怎么持续接工具”。

需要明确：

- 适配器分层：原生适配、脚本适配、Webhook 桥接
- 标准输入：创建 Probe、传递变量、等待 Interaction、导出结果
- 标准输出：JSON、JSONL、SARIF、Webhook Event
- 兼容等级：L1 文档样例、L2 官方脚本、L3 官方插件、L4 双向联动

首批纳入主支持矩阵：

- Nuclei
- Burp Suite
- Yakit/Yak
- ZAP
- xray/rad
- Postman/Apifox

其中要单独定义 Burp Suite 和 Yakit/Yak 的价值，因为它们更接近人工验证与半自动验证场景，是区别于“只跑扫描器”的关键。

### 3. Agent-Native 产品规范

下一阶段必须把 MCP 从“可调用工具”提升为“可治理的 Agent 操作面”。

需要补齐：

- Agent 身份：Agent、Operator、Workspace 的关系
- Agent Run：一次任务的创建、追踪、结束、导出
- 权限范围：按 Case、Payload、工具动作、导出范围做最小授权
- 风险分级：创建探针、等待回连、导出证据、删除/撤销、敏感配置修改
- 审计标准：谁在什么时间，以什么参数，对哪个目标，生成了什么证据

这一部分的核心竞争力不是“接入 MCP”，而是“让企业敢把 MCP 给 Agent 用”。

### 4. 统一控制面规划

当前文档已经有 CLI、MCP、Scanner Hub 三条线，但缺少统一控制面。下一阶段需要定义产品主界面信息架构，至少覆盖：

- Cases 工作台
- Payload 生成与模板中心
- Interaction 时间线与聚类视图
- Evidence 汇总与导出
- Integrations 页面
- Agent Runs 页面
- Audit / API Keys / Workspace 设置

控制面的目标不是展示所有能力，而是让三类用户都能闭环：

- 安全工程师：快速发起验证、筛选证据、导出结论
- 自动化平台：稳定调用 API/CLI
- AI Agent Operator：安全地发放权限、查看回放、控制风险

## 规划原则

### 原则 1：优先统一，不优先堆功能

后续任何新能力，必须优先挂接到统一实体模型：

- Case
- Payload
- Interaction
- Evidence
- Agent Run
- API Key
- Audit Event

### 原则 2：优先做“闭环能力”，不做“点状特性”

优先级应按闭环排序：

1. 生成 Probe
2. 投递到工具或 Agent
3. 捕获 Interaction
4. 汇总为 Evidence
5. 导出与审计

不能再单独追逐某个工具、某个监听协议或某个展示页面。

### 原则 3：先定义成熟度，再扩展工具数

工具支持不按“有没有名字”判断，而按成熟度判断：

- 是否有官方样例
- 是否有标准参数映射
- 是否有结果回读机制
- 是否能进入统一审计链路

## 下一阶段工作流

下一阶段不直接进入实现，按以下顺序推进：

1. 先补一份产品设计规格书：统一产品定位、角色、控制面、差异化叙事。
2. 再补一份集成设计规格书：Scanner Hub 适配器体系与成熟度矩阵。
3. 再补一份 Agent 设计规格书：身份、权限、审计、风险分级。
4. 最后再基于三份规格书生成新的实施计划。

## 阶段验收标准

只有同时满足以下条件，下一阶段规划才算完成：

- 有一份统一的产品定位文档，而不是分散说明
- 有一份 Scanner Hub 支持矩阵，明确主流工具与成熟度等级
- 有一份 Agent-Native 规范，明确权限、风险、审计
- 有一份统一控制面信息架构，能解释 Web/API/CLI/MCP 如何协同
- 每份文档都能映射到明确的后续实施范围，避免再次出现“方向很多，但落地边界不清”

## 预期决策输出

当本阶段规划完成后，应能做出以下明确决策：

- GODNSLOG 2.0 的一句话定位
- 首批官方主支持工具名单
- Agent 默认允许与默认禁止的动作边界
- Web 控制面的一级导航结构
- 下一轮实施应该先做产品面、Agent 面，还是工具适配面
