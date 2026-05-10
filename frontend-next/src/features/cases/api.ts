import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest, CaseUpdateRequest, CaseListResponse } from '@/types'

/** Feature-layer case API wrapper */
export const casesApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<CaseListResponse | undefined> => {
    const response = await caseApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Case | undefined> => {
    const response = await caseApi.get(id)
    return response.data as unknown as Case | undefined
  },

  create: async (data: CaseCreateRequest): Promise<Case | undefined> => {
    const response = await caseApi.create(data)
    return response.data as unknown as Case | undefined
  },

  update: async (id: string, data: CaseUpdateRequest): Promise<Case | undefined> => {
    const response = await caseApi.update(id, data)
    return response.data as unknown as Case | undefined
  },

  delete: async (id: string): Promise<void> => {
    await caseApi.delete(id)
  },
}
