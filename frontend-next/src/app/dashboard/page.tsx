'use client'

import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useCases } from '@/features/cases/hooks/use-cases'
import { usePayloads } from '@/features/payloads/hooks/use-payloads'
import { useInteractions, useInteractionStats } from '@/features/interactions/hooks/use-interactions'
import { useQuery } from '@tanstack/react-query'
import { interactionApi } from '@/lib/api-client'
import type { Interaction } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useI18n } from '@/lib/i18n-context'
import { PayloadCheatSheet } from '@/features/dashboard/payload-cheatsheet'
import { DonutChart, LineChart } from '@/components/charts'

/** Protocol color mapping per design spec */
const PROTOCOL_COLORS: Record<string, string> = {
  dns: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  http: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  smtp: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  ldap: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  smb: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  ftp: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
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
function LegendItem({ color, label, value }: { color: string; label: string; value: number }) {
  return (
    <div className="flex items-center gap-2">
      <span className="w-3 h-3 rounded-sm" style={{ backgroundColor: color }} />
      <span className="text-gray-600 dark:text-gray-400">{label}</span>
      <span className="ml-auto font-medium text-gray-900 dark:text-gray-100">{value}</span>
    </div>
  )
}

/** Dashboard Command Center page */
export default function DashboardPage() {
  const router = useRouter()
  const { t } = useI18n()

  const { data: casesResp, isLoading: casesLoading } = useCases({ page: 1, page_size: 5 })
  const { data: interactionsResp, isLoading: interactionsLoading } = useInteractions({ page: 1, page_size: 10 })
  const { data: statsResp, isLoading: statsLoading } = useInteractionStats()
  const { data: payloadsResp, isLoading: payloadsLoading } = usePayloads({ status: 'deployed', page: 1, page_size: 1 })
  const { data: dailyResp, isLoading: dailyLoading } = useQuery({
    queryKey: ['interactions', 'daily'],
    queryFn: () => interactionApi.dailyStats({ days: 7 }),
  })

  const loading = casesLoading || interactionsLoading || statsLoading || payloadsLoading || dailyLoading

  const cases = casesResp?.data?.items || []
  const recentInteractions: Interaction[] = interactionsResp?.data?.items || []
  const stats = statsResp?.data
  const dnsCount = stats?.dns_count ?? 0
  const httpCount = stats?.http_count ?? 0
  const smtpCount = stats?.smtp_count ?? 0
  const totalInteractions = stats?.total ?? 0
  const activeCases = cases.filter((c) => c.status === 'active').length
  const totalHitsToday = stats?.today ?? stats?.total ?? 0
  const activePayloads = payloadsResp?.data?.total || 0
  const systemOk = true

  const otherCount = totalInteractions - dnsCount - httpCount - smtpCount
  const dailyStats = dailyResp?.data ?? []

  const loadData = () => {
    window.location.reload()
  }

  return (
    <div className="space-y-6">
      {/* Section header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('dashboard.command_center')}</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('dashboard.subtitle')}</p>
        </div>
        <Button variant="outline" size="sm" onClick={loadData}>
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {t('common.refresh')}
        </Button>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => <StatCardSkeleton key={i} />)
        ) : (
          <>
            <StatCard
              title={t('dashboard.active_cases')}
              value={activeCases}
              valueClass="text-emerald-600"
              sub={t('dashboard.currently_active')}
            />
            <StatCard
              title={t('dashboard.hits_today')}
              value={totalHitsToday}
              valueClass="text-indigo-600"
              sub={t('dashboard.across_protocols')}
            />
            <StatCard
              title={t('dashboard.active_payloads')}
              value={activePayloads}
              sub={t('dashboard.deployed_watching')}
            />
            <StatCard
              title={t('dashboard.system_status')}
              value={systemOk ? t('dashboard.status_ok') : t('dashboard.status_error')}
              valueClass={systemOk ? 'text-emerald-600 text-2xl' : 'text-red-600 text-2xl'}
              sub={systemOk ? t('dashboard.all_services_healthy') : t('dashboard.check_system_logs')}
            />
          </>
        )}
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Protocol distribution */}
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
              {t('dashboard.protocol_distribution')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-8 w-full" />
            ) : (
              <div className="flex items-center gap-6">
                <DonutChart
                  size={180}
                  data={[
                    { name: 'DNS', value: dnsCount, color: '#8b5cf6' },
                    { name: 'HTTP', value: httpCount, color: '#3b82f6' },
                    { name: 'SMTP', value: smtpCount, color: '#10b981' },
                    { name: 'Other', value: Math.max(0, otherCount), color: '#6b7280' },
                  ]}
                />
                <div className="space-y-2 text-sm">
                  <LegendItem color="#8b5cf6" label="DNS" value={dnsCount} />
                  <LegendItem color="#3b82f6" label="HTTP" value={httpCount} />
                  <LegendItem color="#10b981" label="SMTP" value={smtpCount} />
                  <LegendItem color="#6b7280" label="Other" value={Math.max(0, otherCount)} />
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 7-day trend */}
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
              {t('dashboard.trend_7d')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-[260px] w-full" />
            ) : (
              <LineChart data={dailyStats} height={260} />
            )}
          </CardContent>
        </Card>

        {/* Quick navigation shortcuts */}
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
              {t('dashboard.quick_actions')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              {[
                { label: t('dashboard.new_case'), href: '/dashboard/cases', color: 'bg-indigo-50 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-900/20 dark:text-indigo-400' },
                { label: t('dashboard.new_payload'), href: '/dashboard/payloads/new', color: 'bg-purple-50 text-purple-700 hover:bg-purple-100 dark:bg-purple-900/20 dark:text-purple-400' },
                { label: t('dashboard.view_timeline'), href: '/dashboard/interactions', color: 'bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400' },
                { label: t('dashboard.canary_tokens'), href: '/dashboard/canary', color: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100 dark:bg-emerald-900/20 dark:text-emerald-400' },
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
            {t('dashboard.live_hit_stream')}
          </CardTitle>
          <Link
            href="/dashboard/interactions"
            className="text-xs text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 font-medium"
          >
            {t('dashboard.view_all')} →
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
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('dashboard.no_interactions')}</p>
              <p className="text-xs text-gray-400 mt-1">{t('dashboard.waiting_payloads')}</p>
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

      {/* Payload Cheat Sheet */}
      <PayloadCheatSheet />
    </div>
  )
}
