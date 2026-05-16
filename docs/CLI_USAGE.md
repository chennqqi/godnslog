# GODNSLOG CLI 使用文档

## 概述

`godnslog-cli` 是 GODNSLOG 2.0 的命令行工具，用于创建 Case、生成 Payload、轮询 Interaction 和导出报告。适用于 Nuclei、CI/CD 和脚本化使用场景。

## 安装

```bash
# 从源码构建
go build -o godnslog-cli ./cmd/cli

# 或使用 Go install
go install github.com/chennqqi/godnslog/cmd/cli@latest
```

## 配置

CLI 通过环境变量或命令行参数配置：

```bash
# 环境变量
export GODNSLOG_API_URL="http://localhost:8080"
export GODNSLOG_API_KEY="your-api-key"

# 或命令行参数
godnslog-cli --api-url http://localhost:8080 --api-key your-api-key
```

## 命令

### 1. 创建 Case

```bash
godnslog-cli case create \
  --title "SSRF 测试" \
  --description "测试目标系统的SSRF漏洞" \
  --target "example.com" \
  --tags "ssrf,oast"
```

### 2. 列出 Case

```bash
godnslog-cli case list
```

### 3. 生成 Payload

```bash
godnslog-cli payload create \
  --template "http://{{.Token}}.example.com" \
  --case-id "case-123" \
  --expires-in "24h" \
  --variables '{"target": "example.com"}'
```

### 4. 等待 Interaction

```bash
godnslog-cli interaction wait \
  --token "your-token" \
  --timeout 300 \
  --format json
```

### 5. 列出 Interaction

```bash
godnslog-cli interaction list \
  --case-id "case-123" \
  --limit 50 \
  --format json
```

### 6. 导出报告

```bash
godnslog-cli report export \
  --case-id "case-123" \
  --format markdown \
  --output report.md
```

## 输出格式

CLI 支持多种输出格式：

- `json`：JSON 格式（默认）
- `yaml`：YAML 格式
- `markdown`：Markdown 格式（仅报告）

## Scanner Hub 集成

GODNSLOG 提供统一的 Scanner Hub 集成接口，支持多种安全扫描器工具。详细的集成合同请参考 [Scanner Hub Integration Contract](./scanner-hub.md)。

### 支持的工具

- Nuclei: CLI wrapper 和模板变量
- Burp Suite: 扩展调用 REST API
- Yakit/Yak: Yak script 调用 REST API 并轮询 token
- ZAP: 脚本或插件调用 REST API 并轮询 token
- xray/rad: CLI 或 webhook 桥接映射扫描器事件到 Case 和 Payload
- Postman/Apifox: 环境变量和预请求脚本

### 集成示例

创建 Probe 并等待结果的完整流程：

```bash
# 1. 创建 Payload (Probe)
curl -X POST http://localhost:8080/api/v2/payloads \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "template": "ssrf-url",
    "case_id": "case-123",
    "variables": {"path": "/callback"},
    "expected_protocols": ["dns", "http"],
    "tool": "burp-suite"
  }'

# 2. 等待 Interaction 结果
curl -X GET "http://localhost:8080/api/v2/interactions?token=<token>&page_size=10" \
  -H "Authorization: Bearer <token>"
```

## Nuclei 集成

### JSONL 输出

```bash
godnslog-cli payload create \
  --template "{{.Token}}.example.com" \
  --case-id "nuclei-scan" \
  --format jsonl > payloads.txt
```

### Nuclei 模板示例

```yaml
id: godnslog-ssrf
info:
  name: GODNSLOG SSRF Test
  severity: medium
  author: GODNSLOG
requests:
  - raw:
      |
      GET /?url={{godnslog-url}} HTTP/1.1
      Host: {{Host}}
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: GODNSLOG Scan

on:
  push:
    branches: [ main ]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Create GODNSLOG Case
        run: |
          godnslog-cli case create \
            --title "CI/CD Scan" \
            --target "${{ github.repository }}" \
            --tags "cicd,automated"
        
      - name: Generate Payloads
        run: |
          TOKEN=$(godnslog-cli payload create \
            --template "{{.Token}}.example.com" \
            --format json | jq -r '.token')
          echo "GODNSLOG_TOKEN=$TOKEN" >> $GITHUB_ENV
        
      - name: Wait for Interactions
        run: |
          godnslog-cli interaction wait \
            --token "${{ env.GODNSLOG_TOKEN }}" \
            --timeout 300
```

## 错误处理

CLI 使用以下退出码：

- `0`：成功
- `1`：错误

错误信息会输出到 stderr，正常输出到 stdout。

## 完整示例

```bash
# 1. 配置环境
export GODNSLOG_API_URL="http://localhost:8080"
export GODNSLOG_API_KEY="your-api-key"

# 2. 创建 Case
CASE_ID=$(godnslog-cli case create \
  --title "完整测试" \
  --description "演示完整CLI工作流" \
  --target "example.com" \
  --format json | jq -r '.id')

# 3. 生成 Payload
TOKEN=$(godnslog-cli payload create \
  --template "http://{{.Token}}.example.com" \
  --case-id "$CASE_ID" \
  --expires-in "1h" \
  --format json | jq -r '.token')

# 4. 使用 Payload 进行测试
echo "Payload: http://$TOKEN.example.com"

# 5. 等待 Interaction
godnslog-cli interaction wait \
  --token "$TOKEN" \
  --timeout 60

# 6. 导出报告
godnslog-cli report export \
  --case-id "$CASE_ID" \
  --format markdown \
  --output "report-$CASE_ID.md"
```

## 帮助

```bash
godnslog-cli --help
godnslog-cli case --help
godnslog-cli payload --help
godnslog-cli interaction --help
godnslog-cli report --help
```
