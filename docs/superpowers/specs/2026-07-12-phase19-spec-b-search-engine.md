# Phase 19 Spec B: 搜索引擎集成

## 概述

在 Scanner Hub 中新增搜索引擎适配器，支持从 ZoomEye、Shodan、Fofa 批量搜索目标并导入为扫描目标。

## 设计

### 适配器模式

遵循现有 scannerhub 的 adapter 模式，新增 3 个搜索引擎适配器：

```go
// SearchAdapter 搜索引擎适配器接口
type SearchAdapter interface {
    Name() string
    Search(query string, page int) (*SearchResult, error)
}
```

### 搜索引擎

| 引擎 | API | 认证方式 | 免费额度 |
|------|-----|---------|---------|
| ZoomEye | `https://api.zoomeye.org/host/search` | API Key (Header) | 每月 10,000 |
| Shodan | `https://api.shodan.io/shodan/host/search` | API Key (Query) | 每月 1,000 |
| Fofa | `https://fofa.info/api/v1/search/all` | Email + Key | 每月 10,000 |

### 数据流

```
用户输入 query → Search Engine Adapter → API 请求 → 返回 IP:Port 列表
                                                          ↓
                                                   创建 ScannerRun
                                                          ↓
                                                   用户执行扫描
```

### 交付物

- `internal/scannerhub/search/` — 搜索引擎适配器包
  - `search.go` — 主入口 + SearchAdapter 接口
  - `zoomeye.go` — ZoomEye 适配器
  - `shodan.go` — Shodan 适配器
  - `fofa.go` — Fofa 适配器
- `server/v2_api.go` — API 端点
- 前端搜索结果展示组件
