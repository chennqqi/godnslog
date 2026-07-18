# Phase 18 Spec A: 来源指纹归属

> 内置扫描器/云厂商 IP 库 + 可选 GeoIP ASN 查询
> 对应 ROADMAP 2.5 工具链深度集成

## 1. 概述

在 Interaction 处理链路中新增指纹模块，自动识别请求来源的类型：扫描器、云厂商、还是真实目标。GeoIP/ASN 增强作为可选能力，依赖外部 .mmdb 文件。

## 2. 架构

```
Interaction Created
       │
       ▼
┌─────────────────────────┐
│   Fingerprint Module    │
│  ┌───────────────────┐  │
│  │ Scanner Detector  │  │  ← 内置已知扫描器 IP + User-Agent
│  │ (static rules)    │  │
│  └────────┬──────────┘  │
│           ↓             │
│  ┌───────────────────┐  │
│  │ Cloud Detector    │  │  ← 内置云厂商 CIDR
│  │ (CIDR matching)   │  │
│  └────────┬──────────┘  │
│           ↓             │
│  ┌───────────────────┐  │
│  │ GeoIP / ASN       │  │  ← 可选 .mmdb 增强
│  │ (optional)        │  │
│  └────────┬──────────┘  │
└───────────┼─────────────┘
            ↓
   Interaction.SourceFingerprint
```

## 3. 包结构

```
internal/interaction/fingerprint/
  fingerprint.go      # 主入口 Lookup(ip, userAgent) → Fingerprint
  scanners.go         # 内置扫描器规则
  clouds.go           # 云厂商 CIDR 列表
  geoip.go            # GeoIP/ASN 查询（可选，无 .mmdb 时降级）
  fingerprint_test.go # 测试
```

## 4. 数据模型

```go
// Fingerprint 表示一个 IP 的来源指纹
type Fingerprint struct {
    IP         string `json:"ip"`
    SourceType string `json:"source_type"` // "scanner", "cloud", "known", "unknown"
    SourceName string `json:"source_name,omitempty"` // "Nuclei", "AWS", "Cloudflare", etc.
    ASN        uint   `json:"asn,omitempty"`
    Org        string `json:"org,omitempty"` // "CLOUDFLARENET", "AMAZON-02", etc.
    Country    string `json:"country,omitempty"`
}

const (
    SourceScanner = "scanner"
    SourceCloud   = "cloud"
    SourceKnown   = "known"   // 已知非扫描器服务（如 DNS 解析器）
    SourceUnknown = "unknown"
)
```

## 5. 扫描器检测

### 5.1 已知扫描器 IP 范围

内置静态规则，覆盖常见安全工具：

```go
var scannerCIDRs = []struct {
    name string
    cidr *net.IPNet
}{
    // Nuclei / ProjectDiscovery (Cloud)
    {"ProjectDiscovery", mustCIDR("45.33.0.0/16")},
    
    // Burp Suite Collaborator (取决于部署)
    
    // Nessus / Tenable
    {"Tenable", mustCIDR("54.0.0.0/8")},
    
    // Acunetix (AWVS)
    // 基于已知扫描节点
    
    // Netsparker
    // 基于已知扫描节点
    
    // Shodan
    {"Shodan", mustCIDR("141.0.0.0/16")},
    
    // Censys
    {"Censys", mustCIDR("162.0.0.0/16")},
    
    // BinaryEdge
    {"BinaryEdge", mustCIDR("185.0.0.0/16")},
    
    // 中国常见扫描器（知道创宇、360、腾讯等）
}
```

### 5.2 User-Agent 指纹

HTTP Interaction 通过 User-Agent 辅助判断：

```go
var scannerUAMap = map[string]string{
    "Nuclei":        "Nuclei",
    "Go-http-client": "Go-http-client", // 常见扫描器底层
    "Mozilla/5.0 (compatible; Nmap Scripting Engine)": "Nmap",
}

var scannerUAKeywords = []struct {
    keyword string
    name    string
}{
    {"nuclei", "Nuclei"},
    {"nessus", "Nessus"},
    {"acunetix", "Acunetix"},
    {"awvs", "AWVS"},
    {"netsparker", "Netsparker"},
    {"python-requests", "Python Scanner"},
    {"zgrab", "ZGrab"},
}
```

### 5.3 DNS 查询特征

某些扫描器使用特定 DNS 解析器 IP，可通过 DNS Interaction 的 SourceIP 识别。

## 6. 云厂商检测

### 6.1 内置 CIDR 列表

```go
var cloudCIDRs = []struct {
    name string
    cidr *net.IPNet
}{
    {"AWS", mustCIDR("13.32.0.0/15")},      // AWS Global Accelerator
    {"AWS", mustCIDR("52.0.0.0/15")},        // AWS US-East
    {"Cloudflare", mustCIDR("103.21.244.0/22")},
    {"Cloudflare", mustCIDR("104.16.0.0/13")},
    {"Google Cloud", mustCIDR("34.0.0.0/15")},
    {"Google Cloud", mustCIDR("35.0.0.0/16")},
    {"Azure", mustCIDR("13.64.0.0/11")},
    {"Azure", mustCIDR("20.0.0.0/10")},
    {"Alibaba", mustCIDR("8.0.0.0/11")},
    {"DigitalOcean", mustCIDR("159.65.0.0/16")},
    {"Vultr", mustCIDR("45.32.0.0/16")},
    {"Linode", mustCIDR("45.33.0.0/16")},
    {"OCI", mustCIDR("129.0.0.0/16")},        // Oracle Cloud
}
```

每个云厂商 2-5 个典型 CIDR 段，覆盖大部分场景。

## 7. GeoIP/ASN 增强

### 7.1 可选依赖

```go
// 启动时检查 .mmdb 文件，不存在则跳过
import "github.com/oschwald/geoip2-golang"
```

### 7.2 文件路径

- 默认查找路径: `./data/GeoLite2-ASN.mmdb`
- 可通过配置文件或环境变量覆盖: `GODNSLOG_MMDB_PATH`
- 文件不存在时，GeoIP 功能自动降级，不影响其他指纹功能

### 7.3 查询逻辑

```go
func (f *Fingerprinter) lookupGeoIP(ip net.IP) {
    if f.asnDB == nil {
        return
    }
    record, err := f.asnDB.ASN(ip)
    if err != nil {
        return
    }
    f.asnDB.ASN(ip)  // 使用 maxminddb-golang 直接解码
}
```

## 8. 主入口

```go
type Fingerprinter struct {
    asnDB     *geoip2.Reader
    mu        sync.RWMutex
}

// NewFingerprinter 创建指纹识别器
// mmdbPath 可选，为空则跳过 GeoIP
func NewFingerprinter(mmdbPath string) *Fingerprinter

// Lookup 对 IP 进行来源指纹识别
func (f *Fingerprinter) Lookup(ip, userAgent string) *Fingerprint

// ReloadDB 重新加载 .mmdb 文件（用于热更新）
func (f *Fingerprinter) ReloadDB(mmdbPath string) error
```

优先级：Scanner → Cloud → GeoIP/ASN → Unknown
- 先匹配扫描器规则（最高优先级）
- 再匹配云厂商 CIDR
- 如果有 GeoIP 数据则补充 ASN/Org
- 都不匹配则为 Unknown

## 9. 集成到 Interaction

在 `internal/interaction/service.go` 的 `CreateInteraction` 中增强：

```go
// 在 enhanceInteraction 之后
if s.fingerprinter != nil && interaction.SourceIP != "" {
    fp := s.fingerprinter.LookUp(interaction.SourceIP, interaction.UserAgent)
    if fp != nil && fp.SourceType != fingerprint.SourceUnknown {
        interaction.SourceType = &fp.SourceType
        interaction.SourceName = &fp.SourceName
    }
}
```

Interaction 模型新增字段：
```go
SourceType *string `json:"source_type,omitempty" xorm:"'source_type' varchar(32)"`
SourceName *string `json:"source_name,omitempty" xorm:"'source_name' varchar(128)"`
```

## 10. 测试

```go
func TestScannerDetection(t *testing.T) {
    f := NewFingerprinter("")
    
    // Shodan
    fp := f.LookUp("141.0.0.1", "")
    if fp == nil || fp.SourceType != SourceScanner {
        t.Error("expected Shodan scanner detection")
    }
    
    // AWS
    fp = f.LookUp("52.0.0.1", "")
    if fp == nil || fp.SourceType != SourceCloud {
        t.Error("expected AWS detection")
    }
    
    // Normal IP
    fp = f.LookUp("8.8.8.8", "curl/7.0")
    if fp == nil || fp.SourceType != SourceUnknown {
        t.Error("expected unknown for 8.8.8.8")
    }
    
    // UA-based scanner
    fp = f.LookUp("1.2.3.4", "nuclei/3.0")
    if fp == nil || fp.SourceType != SourceScanner {
        t.Error("expected scanner detection from UA")
    }
}
```

## 11. 交付物

- `internal/interaction/fingerprint/fingerprint.go` — 主入口
- `internal/interaction/fingerprint/scanners.go` — 扫描器规则
- `internal/interaction/fingerprint/clouds.go` — 云厂商 CIDR
- `internal/interaction/fingerprint/geoip.go` — GeoIP 可选查询
- `internal/interaction/fingerprint/fingerprint_test.go` — 测试
- `internal/interaction/service.go` — 集成指纹到 CreateInteraction
- `internal/models/interaction.go` — 新增 SourceType/SourceName 字段
