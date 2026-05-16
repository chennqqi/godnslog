import { interactionApi } from '@/lib/api-client'
import type { Interaction, InteractionListResponse } from '@/types'

/** Feature-layer interaction API wrapper */
export const interactionsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<InteractionListResponse | undefined> => {
    const response = await interactionApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<Interaction | undefined> => {
    const response = await interactionApi.get(id)
    return response.data as unknown as Interaction | undefined
  },

  getStats: async (): Promise<{ today: number; total: number; high_risk: number } | undefined> => {
    const response = await fetch('/api/v2/interactions/stats', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
    if (!response.ok) {
      return undefined
    }
    const result = await response.json()
    return result.data
  },

  export: async (data: Record<string, unknown>): Promise<Blob> => {
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

  delete: async (ids: string[]): Promise<void> => {
    await interactionApi.delete(ids)
  },
}
