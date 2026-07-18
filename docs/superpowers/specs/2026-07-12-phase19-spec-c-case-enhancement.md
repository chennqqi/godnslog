# Phase 19 Spec C: 任务化 Case 模型优化

## 概述

增强 Case 模型以支持批量检测场景聚合，参考 Antenna 的任务驱动设计。

## 变更

### 数据模型扩展

在 `internal/models/case.go` 的 `Case` 结构体中新增字段：

```go
Type       string   `xorm:"varchar(32) default('single')" json:"type"` // single, batch, scan
Progress   int      `xorm:"int default(0)" json:"progress"`            // 0-100 batch progress
TotalTargets int    `xorm:"int default(0)" json:"total_targets"`       // total targets in batch
HitTargets int      `xorm:"int default(0)" json:"hit_targets"`         // targets with hits
```

### API 扩展

- `GET /api/v2/cases/stats` — 返回 Case 统计（总数/活跃/批量/完成）
- Case 列表响应增加 `type`, `progress`, `total_targets`, `hit_targets` 字段

### 交付物

- `internal/models/case.go` — 字段扩展
- `server/v2_api.go` — stats 端点
