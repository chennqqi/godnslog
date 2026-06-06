// API Response types
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  timestamp?: number
  data?: T
}

// User types
export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  lang?: string
  role: Role
  utime?: string
}

export interface Role {
  id: string
  name: string
  description: string
  permissions: Permission[]
}

export interface Permission {
  roleId: number
  permissionId: string
  permissionName: string
  ActionEntitySet: PermissionActionSet[]
}

export interface PermissionActionSet {
  action: string
  description: string
  defaultCheck: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

// Case types
export interface Case {
  id: string
  title: string
  description: string
  target: string
  status: 'active' | 'archived' | 'completed'
  tags: string[]
  created_by: string
  created_at: string
  updated_at: string
}

export interface CaseCreateRequest {
  title: string
  description?: string
  target?: string
  tags?: string[]
}

export interface CaseUpdateRequest {
  title?: string
  description?: string
  target?: string
  status?: 'active' | 'archived' | 'completed'
  tags?: string[]
}

export interface CaseListResponse {
  items: Case[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Payload types
export interface Payload {
  id: string
  case_id: string
  token: string
  template: string
  rendered_payload: string
  variables: Record<string, string>
  status: 'draft' | 'deployed' | 'hit' | 'archived' | 'expired'
  expected_protocol?: 'dns' | 'http' | 'smtp' | 'ldap'
  expires_at?: string
  created_by: string
  created_at: string
  updated_at: string
  scenario?: string
}

export interface PayloadCreateRequest {
  case_id: string
  template: string
  variables?: Record<string, string>
  expires_at?: string
  expected_protocol?: 'dns' | 'http' | 'smtp' | 'ldap'
}

export interface PayloadListResponse {
  items: Payload[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Interaction types
export interface Interaction {
  id: string
  type: 'dns' | 'http' | 'smtp' | 'ldap' | 'smb' | 'ftp'
  case_id?: string
  payload_id?: string
  token?: string
  timestamp: string
  source_ip: string
  domain?: string
  dns_type?: string
  method?: string
  path?: string
  headers?: Record<string, string>
  body?: string
  user_agent?: string
  content_type?: string
  raw_data: string
  created_at: string
}

export interface InteractionListResponse {
  items: Interaction[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

/** Aggregated interaction counts returned by GET /interactions/stats */
export interface InteractionStats {
  total: number
  by_type?: Record<string, number>
  dns_count?: number
  http_count?: number
  smtp_count?: number
  ldap_count?: number
}

// APIKey types
export interface APIKey {
  id: string
  key: string
  key_prefix: string
  key_masked: string
  name: string
  scopes: string[]
  enabled: boolean
  is_agent?: boolean
  risk_tolerance?: string
  expires_at?: string
  last_used_at?: string
  created_by: string
  created_at: string
  revoked_at?: string
  is_revoked: boolean
}

export interface APIKeyCreateRequest {
  name: string
  scopes: string[]
  is_agent?: boolean
  risk_tolerance?: string
  expires_at?: string
}

export interface APIKeyUpdateRequest {
  name?: string
  scopes?: string[]
  enabled?: boolean
  expires_at?: string
}

export interface APIKeyListResponse {
  items: APIKey[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Settings types
export interface Settings {
  dns_domain: string
  dns_port: number
  dns_ttl: number
  http_port: number
  https_tls_cert: string
  https_tls_key: string
  enable_auth: boolean
  session_timeout: number
  enable_notification: boolean
  notification_url: string
  log_level: string
  log_retention_days: number
}

export type SettingsCreateRequest = Partial<Settings>

export type SettingsUpdateRequest = Partial<Settings>

// User management types
export interface UserCreateRequest {
  username: string
  password: string
  email?: string
  role?: string
}

export interface UserListResponse {
  items: User[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Evidence types
export interface Evidence {
  id: string
  case_id: string
  payload_id?: string
  evidence_strength: 'low' | 'medium' | 'high' | 'critical'
  confidence: number
  interaction_count: number
  unique_sources: number
  timeline: Interaction[]
  explainability: string
  generated_at: string
}

export interface EvidenceRequest {
  case_id?: string
  payload_id?: string
  format: 'json' | 'markdown'
}

export interface EvidenceResponse {
  evidence: Evidence
  format: string
  content: string
  metadata: {
    interaction_count: number
    case_id?: string
    payload_id?: string
  }
}

// Audit types
export interface AuditLog {
  id: string
  user_id?: string
  api_key_id?: string
  api_key_prefix?: string
  is_agent: boolean
  action: string
  resource_type: string
  resource_id?: string
  parameters: string
  result: string
  error_message?: string
  ip_address: string
  user_agent: string
  details?: Record<string, unknown>
  timestamp: string
  created_at: string
}

export interface AuditLogListResponse {
  items: AuditLog[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Scanner Run types
export interface ScannerRun {
  id: string
  case_id: string
  payload_id: string
  scanner: 'nuclei'
  target: string
  template: 'ssrf-basic' | 'xxe-basic' | 'rce-callback'
  delivery_method: 'nuclei-jsonl' | 'nuclei-var'
  command: string
  jsonl: string
  status: 'created' | 'distributed' | 'observed' | 'evidenced'
  created_by: string
  created_at: string
  updated_at: string
}

export interface ScannerRunDetail extends ScannerRun {
  interaction_count: number
  last_interaction_at?: string
  evidence_count: number
  latest_evidence_id?: string
  interactions_url: string
  evidence_url: string
}

export interface ScannerRunCreateRequest {
  case_id: string
  payload_id: string
  scanner: 'nuclei'
  target: string
  template: string
  delivery_method: 'nuclei-jsonl' | 'nuclei-var'
}

export interface ScannerRunUpdateStatusRequest {
  status: 'created' | 'distributed' | 'observed' | 'evidenced'
}

export interface ScannerRunListResponse {
  items: ScannerRun[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Agent Run types
export type AgentRunStatus = 'created' | 'running' | 'waiting' | 'completed' | 'failed' | 'cancelled' | 'timed_out'

export interface AgentRun {
  id: string
  agent_id: string
  operator_id: string
  case_id: string
  payload_id: string
  target: string
  title: string
  status: AgentRunStatus
  started_at?: string
  ended_at?: string
  created_at: string
  updated_at: string
}

export interface AgentOperation {
  id: string
  agent_run_id: string
  agent_id: string
  action: string
  risk_level?: string
  request: string
  result: string
  error?: string
  started_at: string
  ended_at?: string
  created_at: string
}

export interface AgentRunDetail extends AgentRun {
  interaction_count: number
  last_interaction_at?: string
  operations: AgentOperation[]
  case_url: string
  payload_url: string
  interactions_url: string
  evidence_url: string
}

export interface AgentRunCreateRequest {
  agent_id: string
  operator_id: string
  case_id?: string
  payload_id?: string
  target: string
  title: string
}

export interface AgentRunUpdateStatusRequest {
  status: AgentRunStatus
}

export interface AgentRunInteractionSummary {
  total: number
  dns_count: number
  http_count: number
  unique_sources: number
  last_interaction_at?: string
}

export interface AgentRunAuditRef {
  id: string
  action: string
  resource_type: string
  resource_id?: string
  timestamp: string
}

export interface AgentRunReviewPacket {
  id: string
  agent_run: AgentRunDetail
  case_id?: string
  payload_id?: string
  target?: string
  interaction_summary: AgentRunInteractionSummary
  evidence?: {
    id: string
    case_id?: string
    payload_id?: string
    evidence_strength: string
    confidence: number
    interaction_count: number
    unique_sources: number
    explainability: string
    generated_at: string
  }
  audit_refs: AgentRunAuditRef[]
  generated_at: string
  format: string
  content?: string
}

export interface AgentRunListRequest {
  agent_id?: string
  case_id?: string
  payload_id?: string
  status?: string
  page?: number
  page_size?: number
}

export interface AgentRunListResponse {
  items: AgentRunDetail[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface AgentOperationCreateRequest {
  action: string
  risk_level?: string
  request?: Record<string, unknown>
  result?: Record<string, unknown>
  error?: string
}

export type AgentRunFollowupActionType = 'recheck_evidence' | 'wait_more_interactions' | 'create_followup_note'

export interface AgentRunFollowupRequest {
  action_type: AgentRunFollowupActionType
  reason: string
  review_packet_id?: string
}

export interface AgentRunFollowupResponse {
  agent_run_id: string
  operation_id: string
  action_type: AgentRunFollowupActionType
  reason: string
  review_packet_id?: string
  operation: AgentOperation
  created_at: string
}

