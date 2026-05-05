'use client'

import { useAuthStore } from '../store'

export function useAuth() {
  return useAuthStore()
}
