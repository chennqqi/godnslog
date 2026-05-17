# GODNSLOG 2.0 Implementation Dependency Graph and Acceptance Criteria

**Purpose**: Define the implementation dependencies and acceptance criteria for the next development phase. This document transforms the design documents into an actionable implementation plan with clear dependencies and validation checkpoints.

**Principle**: Dependency-first implementation. Work that must be done first is identified and sequenced to avoid blocking and rework.

## Implementation Phases

### Phase 1: Unified Domain Model and API Contract
**Goal**: Establish the data model and API contract that all other work depends on.

**Dependencies**: None (foundational work)

**Duration**: 2-3 weeks

**Deliverables**:
- Unified data models (Case, Payload, Interaction, Evidence, Agent, Agent Run, Workspace, Audit Event)
- API contract specifications for `/api/v2` endpoints
- Database schema migrations
- API authentication and authorization middleware
- API response format standardization

**Acceptance Criteria**:
- [ ] All entity models are defined in `internal/models/` with consistent field names
- [ ] All API endpoints follow REST conventions and return consistent response format
- [ ] Database schema is migrated and backward-compatible
- [ ] API authentication uses JWT tokens with `IsAgent` flag support
- [ ] API authorization respects role-based access control
- [ ] All API endpoints are documented in OpenAPI/Swagger spec
- [ ] API tests pass for all endpoints

**Validation Questions**:
- Does this capability enter the unified entity model? ✅ Yes (all entities)
- Does it enter the unified audit chain? ✅ Yes (audit event model)
- Does it have clear role and permission boundaries? ✅ Yes (RBAC middleware)
- Does it enter the first MVP closed loop? ✅ Yes (core entities)

---

### Phase 2: Probe → Interaction → Evidence → Export → Audit Closed Loop
**Goal**: Implement the core OAST evidence chain end-to-end.

**Dependencies**: Phase 1 (must have data models and API contract)

**Duration**: 4-5 weeks

**Deliverables**:
- Case CRUD operations (API + backend logic)
- Payload generation with template rendering (API + backend logic)
- Interaction capture for DNS and HTTP (backend listener)
- Evidence aggregation and scoring (backend logic)
- Evidence export in JSON and Markdown (API + backend logic)
- Audit logging for all operations (backend middleware)
- CLI commands for all operations (CLI tool)

**Acceptance Criteria**:
- [ ] Can create a Case via API and CLI
- [ ] Can create a Payload with template rendering via API and CLI
- [ ] DNS queries are captured and attributed to Payload and Case
- [ ] HTTP requests are captured and attributed to Payload and Case
- [ ] Evidence is generated with strength and confidence scores
- [ ] Evidence can be exported in JSON and Markdown formats
- [ ] All operations are logged in audit trail with full parameters
- [ ] CLI commands work for all operations
- [ ] Integration tests pass for the full closed loop

**Validation Questions**:
- Does this capability enter the unified entity model? ✅ Yes
- Does it enter the unified audit chain? ✅ Yes
- Does it have clear role and permission boundaries? ✅ Yes
- Does it enter the first MVP closed loop? ✅ Yes (this IS the closed loop)

---

### Phase 3: First Control Plane Pages (Web UI)
**Goal**: Implement the minimum viable Web UI for the MVP closed loop.

**Dependencies**: Phase 2 (must have working API endpoints)

**Duration**: 3-4 weeks

**Deliverables**:
- Case List page
- Case Detail page
- Payload Studio page
- Interaction Timeline page
- Evidence Timeline page
- Audit Log page (admin only)
- Authentication and authorization in Web UI
- Responsive design for desktop and tablet

**Acceptance Criteria**:
- [ ] User can log in via Web UI
- [ ] User can create and view Cases
- [ ] User can create Payloads with template selection
- [ ] User can view Interaction timeline filtered by Case
- [ ] User can view Evidence with strength and confidence
- [ ] User can export Evidence in JSON and Markdown
- [ ] Admin can view Audit Log with filters
- [ ] Web UI is responsive and accessible
- [ ] E2E tests pass for all pages

**Validation Questions**:
- Does this capability enter the unified entity model? ✅ Yes (uses existing entities)
- Does it enter the unified audit chain? ✅ Yes (all UI actions logged)
- Does it have clear role and permission boundaries? ✅ Yes (admin-only audit page)
- Does it enter the first MVP closed loop? ✅ Yes (UI for closed loop)

**Parallelization**: Web UI development can start in parallel with Phase 2 backend work, using API mocks initially. However, final integration requires Phase 2 completion.

---

### Phase 4: Primary Scanner Integration (Nuclei, Burp Suite, Yakit/Yak)
**Goal**: Implement official integrations for Primary tier scanners.

**Dependencies**: Phase 2 (must have stable Probe creation and Interaction capture)

**Duration**: 3-4 weeks

**Deliverables**:
- Nuclei official script with JSONL export
- Burp Suite extension with probe generation
- Yakit/Yak official script with standard contract
- Documentation for all three integrations
- Integration tests for each scanner

**Acceptance Criteria**:
- [ ] Nuclei can create GODNSLOG probes via script
- [ ] Nuclei can export results in GODNSLOG JSONL format
- [ ] Burp Suite extension can create payloads within Burp UI
- [ ] Burp Suite extension shows interaction feedback
- [ ] Yakit/Yak script can create probes and export results
- [ ] All three integrations have step-by-step documentation
- [ ] Integration tests validate each scanner workflow

**Validation Questions**:
- Does this capability enter the unified entity model? ✅ Yes (uses Case/Payload/Interaction)
- Does it enter the unified audit chain? ✅ Yes (scanner operations logged)
- Does it have clear role and permission boundaries? ✅ Yes (scanner uses API keys)
- Does it enter the first MVP closed loop? ✅ Yes (scanner creates probes, captures interactions)

**Parallelization**: Scanner integrations can be developed in parallel by different developers, as they depend on the same stable API but don't depend on each other.

---

### Phase 5: Agent Governance and Replay
**Goal**: Implement Agent-Native capabilities for AI agent integration.

**Dependencies**: Phase 1 (must have API contract with Agent support), Phase 2 (must have stable audit chain)

**Duration**: 3-4 weeks

**Deliverables**:
- Agent entity and API key management
- Agent Run lifecycle tracking
- MCP Server with agent-specific tools
- Agent-specific API key scopes
- Agent Run dashboard (Web UI)
- Audit logging for all agent operations
- Agent revocation and risk controls

**Acceptance Criteria**:
- [ ] Operator can create Agent with scoped permissions
- [ ] Agent has unique API key with `IsAgent` flag
- [ ] Agent Run lifecycle is tracked from creation to completion
- [ ] MCP Server exposes `create_oast_probe`, `wait_for_interaction`, `export_evidence`
- [ ] All agent operations are logged in audit trail
- [ ] Agent Run dashboard shows operation history
- [ ] Operator can revoke agent API key
- [ ] Agent operations respect risk-based controls

**Validation Questions**:
- Does this capability enter the unified entity model? ✅ Yes (Agent, Agent Run entities)
- Does it enter the unified audit chain? ✅ Yes (agent operations fully audited)
- Does it have clear role and permission boundaries? ✅ Yes (scoped API keys)
- Does it enter the first MVP closed loop? ❌ No (separate track, can be parallel to Phase 4)

**Parallelization**: Agent governance can be developed in parallel with Phase 4 (Scanner Integration), as it depends on Phase 1 and Phase 2 but not on scanner work.

---

## Dependency Graph

```
Phase 1: Unified Domain Model & API Contract (Foundational)
  ↓
Phase 2: Closed Loop Backend (Probe → Interaction → Evidence → Export → Audit)
  ↓
  ├─→ Phase 3: Web UI (depends on Phase 2 API)
  ├─→ Phase 4: Scanner Integration (depends on Phase 2 API)
  └─→ Phase 5: Agent Governance (depends on Phase 1 API + Phase 2 Audit)
```

**Key Dependencies**:
- **Phase 1 must be first**: All other phases depend on the unified data model and API contract
- **Phase 2 must be second**: Web UI, Scanner Integration, and Agent Governance all depend on stable backend APIs
- **Phase 3, 4, 5 can be parallel**: Once Phase 1 and Phase 2 are complete, the remaining phases can be developed in parallel

## Acceptance Criteria for All Implementation Work

### Universal Questions
Before any capability is considered complete, it must answer "Yes" to these questions:

1. **Does this capability enter the unified entity model?**
   - Yes: Uses entities defined in `docs/unified-terminology.md`
   - No: Must be redesigned to use unified entities

2. **Does this capability enter the unified audit chain?**
   - Yes: All operations are logged in Audit Event with full parameters
   - No: Must add audit logging before completion

3. **Does it have clear role and permission boundaries?**
   - Yes: RBAC is defined and enforced
   - No: Must define roles and permissions before completion

4. **Does it enter the first MVP closed loop?**
   - Yes: Part of Probe → Interaction → Evidence → Export → Audit
   - No: Must be moved to post-MVP backlog or separate track

### Phase-Specific Acceptance Criteria

#### Phase 1 Acceptance
- [ ] All entity models match `docs/unified-terminology.md` exactly
- [ ] API endpoints follow REST conventions
- [ ] API responses have consistent format (code, message, data)
- [ ] Database migrations are reversible
- [ ] API tests have >80% coverage
- [ ] OpenAPI spec is auto-generated and up-to-date

#### Phase 2 Acceptance
- [ ] End-to-end integration test passes (create probe → capture interaction → generate evidence → export)
- [ ] DNS and HTTP listeners are stable under load
- [ ] Evidence scoring logic is deterministic and documented
- [ ] Audit log captures all operations with full parameters
- [ ] CLI commands work identically to API endpoints
- [ ] Error handling covers all failure modes

#### Phase 3 Acceptance
- [ ] All MVP pages are implemented per `docs/mvp-closed-loop.md`
- [ ] Web UI is responsive on desktop (1920x1080) and tablet (768x1024)
- [ ] Authentication flow works with JWT tokens
- [ ] Authorization respects RBAC (admin-only pages enforced)
- [ ] E2E tests cover all user workflows
- [ ] Page load time < 2 seconds on 3G connection

#### Phase 4 Acceptance
- [ ] All Primary scanner integrations are implemented per `docs/official-support-boundary.md`
- [ ] Scanner integrations are tested with real scanner instances
- [ ] Documentation includes step-by-step workflows
- [ ] Scanner operations are logged in audit trail
- [ ] Scanner integrations handle errors gracefully
- [ ] Scanner integrations are versioned and backward-compatible

#### Phase 5 Acceptance
- [ ] Agent entity matches `docs/agent-native-specification.md` exactly
- [ ] Agent API keys have `IsAgent` flag and scopes
- [ ] Agent Run lifecycle is tracked end-to-end
- [ ] MCP Server exposes all required tools
- [ ] All agent operations are logged with agent_id and agent_run_id
- [ ] Agent revocation works immediately
- [ ] Risk-based controls are enforced

## Risk Mitigation

### Dependency Risk
**Risk**: Phase 1 delays block all subsequent phases
**Mitigation**: Phase 1 is the highest priority. Assign most experienced developers. Daily standups. No scope changes in Phase 1.

### Integration Risk
**Risk**: Phase 3, 4, 5 integration issues due to API changes
**Mitigation**: Phase 2 API contract is frozen once Phase 2 starts. Any API changes require impact assessment on all dependent phases.

### Parallelization Risk
**Risk**: Parallel development of Phase 3, 4, 5 leads to merge conflicts
**Mitigation**: Clear ownership of code modules. Regular sync meetings. Continuous integration to catch conflicts early.

### Scope Creep Risk
**Risk**: Adding features beyond MVP closed loop during implementation
**Mitigation**: Strict adherence to `docs/mvp-closed-loop.md`. Any new feature must be evaluated against the 4 universal acceptance criteria.

## Implementation Sequence Recommendation

Based on dependencies and risk mitigation, the recommended implementation sequence is:

1. **Week 1-3**: Phase 1 (Unified Domain Model & API Contract) - All hands on deck
2. **Week 4-8**: Phase 2 (Closed Loop Backend) - Core team focus
3. **Week 6-9**: Phase 3 (Web UI) - Frontend team (parallel with Phase 2)
4. **Week 9-12**: Phase 4 (Scanner Integration) - Integration team (parallel with Phase 5)
5. **Week 9-12**: Phase 5 (Agent Governance) - Agent team (parallel with Phase 4)
6. **Week 13**: Integration testing and bug fixes
7. **Week 14**: Documentation finalization and release preparation

**Total Duration**: 14 weeks (3.5 months)

## Post-MVP Roadmap

After MVP completion, the next phases should be:

1. **Multi-Protocol Listeners** (SMTP, LDAP, SMB, FTP) - 2-3 weeks
2. **Secondary Scanner Integration** (ZAP, xray/rad, Postman/Apifox) - 3-4 weeks
3. **Workspace Isolation** - 2-3 weeks
4. **Workflow Automation** - 3-4 weeks
5. **Canary Tokens** - 2-3 weeks
6. **DNS Rebinding Lab** - 2-3 weeks
7. **Advanced Evidence Features** (clustering, SARIF, PDF) - 3-4 weeks

Each post-MVP phase must also answer the 4 universal acceptance criteria before completion.

## Version History

- v1.0 (2026-05-17): Initial implementation dependency graph for GODNSLOG 2.0 MVP
