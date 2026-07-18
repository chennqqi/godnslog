import { AuthLayoutClient } from './auth-layout-client'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return <AuthLayoutClient>{children}</AuthLayoutClient>
}
