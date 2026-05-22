// API Response types
export interface ApiResponse<T = any> {
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

export interface SettingsCreateRequest extends Partial<Settings> {}

export interface SettingsUpdateRequest extends Partial<Settings> {}

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
