# GODNSLOG 2.0 Sprint F Design

## 设计主题

Sprint F 聚焦首批控制面页面收口，范围限定为：

- `Cases Board`
- `Case Detail`
- `New Payload`
- `Payload Detail`

目标不是扩展更多页面，而是把 `Case -> Payload` 主操作链打磨成稳定、清晰、可验收的控制面闭环。

## 设计结论

Sprint F 采用 **主链路优先** 方案：

`Cases Board -> Case Detail -> New Payload -> Payload Detail`

该方案优先保证主业务路径顺滑，而不是把更多高密度操作堆回列表页。

## 为什么不用其他方案

### 不采用“控制台优先”

如果把更多快捷操作、统计和批量动作堆回 `Cases Board`，会导致：

- 列表页承担过多职责
- `Case Detail` 的价值被削弱
- Sprint F 范围变大，验收边界变差

### 不采用“Payload 优先”

如果让 `Payloads` 成为主入口，会弱化 `Case` 作为统一工作单元的产品叙事，不符合当前 2.0 的核心模型。

## 页面边界

### 1. Cases Board

定位：稳定入口，不是控制塔。

Sprint F 只保留以下能力：

- 搜索
- 状态筛选
- 进入 `Case Detail`
- 新建 `Case`

明确不做：

- 批量删除
- 重型快捷操作
- 大量统计卡
- 多层操作面板

设计要求：

- 列表页应轻量、可扫读
- 主 CTA 清晰
- 每一行的“进入详情”语义明确

### 2. Case Detail

定位：本 Sprint 的主工作页。

必须展示：

- Case 基本信息
- `payload_count`
- `interaction_count`
- `hit_payload_count`
- Payload 列表

必须提供：

- `Create Payload`
- 进入 `Evidence`
- 进入 `Interactions`

设计要求：

- `Case Detail` 要成为从组织对象进入执行对象的桥
- 页面应优先支持“围绕一个 Case 开展操作”

### 3. New Payload

定位：从 `Case` 进入执行面的创建向导。

保留：

- 向导式创建流程
- 模板选择
- 变量配置
- 创建前预览

必须增强：

- 从 `Case Detail` 进入时自动带入 `case_id`
- 页面上明确展示当前关联的 Case

不做：

- 扩展到 Payload 生命周期治理
- 自定义模板中心
- 批量复杂创建策略

### 4. Payload Detail

定位：闭环层页面，而不是纯展示页。

必须展示：

- token
- template
- rendered payload
- status
- `created_at`
- `expires_at`
- 关联 Case
- 最近交互

必须提供：

- 跳转到 `Interactions`
- 跳转到 `Evidence`

不做：

- revoke
- expire
- 手工改状态
- 生命周期管理操作

这页的目标是让 Payload 不只是“被创建出来”，而是能自然进入后续验证与证据查看。

## 交互与导航原则

Sprint F 的页面关系遵循以下原则：

1. `Cases Board` 负责进入
2. `Case Detail` 负责组织与分发操作
3. `New Payload` 负责创建
4. `Payload Detail` 负责连接后续闭环

页面不能相互抢职责：

- `Cases Board` 不替代 `Case Detail`
- `Payload Detail` 不承担生命周期管理
- `New Payload` 不扩成模板平台

## 视觉与信息架构原则

### Cases Board

- 以列表为主
- 搜索与筛选紧邻列表
- 主按钮清晰但不过度突出

### Case Detail

- 顶部为 Case 主信息
- 中部为统计摘要
- 下部为 Payload 列表和主操作

### New Payload

- 继续沿用当前向导结构
- 重点收口关联 Case 的上下文提示

### Payload Detail

- 顶部为 Payload 基本信息
- 中部为关联 Case 与最近交互
- 下部提供进入 `Interactions` / `Evidence` 的明确入口

## 验收导向

Sprint F 完成后，必须能回答“是”的问题：

1. 用户能否从 `Cases Board` 顺滑进入 `Case Detail`
2. 用户能否从 `Case Detail` 发起 `Create Payload`
3. `New Payload` 是否清楚地绑定当前 Case
4. `Payload Detail` 是否不再只是静态信息，而是能连接后续闭环
5. 当前收口是否没有扩展到批量操作、生命周期治理、模板平台等额外范围

## 明确留给后续 Sprint 的内容

以下内容不属于 Sprint F：

- Cases 批量操作增强
- Payload 生命周期管理
- 模板中心
- 更复杂的 Payload Studio
- Dashboard 全局壳层重构
- 更多控制面页面统一重排

## 最终范围结论

Sprint F 是一个 **主链路收口 Sprint**，不是能力扩展 Sprint。

它的价值在于把：

`Case -> Payload`

这一段控制面操作做成真正可用、可理解、可继续接入 `Interaction / Evidence` 的前置工作流。
