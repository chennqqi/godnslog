# Scanner Hub Adapter System Design

## Design Goal

Define a sustainable, extensible system for integrating security scanners with GODNSLOG 2.0. The goal is not to "list tools" but to define "how to continuously add tools" with clear maturity levels and standard interfaces.

## Adapter Layer Architecture

### Layer 1: Native Adapter (原生适配)
**Definition**: Deep integration with scanner-specific APIs or plugin systems.

**Characteristics**:
- Bidirectional communication
- Scanner-specific UI integration
- Real-time interaction feedback
- Native authentication flow

**Examples**:
- Burp Suite Extension (plugin)
- Yakit/Yak Script (native script API)

**Maturity**: L4 - Highest integration level

### Layer 2: Script Adapter (脚本适配)
**Definition**: Integration through scanner scripting languages or automation APIs.

**Characteristics**:
- Unidirectional (scanner → GODNSLOG)
- Script-based probe creation and result export
- Standard parameter mapping
- Requires scanner scripting knowledge

**Examples**:
- Burp Suite Python/JavaScript Extension
- Yakit/Yak Script
- ZAP Script (ZAPy/ZAP extension)

**Maturity**: L3 - Official script support

### Layer 3: Webhook Bridge (Webhook 桥接)
**Definition**: Integration through scanner webhook capabilities or custom webhooks.

**Characteristics**:
- Event-driven communication
- Scanner-agnostic contract
- JSON payload standardization
- Requires scanner webhook support or custom bridge

**Examples**:
- CI/CD webhook integration
- Custom scanner webhook handlers
- Generic webhook receiver

**Maturity**: L2 - Standardized webhook interface

### Layer 4: Documentation Example (文档样例)
**Definition**: Manual integration patterns documented for reference.

**Characteristics**:
- Manual probe creation
- Manual result export
- No automation
- Educational purpose

**Examples**:
- Postman Collection
- cURL examples
- Manual workflow documentation

**Maturity**: L1 - Documentation-level support

## Standard Input Contract

All adapters must support the following standard operations:

### 1. Create Probe
**Input Parameters**:
```json
{
  "title": "Engagement title",
  "description": "Engagement description",
  "template": "ssrf-basic",
  "target": "target.com",
  "variables": {
    "custom_var": "value"
  },
  "expires_in": "24h",
  "expected_protocols": ["dns", "http"]
}
```

**Output**:
```json
{
  "probe_id": "case_id:payload_id",
  "case_id": "uuid",
  "payload_id": "uuid",
  "token": "random-token",
  "urls": {
    "dns": "token.example.com",
    "http": "http://example.com/callback/token"
  },
  "template_rendered": "rendered-payload"
}
```

### 2. Wait for Interaction
**Input Parameters**:
```json
{
  "token": "random-token",
  "timeout": 300,
  "expected_count": 1
}
```

**Output**:
```json
{
  "interactions": [
    {
      "type": "dns",
      "source_ip": "10.0.0.1",
      "timestamp": "2026-05-17T12:00:00Z",
      "data": {...}
    }
  ],
  "total_count": 1,
  "evidence_strength": "high",
  "confidence": 85
}
```

### 3. Export Results
**Input Parameters**:
```json
{
  "probe_id": "case_id:payload_id",
  "format": "jsonl",
  "include_raw": true
}
```

**Output**:
- JSON: Structured result object
- JSONL: Line-delimited JSON for streaming
- SARIF: Static Analysis Results Interchange Format
- Webhook Event: Real-time notification payload

## Standard Output Contract

### JSON Format
```json
{
  "probe_id": "case_id:payload_id",
  "case_id": "uuid",
  "payload_id": "uuid",
  "interactions": [...],
  "evidence_summary": {
    "total": 10,
    "dns_count": 5,
    "http_count": 3,
    "smtp_count": 2,
    "evidence_strength": "high",
    "confidence": 85
  },
  "timeline": [...],
  "exported_at": "2026-05-17T12:00:00Z"
}
```

### JSONL Format
```jsonl
{"type":"dns","source_ip":"10.0.0.1","timestamp":"2026-05-17T12:00:00Z","token":"tok123"}
{"type":"http","source_ip":"10.0.0.1","timestamp":"2026-05-17T12:00:05Z","token":"tok123","path":"/callback"}
```

### SARIF Format
```json
{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "runs": [{
    "tool": {
      "name": "godnslog",
      "version": "2.0.0"
    },
    "results": [...]
  }]
}
```

### Webhook Event Format
```json
{
  "event_type": "interaction.created",
  "timestamp": "2026-05-17T12:00:00Z",
  "data": {
    "interaction": {...},
    "probe_id": "case_id:payload_id"
  }
}
```

## Maturity Levels

### L1: Documentation Example
**Criteria**:
- ✅ Official documentation exists
- ✅ Manual workflow documented
- ✅ Example payloads provided
- ❌ No automation
- ❌ No standard interface

**Status**: Educational reference only

### L2: Webhook Bridge
**Criteria**:
- ✅ Standard webhook contract defined
- ✅ Webhook receiver implemented
- ✅ JSON payload specification
- ✅ Example webhook handlers provided
- ❌ No scanner-specific optimization

**Status**: Basic integration via webhooks

### L3: Official Script
**Criteria**:
- ✅ Official script for scanner automation API
- ✅ Standard parameter mapping
- ✅ Result export capability
- ✅ Error handling and retry logic
- ❌ No bidirectional communication

**Status**: Script-based automation

### L4: Official Plugin
**Criteria**:
- ✅ Native plugin/extension for scanner
- ✅ Bidirectional communication
- ✅ Scanner-specific UI integration
- ✅ Real-time interaction feedback
- ✅ Native authentication flow

**Status**: Deep integration, highest maturity

## Primary Support Matrix

| Tool | Current Status | Target Maturity | Integration Type | Value Proposition |
|------|---------------|-----------------|------------------|-------------------|
| **Nuclei** | L2 (JSONL export) | L3 (Official Script) | Script | Automated vulnerability scanning with OAST |
| **Burp Suite** | L2 (Extension) | L4 (Native Plugin) | Native | Manual testing with OAST integration |
| **Yakit/Yak** | L2 (Script) | L3 (Official Script) | Script | Hybrid manual/automated testing |
| **ZAP** | L1 (Documentation) | L3 (Official Script) | Script | OWASP-focused security testing |
| **xray/rad** | L1 (Documentation) | L2 (Webhook) | Webhook | Chinese security scanner ecosystem |
| **Postman/Apifox** | L1 (Documentation) | L2 (Webhook) | Webhook | API testing with OAST verification |

## Special Focus: Burp Suite and Yakit/Yak

### Burp Suite
**Why Special**: Burp Suite is the de facto standard for manual web security testing. It represents the "human-in-the-loop" verification scenario that pure automated scanners miss.

**Integration Value**:
- Manual penetration testing with OAST verification
- Repeater/Intruder integration for systematic testing
- Proxy-based workflow for real-world scenarios
- Extension ecosystem for deep integration

**Target Maturity**: L4 (Native Plugin)
- Burp Suite Extension with native UI
- Bidirectional communication with GODNSLOG
- Real-time interaction feedback in Burp UI
- Native authentication flow

### Yakit/Yak
**Why Special**: Yakit/Yak represents the hybrid manual/automated testing approach popular in Chinese security community. It bridges the gap between fully automated scanners and manual tools.

**Integration Value**:
- Script-based automation with manual control
- Chinese security ecosystem integration
- Flexible workflow for different testing scenarios
- Community-driven script library

**Target Maturity**: L3 (Official Script)
- Official Yak script for GODNSLOG integration
- Standard parameter mapping for probe creation
- Result export in JSON/JSONL format
- Error handling and retry logic

## Extensibility Design

### Adding New Tools

**Step 1: Assess Tool Capabilities**
- Does the tool have automation APIs?
- Does the tool support webhooks?
- Does the tool have scripting languages?
- What is the tool's primary use case?

**Step 2: Choose Adapter Layer**
- If tool has plugin system → L4 (Native Adapter)
- If tool has scripting API → L3 (Script Adapter)
- If tool has webhook support → L2 (Webhook Bridge)
- If tool has no automation → L1 (Documentation)

**Step 3: Implement Standard Contract**
- Implement Create Probe operation
- Implement Wait for Interaction operation
- Implement Export Results operation
- Support standard output formats (JSON, JSONL, SARIF)

**Step 4: Document and Test**
- Create integration documentation
- Provide example workflows
- Test with real scenarios
- Validate maturity level criteria

### Adapter Registry

All adapters are registered in `docs/scanner-hub-registry.md` with:
- Tool name and version
- Current maturity level
- Integration type
- Documentation link
- Example scripts/configurations
- Known limitations

## Implementation Roadmap

### Phase 1: L2 Webhook Bridge
- Standardize webhook contract
- Implement webhook receiver
- Create Postman/Apifox webhook integration
- Document webhook usage patterns

### Phase 2: L3 Official Scripts
- Nuclei official script with JSONL export
- ZAP official script for OAST verification
- Yakit/Yak official script
- xray/rad webhook integration

### Phase 3: L4 Native Plugins
- Burp Suite native extension
- Yakit/Yak native plugin (if API supports)
- Scanner-specific UI integrations

### Phase 4: Ecosystem Expansion
- Community-driven adapter contributions
- Adapter marketplace (if applicable)
- Continuous integration with new scanner releases

## Success Metrics

- **Coverage**: Number of tools at L3+ maturity
- **Adoption**: Usage statistics per adapter
- **Quality**: Bug reports and issues per adapter
- **Extensibility**: Time to add new tool (target: < 1 week)
- **Community**: Community contributions to adapter ecosystem

## Conclusion

The Scanner Hub adapter system provides a clear, sustainable path for integrating security scanners with GODNSLOG 2.0. By defining maturity levels, standard contracts, and implementation guidelines, we ensure that tool integration is systematic, maintainable, and extensible.
