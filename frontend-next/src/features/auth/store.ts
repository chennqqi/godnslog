import { create } from 'zustand'

export interface AuthState {
  token: string | null
  user: { id: string; name: string; email: string; role: string } | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  setUser: (user: AuthState['user']) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: typeof window !== 'undefined' ? localStorage.getItem('token') : null,
  user: null,
  isAuthenticated: false,
  setToken: (token) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token)
    }
    set({ token, isAuthenticated: true })
  },
  setUser: (user) => set({ user }),
  logout: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
    set({ token: null, user: null, isAuthenticated: false })
  },
}))
