import axios, { AxiosInstance, AxiosError } from 'axios'
import type { ApiResponse } from '@/types'

export interface ApiError {
  code: number
  message: string
  originalError?: AxiosError<ApiResponse>
}

/**
 * Get token from localStorage safely (works in browser only).
 */
function getToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('token')
  }
  return null
}

/**
 * Clear auth state and redirect to login.
 */
function clearAuthAndRedirect() {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.href = '/login'
  }
}

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v2',
      timeout: 30000,
    })

    // Request interceptor: attach Bearer token
    this.client.interceptors.request.use(
      (config) => {
        const token = getToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor: handle 401 and API-level errors
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiResponse>) => {
        if (error.response?.status === 401) {
          clearAuthAndRedirect()
        }
        if (error.response?.data?.code !== undefined) {
          const apiError: ApiError = {
            code: error.response.data.code,
            message: error.response.data.message || 'Unknown error',
            originalError: error,
          }
          return Promise.reject(apiError)
        }
        return Promise.reject(error)
      }
    )
  }

  async get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
    const response = await this.client.get<ApiResponse<T>>(url, { params })
    return response.data
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.post<ApiResponse<T>>(url, data)
    return response.data
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.put<ApiResponse<T>>(url, data)
    return response.data
  }

  async delete<T>(url: string): Promise<ApiResponse<T>> {
    const response = await this.client.delete<ApiResponse<T>>(url)
    return response.data
  }
}

export const api = new ApiClient()
