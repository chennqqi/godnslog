# Phase 19 Spec C: 任务化 Case 模型优化 — 实现计划

- [ ] **Step 1: 扩展 Case 数据模型**

在 `internal/models/case.go` 的 `Case` 结构体中添加字段（在 `Tags` 之后）：

```go
	Type         string `xorm:"varchar(32) default('single')" json:"type"`          // single, batch, scan
	Progress     int    `xorm:"int default(0)" json:"progress"`                     // 0-100
	TotalTargets int    `xorm:"int default(0)" json:"total_targets"`
	HitTargets   int    `xorm:"int default(0)" json:"hit_targets"`
```

- [ ] **Step 2: 验证编译**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/models/
```

- [ ] **Step 3: 添加 Case Stats API**

在 `server/v2_api.go` 中添加：

```go
// GET /api/v2/cases/stats
func (self *WebServer) v2CaseStats(c *gin.Context) {
	type CaseStats struct {
		Total    int64 `json:"total"`
		Active   int64 `json:"active"`
		Archived int64 `json:"archived"`
		Batch    int64 `json:"batch"`
	}
	var stats CaseStats
	stats.Total, _ = self.orm.Count(&models.Case{})
	stats.Active, _ = self.orm.Where("status = 'active'").Count(&models.Case{})
	stats.Archived, _ = self.orm.Where("status = 'archived'").Count(&models.Case{})
	stats.Batch, _ = self.orm.Where("type = 'batch' OR type = 'scan'").Count(&models.Case{})
	c.JSON(200, gin.H{"code": 0, "data": stats})
}
```

路由注册：
```go
caseRoutes.GET("/stats", self.v2CaseStats)
```

- [ ] **Step 4: 编译 + 测试**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./...
go test ./internal/case/ -v 2>&1 | grep -E "PASS|FAIL|ok"
```

- [ ] **Step 5: Commit**

```bash
git add internal/models/case.go server/v2_api.go
git commit -m "feat: enhance Case model with batch task fields and stats API"
```
