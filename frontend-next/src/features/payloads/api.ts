import { payloadApi } from '@/lib/api-client'
import type { Payload, PayloadCreateRequest, PayloadListResponse } from '@/types'

/** Feature-layer payload API wrapper */
export const payloadsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<PayloadListResponse | undefined> => {
    const response = await payloadApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Payload | undefined> => {
    const response = await payloadApi.get(id)
    return response.data as unknown as Payload | undefined
  },

  create: async (data: PayloadCreateRequest): Promise<Payload | undefined> => {
    const response = await payloadApi.create(data)
    return response.data as unknown as Payload | undefined
  },

  revoke: async (id: string): Promise<void> => {
    await payloadApi.revoke(id)
  },
}
