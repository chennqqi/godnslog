'use client'

import { useEffect } from 'react'
import { AppShell } from '@/components/app-shell'
import { ErrorBoundary } from '@/components/error-boundary'

export function AuthLayoutClient({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      window.location.href = '/login'
    }
  }, [])

  return <AppShell><ErrorBoundary>{children}</ErrorBoundary></AppShell>
}
