# Phase 18 Spec B: 轮询 API + 请求回显

> 对应 ROADMAP 2.5 工具链深度集成

## 1. 轮询 API

### 1.1 端点

```
GET /api/v2/poll?cursor={timestamp}&limit={n}
```

### 1.2 行为

- `cursor`：游标时间戳（ISO8601 格式），只返回该时间戳**之后**创建的 Interaction
- `limit`：最大返回条数（默认 20，最大 100）
- 返回：Interaction 列表 + 新游标值（下次请求时传入）

### 1.3 响应格式

```json
{
  "code": 0,
  "data": {
    "interactions": [...],
    "next_cursor": "2026-07-12T10:00:05Z",
    "has_more": false
  }
}
```

### 1.4 实现

```go
// server/v2_api.go
func (self *WebServer) v2Poll(c *gin.Context) {
    cursor := c.Query("cursor") // ISO8601
    limitStr := c.DefaultQuery("limit", "20")
    limit, _ := strconv.Atoi(limitStr)
    if limit < 1 || limit > 100 {
        limit = 20
    }

    var cursorTime time.Time
    if cursor != "" {
        var err error
        cursorTime, err = time.Parse(time.RFC3339, cursor)
        if err != nil {
            cursorTime = time.Now().Add(-24 * time.Hour) // fallback to 24h ago
        }
    } else {
        cursorTime = time.Now().Add(-24 * time.Hour) // default: last 24h
    }

    iaSvc := interaction.NewService(self.orm, nil, nil)
    interactions, err := iaSvc.ListInteractions("", "", "", &cursorTime, nil, 1, limit)
    // ... response with next_cursor
}
```

## 2. 请求回显

### 2.1 行为

HTTP Listener 在响应体中回显收到的完整请求（请求行 + Header + Body）。

### 2.2 实现

在 `internal/listener/` 的 HTTP handler 中，记录请求后返回回显内容：

```
HTTP/1.1 200 OK
Content-Type: text/plain

=== Request Echo ===
POST /path HTTP/1.1
Host: token.domain.com
User-Agent: curl/7.0
Content-Type: application/x-www-form-urlencoded

body content here
===================
```

### 2.3 配置

通过 Listener 配置选项控制是否启用回显，默认开启。

## 3. 交付物

- `server/v2_api.go` — 轮询 API 端点
- `internal/listener/handler.go` — HTTP 请求回显
