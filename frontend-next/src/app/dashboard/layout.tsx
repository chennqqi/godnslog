'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/app-shell'
import { useAuthStore } from '@/features/auth/store'

/** Dashboard layout: guards auth and wraps all sub-pages with the enterprise AppShell */
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const { token } = useAuthStore()
  const [authChecked, setAuthChecked] = useState(false)

  useEffect(() => {
    const storedToken = typeof window !== 'undefined' ? localStorage.getItem('token') : null
    if (!storedToken && !token) {
      router.replace('/login')
      return
    }
    setAuthChecked(true)
  }, [router, token])

  if (!authChecked) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    )
  }

  return <AppShell>{children}</AppShell>
}
