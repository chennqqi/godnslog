'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/app-shell'
import { ErrorBoundary } from '@/components/error-boundary'

export function AuthLayoutClient({ children }: { children: React.ReactNode }) {
  const router = useRouter()

  useEffect(() => {
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      router.replace('/login')
    }
  }, [router])

  return <AppShell><ErrorBoundary>{children}</ErrorBoundary></AppShell>
}
