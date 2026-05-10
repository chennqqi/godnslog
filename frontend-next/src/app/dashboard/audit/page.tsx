'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { api } from '@/lib/api'

/** Audit log entry shape */
interface AuditEntry {
  id: string
  timestamp: string
  actor: string
  actor_ip: string
  action: string
  resource_type: string
  resource_id: string
  result: 'success' | 'failure' | 'denied'
  detail: string
}

/** Colour mapping for audit result badges */
const RESULT_STYLES: Record<string, string> = {
  success: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  failure: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  denied:  'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
}

/** Sentinel for unfiltered Radix Select rows (SelectItem must not use value="") */
const FILTER_ALL = 'all'

/** Colour mapping for action category badges */
const ACTION_STYLES: Record<string, string> = {
  auth:     'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  case:     'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
  payload:  'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  apikey:   'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
  user:     'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  settings: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
}

/** Returns an action category derived from the action string */
function actionCategory(action: string): string {
  if (action.startsWith('auth.')) return 'auth'
  if (action.startsWith('case.')) return 'case'
  if (action.startsWith('payload.')) return 'payload'
  if (action.startsWith('apikey.')) return 'apikey'
  if (action.startsWith('user.')) return 'user'
  if (action.startsWith('settings.')) return 'settings'
  return 'settings'
}

/** Skeleton row for loading state */
function AuditRowSkeleton() {
  return (
    <tr className="animate-pulse">
      {[1, 2, 3, 4, 5, 6].map((i) => (
        <td key={i} className="px-4 py-3">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full" />
        </td>
      ))}
    </tr>
  )
}

/** Audit Log page — displays a filterable table of system-level audit events */
export default function AuditPage() {
  const router = useRouter()
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [searchActor, setSearchActor] = useState('')
  const [resultFilter, setResultFilter] = useState(FILTER_ALL)
  const [categoryFilter, setCategoryFilter] = useState(FILTER_ALL)

  const loadAudit = useCallback(async () => {
    setLoading(true)
    try {
      const resp = await api.get<{ items: AuditEntry[]; total: number }>('/audit', {
        page: 1,
        page_size: 100,
      })
      setEntries(resp.data?.items ?? [])
    } catch (err) {
      console.error('Failed to load audit log:', err)
      // Fall back to empty list so the page remains usable
      setEntries([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadAudit()
  }, [router, loadAudit])

  const filtered = entries.filter((e) => {
    if (searchActor && !e.actor.toLowerCase().includes(searchActor.toLowerCase())) return false
    if (resultFilter !== FILTER_ALL && e.result !== resultFilter) return false
    if (categoryFilter !== FILTER_ALL && actionCategory(e.action) !== categoryFilter) return false
    return true
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Audit Log</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            System-level activity trail for security and compliance review
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={loadAudit}>
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardContent className="p-4">
          <div className="flex flex-wrap gap-3">
            <Input
              placeholder="Filter by actor / user..."
              className="w-56"
              value={searchActor}
              onChange={(e) => setSearchActor(e.target.value)}
            />
            <Select value={resultFilter} onValueChange={setResultFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="All results" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={FILTER_ALL}>All results</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failure">Failure</SelectItem>
                <SelectItem value="denied">Denied</SelectItem>
              </SelectContent>
            </Select>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-44">
                <SelectValue placeholder="All categories" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={FILTER_ALL}>All categories</SelectItem>
                <SelectItem value="auth">Auth</SelectItem>
                <SelectItem value="case">Case</SelectItem>
                <SelectItem value="payload">Payload</SelectItem>
                <SelectItem value="apikey">API Key</SelectItem>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="settings">Settings</SelectItem>
              </SelectContent>
            </Select>
            {(searchActor || resultFilter !== FILTER_ALL || categoryFilter !== FILTER_ALL) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSearchActor('')
                  setResultFilter(FILTER_ALL)
                  setCategoryFilter(FILTER_ALL)
                }}
              >
                Clear filters
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="dark:bg-gray-800 dark:border-gray-700 overflow-hidden">
        <CardHeader className="pb-0 border-b border-gray-100 dark:border-gray-700">
          <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
            {loading ? 'Loading events...' : `${filtered.length} event${filtered.length !== 1 ? 's' : ''}`}
          </CardTitle>
        </CardHeader>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 dark:bg-gray-900/40 text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <th className="px-4 py-3 text-left">Timestamp</th>
                <th className="px-4 py-3 text-left">Actor</th>
                <th className="px-4 py-3 text-left">IP</th>
                <th className="px-4 py-3 text-left">Action</th>
                <th className="px-4 py-3 text-left">Resource</th>
                <th className="px-4 py-3 text-left">Result</th>
                <th className="px-4 py-3 text-left">Detail</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {loading ? (
                Array.from({ length: 8 }).map((_, i) => <AuditRowSkeleton key={i} />)
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-16 text-center">
                    <div className="text-4xl mb-3">📋</div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">No audit events found</p>
                    <p className="text-xs text-gray-400 mt-1">
                      {entries.length === 0
                        ? 'Audit events will appear here once the backend audit endpoint is available'
                        : 'Try adjusting your filters'}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((entry) => {
                  const cat = actionCategory(entry.action)
                  return (
                    <tr
                      key={entry.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                    >
                      <td className="px-4 py-3 text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap font-mono">
                        {new Date(entry.timestamp).toLocaleString()}
                      </td>
                      <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">
                        {entry.actor}
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs">
                        {entry.actor_ip}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${ACTION_STYLES[cat]}`}>
                          {entry.action}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs truncate max-w-[8rem]">
                        {entry.resource_type}{entry.resource_id ? ` / ${entry.resource_id}` : ''}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${RESULT_STYLES[entry.result]}`}>
                          {entry.result}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs truncate max-w-[12rem]" title={entry.detail}>
                        {entry.detail}
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}
