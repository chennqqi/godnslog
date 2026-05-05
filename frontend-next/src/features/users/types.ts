import type { User, UserCreateRequest } from '@/types'

export type { User, UserCreateRequest }

export interface UserFormData {
  username: string
  email: string
  password: string
  role: number
}
