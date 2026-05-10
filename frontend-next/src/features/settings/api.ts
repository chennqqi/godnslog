/** Settings configuration key-value type */
export interface Settings {
  key: string
  value: string
  description?: string
}

export interface SettingsListResponse {
  items: Settings[]
  total: number
}

/** Feature-layer settings API wrapper */
export const settingsApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<SettingsListResponse> => {
    const response = await fetch(
      `/api/v2/settings?page=${params.page || 1}&page_size=${params.page_size || 20}`,
      { headers: { Authorization: `Bearer ${localStorage.getItem('token')}` } }
    )
    const data = await response.json()
    return data.data ?? { items: [], total: 0 }
  },

  get: async (key: string): Promise<Settings | undefined> => {
    const response = await fetch(`/api/v2/settings/${key}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
    const data = await response.json()
    return data.data
  },

  update: async (key: string, value: string): Promise<Settings | undefined> => {
    const response = await fetch(`/api/v2/settings/${key}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      body: JSON.stringify({ value }),
    })
    const result = await response.json()
    return result.data
  },
}
