'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type Theme = 'light' | 'dark' | 'system'

interface ThemeState {
  theme: Theme
  resolvedTheme: 'light' | 'dark'
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  setResolvedTheme: (theme: 'light' | 'dark') => void
}

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(theme: 'light' | 'dark') {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  if (theme === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'system',
      resolvedTheme: 'light',
      setTheme: (theme) => {
        const resolved = theme === 'system' ? getSystemTheme() : theme
        applyTheme(resolved)
        set({ theme, resolvedTheme: resolved })
      },
      toggleTheme: () => {
        const current = get().resolvedTheme
        const next = current === 'dark' ? 'light' : 'dark'
        applyTheme(next)
        set({ theme: next, resolvedTheme: next })
      },
      setResolvedTheme: (resolvedTheme) => {
        applyTheme(resolvedTheme)
        set({ resolvedTheme })
      },
    }),
    {
      name: 'godnslog-theme',
      partialize: (state) => ({ theme: state.theme }),
    }
  )
)

/**
 * Initialize theme on page load. Call this in a useEffect in the root layout.
 * Applies the persisted theme or system theme to the document.
 */
export function initTheme() {
  if (typeof window === 'undefined') return
  const state = useThemeStore.getState()
  const resolved = state.theme === 'system' ? getSystemTheme() : state.theme
  applyTheme(resolved)
  useThemeStore.getState().setResolvedTheme(resolved)

  // Listen for system theme changes if in system mode
  const mql = window.matchMedia('(prefers-color-scheme: dark)')
  const handler = (e: MediaQueryListEvent) => {
    const current = useThemeStore.getState()
    if (current.theme === 'system') {
      const newResolved = e.matches ? 'dark' : 'light'
      applyTheme(newResolved)
      current.setResolvedTheme(newResolved)
    }
  }
  mql.addEventListener('change', handler)
}
