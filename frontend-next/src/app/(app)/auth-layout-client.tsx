'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { ErrorBoundary } from '@/components/error-boundary'
import { Spinner } from '@/components/ui/spinner'

export function AuthLayoutClient({ children }: { children: React.ReactNode }) {
  const [authChecked, setAuthChecked] = useState(false)

  useEffect(() => {
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      window.location.href = '/login'
      return
    }
    setAuthChecked(true)
  }, [])

  if (!authChecked) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner className="w-8 h-8" />
      </div>
    )
  }

  return <AppShell><ErrorBoundary>{children}</ErrorBoundary></AppShell>
}
