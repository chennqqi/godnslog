import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest, CaseUpdateRequest, CaseListResponse } from '@/types'

export const casesApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<CaseListResponse> => {
    const response = await caseApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Case> => {
    const response = await caseApi.get(id)
    return response.data
  },

  create: async (data: CaseCreateRequest): Promise<Case> => {
    const response = await caseApi.create(data)
    return response.data
  },

  update: async (id: string, data: CaseUpdateRequest): Promise<Case> => {
    const response = await caseApi.update(id, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await caseApi.delete(id)
  },
}
