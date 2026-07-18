# Phase 19 Spec D: Template 插件化执行引擎

> 参考 Antenna 的 Template 组件系统设计

## 概述

在已有 marketplace Template CRUD 基础上，新增模板执行引擎，让用户编写的检测模板能实际处理交互事件。

## 模板格式

每个模板定义为一个 JSON/YAML，包含：

```json
{
  "name": "log4shell-detector",
  "description": "Detects Log4Shell JNDI lookups",
  "protocols": ["dns", "http"],
  "match": {
    "type": "regex",
    "pattern": "\\$\\{jndi:(ldap|rmi|dns)://"
  },
  "action": {
    "type": "tag",
    "value": "log4shell"
  },
  "severity": "critical"
}
```

## 架构

```
Interaction Created
       ↓
┌──────────────────┐
│  Template Engine │  ← 遍历已激活的模板，匹配执行
│  (executor)      │
└────────┬─────────┘
         ↓
   匹配成功?
   ┌──┴──┐
  YES    NO
   ↓     └→ 继续
  执行动作 (打标签/通知/拦截)
```

## 交付物

- `internal/marketplace/executor/` — 执行引擎包
  - `engine.go` — 主引擎
  - `executor.go` — 执行逻辑
- `internal/interaction/service.go` — 集成点（可选，补充 enhanceInteraction）
