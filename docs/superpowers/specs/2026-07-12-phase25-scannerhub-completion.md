# Phase 2.5 Remaining: Scanner Hub Integration Completion

## 问题

Scanner Hub 的 5 个扫描器集成的 3 个缺陷。

## 修改方案

### 1. 修复命令生成（`generateScannerCommand`）

对 burp/zap/xray 生成真实可执行的命令，替换当前的纯文本说明。

**Burp**：Burp Suite 提供了 REST API 用于扩展集成，可生成以下命令形式：
```
# 创建测试 Payload
curl -s -X POST "$GODNSLOG_URL/api/v2/payloads" \
  -H "Authorization: Bearer $GODNSLOG_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"case_id":"...","template":"...","scenario":"Burp scanner probe"}'

# 轮询交互结果
curl -s "$GODNSLOG_URL/api/v2/interactions?payload_id=..." \
  -H "Authorization: Bearer $GODNSLOG_API_KEY" | jq .
```

**ZAP**：使用 `zap-cli` 或 `zap.sh -cmd` + `-script` 参数：
```
zap-cli quick-scan -u $TARGET -s godnslog-oast.js \
  -z "godnslog_payload=$PAYLOAD godnslog_url=$GODNSLOG_URL"
```

**xray**：生成 xray webhook 配置 JSON 文件 + 运行命令：
```
# xray webhook 配置 (godnslog-webhook.yml)
echo '{"url":"$GODNSLOG_URL/api/v2/webhook/xray","payload":"$PAYLOAD"}' > xray-webhook-config.json
./xray webhook --config xray-webhook-config.json --url $TARGET
```

### 2. 扩展回填解析器

**Burp Suite XML/JSON**：Burp Suite 可导出 Issue 为 XML 或 JSON 格式。
- JSON 格式：解析 `issues[]` 数组，提取 `name`、`severity`、`path`、`confidence`
- 格式标识：`"burp-json"`

**xray JSON**：xray `--json-output` 输出的 JSON 数组格式。
- 每条记录：`vuln_id`、`plugin`、`severity`、`url`、`payload`
- 格式标识：`"xray-json"`

### 3. 搜索引擎串联 Scanner Run

**后端**：新增 `POST /api/v2/scanner-runs/from-search` 端点。
- 输入：`{ search_source, search_result: { ip, port, hostname }, scanner, delivery_method, case_id }`
- 自动提取目标 URL，创建对应 Scanner Run

**前端**：在搜索引擎结果区域的每个 IP/端口项目上，增加下拉菜单 → "创建 Scanner Run"

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/scannerhub/service.go` | 修复 `generateScannerCommand`：burp/zap/xray 真命令 |
| `internal/scannerhub/service_test.go` | 新增命令生成测试 |
| `internal/scannerhub/backfill.go` | 新增 `parseBurpJSON`、`parseXrayJSON` |
| `internal/scannerhub/backfill_test.go` | 新增回填解析测试 |
| `server/v2_api.go` | 新增 `v2CreateScannerRunFromSearch` handler + 扩展 backfill 支持的格式 |
| `frontend-next/src/app/dashboard/scanner-hub/page.tsx` | 搜索结果显示"创建 Scanner Run"操作 |
| `frontend-next/src/lib/api-client.ts` | 新增 `fromSearch` API 调用 |
| `frontend-next/src/types/index.ts` | 新增相关类型 |
