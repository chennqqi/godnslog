# GODNSLOG 2.0 需求文档

## Phase 11: 测试框架开发

为项目开发完整的测试框架，包括后端单元测试和前端端到端测试。

### 后端单元测试

已创建以下测试文件：
- internal/auth/apikey_test.go
- internal/payload/payload_test.go
- internal/canary/model_test.go
- internal/listener/model_test.go

### 前端E2E测试

已创建以下测试文件：
- frontend-next/playwright.config.ts
- frontend-next/e2e/login.spec.ts
- frontend-next/e2e/dashboard.spec.ts

### 运行测试

```bash
# 后端测试
go test ./...

# 前端E2E测试
cd frontend-next
cnpm install
cnpm run test:e2e
```

## 2026-05-03
按照2.0的规划开始开发，每一轮开发完成后在开发计划中标记进度

- 按照 ROADMAP_2.0.md 和 DEVELOPMENT_PLAN_2.0.md 执行开发
- 从 Phase 0 开始，逐步推进各个阶段
- 每轮开发完成后在 DEVELOPMENT_PLAN_2.0.md 中标记进度

## 2026-05-04
Phase 12：测试框架扩展
- 扩展interaction模块测试
- 扩展mcp模块测试
- 扩展listener模块测试
- 添加CLI测试
- 修复所有编译错误
- 所有测试通过
- 创建CLI使用文档
- 创建MCP Server使用文档
- 更新开发计划标记交付物完成

Phase 13：文档完善
- 修复所有编译错误
- 扩展测试覆盖
- 创建CLI使用文档
- 创建MCP Server使用文档
- 更新开发计划标记交付物完成
- 验证MVP验收标准

Phase 14：SMB/FTP Listener
- 扩展数据模型（SMBRequest、FTPCommand）
- 实现SMB Listener逻辑
- 实现FTP Listener逻辑
- 扩展Store接口
- 添加Handler端点
- 添加模型测试
- 更新文档
- 所有测试通过

Phase 15：企业级功能
- 实现数据保留策略
- 实现数据归档功能
- 实现高可用配置
- 添加企业级测试
- 创建文档
- 所有测试通过

Phase 16：插件市场/模板市场
- 实现插件模型
- 实现模板模型
- 实现插件服务
- 实现模板服务
- 添加插件/模板测试
- 创建文档
- 所有测试通过

## 2026-05-05
用户反馈2.0版本虽然标记完成，但存在严重问题：
1. 很多1.0的核心功能在2.0都丢失了：
   - 日志记录
   - API自动化操作
   - 多用户支持
   - 一些payload可见性
   - 一些文档
2. 2.0的功能实现都很草率，根本没有深入理解意思，去认真实现
3. 前端实现很粗糙，跟demo html没有区别，完全看不出来是一个生产可用的项目

根据ROADMAP_2.0.md开始完整的开发流程，分步骤以保证质量：
- 需求梳理 ✅ 已完成（doc/requirements-analysis-2026-05-05.md）
- 程序设计 ✅ 已完成（doc/design-plan-2026-05-05.md）
- 程序开发
- 测试验收
- 程序修复
- 完成验收

## 2026-05-05 (下午)
实现v2_api.go中的TODO标记API端点业务逻辑：

### 创建Service层
- internal/canary/service.go - Canary服务层，提供Canary的CRUD操作
- internal/rebinding/service.go - Rebinding服务层，提供Rebinding Rule的CRUD操作
- internal/listener/service.go - Listener服务层，提供Listener的CRUD操作

### 实现Canary API端点
- v2ListCanaries - 列出所有Canary Token
- v2CreateCanary - 创建新的Canary Token
- v2GetCanary - 获取指定Canary Token
- v2UpdateCanary - 更新Canary Token
- v2DeleteCanary - 删除Canary Token
- v2ListCanaryHits - 列出Canary Token的命中记录

### 实现Rebinding API端点
- v2ListRebindingRules - 列出所有Rebinding规则
- v2CreateRebindingRule - 创建新的Rebinding规则
- v2GetRebindingRule - 获取指定Rebinding规则
- v2UpdateRebindingRule - 更新Rebinding规则
- v2DeleteRebindingRule - 删除Rebinding规则
- v2ListRebindingSessions - 列出Rebinding会话

### 实现Listener API端点
- v2ListListeners - 列出所有Listener
- v2CreateListener - 创建新的Listener
- v2GetListener - 获取指定Listener
- v2UpdateListener - 更新Listener
- v2DeleteListener - 删除Listener
- v2ListListenerInteractions - 列出Listener的交互记录

### 实现Evidence API端点
- v2GetEvidence - 获取Evidence报告（添加说明：Evidence报告按需生成，不持久化存储）

