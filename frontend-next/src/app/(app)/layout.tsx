/**
 * Dashboard layout (server component).
 * Exports route segment config for dynamic rendering.
 * Delegates to AuthLayoutClient for the actual UI and auth guard.
 */
export const dynamic = 'force-dynamic'

import { AuthLayoutClient } from './auth-layout-client'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return <AuthLayoutClient>{children}</AuthLayoutClient>
}
