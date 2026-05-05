import type { Settings, SettingsCreateRequest, SettingsUpdateRequest, SettingsListResponse } from '@/types'

export const settingsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<SettingsListResponse> => {
    const response = await fetch(`/api/v2/settings?page=${params.page || 1}&page_size=${params.page_size || 20}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
    const data = await response.json()
    return data.data
  },

  get: async (key: string): Promise<Settings> => {
    const response = await fetch(`/api/v2/settings/${key}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
    const data = await response.json()
    return data.data
  },

  create: async (data: SettingsCreateRequest): Promise<Settings> => {
    const response = await fetch('/api/v2/settings', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
      body: JSON.stringify(data),
    })
    const result = await response.json()
    return result.data
  },

  update: async (key: string, data: SettingsUpdateRequest): Promise<Settings> => {
    const response = await fetch(`/api/v2/settings/${key}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
      body: JSON.stringify(data),
    })
    const result = await response.json()
    return result.data
  },

  delete: async (key: string): Promise<void> => {
    await fetch(`/api/v2/settings/${key}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
  },
}
