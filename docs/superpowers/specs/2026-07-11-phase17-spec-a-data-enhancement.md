# Phase 17 Spec A: 数据增强层

> 解码引擎 + 利用类型分类器
> 对应 ROADMAP 2.4 智能增强版

## 1. 概述

在 Interaction 写入链路中插入两个增强模块，对原始交互数据进行自动化处理：

- **解码引擎**：自动识别并还原 DNS/HTTP 外带数据中的编码内容
- **类型分类器**：基于规则自动标注交互的利用类型

两者都是纯函数式处理，无外部依赖，无副作用，可以独立测试。

## 2. 架构位置

```
External Request
     ↓
Listener (DNS/HTTP/LDAP/...)
     ↓
Interaction Record Created
     ↓
┌─────────────────────────────┐
│   Data Enhancement Pipeline │
│  ┌───────────────────────┐  │
│  │  Decoder Engine       │  │  ← 新
│  │  (auto-detect +       │  │
│  │   decode base32/hex/  │  │
│  │   base64)             │  │
│  └────────┬──────────────┘  │
│           ↓                 │
│  ┌───────────────────────┐  │
│  │  Exploit Classifier   │  │  ← 新
│  │  (rule-based pattern  │  │
│  │   matching → type +   │  │
│  │   confidence)         │  │
│  └────────┬──────────────┘  │
└───────────┼─────────────────┘
            ↓
Interaction Enriched
(decoded_data, exploit_type,
 confidence, enrichments)
     ↓
Store to DB + WebSocket Broadcast
```

## 3. 解码引擎 (Decoder)

### 3.1 包结构

```
internal/interaction/decoder/
  decoder.go       # 主入口 DetectAndDecode()
  base32.go        # base32 编解码
  hex.go           # hex 编解码
  base64.go        # base64 + URL-safe base64 编解码
  decoder_test.go  # 测试
```

### 3.2 核心接口

```go
// Result 表示一次解码的结果
type Result struct {
    Original  string `json:"original"`
    Decoded   string `json:"decoded"`
    Encoding  string `json:"encoding"` // "base32", "hex", "base64", ""
    Confident bool   `json:"confident"` // true=高置信度, false=猜测
}

// Decode 对单个字符串进行自动解码检测
// 依次尝试 base64 → base32 → hex，返回第一个高置信度结果
// 若无匹配，返回 Original="" 且 Encoding="" 的 Result
func Decode(input string) Result

// DecodeAll 对输入尝试所有编码并返回全部可能结果（用于展示）
func DecodeAll(input string) []Result

// 内部检测器接口
type detector interface {
    Detect(input string) bool     // 是否可能是这种编码
    Decode(input string) (string, bool) // 解码，返回 (明文, 是否成功)
}
```

### 3.3 编码检测策略

| 编码 | 检测条件 | 置信度条件 |
|------|---------|-----------|
| **base64** | 长度 % 3 特征 + `[A-Za-z0-9+/=]` 或 `[A-Za-z0-9\-_=]` | 解码后为可打印 ASCII 或 UTF-8 |
| **base32** | 长度 % 8 特征 + `[A-Z2-7=]` | 解码后全为可读文本 |
| **hex** | 全部 `[0-9a-fA-F]`，偶数长度 | 解码后可打印 ASCII 或 UTF-8 |

**优先级**: base64 → base32 → hex（按实际出现频率排序，base64 最常见）

### 3.4 输入源映射

解码引擎针对 Interaction 的不同字段自动提取输入：

| Interaction Type | 扫描字段 | 优先级 |
|-----------------|---------|--------|
| DNS | Domain 中 token 后的子域名部分 | 高 |
| HTTP | Body、Path、特定 Header | 高 |
| LDAP | Filter、BaseDN | 中 |
| 其他 | RawData | 低 |

### 3.5 边界情况

- **空输入**：直接返回空 Result
- **短输入 (< 4 chars)**：跳过 base64 检测（太短误报高），尝试 hex
- **混合内容**：如 `abc_base64data_xyz`，用正则提取候选段再尝试
- **URL 编码**：先做 URLDecode 再做编码检测

## 4. 类型分类器 (Classifier)

### 4.1 包结构

```
internal/interaction/classifier/
  classifier.go       # 主入口 Classify()
  classifier_test.go  # 测试
  rules.go            # 规则定义
  patterns.go         # 模式常量
```

### 4.2 核心接口

```go
// Classification 分类结果
type Classification struct {
    ExploitType string `json:"exploit_type"` // "log4shell", "fastjson", "ssrf", "xxe", "sqli", "rce"...
    Confidence  string `json:"confidence"`   // "high", "medium", "low"
    RuleID      string `json:"rule_id"`      // 匹配的规则 ID，用于追溯
    Evidence    string `json:"evidence"`      // 匹配到的具体特征片段
}

// Classify 对 Interaction 进行分类
// 返回匹配到的第一个高置信度分类，若无则继续降置信度匹配
// 完全不匹配返回 nil
func Classify(interaction *Interaction) *Classification

// ClassifyAll 返回所有匹配的分类（从高到低）
func ClassifyAll(interaction *Interaction) []*Classification
```

### 4.3 规则定义

每条规则包含：

```go
type Rule struct {
    ID          string             // 唯一标识
    Name        string             // 规则名
    ExploitType string             // 分类结果
    Confidence  string             // high/medium/low
    Priority    int                // 优先级（数字越小越优先）
    Matchers    []Condition        // 匹配条件（AND 关系）
}

type Condition interface {
    Match(interaction *Interaction) (matched bool, evidence string)
}
```

### 4.4 规则表（初始版本）

| ID | 类型 | 置信度 | 匹配条件 | 优先级 |
|----|------|--------|---------|--------|
| `log4shell-jndi` | log4shell | high | DNS Domain 含 `${jndi:ldap://}` 或 `${jndi:rmi://}` | 10 |
| `log4shell-http` | log4shell | high | HTTP User-Agent 含 `${jndi:` 或 Header 含 `${jndi:` | 10 |
| `log4shell-dns` | log4shell | high | DNS Domain 含 `jndi` 且长度有明显编码特征 | 20 |
| `fastjson` | fastjson | high | HTTP Body 含 `{"@type":"` 或 `parseObject` | 10 |
| `fastjson-header` | fastjson | medium | Content-Type `application/json` 且 Body 含 `@type` | 20 |
| `xxe-dns` | xxe | high | DNS Domain 含 `xxe` 且在 Payload 模板的 xxe 分组内 | 10 |
| `xxe-http` | xxe | high | HTTP Path 含 `xxe` 或 `xml` | 10 |
| `ssrf-http` | ssrf | medium | HTTP Path 含 `/redirect` `/proxy` `/fetch` | 20 |
| `ssrf-dns` | ssrf | medium | 由 ssrf 类 Payload 模板生成且命中 | 30 |
| `sqli-http` | sqli | medium | HTTP Path 含 `?id=` `?page=` `?query=` | 20 |
| `sqli-dns` | sqli | medium | DNS Domain 符合盲注外带特征（长 subdomain、编码数据） | 30 |
| `rce-dns` | rce | medium | DNS Domain 含 `cmd` `exec` `whoami` 等关键字 | 20 |
| `rce-http` | rce | medium | HTTP Path 含 `/cmd` `/exec` 或 Body 含命令执行特征 | 20 |
| `deserialization` | deserialization | medium | HTTP Body 含 Java 序列化魔数 `aced0005` | 10 |
| `ldap-injection` | ldap | low | DNS Domain 或 HTTP Path 含 `ldap` 关键词 | 40 |

### 4.5 分类增强策略

除了规则匹配外，还结合以下信息辅助分类：

- **Payload 模板关联**：如果 Interaction 关联了已知 Payload，直接从 Payload 的 TemplateID 映射到利用类型
- **路径关键词**：HTTP Path 中出现已知漏洞关键词则提升置信度
- **域名模式**：DNS Domain 中出现 `r2` `rce` `payload` 等模式

### 4.6 未匹配回退

当所有规则都不匹配时：
- 返回 `nil`（不标注）
- Interaction 的 `ExploitType` 字段保持空值
- 前端展示为"未识别"

## 5. Integration 集成点

### 5.1 在 Interaction Service 中调用

```go
// internal/interaction/service.go 增强

func (s *Service) CreateInteraction(ctx context.Context, req *CreateInteractionRequest) (*Interaction, error) {
    // 1. 创建原始 Interaction
    interaction := buildInteraction(req)
    
    // 2. 解码增强
    decodedField := extractDecodeCandidate(interaction)
    if decodedField != "" {
        result := decoder.Decode(decodedField)
        if result.Confident {
            interaction.DecodedData = &result.Decoded
            interaction.Encoding = &result.Encoding
        }
    }
    
    // 3. 分类增强
    classification := classifier.Classify(interaction)
    if classification != nil {
        interaction.ExploitType = classification.ExploitType
        interaction.Confidence = classification.Confidence
    }
    
    // 4. 持久化
    _, err := s.store.Insert(interaction)
    return interaction, err
}
```

### 5.2 数据模型扩展

在现有的 `Interaction` 结构体中新增字段：

```go
// Interaction 新增字段
DecodedData *string `xorm:"text" json:"decoded_data"`          // 解码后的明文
Encoding    *string `xorm:"varchar(32)" json:"encoding"`       // 检测到的编码类型
ExploitType *string `xorm:"varchar(64) index" json:"exploit_type"` // 利用类型
Confidence  *string `xorm:"varchar(16)" json:"confidence"`     // 置信度
```

### 5.3 数据库 Migration

在 `internal/interaction/migration.go` 中添加新字段的 migration：

```sql
ALTER TABLE interactions ADD COLUMN decoded_data TEXT;
ALTER TABLE interactions ADD COLUMN encoding VARCHAR(32);
ALTER TABLE interactions ADD COLUMN exploit_type VARCHAR(64);
ALTER TABLE interactions ADD COLUMN confidence VARCHAR(16);
CREATE INDEX idx_interactions_exploit_type ON interactions(exploit_type);
```

## 6. 测试策略

### 解码引擎测试

```go
// 解码成功场景
Decode("eyJhZG1pbiI6InRydWUifQ==") → {Decoded: `{"admin":"true"}`, Encoding: "base64", Confident: true}
Decode("NZ2WY3DPFYZW44TQJ5AUQ")      → {Decoded: "hello", Encoding: "base32", Confident: true}
Decode("68656c6c6f")                  → {Decoded: "hello", Encoding: "hex", Confident: true}

// 解码失败场景
Decode("hello")                       → {Original: "hello", Encoding: "", Confident: false}
Decode("123")                         → {Original: "123", Encoding: "", Confident: false}
Decode("")                            → {Original: "", Encoding: "", Confident: false}

// 混合内容
Decode("data:eyJhIjoxfQ==")           → 提取 base64 段并解码
```

### 分类器测试

```go
// DNS log4shell
interaction := &Interaction{
    Type:   "dns",
    Domain: "exp.jndi.payload.abc123.dnslog.fun",
}
Classify(interaction) → {ExploitType: "log4shell", Confidence: "high"}

// HTTP XXE
interaction := &Interaction{
    Type: "http",
    Path: "/xxe",
}
Classify(interaction) → {ExploitType: "xxe", Confidence: "high"}

// 无匹配
interaction := &Interaction{
    Type:   "dns",
    Domain: "random.test.example.com",
}
Classify(interaction) → nil
```

### 边缘测试

- 解码器：超大输入（>10KB）、二进制输入、Unicode 输入
- 分类器：空 Interaction、无 Domain 和 Path 的 Interaction、所有字段全空

## 7. 交付物清单

- `internal/interaction/decoder/decoder.go` — 解码器入口
- `internal/interaction/decoder/base32.go` — base32 解码
- `internal/interaction/decoder/hex.go` — hex 解码
- `internal/interaction/decoder/base64.go` — base64 解码
- `internal/interaction/decoder/decoder_test.go` — 解码器测试
- `internal/interaction/classifier/classifier.go` — 分类器入口
- `internal/interaction/classifier/rules.go` — 规则定义
- `internal/interaction/classifier/patterns.go` — 模式常量
- `internal/interaction/classifier/classifier_test.go` — 分类器测试
- `internal/interaction/interaction.go` — 数据模型扩展（新增字段）
- `internal/interaction/migration.go` — 数据库 migration
- `internal/interaction/service.go` — 集成解码和分类到创建流程

## 8. 不纳入范围

- ❌ 不涉及前端展示（前端在 Spec C 攻击链时间线中统一使用这些字段）
- ❌ 不涉及 ML/AI 分类器（未来可替换，当前规则引擎足够）
- ❌ 不涉及解码历史记录重算（只对新写入的 Interaction 生效）
