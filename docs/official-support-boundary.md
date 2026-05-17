# GODNSLOG 2.0 Official Support Boundary and Maturity List

**Purpose**: Define the official support boundary for scanner integrations to prevent scope creep and ensure focused delivery. This document establishes which tools receive what level of support in the first release cycle.

**Principle**: Explicit boundaries over implicit expansion. Only tools in the Primary tier receive full support in MVP. Secondary and Backlog tools are documented but not implemented in the first cycle.

## Support Tiers

### Primary (首批官方主支持)
**Definition**: Tools that receive full integration in the first MVP release. These tools have official scripts, documentation, and are part of the core OAST evidence chain.

**Commitment Level**: L3 (Official Script) for MVP, with roadmap to L4 (Native Plugin) in Phase 2.

**Support Scope**:
- Official integration script/example
- Standard input/output contract
- Documentation with step-by-step guide
- Active maintenance and bug fixes
- Part of MVP testing and validation
- Priority for feature requests

### Secondary (官方提供脚本/桥接，但不承诺深度集成)
**Definition**: Tools that receive documented integration patterns but are not part of the MVP release. These tools have community examples or basic scripts but no deep integration.

**Commitment Level**: L1 (Documentation) or L2 (Webhook Bridge) initially.

**Support Scope**:
- Documented integration pattern
- Example script or configuration (if applicable)
- Community support (best-effort)
- Bug fixes accepted as contributions
- Not part of MVP testing
- Lower priority for feature requests

### Backlog (已识别，但不进入当前周期)
**Definition**: Tools that have been identified as potentially useful but are not scheduled for any near-term development. These tools are documented for future reference only.

**Commitment Level**: L0 (Identified only).

**Support Scope**:
- Listed in backlog for reference
- No documentation
- No scripts
- No active development
- May be promoted to Secondary/Primary based on demand

## Official Support Matrix

| Tool | Tier | Current Maturity | Target Maturity | Integration Type | Status | Notes |
|------|------|------------------|-----------------|------------------|--------|-------|
| **Nuclei** | Primary | L2 (JSONL export) | L3 (Official Script) | Script | MVP | Core automation scanner |
| **Burp Suite** | Primary | L2 (Extension) | L4 (Native Plugin) | Native | MVP | Manual testing standard |
| **Yakit/Yak** | Primary | L2 (Script) | L3 (Official Script) | Script | MVP | Chinese ecosystem hybrid tool |
| **ZAP** | Secondary | L1 (Documentation) | L2 (Webhook Bridge) | Webhook | Phase 2 | OWASP scanner |
| **xray/rad** | Secondary | L1 (Documentation) | L2 (Webhook Bridge) | Webhook | Phase 2 | Chinese DAST scanner |
| **Postman/Apifox** | Secondary | L1 (Documentation) | L2 (Webhook Bridge) | Webhook | Phase 2 | API testing tools |
| **SQLMap** | Backlog | L0 | - | - | Future | SQL injection scanner |
| **Metasploit** | Backlog | L0 | - | - | Future | Exploitation framework |
| **Nmap** | Backlog | L0 | - | - | Future | Network scanner |
| **Masscan** | Backlog | L0 | - | - | Future | Fast port scanner |

## Primary Tool Details

### Nuclei
**Why Primary**: Nuclei is the de facto standard for automated vulnerability scanning in the security community. It has a large template ecosystem and is widely used in CI/CD pipelines.

**Current State**: L2 - JSONL export support exists
**Target State**: L3 - Official script with standard contract
**Integration Type**: Script (YAML template integration)

**MVP Deliverables**:
- Official Nuclei template for GODNSLOG OAST
- Script to convert Nuclei output to GODNSLOG probe format
- Documentation: "Using Nuclei with GODNSLOG 2.0"
- Example: Nuclei YAML template with GODNSLOG variables

**Phase 2 Roadmap**:
- Bidirectional integration (Nuclei reads GODNSLOG interactions)
- Native Nuclei plugin (if API supports)

### Burp Suite
**Why Primary**: Burp Suite is the standard for manual web security testing. It represents the "human-in-the-loop" verification scenario that automated scanners miss. This is critical for the security engineer persona.

**Current State**: L2 - Extension exists
**Target State**: L4 - Native plugin with UI integration
**Integration Type**: Native (Burp Extension)

**MVP Deliverables**:
- Burp Suite extension for GODNSLOG probe generation
- Extension UI for creating payloads within Burp
- Real-time interaction feedback in Burp UI
- Documentation: "Burp Suite Extension for GODNSLOG 2.0"
- Example: Workflow for manual SSRF testing with Burp + GODNSLOG

**Phase 2 Roadmap**:
- Bidirectional communication (Burp reads GODNSLOG interactions)
- Native authentication flow
- Scanner-specific UI integration

### Yakit/Yak
**Why Primary**: Yakit/Yak represents the hybrid manual/automated testing approach popular in the Chinese security community. It bridges the gap between fully automated scanners and manual tools.

**Current State**: L2 - Script exists
**Target State**: L3 - Official script with standard contract
**Integration Type**: Script (Yak script)

**MVP Deliverables**:
- Official Yak script for GODNSLOG integration
- Standard parameter mapping for probe creation
- Result export in JSON format
- Documentation: "Using Yakit/Yak with GODNSLOG 2.0"
- Example: Yak script for automated SSRF testing

**Phase 2 Roadmap**:
- Native Yakit plugin (if API supports)
- Real-time interaction feedback

## Secondary Tool Details

### ZAP
**Why Secondary**: ZAP is a popular OWASP scanner, but it has lower priority than Nuclei for the initial release. ZAP integration is valuable for OWASP-focused security teams.

**Current State**: L1 - Documentation only
**Target State**: L2 - Webhook bridge
**Integration Type**: Webhook

**Phase 2 Deliverables**:
- Webhook contract for ZAP alerts
- Example ZAP script to send alerts to GODNSLOG
- Documentation: "ZAP Integration with GODNSLOG 2.0"

### xray/rad
**Why Secondary**: xray/rad are popular Chinese DAST scanners. They are important for the Chinese market but have lower priority than Yakit/Yak for the initial release.

**Current State**: L1 - Documentation only
**Target State**: L2 - Webhook bridge
**Integration Type**: Webhook

**Phase 2 Deliverables**:
- Webhook contract for xray/rad alerts
- Example configuration for xray/rad to send alerts to GODNSLOG
- Documentation: "xray/rad Integration with GODNSLOG 2.0"

### Postman/Apifox
**Why Secondary**: Postman/Apifox are API testing tools. They are useful for API security testing but are not core vulnerability scanners. They have lower priority for the initial release.

**Current State**: L1 - Documentation only
**Target State**: L2 - Webhook bridge
**Integration Type**: Webhook

**Phase 2 Deliverables**:
- Webhook contract for API test results
- Example Postman collection with GODNSLOG variables
- Documentation: "Postman/Apifox Integration with GODNSLOG 2.0"

## Backlog Tool Details

### SQLMap
**Why Backlog**: SQLMap is a specialized SQL injection scanner. While useful, it has a narrower use case than general-purpose scanners like Nuclei.

**Future Consideration**: May be promoted to Secondary if there is strong demand for SQL injection-specific OAST verification.

### Metasploit
**Why Backlog**: Metasploit is an exploitation framework. While it could use OAST for exploitation verification, this is a more advanced use case beyond the initial MVP scope.

**Future Consideration**: May be promoted to Secondary if exploitation verification becomes a priority.

### Nmap
**Why Backlog**: Nmap is a network scanner. While it could use OAST for port verification, this is outside the core web application security focus of the MVP.

**Future Consideration**: May be promoted to Secondary if network security becomes a priority.

### Masscan
**Why Backlog**: Masscan is a fast port scanner. Similar to Nmap, it is outside the core web application security focus.

**Future Consideration**: May be promoted to Secondary if network security becomes a priority.

## Support Tier Promotion Criteria

### Promotion from Backlog to Secondary
- Community demand (GitHub issues, forum requests)
- Clear use case identified
- Feasibility assessment completed
- Resource availability confirmed

### Promotion from Secondary to Primary
- High adoption rate (usage statistics)
- Critical for security engineer workflow
- Stable integration pattern established
- Resource availability for full support

### Demotion from Primary to Secondary
- Low adoption rate over 6 months
- Maintenance burden exceeds value
- Better alternative emerges
- Strategic shift in focus

## Integration Maturity Levels

### L0: Identified (Backlog)
- Tool identified as potentially useful
- No documentation
- No scripts
- No active development

### L1: Documentation (Secondary)
- Documented integration pattern
- Example workflow documented
- No official script
- Community support only

### L2: Webhook Bridge (Secondary)
- Standard webhook contract defined
- Example webhook handler provided
- JSON payload specification
- Basic integration possible

### L3: Official Script (Primary/Secondary)
- Official script for scanner automation API
- Standard parameter mapping
- Result export in standard format
- Error handling and retry logic
- Active maintenance

### L4: Native Plugin (Primary)
- Native plugin/extension for scanner
- Bidirectional communication
- Scanner-specific UI integration
- Real-time interaction feedback
- Native authentication flow
- Highest integration maturity

## Support Commitment

### Primary Tier Commitments
- Bug fixes within 7 days
- Feature requests considered and prioritized
- Documentation updates with each release
- Breaking changes communicated 30 days in advance
- Security vulnerabilities addressed within 24 hours

### Secondary Tier Commitments
- Bug fixes accepted as contributions
- Feature requests considered but lower priority
- Documentation updates as time permits
- Breaking changes communicated 60 days in advance
- Security vulnerabilities addressed within 7 days

### Backlog Tier Commitments
- No active support
- Contributions accepted but not prioritized
- No documentation maintenance
- No breaking change communication
- Security vulnerabilities addressed when resources allow

## Adding New Tools to Support Matrix

### Process for Adding New Tool
1. **Proposal**: Submit GitHub issue with tool name, use case, and integration feasibility
2. **Assessment**: Core team evaluates tool against criteria (market relevance, technical feasibility, resource availability)
3. **Decision**: Tool assigned to Backlog, Secondary, or Primary tier
4. **Planning**: Integration added to appropriate roadmap phase
5. **Implementation**: Development according to tier commitment level

### Evaluation Criteria
- Market relevance (adoption rate, community size)
- Technical feasibility (API availability, integration complexity)
- Resource availability (developer time, maintenance burden)
- Strategic alignment (fits product positioning, target personas)
- Differentiation value (unique capabilities vs. existing tools)

## Removal from Support Matrix

### Process for Removing Tool
1. **Proposal**: Submit GitHub issue proposing removal with justification
2. **Assessment**: Core team evaluates impact on users and product
3. **Decision**: Tool removed or retained with deprecation notice
4. **Communication**: Announcement of removal with timeline
5. **Decommission**: Documentation archived, support withdrawn

### Removal Criteria
- Tool discontinued by vendor
- Zero adoption over 12 months
- Maintenance burden exceeds value
- Strategic shift in product direction
- Security vulnerability with no feasible fix

## Version History

- v1.0 (2026-05-17): Initial official support boundary for GODNSLOG 2.0 MVP
