import { interactionApi } from '@/lib/api-client'
import type { Interaction, InteractionListResponse, ExportRequest, DeleteRequest } from '@/types'

export const interactionsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<InteractionListResponse> => {
    const response = await interactionApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Interaction> => {
    const response = await interactionApi.get(id)
    return response.data
  },

  export: async (data: ExportRequest): Promise<Blob> => {
    const response = await fetch('/api/v2/interactions/export', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
      body: JSON.stringify(data),
    })
    return response.blob()
  },

  delete: async (data: DeleteRequest): Promise<void> => {
    await interactionApi.delete(data)
  },
}
