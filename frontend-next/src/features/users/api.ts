import { usersApi } from '@/lib/api-client'
import type { User, UserCreateRequest, UserListResponse } from '@/types'

export const userFeaturesApi = {
  list: async (params: { page?: number; page_size?: number }): Promise<UserListResponse> => {
    const response = await usersApi.list(params)
    return response.data
  },

  get: async (id: string): Promise<User> => {
    const response = await fetch(`/api/v2/users/${id}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
    const data = await response.json()
    return data.data
  },

  create: async (data: UserCreateRequest): Promise<User> => {
    const response = await fetch('/api/v2/users', {
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

  delete: async (id: string): Promise<void> => {
    await fetch(`/api/v2/users/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
      },
    })
  },
}
