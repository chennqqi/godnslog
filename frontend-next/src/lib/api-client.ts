import { api } from './api'
import type {
  LoginRequest,
  LoginResponse,
  Case,
  CaseCreateRequest,
  CaseUpdateRequest,
  CaseListResponse,
  Payload,
  PayloadCreateRequest,
  PayloadListResponse,
  Interaction,
  InteractionListResponse,
  InteractionStats,
  DailyStat,
  APIKey,
  APIKeyCreateRequest,
  APIKeyUpdateRequest,
  APIKeyListResponse,
  AgentScopeCatalog,
  EvidenceRequest,
  EvidenceResponse,
  EvidenceSummaryRequest,
  EvidenceSummaryResponse,
  AuditLogListResponse,
  ScannerRun,
  ScannerRunDetail,
  ScannerRunCreateRequest,
  ScannerRunUpdateStatusRequest,
  ScannerRunListResponse,
  ScannerAdapterListResponse,
  ScannerRunCreateFromSearchRequest,
  SearchResult,
  SearchResultItem,
  AgentRun,
  AgentRunDetail,
  AgentRunCreateRequest,
  AgentRunUpdateStatusRequest,
  AgentRunReviewDeliveryRequest,
  AgentRunReviewDeliveryResponse,
  AgentRunReviewDeliveryHistoryResponse,
  AgentRunListRequest,
  AgentRunListResponse,
  AgentOperationCreateRequest,
  AgentRunReviewPacket,
  AgentRunFollowupRequest,
  AgentRunFollowupResponse,
  ReviewQueueFilters,
  AgentRunReviewQueueResponse,
  AgentRunFollowupHistoryItem,
  AgentRunReviewDecisionRequest,
  AgentRunReviewDecisionResponse,
  AgentRunReviewExportRequest,
  AgentRunReviewExportResponse,
  AgentRunReviewPackageTraceResponse,
} from '@/types'

interface UserListItem {
  id: string
  username: string
  email: string
  role: number
  created_at: string
}

interface UserListResponse {
  items: UserListItem[]
  total: number
  page: number
  page_size: number
}

interface UnknownItemListResponse {
  items: never[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Auth API
export const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  info: () => api.get('/auth/info'),
}

/** Summary counts for a case from GET /cases/:id/stats */
export interface CaseStats {
  payload_count: number
  interaction_count: number
  hit_payload_count: number
}

// Case API
export const caseApi = {
  list: (params?: { status?: string; search?: string; page?: number; page_size?: number }) =>
    api.get<CaseListResponse>('/cases', params),
  get: (id: string) => api.get<{ data: Case }>(`/cases/${id}`),
  /** Returns aggregate payload and interaction counts for the case. */
  stats: (id: string) => api.get<CaseStats>(`/cases/${id}/stats`),
  create: (data: CaseCreateRequest) => api.post<{ data: Case }>('/cases', data),
  update: (id: string, data: CaseUpdateRequest) => api.put<{ data: Case }>(`/cases/${id}`, data),
  delete: (id: string) => api.delete(`/cases/${id}`),
}

// Payload API
export const payloadApi = {
  list: (params?: { case_id?: string; status?: string; page?: number; page_size?: number }) =>
    api.get<PayloadListResponse>('/payloads', params),
  get: (id: string) => api.get<{ data: Payload }>(`/payloads/${id}`),
  create: (data: PayloadCreateRequest) => api.post<{ data: Payload }>('/payloads', data),
  revoke: (id: string) => api.post(`/payloads/${id}/revoke`),
  preview: (id: string) => api.post<{ data: { rendered: string } }>(`/payloads/${id}/preview`),
  batchCreate: (data: { case_id: string; template: string; variables?: Record<string, string>; count: number }) =>
    api.post<{ data: Payload[] }>('/payloads/batch', data),
}

// Interaction API
export const interactionApi = {
  list: (params?: {
    case_id?: string
    payload_id?: string
    type?: string
    start_time?: string
    end_time?: string
    page?: number
    page_size?: number
  }) => api.get<InteractionListResponse>('/interactions', params),
  stats: (params?: { case_id?: string; payload_id?: string; period?: string }) =>
    api.get<InteractionStats>('/interactions/stats', params),
  dailyStats: (params?: { case_id?: string; payload_id?: string; days?: number }) =>
    api.get<DailyStat[]>('/interactions/stats/daily', params),
  get: (id: string) => api.get<{ data: Interaction }>(`/interactions/${id}`),
  delete: (ids: string[]) => api.post('/interactions/delete', { ids }),
  export: (data: Record<string, unknown>) => api.post('/interactions/export', data),
}

// APIKey API
export const apiKeyApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<APIKeyListResponse>('/apikeys', params),
  create: (data: APIKeyCreateRequest) => api.post<APIKey>('/apikeys', data),
  get: (id: string) => api.get<APIKey>(`/apikeys/${id}`),
  update: (id: string, data: APIKeyUpdateRequest) => api.put<APIKey>(`/apikeys/${id}`, data),
  delete: (id: string) => api.delete(`/apikeys/${id}`),
}

export const agentPolicyApi = {
  listScopes: () => api.get<AgentScopeCatalog>('/agent-policy/scopes'),
}

// Users API
export const usersApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<UserListResponse>('/users', params),
  create: (data: { username: string; email: string; password: string; role: number }) =>
    api.post<unknown>('/users', data),
  update: (id: string, data: { email?: string; role?: number; password?: string }) =>
    api.put<unknown>(`/users/${id}`, data),
  delete: (id: string) => api.delete(`/users/${id}`),
}

// Marketplace API
export const marketplaceApi = {
  listPlugins: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/marketplace/plugins', params),
  createPlugin: (data: Record<string, unknown>) =>
    api.post<{ data: unknown }>('/marketplace/plugins', data),
  getPlugin: (id: string) => api.get<unknown>(`/marketplace/plugins/${id}`),
  installPlugin: (id: string, body?: { version?: string; config?: string }) =>
    api.post<unknown>(`/marketplace/plugins/${id}/install`, body),
  listTemplates: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/marketplace/templates', params),
  createTemplate: (data: Record<string, unknown>) =>
    api.post<{ data: unknown }>('/marketplace/templates', data),
  getTemplate: (id: string) => api.get<unknown>(`/marketplace/templates/${id}`),
  listInstalled: () =>
    api.get<UnknownItemListResponse>('/marketplace/installed'),
  uninstallPlugin: (id: string) =>
    api.delete<unknown>(`/marketplace/installed/${id}`),
}

// Rules/Workflow API
export const rulesApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/rules', params),
  get: (id: string) => api.get<unknown>(`/rules/${id}`),
  create: (data: Record<string, unknown>) => api.post<never>('/rules', data),
  update: (id: string, data: Record<string, unknown>) => api.put<unknown>(`/rules/${id}`, data),
  delete: (id: string) => api.delete(`/rules/${id}`),
}

// Evidence API
export const evidenceApi = {
  generate: (data: EvidenceRequest) => api.post<EvidenceResponse>('/evidence/generate', data),
  summary: (data: EvidenceSummaryRequest) => api.post<EvidenceSummaryResponse>('/evidence/summary', data),
}

// Audit API
export const auditApi = {
  list: (params?: {
    user_id?: string
    action?: string
    resource_type?: string
    resource_id?: string
    start_time?: string
    end_time?: string
    page?: number
    page_size?: number
  }) => api.get<AuditLogListResponse>('/audit/logs', params),
}

// Settings API
export const settingsApi = {
  get: () => api.get<unknown>('/settings'),
  update: (data: Record<string, unknown>) => api.put<unknown>('/settings', data),
}

// Scanner Run API
export const scannerRunApi = {
  listAdapters: () => api.get<ScannerAdapterListResponse>('/scanner-hub/adapters'),
  list: (params?: {
    case_id?: string
    payload_id?: string
    scanner?: string
    status?: string
    page?: number
    page_size?: number
  }) => api.get<ScannerRunListResponse>('/scanner-runs', params),
  get: (id: string) => api.get<{ data: ScannerRunDetail }>(`/scanner-runs/${id}`),
  create: (data: ScannerRunCreateRequest) => api.post<ScannerRun>('/scanner-runs', data),
  createFromSearch: (data: ScannerRunCreateFromSearchRequest) =>
    api.post<{ items: ScannerRun[]; total: number }>('/scanner-runs/from-search', data),
  updateStatus: (id: string, data: ScannerRunUpdateStatusRequest) =>
    api.put<{ data: ScannerRun }>(`/scanner-runs/${id}/status`, data),
}

// Search Engine API
export const searchApi = {
  zoomeye: (params: { q: string; page?: number }) =>
    api.get<SearchResult>('/search/zoomeye', params),
  shodan: (params: { q: string; page?: number }) =>
    api.get<SearchResult>('/search/shodan', params),
  fofa: (params: { q: string; page?: number }) =>
    api.get<SearchResult>('/search/fofa', params),
}

// Agent Run API
export const agentRunApi = {
  list: (params?: AgentRunListRequest) =>
    api.get<AgentRunListResponse>('/agent-runs', params),
  get: (id: string) => api.get<{ data: AgentRunDetail }>(`/agent-runs/${id}`),
  getReview: (id: string, format: 'json' | 'markdown' = 'json') =>
    api.get<{ data: AgentRunReviewPacket }>(`/agent-runs/${id}/review`, { format }),
  create: (data: AgentRunCreateRequest) => api.post<{ data: AgentRun }>('/agent-runs', data),
  updateStatus: (id: string, data: AgentRunUpdateStatusRequest) =>
    api.put<{ data: AgentRun }>(`/agent-runs/${id}/status`, data),
  appendOperation: (id: string, data: AgentOperationCreateRequest) =>
    api.post<{ data: AgentRun }>(`/agent-runs/${id}/operations`, data),
  createFollowup: (id: string, data: AgentRunFollowupRequest) =>
    api.post<{ data: AgentRunFollowupResponse }>(`/agent-runs/${id}/followups`, data),
  createReviewDecision: (id: string, data: AgentRunReviewDecisionRequest) =>
    api.post<{ data: AgentRunReviewDecisionResponse }>(`/agent-runs/${id}/review-decision`, data),
  exportReview: (id: string, data: AgentRunReviewExportRequest) =>
    api.post<{ data: AgentRunReviewExportResponse }>(`/agent-runs/${id}/review-export`, data),
  deliverReview: (id: string, data: AgentRunReviewDeliveryRequest) =>
    api.post<{ data: AgentRunReviewDeliveryResponse }>(`/agent-runs/${id}/review-delivery`, data),
  listReviewDeliveries: (id: string) =>
    api.get<{ data: AgentRunReviewDeliveryHistoryResponse }>(`/agent-runs/${id}/review-deliveries`),
  listReviewQueue: (params?: ReviewQueueFilters) =>
    api.get<AgentRunReviewQueueResponse>('/agent-runs/review-queue', params),
  listFollowupHistory: (id: string) =>
    api.get<{ data: AgentRunFollowupHistoryItem[] }>(`/agent-runs/${id}/followups`),
  traceReviewPackage: (packageHash: string) =>
    api.get<AgentRunReviewPackageTraceResponse>('/agent-runs/review-package-trace', { package_hash: packageHash }),
}

// Listener API
export interface ProtocolListener {
  id: string
  protocol: 'smtp' | 'ldap' | 'smb' | 'ftp'
  host: string
  port: number
  token: string
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export const listenerApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<{ items: ProtocolListener[]; total: number }>('/listeners', params),
  get: (id: string) => api.get<{ data: ProtocolListener }>(`/listeners/${id}`),
  create: (data: Partial<ProtocolListener>) => api.post<{ data: ProtocolListener }>('/listeners', data),
  update: (id: string, data: Partial<ProtocolListener>) => api.put<{ data: ProtocolListener }>(`/listeners/${id}`, data),
  delete: (id: string) => api.delete(`/listeners/${id}`),
}

// Canary API
export interface CanaryToken {
  id: string
  type: string
  token: string
  description: string
  context: string
  expires_at: string
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface CanaryHit {
  id: string
  canary_id: string
  source_ip: string
  user_agent: string
  headers: string
  body: string
  timestamp: string
  is_compressed: boolean
}

export const canaryApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<{ items: CanaryToken[]; total: number }>('/canary', params),
  create: (data: { type: string; token: string; description?: string; context?: string; expires_at?: string }) =>
    api.post<{ data: CanaryToken }>('/canary', data),
  get: (id: string) => api.get<{ data: CanaryToken }>(`/canary/${id}`),
  update: (id: string, data: { description?: string; is_enabled?: boolean }) =>
    api.put<{ data: CanaryToken }>(`/canary/${id}`, data),
  delete: (id: string) => api.delete(`/canary/${id}`),
  listHits: (id: string) => api.get<{ data: CanaryHit[] }>(`/canary/${id}/hits`),
}

// Rebinding API
export interface RebindingStage {
  order: number
  target_ip: string
  ttl: number
  hit_count: number
  max_hits: number
  condition: string
  description: string
}

export interface RebindingRule {
  id: string
  domain: string
  stages: RebindingStage[]
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface RebindingScenario {
  name: string
  description: string
  stages: RebindingStage[]
}

export interface RebindingSession {
  id: string
  rule_id: string
  source_ip: string
  current_stage: number
  hit_count: number
  started_at: string
  last_hit: string
}

export const rebindingApi = {
  listRules: (params?: { page?: number; page_size?: number }) =>
    api.get<{ items: RebindingRule[]; total: number }>('/rebinding/rules', params),
  createRule: (data: Partial<RebindingRule>) => api.post<{ data: RebindingRule }>('/rebinding/rules', data),
  getRule: (id: string) => api.get<{ data: RebindingRule }>(`/rebinding/rules/${id}`),
  updateRule: (id: string, data: Partial<RebindingRule>) => api.put<{ data: RebindingRule }>(`/rebinding/rules/${id}`, data),
  deleteRule: (id: string) => api.delete(`/rebinding/rules/${id}`),
  listSessions: (id: string) => api.get<{ data: { rule_id: string; sessions: RebindingSession[]; total: number } }>(`/rebinding/rules/${id}/sessions`),
  listScenarios: () => api.get<RebindingScenario[]>('/rebinding/scenarios'),
  createFromScenario: (name: string, data: { domain: string }) =>
    api.post<{ data: RebindingRule }>(`/rebinding/scenarios/${name}/rules`, data),
}

// Retention API
export interface RetentionPolicy {
  id: string
  name: string
  description: string
  apply_to_interactions: boolean
  apply_to_cases: boolean
  apply_to_payloads: boolean
  apply_to_evidence: boolean
  apply_to_logs: boolean
  retention_days: number
  max_records: number
  archive_after_days: number
  archive_to_storage: string
  delete_after_archive: boolean
  run_hourly: boolean
  run_daily: boolean
  run_weekly: boolean
  run_monthly: boolean
  run_interval_hours: number
  is_enabled: boolean
  created_at: string
  updated_at: string
  last_run_at: string | null
}

export interface RetentionJob {
  id: string
  policy_id: string
  job_type: string
  status: string
  records_processed: number
  records_deleted: number
  records_archived: number
  error_message: string
  started_at: string
  completed_at: string | null
  duration: number
  created_at: string
}

export interface RetentionArchive {
  id: string
  policy_id: string
  data_type: string
  record_count: number
  storage_path: string
  file_size: number
  checksum: string
  compression: string
  status: string
  created_at: string
  completed_at: string | null
}

export const retentionApi = {
  listPolicies: () => api.get<{ items: RetentionPolicy[]; total: number }>('/retention/policies'),
  createPolicy: (data: Partial<RetentionPolicy>) => api.post<{ data: RetentionPolicy }>('/retention/policies', data),
  getPolicy: (id: string) => api.get<{ data: RetentionPolicy }>(`/retention/policies/${id}`),
  updatePolicy: (id: string, data: Partial<RetentionPolicy>) => api.put<{ data: RetentionPolicy }>(`/retention/policies/${id}`, data),
  deletePolicy: (id: string) => api.delete(`/retention/policies/${id}`),
  runPolicy: (id: string) => api.post<{ data: RetentionJob }>(`/retention/policies/${id}/run`),
  listJobs: () => api.get<{ items: RetentionJob[]; total: number }>('/retention/jobs'),
  listArchives: () => api.get<{ items: RetentionArchive[]; total: number }>('/retention/archives'),
}

/**
 * Raw API client that auto-unwraps response data.
 * Accepts full API paths including /api/v2 prefix.
 * Strips the /api/v2 prefix internally since the underlying api has it as baseURL.
 */
export const apiClient = {
  async get<T>(url: string): Promise<T> {
    // Strip /api/v2 prefix since api already has it as baseURL
    const path = url.startsWith('/api/v2') ? url.slice(7) : url
    const response = await api.get<T>(path)
    return response.data as T
  },
  async post<T>(url: string, data?: unknown): Promise<T> {
    const path = url.startsWith('/api/v2') ? url.slice(7) : url
    const response = await api.post<T>(path, data)
    return response.data as T
  },
  async put<T>(url: string, data?: unknown): Promise<T> {
    const path = url.startsWith('/api/v2') ? url.slice(7) : url
    const response = await api.put<T>(path, data)
    return response.data as T
  },
  async delete<T>(url: string): Promise<T> {
    const path = url.startsWith('/api/v2') ? url.slice(7) : url
    const response = await api.delete<T>(path)
    return response.data as T
  },
}
