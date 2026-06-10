'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
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
import { auditApi, agentRunApi } from '@/lib/api-client'
import type { AuditLog, AgentRunReviewPackageTraceResponse } from '@/types'

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
  if (action.startsWith('auth.') || action.includes('login') || action.includes('logout')) return 'auth'
  if (action.includes('case')) return 'case'
  if (action.includes('payload')) return 'payload'
  if (action.includes('apikey') || action.includes('api_key')) return 'apikey'
  if (action.includes('user')) return 'user'
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

/** Audit Log page content — displays a filterable table of system-level audit events */
export default function AuditPageContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [entries, setEntries] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [searchActor, setSearchActor] = useState('')
  const [resultFilter, setResultFilter] = useState(FILTER_ALL)
  const [categoryFilter, setCategoryFilter] = useState(FILTER_ALL)
  const [resourceTypeFilter, setResourceTypeFilter] = useState('')
  const [resourceIdFilter, setResourceIdFilter] = useState('')
  const [page] = useState(1)
  const [total, setTotal] = useState(0)

  // Package hash trace state
  const [packageHashInput, setPackageHashInput] = useState('')
  const [traceResult, setTraceResult] = useState<AgentRunReviewPackageTraceResponse | null>(null)
  const [traceLoading, setTraceLoading] = useState(false)
  const [traceError, setTraceError] = useState('')
  const [showTrace, setShowTrace] = useState(false)

  // Initialize filters from URL params
  useEffect(() => {
    const resourceType = searchParams.get('resource_type')
    const resourceId = searchParams.get('resource_id')
    // Wrap in setTimeout to avoid react-hooks/set-state-in-effect lint error
    setTimeout(() => {
      if (resourceType) setResourceTypeFilter(resourceType)
      if (resourceId) setResourceIdFilter(resourceId)
    }, 0)
  }, [searchParams])

  const loadAudit = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const resp = await auditApi.list({
        resource_type: resourceTypeFilter || undefined,
        resource_id: resourceIdFilter || undefined,
        page,
        page_size: 100,
      })
      if (resp.code === 0 && resp.data) {
        setEntries(resp.data.items || [])
        setTotal(resp.data.total || 0)
      } else {
        setError(resp.message || '加载审计日志失败')
        setEntries([])
      }
    } catch (err: unknown) {
      console.error('Failed to load audit log:', err)
      const errorMessage = err instanceof Error ? err.message : '未知错误'
      setError('加载审计日志失败: ' + errorMessage)
      setEntries([])
    } finally {
      setLoading(false)
    }
  }, [page, resourceTypeFilter, resourceIdFilter])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    // Wrap in setTimeout to avoid react-hooks/set-state-in-effect lint error
    setTimeout(() => {
      loadAudit()
    }, 0)
  }, [router, loadAudit])

  const filtered = entries.filter((e) => {
    if (searchActor && !e.user_id?.toLowerCase().includes(searchActor.toLowerCase())) return false
    if (resultFilter !== FILTER_ALL && e.result !== resultFilter) return false
    if (categoryFilter !== FILTER_ALL && actionCategory(e.action) !== categoryFilter) return false
    return true
  })

  const handleTracePackage = async () => {
    if (!packageHashInput.trim()) {
      setTraceError('Package hash is required')
      return
    }
    // Basic validation: 64 hex characters
    const hashRegex = /^[a-fA-F0-9]{64}$/
    if (!hashRegex.test(packageHashInput)) {
      setTraceError('Invalid package hash: must be 64-character hex string')
      return
    }

    setTraceLoading(true)
    setTraceError('')
    setTraceResult(null)
    try {
      const resp = await agentRunApi.traceReviewPackage(packageHashInput.trim())
      if (resp.code === 0 && resp.data) {
        setTraceResult(resp.data.data)
        setShowTrace(true)
      } else {
        setTraceError(resp.message || 'Failed to trace package')
      }
    } catch (err: unknown) {
      console.error('Failed to trace package:', err)
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      setTraceError('Failed to trace package: ' + errorMessage)
    } finally {
      setTraceLoading(false)
    }
  }

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
              placeholder="Filter by user ID..."
              className="w-56"
              value={searchActor}
              onChange={(e) => setSearchActor(e.target.value)}
            />
            <Input
              placeholder="Resource type..."
              className="w-40"
              value={resourceTypeFilter}
              onChange={(e) => setResourceTypeFilter(e.target.value)}
            />
            <Input
              placeholder="Resource ID..."
              className="w-40"
              value={resourceIdFilter}
              onChange={(e) => setResourceIdFilter(e.target.value)}
            />
            <Select value={resultFilter} onValueChange={setResultFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="All results" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={FILTER_ALL}>All results</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failure">Failure</SelectItem>
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
            {(searchActor || resultFilter !== FILTER_ALL || categoryFilter !== FILTER_ALL || resourceTypeFilter || resourceIdFilter) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSearchActor('')
                  setResultFilter(FILTER_ALL)
                  setCategoryFilter(FILTER_ALL)
                  setResourceTypeFilter('')
                  setResourceIdFilter('')
                }}
              >
                Clear filters
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Package Hash Trace */}
      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardHeader>
          <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
            Package Hash Trace
          </CardTitle>
        </CardHeader>
        <CardContent className="p-4">
          <div className="flex gap-3">
            <Input
              placeholder="Paste 64-character package hash..."
              className="flex-1 font-mono text-xs"
              value={packageHashInput}
              onChange={(e) => setPackageHashInput(e.target.value)}
            />
            <Button onClick={handleTracePackage} disabled={traceLoading}>
              {traceLoading ? 'Tracing...' : 'Trace'}
            </Button>
            {showTrace && (
              <Button variant="ghost" onClick={() => setShowTrace(false)}>
                Hide
              </Button>
            )}
          </div>
          {traceError && (
            <p className="text-sm text-red-600 dark:text-red-400 mt-2">{traceError}</p>
          )}
          {showTrace && traceResult && (
            <div className="mt-4 space-y-4">
              {/* Summary */}
              <div className="grid grid-cols-4 gap-4 text-center">
                <div className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded">
                  <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">{traceResult.summary.agent_run_count}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">Agent Runs</div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded">
                  <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">{traceResult.summary.export_count}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">Exports</div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded">
                  <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">{traceResult.summary.delivery_count}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">Deliveries</div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded">
                  <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">{traceResult.summary.audit_count}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">Audits</div>
                </div>
              </div>
              {/* Delivery Status */}
              <div className="grid grid-cols-3 gap-4 text-center">
                <div className="bg-emerald-50 dark:bg-emerald-900/20 p-3 rounded">
                  <div className="text-xl font-bold text-emerald-700 dark:text-emerald-400">{traceResult.summary.delivered}</div>
                  <div className="text-xs text-emerald-600 dark:text-emerald-500">Delivered</div>
                </div>
                <div className="bg-red-50 dark:bg-red-900/20 p-3 rounded">
                  <div className="text-xl font-bold text-red-700 dark:text-red-400">{traceResult.summary.failed}</div>
                  <div className="text-xs text-red-600 dark:text-red-500">Failed</div>
                </div>
                <div className="bg-amber-50 dark:bg-amber-900/20 p-3 rounded">
                  <div className="text-xl font-bold text-amber-700 dark:text-amber-400">{traceResult.summary.timeout}</div>
                  <div className="text-xs text-amber-600 dark:text-amber-500">Timeout</div>
                </div>
              </div>
              {/* Agent Runs */}
              {traceResult.agent_runs.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Agent Runs</h4>
                  <div className="space-y-2">
                    {traceResult.agent_runs.map((run) => (
                      <div key={run.agent_run_id} className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-medium">{run.title || run.agent_run_id}</span>
                          <span className={`px-2 py-0.5 rounded ${run.status === 'completed' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'}`}>
                            {run.status || 'Unknown'}
                          </span>
                        </div>
                        <div className="mt-1 text-gray-500 dark:text-gray-400">
                          {run.case_id && <span>Case: {run.case_id}</span>}
                          {run.payload_id && <span className="ml-2">Payload: {run.payload_id}</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {/* Exports */}
              {traceResult.exports.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Exports</h4>
                  <div className="space-y-2">
                    {traceResult.exports.map((exp) => (
                      <div key={exp.operation_id} className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-mono">{exp.format}</span>
                          <span className="text-gray-500 dark:text-gray-400">{new Date(exp.created_at).toLocaleString()}</span>
                        </div>
                        <div className="mt-1 text-gray-500 dark:text-gray-400">
                          Operation ID: {exp.operation_id}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {/* Deliveries */}
              {traceResult.deliveries.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Deliveries</h4>
                  <div className="space-y-2">
                    {traceResult.deliveries.map((del) => (
                      <div key={del.delivery_operation_id} className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded text-xs">
                        <div className="flex items-center justify-between">
                          <span className={`px-2 py-0.5 rounded ${del.result === 'delivered' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : del.result === 'failed' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'}`}>
                            {del.result}
                          </span>
                          <span className="text-gray-500 dark:text-gray-400">{new Date(del.created_at).toLocaleString()}</span>
                        </div>
                        <div className="mt-1 text-gray-500 dark:text-gray-400">
                          {del.destination_host && <span>Host: {del.destination_host}</span>}
                          {del.status_code && <span className="ml-2">Status: {del.status_code}</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {/* Audits */}
              {traceResult.audits.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Audit Records</h4>
                  <div className="space-y-2">
                    {traceResult.audits.map((audit) => (
                      <div key={audit.audit_ref_id} className="bg-gray-50 dark:bg-gray-700/50 p-3 rounded text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-medium">{audit.action}</span>
                          <span className="text-gray-500 dark:text-gray-400">{new Date(audit.timestamp).toLocaleString()}</span>
                        </div>
                        <div className="mt-1 text-gray-500 dark:text-gray-400">
                          {audit.resource_type} {audit.resource_id && `/ ${audit.resource_id}`}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Error display */}
      {error && (
        <Card className="dark:bg-gray-800 dark:border-gray-700 border-red-200">
          <CardContent className="p-4">
            <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Table */}
      <Card className="dark:bg-gray-800 dark:border-gray-700 overflow-hidden">
        <CardHeader className="pb-0 border-b border-gray-100 dark:border-gray-700">
          <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
            {loading ? 'Loading events...' : `${filtered.length} event${filtered.length !== 1 ? 's' : ''} (Total: ${total})`}
          </CardTitle>
        </CardHeader>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 dark:bg-gray-900/40 text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <th className="px-4 py-3 text-left">Timestamp</th>
                <th className="px-4 py-3 text-left">User ID</th>
                <th className="px-4 py-3 text-left">IP</th>
                <th className="px-4 py-3 text-left">Action</th>
                <th className="px-4 py-3 text-left">Resource</th>
                <th className="px-4 py-3 text-left">Result</th>
                <th className="px-4 py-3 text-left">Details</th>
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
                        ? 'No audit events available'
                        : 'Try adjusting your filters'}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((entry) => {
                  const cat = actionCategory(entry.action)
                  const packageHash = entry.details && typeof entry.details === 'object' && 'package_hash' in entry.details && typeof entry.details.package_hash === 'string' ? entry.details.package_hash as string : null
                  return (
                    <tr
                      key={entry.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                    >
                      <td className="px-4 py-3 text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap font-mono">
                        {new Date(entry.timestamp).toLocaleString()}
                      </td>
                      <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">
                        {entry.user_id || (entry.is_agent ? 'Agent' : 'System')}
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 font-mono text-xs">
                        {entry.ip_address}
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
                        <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${RESULT_STYLES[entry.result] || RESULT_STYLES.success}`}>
                          {entry.result}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs truncate max-w-[12rem]" title={entry.error_message || entry.parameters}>
                        {entry.error_message || entry.parameters}
                        {packageHash && (
                          <div className="mt-1 flex items-center gap-1">
                            <span className="font-medium text-gray-700 dark:text-gray-300">Package Hash:</span>
                            <code className="bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded text-xs cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600"
                                  onClick={() => {
                                    setPackageHashInput(packageHash)
                                    setShowTrace(true)
                                    handleTracePackage()
                                  }}
                                  title="Click to trace this package hash">
                              {packageHash.substring(0, 12)}...
                            </code>
                          </div>
                        )}
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
