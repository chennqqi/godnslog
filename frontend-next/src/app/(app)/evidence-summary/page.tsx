'use client'

/* eslint-disable react-hooks/set-state-in-effect */
import { useEffect, useState, Suspense, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { LoadingState } from '@/components/loading-state'
import { caseApi, evidenceApi, scannerRunApi } from '@/lib/api-client'
import type { Case, EvidenceSummaryResponse, ScannerRun } from '@/types'

function EvidenceSummaryContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [cases, setCases] = useState<Case[]>([])
  const [scannerRuns, setScannerRuns] = useState<ScannerRun[]>([])
  const [loading, setLoading] = useState(true)
  const [querying, setQuerying] = useState(false)
  const [summary, setSummary] = useState<EvidenceSummaryResponse | null>(null)
  const [error, setError] = useState<string>('')

  const [scopeType, setScopeType] = useState<'case' | 'payload' | 'scanner_run'>('case')
  const [caseId, setCaseId] = useState<string>('')
  const [payloadId, setPayloadId] = useState<string>('')
  const [scannerRunId, setScannerRunId] = useState<string>('')

  const urlCaseId = searchParams.get('case_id')
  const urlPayloadId = searchParams.get('payload_id')
  const urlScannerRunId = searchParams.get('scanner_run_id')

  const loadCases = useCallback(async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load cases:', err)
      setError('Failed to load cases')
    } finally {
      setLoading(false)
    }
  }, [])

  const loadScannerRuns = useCallback(async () => {
    try {
      const response = await scannerRunApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setScannerRuns(response.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load scanner runs:', err)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      window.location.href = '/login'
      return
    }
    loadCases()
    loadScannerRuns()
  }, [router, loadCases, loadScannerRuns])

  const handleQuery = useCallback(async () => {
    setQuerying(true)
    setError('')
    setSummary(null)
    try {
      const params: Record<string, string> = {}
      if (scopeType === 'case' && caseId) params.case_id = caseId
      else if (scopeType === 'payload' && payloadId) params.payload_id = payloadId
      else if (scopeType === 'scanner_run' && scannerRunId) params.scanner_run_id = scannerRunId

      if (Object.keys(params).length === 0) {
        setError('Please select or enter a scope identifier')
        return
      }

      const response = await evidenceApi.summary(params)
      if (response.code === 0 && response.data) {
        setSummary(response.data)
      } else {
        setError(response.message || 'Failed to fetch evidence summary')
      }
    } catch (err: unknown) {
      console.error('Failed to fetch evidence summary:', err)
      if (err && typeof err === 'object' && 'response' in err) {
        const e = err as { response?: { status?: number } }
        if (e.response?.status === 404) {
          setError('No evidence data found for the given scope')
        } else {
          setError('Failed to fetch evidence summary')
        }
      } else {
        setError('Failed to fetch evidence summary')
      }
    } finally {
      setQuerying(false)
    }
  }, [scopeType, caseId, payloadId, scannerRunId])

  // Auto-query from URL params
  useEffect(() => {
    if (urlCaseId && !loading) {
      setScopeType('case')
      setCaseId(urlCaseId)
      handleQuery()
    } else if (urlPayloadId && !loading) {
      setScopeType('payload')
      setPayloadId(urlPayloadId)
      handleQuery()
    } else if (urlScannerRunId && !loading) {
      setScopeType('scanner_run')
      setScannerRunId(urlScannerRunId)
      handleQuery()
    }
  }, [urlCaseId, urlPayloadId, urlScannerRunId, loading, handleQuery])

  const getStrengthColor = (strength?: string) => {
    switch (strength) {
      case 'critical':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
      case 'high':
        return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      case 'low':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
    }
  }

  if (loading) {
    return <LoadingState />
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Evidence Summary</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Structured evidence bundle for agents and CI
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left panel: Query controls */}
        <div className="lg:col-span-1 bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Query Scope</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Scope Type
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={scopeType}
                  onChange={(e) => setScopeType(e.target.value as 'case' | 'payload' | 'scanner_run')}
                >
                  <option value="case">By Case</option>
                  <option value="payload">By Payload</option>
                  <option value="scanner_run">By Scanner Run</option>
                </select>
              </div>

              {scopeType === 'case' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Select Case
                  </label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    value={caseId}
                    onChange={(e) => setCaseId(e.target.value)}
                  >
                    <option value="">Select a case</option>
                    {cases.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.title} ({c.status})
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {scopeType === 'payload' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Payload ID
                  </label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    placeholder="Enter payload ID"
                    value={payloadId}
                    onChange={(e) => setPayloadId(e.target.value)}
                  />
                </div>
              )}

              {scopeType === 'scanner_run' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Select Scanner Run
                  </label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    value={scannerRunId}
                    onChange={(e) => setScannerRunId(e.target.value)}
                  >
                    <option value="">Select a scanner run</option>
                    {scannerRuns.map((sr) => (
                      <option key={sr.id} value={sr.id}>
                        {sr.scanner} — {sr.status} ({sr.id.slice(0, 8)})
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <button
                onClick={handleQuery}
                disabled={querying}
                className="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
              >
                {querying ? 'Querying...' : 'Get Evidence Summary'}
              </button>

              {error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded text-red-700 dark:text-red-300 text-sm">
                  {error}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right panel: Summary display */}
        <div className="lg:col-span-2 space-y-6">
          {summary ? (
            <>
              {/* Scope & Hash */}
              <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Summary Scope</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">Source</div>
                      <div className="text-sm font-medium text-gray-900 dark:text-gray-200 mt-1">
                        {summary.scope.source}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">Summary Hash</div>
                      <div className="text-sm font-mono text-gray-900 dark:text-gray-200 mt-1 break-all">
                        {summary.summary_hash}
                      </div>
                    </div>
                    {summary.scope.case_id && (
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Case ID</div>
                        <div className="text-sm font-mono text-gray-900 dark:text-gray-200 mt-1">
                          {summary.scope.case_id}
                        </div>
                      </div>
                    )}
                    {summary.scope.payload_id && (
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Payload ID</div>
                        <div className="text-sm font-mono text-gray-900 dark:text-gray-200 mt-1">
                          {summary.scope.payload_id}
                        </div>
                      </div>
                    )}
                    {summary.scope.scanner_run_id && (
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Scanner Run ID</div>
                        <div className="text-sm font-mono text-gray-900 dark:text-gray-200 mt-1">
                          {summary.scope.scanner_run_id}
                        </div>
                      </div>
                    )}
                    <div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">Generated At</div>
                      <div className="text-sm text-gray-900 dark:text-gray-200 mt-1">
                        {new Date(summary.generated_at).toLocaleString()}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Evidence Stats */}
              {summary.evidence && (
                <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Evidence</h3>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Strength</div>
                        <div className={`inline-block px-2 py-1 rounded text-sm font-medium mt-1 ${getStrengthColor(summary.evidence.evidence_strength)}`}>
                          {summary.evidence.evidence_strength}
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Confidence</div>
                        <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                          {summary.evidence.confidence}%
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Interactions</div>
                        <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                          {summary.evidence.interaction_count}
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">Unique Sources</div>
                        <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                          {summary.evidence.unique_sources}
                        </div>
                      </div>
                    </div>
                    {summary.evidence.explainability && (
                      <div className="mt-4">
                        <div className="text-sm text-gray-500 dark:text-gray-400 mb-1">Explainability</div>
                        <p className="text-gray-700 dark:text-gray-300 text-sm whitespace-pre-wrap">
                          {summary.evidence.explainability}
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Scanner Runs */}
              {summary.scanner_runs && summary.scanner_runs.length > 0 && (
                <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                      Scanner Runs ({summary.scanner_runs.length})
                    </h3>
                    <div className="space-y-2">
                      {summary.scanner_runs.map((run) => (
                        <div key={run.id} className="p-3 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600">
                          <div className="flex items-center justify-between">
                            <span className="text-sm font-medium text-gray-900 dark:text-gray-200">
                              {run.scanner}
                            </span>
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              {run.status}
                            </span>
                          </div>
                          <div className="text-xs font-mono text-gray-500 dark:text-gray-400 mt-1">
                            {run.id}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Package Hashes */}
              {summary.package_hashes && summary.package_hashes.length > 0 && (
                <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                      Package Hashes ({summary.package_hashes.length})
                    </h3>
                    <div className="space-y-1">
                      {summary.package_hashes.map((hash) => (
                        <div key={hash} className="text-sm font-mono text-gray-700 dark:text-gray-300 break-all">
                          {hash}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Next Actions */}
              {summary.next_actions && summary.next_actions.length > 0 && (
                <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Next Actions</h3>
                    <ul className="space-y-2">
                      {summary.next_actions.map((action, idx) => (
                        <li key={idx} className="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300">
                          <span className="text-indigo-600 dark:text-indigo-400 mt-0.5">→</span>
                          <span>{action}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              )}

              {/* Metadata */}
              <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Metadata</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">Scanner Run Count</div>
                      <div className="text-sm font-medium text-gray-900 dark:text-gray-200 mt-1">
                        {summary.metadata.scanner_run_count}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">Package Count</div>
                      <div className="text-sm font-medium text-gray-900 dark:text-gray-200 mt-1">
                        {summary.metadata.package_count}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6 text-center">
                <p className="text-gray-500 dark:text-gray-400 py-8">
                  {querying ? 'Querying...' : 'Select a scope and click "Get Evidence Summary"'}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function EvidenceSummaryPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <EvidenceSummaryContent />
    </Suspense>
  )
}
