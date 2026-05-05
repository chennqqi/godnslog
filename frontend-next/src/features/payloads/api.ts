import { payloadApi } from '@/lib/api-client'
import type { Payload, PayloadCreateRequest, PayloadUpdateRequest, PayloadListResponse } from '@/types'

export const payloadsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<PayloadListResponse> => {
    const response = await payloadApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Payload> => {
    const response = await payloadApi.get(id)
    return response.data
  },

  create: async (data: PayloadCreateRequest): Promise<Payload> => {
    const response = await payloadApi.create(data)
    return response.data
  },

  update: async (id: string, data: PayloadUpdateRequest): Promise<Payload> => {
    const response = await payloadApi.update(id, data)
    return response.data
  },

  revoke: async (id: string): Promise<void> => {
    await payloadApi.revoke(id)
  },
}
