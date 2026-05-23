'use client'

/* eslint-disable react-hooks/set-state-in-effect */
import { useEffect, useState, Suspense, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { caseApi, evidenceApi } from '@/lib/api-client'
import type { Case, Evidence } from '@/types'

function EvidenceReportContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedCase, setSelectedCase] = useState<string>('')
  const [format, setFormat] = useState<'json' | 'markdown'>('markdown')
  const [generating, setGenerating] = useState(false)
  const [evidence, setEvidence] = useState<Evidence | null>(null)
  const [reportContent, setReportContent] = useState('')
  const [error, setError] = useState<string>('')

  // Get scope from URL
  const caseId = searchParams.get('case_id')
  const payloadId = searchParams.get('payload_id')
  const formatParam = searchParams.get('format')

  const loadCases = useCallback(async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
      setError('加载Case列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  const handleGenerateWithScope = useCallback(async (id: string, scope: 'case' | 'payload') => {
    setGenerating(true)
    setError('')
    setEvidence(null)
    setReportContent('')
    try {
      const params: { format: 'json' | 'markdown'; case_id?: string; payload_id?: string } = { format }
      if (scope === 'case') {
        params.case_id = id
      } else {
        params.payload_id = id
      }
      const response = await evidenceApi.generate(params)
      if (response.code === 0 && response.data) {
        setEvidence(response.data.evidence)
        setReportContent(response.data.content)
      } else {
        setError(response.message || '生成证据失败')
      }
    } catch (error: unknown) {
      console.error('Failed to generate evidence:', error)
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response?: { status?: number } }
        if (err.response?.status === 404) {
          setError('未找到该证据数据')
        } else {
          setError('生成证据失败: 未知错误')
        }
      } else {
        setError('生成证据失败: 未知错误')
      }
    } finally {
      setGenerating(false)
    }
  }, [format])

  // Set format from URL param if present
  useEffect(() => {
    if (formatParam === 'json' || formatParam === 'markdown') {
      setFormat(formatParam)
    }
  }, [formatParam])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadCases()
  }, [router, loadCases])

  // Auto-generate evidence if case_id or payload_id is in URL
  useEffect(() => {
    if (caseId && cases.length > 0) {
      setSelectedCase(caseId)
      handleGenerateWithScope(caseId, 'case')
    } else if (payloadId) {
      handleGenerateWithScope(payloadId, 'payload')
    }
  }, [caseId, payloadId, cases, handleGenerateWithScope])

  const handleGenerate = async () => {
    if (!selectedCase) return
    setGenerating(true)
    setError('')
    setEvidence(null)
    setReportContent('')
    try {
      const response = await evidenceApi.generate({
        case_id: selectedCase,
        format: format,
      })
      if (response.code === 0 && response.data) {
        setEvidence(response.data.evidence)
        setReportContent(response.data.content)
      } else {
        setError(response.message || '生成证据失败')
      }
    } catch (error: unknown) {
      console.error('Failed to generate evidence:', error)
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response?: { status?: number } }
        if (err.response?.status === 404) {
          setError('未找到该Case的证据数据')
        } else {
          setError('生成证据失败: 未知错误')
        }
      } else {
        setError('生成证据失败: 未知错误')
      }
    } finally {
      setGenerating(false)
    }
  }

  const handleDownload = () => {
    if (!reportContent) return
    const blob = new Blob([reportContent], { type: format === 'json' ? 'application/json' : 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `evidence-${selectedCase}.${format === 'json' ? 'json' : 'md'}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const getStrengthColor = (strength: string) => {
    switch (strength) {
      case 'critical':
        return 'bg-red-100 text-red-800'
      case 'high':
        return 'bg-orange-100 text-orange-800'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800'
      case 'low':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStrengthLabel = (strength: string) => {
    switch (strength) {
      case 'critical':
        return '严重'
      case 'high':
        return '高'
      case 'medium':
        return '中'
      case 'low':
        return '低'
      default:
        return strength
    }
  }

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">证据报告</h2>
          {caseId && (
            <p className="text-sm text-gray-500 mt-1">
              Case scoped: {caseId}
              <button
                onClick={() => router.push('/dashboard/evidence')}
                className="ml-2 text-indigo-600 hover:text-indigo-800"
              >
                Clear scope
              </button>
            </p>
          )}
          {payloadId && (
            <p className="text-sm text-gray-500 mt-1">
              Payload scoped: {payloadId}
              <button
                onClick={() => router.push('/dashboard/evidence')}
                className="ml-2 text-indigo-600 hover:text-indigo-800"
              >
                Clear scope
              </button>
            </p>
          )}
          {!caseId && !payloadId && (
            <p className="text-sm text-gray-500 mt-1">All Evidence</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left panel: Controls */}
        <div className="lg:col-span-1 bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">生成证据</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  选择Case
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={selectedCase}
                  onChange={(e) => setSelectedCase(e.target.value)}
                >
                  <option value="">选择Case</option>
                  {cases.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.title} ({c.status})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  报告格式
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={format}
                  onChange={(e) => setFormat(e.target.value as 'json' | 'markdown')}
                >
                  <option value="markdown">Markdown</option>
                  <option value="json">JSON</option>
                </select>
              </div>

              <div className="flex space-x-2">
                <button
                  onClick={handleGenerate}
                  disabled={!selectedCase || generating}
                  className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {generating ? '生成中...' : '生成证据'}
                </button>
                {reportContent && (
                  <button
                    onClick={handleDownload}
                    className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                  >
                    下载
                  </button>
                )}
              </div>

              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
                  {error}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right panel: Evidence display */}
        <div className="lg:col-span-2 space-y-6">
          {evidence ? (
            <>
              {/* Evidence Summary */}
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">证据摘要</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <div className="text-sm text-gray-500">证据强度</div>
                      <div className={`inline-block px-2 py-1 rounded text-sm font-medium mt-1 ${getStrengthColor(evidence.evidence_strength)}`}>
                        {getStrengthLabel(evidence.evidence_strength)}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">置信度</div>
                      <div className="text-2xl font-bold text-gray-900">{evidence.confidence}%</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">交互数量</div>
                      <div className="text-2xl font-bold text-gray-900">{evidence.interaction_count}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">唯一来源</div>
                      <div className="text-2xl font-bold text-gray-900">{evidence.unique_sources}</div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Explainability */}
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">可解释性</h3>
                  <p className="text-gray-700 whitespace-pre-wrap">{evidence.explainability}</p>
                </div>
              </div>

              {/* Timeline */}
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">时间线 ({evidence.timeline.length} 条交互)</h3>
                  <div className="space-y-3 max-h-96 overflow-y-auto">
                    {evidence.timeline.map((interaction) => (
                      <div key={interaction.id} className="p-3 bg-gray-50 rounded border border-gray-200">
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            <div className="flex items-center space-x-2">
                              <span className="inline-block px-2 py-1 text-xs font-medium rounded bg-blue-100 text-blue-800">
                                {interaction.type.toUpperCase()}
                              </span>
                              <span className="text-sm text-gray-500">{interaction.timestamp}</span>
                            </div>
                            <div className="mt-1 text-sm text-gray-700">
                              <span className="font-medium">来源:</span> {interaction.source_ip}
                              {interaction.domain && <span className="ml-2">| 域名: {interaction.domain}</span>}
                              {interaction.method && <span className="ml-2">| 方法: {interaction.method}</span>}
                              {interaction.path && <span className="ml-2">| 路径: {interaction.path}</span>}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Report Preview */}
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">报告预览</h3>
                  <pre className="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-96">{reportContent}</pre>
                </div>
              </div>
            </>
          ) : (
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6 text-center">
                <p className="text-gray-500 py-8">
                  {generating ? '生成中...' : 
                   caseId || payloadId ? '暂无证据数据' : 
                   selectedCase ? '选择Case后点击"生成证据"按钮' : '请先选择一个Case'}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function EvidenceReportPage() {
  return (
    <Suspense fallback={<div className="text-center py-12">加载中...</div>}>
      <EvidenceReportContent />
    </Suspense>
  )
}
