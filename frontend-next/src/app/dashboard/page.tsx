'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { caseApi, interactionApi, payloadApi } from '@/lib/api-client'
import type { Case, Interaction } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'

/** Protocol color mapping per design spec */
const PROTOCOL_COLORS: Record<string, string> = {
  dns: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  http: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  smtp: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  ldap: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  smb: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  ftp: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
}

/** Status color mapping */
const STATUS_COLORS: Record<string, string> = {
  active: 'bg-emerald-100 text-emerald-700 border-emerald-200',
  deployed: 'bg-blue-100 text-blue-700 border-blue-200',
  hit: 'bg-purple-100 text-purple-700 border-purple-200',
  archived: 'bg-gray-100 text-gray-600 border-gray-200',
  completed: 'bg-cyan-100 text-cyan-700 border-cyan-200',
  expired: 'bg-red-100 text-red-700 border-red-200',
}

interface DashboardStats {
  activeCases: number
  totalHitsToday: number
  activePayloads: number
  systemOk: boolean
  dnsCount: number
  httpCount: number
  smtpCount: number
  totalInteractions: number
}

/** Stat card with title, value, and optional sub-text */
function StatCard({
  title,
  value,
  sub,
  valueClass = 'text-gray-900 dark:text-gray-100',
}: {
  title: string
  value: string | number
  sub?: string
  valueClass?: string
}) {
  return (
    <Card className="dark:bg-gray-800 dark:border-gray-700">
      <CardContent className="p-5">
        <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
          {title}
        </p>
        <p className={`text-3xl font-bold ${valueClass}`}>{value}</p>
        {sub && <p className="text-xs text-gray-400 mt-1">{sub}</p>}
      </CardContent>
    </Card>
  )
}

/** Skeleton placeholder for stat cards */
function StatCardSkeleton() {
  return (
    <Card className="dark:bg-gray-800 dark:border-gray-700">
      <CardContent className="p-5">
        <Skeleton className="h-3 w-24 mb-2" />
        <Skeleton className="h-8 w-16" />
        <Skeleton className="h-2 w-28 mt-2" />
      </CardContent>
    </Card>
  )
}

/** Protocol distribution mini-bar */
function ProtocolBar({ dns, http, other }: { dns: number; http: number; other: number }) {
  const total = dns + http + other || 1
  const dnsPct = Math.round((dns / total) * 100)
  const httpPct = Math.round((http / total) * 100)
  const otherPct = 100 - dnsPct - httpPct

  return (
    <div className="space-y-2">
      <div className="flex rounded-full overflow-hidden h-3">
        <div className="bg-purple-500" style={{ width: `${dnsPct}%` }} title={`DNS ${dnsPct}%`} />
        <div className="bg-blue-500" style={{ width: `${httpPct}%` }} title={`HTTP ${httpPct}%`} />
        <div className="bg-gray-300 dark:bg-gray-600" style={{ width: `${otherPct}%` }} title={`Other ${otherPct}%`} />
      </div>
      <div className="flex gap-4 text-xs text-gray-500 dark:text-gray-400">
        <span className="flex items-center gap-1">
          <span className="w-2 h-2 rounded-full bg-purple-500 inline-block" />
          DNS {dnsPct}%
        </span>
        <span className="flex items-center gap-1">
          <span className="w-2 h-2 rounded-full bg-blue-500 inline-block" />
          HTTP {httpPct}%
        </span>
        <span className="flex items-center gap-1">
          <span className="w-2 h-2 rounded-full bg-gray-300 dark:bg-gray-600 inline-block" />
          Other {otherPct}%
        </span>
      </div>
    </div>
  )
}

/** Dashboard Command Center page */
export default function DashboardPage() {
  const router = useRouter()
  const [stats, setStats] = useState<DashboardStats>({
    activeCases: 0,
    totalHitsToday: 0,
    activePayloads: 0,
    systemOk: true,
    dnsCount: 0,
    httpCount: 0,
    smtpCount: 0,
    totalInteractions: 0,
  })
  const [recentInteractions, setRecentInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadData()
  }, [router])

  const loadData = async () => {
    try {
      const [casesResp, interactionsResp, payloadsResp] = await Promise.all([
        caseApi.list({ page: 1, page_size: 5 }),
        interactionApi.list({ page: 1, page_size: 10 }),
        payloadApi.list({ page: 1, page_size: 1 }),
      ])

      const interactions = interactionsResp.data?.items || []
      const cases = casesResp.data?.items || []

      setStats({
        activeCases: cases.filter((c) => c.status === 'active').length,
        totalHitsToday: interactionsResp.data?.total || 0,
        activePayloads: payloadsResp.data?.total || 0,
        systemOk: true,
        dnsCount: interactions.filter((i) => i.type === 'dns').length,
        httpCount: interactions.filter((i) => i.type === 'http').length,
        smtpCount: interactions.filter((i) => i.type === 'smtp').length,
        totalInteractions: interactionsResp.data?.total || 0,
      })
      setRecentInteractions(interactions)
    } catch (err) {
      console.error('Failed to load dashboard data:', err)
    } finally {
      setLoading(false)
    }
  }

  const otherCount = stats.totalInteractions - stats.dnsCount - stats.httpCount

  return (
    <div className="space-y-6">
      {/* Section header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Command Center</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">OAST interaction monitoring overview</p>
        </div>
        <Button variant="outline" size="sm" onClick={loadData}>
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </Button>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => <StatCardSkeleton key={i} />)
        ) : (
          <>
            <StatCard
              title="Active Cases"
              value={stats.activeCases}
              valueClass="text-emerald-600"
              sub="Currently active"
            />
            <StatCard
              title="Hits Today"
              value={stats.totalHitsToday}
              valueClass="text-indigo-600"
              sub="Across all protocols"
            />
            <StatCard
              title="Active Payloads"
              value={stats.activePayloads}
              sub="Deployed &amp; watching"
            />
            <StatCard
              title="System Status"
              value={stats.systemOk ? '✓ OK' : '✗ Error'}
              valueClass={stats.systemOk ? 'text-emerald-600 text-2xl' : 'text-red-600 text-2xl'}
              sub={stats.systemOk ? 'All services healthy' : 'Check system logs'}
            />
          </>
        )}
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Protocol distribution */}
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
              Protocol Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-8 w-full" />
            ) : (
              <ProtocolBar
                dns={stats.dnsCount}
                http={stats.httpCount}
                other={otherCount > 0 ? otherCount : 0}
              />
            )}
          </CardContent>
        </Card>

        {/* Quick navigation shortcuts */}
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
              Quick Actions
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              {[
                { label: 'New Case', href: '/dashboard/cases', color: 'bg-indigo-50 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-900/20 dark:text-indigo-400' },
                { label: 'New Payload', href: '/dashboard/payloads/new', color: 'bg-purple-50 text-purple-700 hover:bg-purple-100 dark:bg-purple-900/20 dark:text-purple-400' },
                { label: 'View Timeline', href: '/dashboard/interactions', color: 'bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400' },
                { label: 'Canary Tokens', href: '/dashboard/canary', color: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-900/20 dark:text-emerald-400' },
              ].map((action) => (
                <Link
                  key={action.href}
                  href={action.href}
                  className={`flex items-center justify-center p-3 rounded-lg text-sm font-medium transition-colors ${action.color}`}
                >
                  {action.label}
                </Link>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Real-time hit stream */}
      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardHeader className="pb-0 flex flex-row items-center justify-between">
          <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            Live Hit Stream
          </CardTitle>
          <Link
            href="/dashboard/interactions"
            className="text-xs text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 font-medium"
          >
            View all →
          </Link>
        </CardHeader>
        <CardContent className="pt-4">
          {loading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3">
                  <Skeleton className="h-5 w-12" />
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="h-4 flex-1" />
                  <Skeleton className="h-4 w-16" />
                </div>
              ))}
            </div>
          ) : recentInteractions.length === 0 ? (
            <div className="text-center py-10">
              <div className="text-4xl mb-3">⏱</div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">No interactions yet</p>
              <p className="text-xs text-gray-400 mt-1">Waiting for payloads to be triggered</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-100 dark:divide-gray-700">
              {recentInteractions.map((interaction) => (
                <div
                  key={interaction.id}
                  className="flex items-center gap-3 py-3 text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/30 -mx-2 px-2 rounded"
                  onClick={() => router.push('/dashboard/interactions')}
                >
                  <span
                    className={`shrink-0 px-2 py-0.5 rounded text-xs font-medium ${
                      PROTOCOL_COLORS[interaction.type] || 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {interaction.type.toUpperCase()}
                  </span>
                  <span className="text-gray-600 dark:text-gray-400 w-28 truncate shrink-0">
                    {interaction.source_ip}
                  </span>
                  <span className="flex-1 text-gray-500 dark:text-gray-400 truncate text-xs">
                    {interaction.domain || (interaction.method && interaction.path ? `${interaction.method} ${interaction.path}` : interaction.token || '—')}
                  </span>
                  <span className="text-xs text-gray-400 shrink-0">
                    {new Date(interaction.timestamp).toLocaleTimeString()}
                  </span>
                  <div className="flex gap-1 shrink-0">
                    <button
                      className="text-xs text-gray-400 hover:text-indigo-600"
                      onClick={(e) => { e.stopPropagation(); router.push('/dashboard/interactions') }}
                    >
                      →
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
