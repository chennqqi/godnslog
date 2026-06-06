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
  APIKey,
  APIKeyCreateRequest,
  APIKeyUpdateRequest,
  APIKeyListResponse,
  EvidenceRequest,
  EvidenceResponse,
  AuditLogListResponse,
  ScannerRun,
  ScannerRunDetail,
  ScannerRunCreateRequest,
  ScannerRunUpdateStatusRequest,
  ScannerRunListResponse,
  AgentRun,
  AgentRunDetail,
  AgentRunCreateRequest,
  AgentRunUpdateStatusRequest,
  AgentRunListRequest,
  AgentRunListResponse,
  AgentOperationCreateRequest,
  AgentRunReviewPacket,
  AgentRunFollowupRequest,
  AgentRunFollowupResponse,
} from '@/types'

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
  get: (id: string) => api.get<{ data: Interaction }>(`/interactions/${id}`),
  delete: (ids: string[]) => api.post('/interactions/delete', { ids }),
  export: (data: Record<string, unknown>) => api.post('/interactions/export', data),
}

// APIKey API
export const apiKeyApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<APIKeyListResponse>('/apikeys', params),
  create: (data: APIKeyCreateRequest) => api.post<{ data: APIKey }>('/apikeys', data),
  get: (id: string) => api.get<{ data: APIKey }>(`/apikeys/${id}`),
  update: (id: string, data: APIKeyUpdateRequest) => api.put<{ data: APIKey }>(`/apikeys/${id}`, data),
  delete: (id: string) => api.delete(`/apikeys/${id}`),
}

// Users API
export const usersApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/users', params),
}

// Marketplace API
export const marketplaceApi = {
  listPlugins: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/marketplace/plugins', params),
  getPlugin: (id: string) => api.get<unknown>(`/marketplace/plugins/${id}`),
  listTemplates: (params?: { page?: number; page_size?: number }) =>
    api.get<UnknownItemListResponse>('/marketplace/templates', params),
  getTemplate: (id: string) => api.get<unknown>(`/marketplace/templates/${id}`),
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
}

// Audit API
export const auditApi = {
  list: (params?: {
    user_id?: string
    action?: string
    resource_type?: string
    start_time?: string
    end_time?: string
    page?: number
    page_size?: number
  }) => api.get<AuditLogListResponse>('/audit/logs', params),
}

// Scanner Run API
export const scannerRunApi = {
  list: (params?: {
    case_id?: string
    payload_id?: string
    scanner?: string
    status?: string
    page?: number
    page_size?: number
  }) => api.get<ScannerRunListResponse>('/scanner-runs', params),
  get: (id: string) => api.get<{ data: ScannerRunDetail }>(`/scanner-runs/${id}`),
  create: (data: ScannerRunCreateRequest) => api.post<{ data: ScannerRun }>('/scanner-runs', data),
  updateStatus: (id: string, data: ScannerRunUpdateStatusRequest) =>
    api.put<{ data: ScannerRun }>(`/scanner-runs/${id}/status`, data),
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
}
