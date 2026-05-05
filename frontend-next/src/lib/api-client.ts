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
  APIKey,
  APIKeyCreateRequest,
  APIKeyListResponse,
} from '@/types'

// Auth API
export const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  info: () => api.get('/auth/info'),
}

// Case API
export const caseApi = {
  list: (params?: { status?: string; search?: string; page?: number; page_size?: number }) =>
    api.get<CaseListResponse>('/cases', params),
  get: (id: string) => api.get<{ data: Case }>(`/cases/${id}`),
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
  get: (id: string) => api.get<{ data: Interaction }>(`/interactions/${id}`),
  delete: (ids: string[]) => api.post('/interactions/delete', { ids }),
  export: (data: any) => api.post('/interactions/export', data),
}

// APIKey API
export const apiKeyApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<APIKeyListResponse>('/apikeys', params),
  create: (data: APIKeyCreateRequest) => api.post<{ data: APIKey }>('/apikeys', data),
  delete: (id: string) => api.delete(`/apikeys/${id}`),
}

// Users API
export const usersApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<any>('/users', params),
}

// Marketplace API
export const marketplaceApi = {
  listPlugins: (params?: { page?: number; page_size?: number }) =>
    api.get<any>('/marketplace/plugins', params),
  getPlugin: (id: string) => api.get<any>(`/marketplace/plugins/${id}`),
  listTemplates: (params?: { page?: number; page_size?: number }) =>
    api.get<any>('/marketplace/templates', params),
  getTemplate: (id: string) => api.get<any>(`/marketplace/templates/${id}`),
}

// Rules/Workflow API
export const rulesApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get<any>('/rules', params),
  get: (id: string) => api.get<any>(`/rules/${id}`),
  create: (data: any) => api.post<any>('/rules', data),
  update: (id: string, data: any) => api.put<any>(`/rules/${id}`, data),
  delete: (id: string) => api.delete(`/rules/${id}`),
}
