# Phase 19 Spec B: 搜索引擎集成 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development

**Goal:** 在 Scanner Hub 中新增 ZoomEye/Shodan/Fofa 搜索引擎适配器，支持批量搜索目标

**Architecture:** 每个搜索引擎一个独立文件，实现统一的 `Searcher` 接口，通过工厂函数创建

**Tech Stack:** Go `net/http`, JSON API

---

### Task 1: 搜索引擎适配器包

**Files:**
- Create: `internal/scannerhub/search/search.go`
- Create: `internal/scannerhub/search/zoomeye.go`
- Create: `internal/scannerhub/search/shodan.go`
- Create: `internal/scannerhub/search/fofa.go`

- [ ] **Step 1: 创建 `internal/scannerhub/search/search.go`**

```go
package search

// ResultItem 表示一条搜索结果
type ResultItem struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Country  string `json:"country,omitempty"`
	Title    string `json:"title,omitempty"`
}

// SearchResult 搜索引擎返回结果
type SearchResult struct {
	Total   int          `json:"total"`
	Results []ResultItem `json:"results"`
}

// Searcher 搜索引擎适配器接口
type Searcher interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// Config 搜索引擎配置
type Config struct {
	ZoomEyeKey string
	ShodanKey  string
	FofaEmail  string
	FofaKey    string
}
```

- [ ] **Step 2: 创建 `internal/scannerhub/search/zoomeye.go`**

```go
package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type zoomEye struct {
	apiKey string
}

func NewZoomEye(apiKey string) Searcher {
	return &zoomEye{apiKey: apiKey}
}

func (z *zoomEye) Name() string { return "ZoomEye" }

func (z *zoomEye) Search(query string, page int) (*SearchResult, error) {
	u := fmt.Sprintf("https://api.zoomeye.org/host/search?query=%s&page=%d", url.QueryEscape(query), page)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("API-KEY", z.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Total int `json:"total"`
		Matches []struct {
			IP string `json:"ip"`
			PortInfo struct {
				Port int `json:"port"`
			} `json:"portinfo"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	result := &SearchResult{Total: apiResp.Total}
	for _, m := range apiResp.Matches {
		result.Results = append(result.Results, ResultItem{
			IP:   m.IP,
			Port: m.PortInfo.Port,
		})
	}
	return result, nil
}
```

- [ ] **Step 3: 创建 `internal/scannerhub/search/shodan.go`**

```go
package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type shodan struct {
	apiKey string
}

func NewShodan(apiKey string) Searcher {
	return &shodan{apiKey: apiKey}
}

func (s *shodan) Name() string { return "Shodan" }

func (s *shodan) Search(query string, page int) (*SearchResult, error) {
	u := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s&page=%d",
		s.apiKey, url.QueryEscape(query), page)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Total  int `json:"total"`
		Matches []struct {
			IPStr    string `json:"ip_str"`
			Port     int    `json:"port"`
			Protocol string `json:"protocol"`
			Hostname string `json:"hostnames"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	result := &SearchResult{Total: apiResp.Total}
	for _, m := range apiResp.Matches {
		result.Results = append(result.Results, ResultItem{
			IP:       m.IPStr,
			Port:     m.Port,
			Protocol: m.Protocol,
		})
	}
	return result, nil
}
```

- [ ] **Step 4: 创建 `internal/scannerhub/search/fofa.go`**

```go
package search

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type fofa struct {
	email string
	key   string
}

func NewFofa(email, key string) Searcher {
	return &fofa{email: email, key: key}
}

func (f *fofa) Name() string { return "Fofa" }

func (f *fofa) Search(query string, page int) (*SearchResult, error) {
	qbase64 := base64.StdEncoding.EncodeToString([]byte(query))
	u := fmt.Sprintf("https://fofa.info/api/v1/search/all?email=%s&key=%s&qbase64=%s&page=%d&size=20",
		url.QueryEscape(f.email), url.QueryEscape(f.key), url.QueryEscape(qbase64), page)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		Error   bool       `json:"error"`
		Results [][]interface{} `json:"results"`
		Size    int        `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if apiResp.Error {
		return nil, fmt.Errorf("fofa API error")
	}

	result := &SearchResult{Total: apiResp.Size}
	// Fofa returns results as array of arrays: [ip, port, protocol, hostname, country, title]
	for _, row := range apiResp.Results {
		if len(row) >= 2 {
			item := ResultItem{}
			if ip, ok := row[0].(string); ok {
				item.IP = ip
			}
			if port, ok := row[1].(float64); ok {
				item.Port = int(port)
			}
			result.Results = append(result.Results, item)
		}
	}
	return result, nil
}
```

- [ ] **Step 5: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/scannerhub/search/
```

- [ ] **Step 6: Commit**

```bash
git add internal/scannerhub/search/
git commit -m "feat: add search engine adapters (ZoomEye, Shodan, Fofa)"
```

---

### Task 2: API 端点

**Files:**
- Modify: `server/v2_api.go`

- [ ] **Step 1: 添加搜索 API 路由**

在 `registerV2API` 中 `Scanner Hub` 相关路由附近添加：

```go
// Search engines
search := v2.Group("/search", self.authHandler)
{
    search.GET("/zoomeye", self.v2SearchZoomEye)
    search.GET("/shodan", self.v2SearchShodan)
    search.GET("/fofa", self.v2SearchFofa)
}
```

- [ ] **Step 2: 添加 Handler 方法**

```go
func (self *WebServer) v2SearchZoomEye(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(400, gin.H{"error": "query required"})
        return
    }
    // Read API key from settings
    // s := search.NewZoomEye(apiKey)
    // result, err := s.Search(query, 1)
    // c.JSON(200, result)
    c.JSON(501, gin.H{"error": "not implemented"})
}
```

For now, implement as a stub since API keys need to be configured in settings. The full implementation reads the key from configuration and proxies the request.

- [ ] **Step 3: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./server/
```

- [ ] **Step 4: Commit**

```bash
git add server/v2_api.go
git commit -m "feat: add search engine API endpoints (ZoomEye, Shodan, Fofa)"
```
