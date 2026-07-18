'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { ErrorBoundary } from '@/components/error-boundary'
import { Spinner } from '@/components/ui/spinner'
import { useAuthStore } from '@/features/auth/store'

/** Dashboard layout: guards auth and wraps all sub-pages with the enterprise AppShell */
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { token } = useAuthStore()
  const [authChecked, setAuthChecked] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const storedToken = localStorage.getItem('token')
    if (!storedToken && !token) {
      window.location.href = '/login'
      return
    }
    // Defer state update to avoid cascading renders
    const timer = setTimeout(() => setAuthChecked(true), 0)
    return () => clearTimeout(timer)
  }, [token])

  if (!authChecked) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="flex flex-col items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center font-bold text-white text-lg animate-pulse">
            G
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-400 dark:text-gray-600">
            <Spinner size="sm" />
            Loading...
          </div>
        </div>
      </div>
    )
  }

  return <AppShell><ErrorBoundary>{children}</ErrorBoundary></AppShell>
}
