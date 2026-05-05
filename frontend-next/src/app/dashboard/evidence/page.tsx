'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi, interactionApi } from '@/lib/api-client'
import type { Case } from '@/types'

export default function EvidenceReportPage() {
  const router = useRouter()
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedCase, setSelectedCase] = useState<string>('')
  const [format, setFormat] = useState('markdown')
  const [includeRaw, setIncludeRaw] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [reportContent, setReportContent] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadCases()
  }, [router])

  const loadCases = async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleGenerate = async () => {
    if (!selectedCase) return
    setGenerating(true)
    try {
      const response = await interactionApi.export({
        case_id: selectedCase,
        format: format,
        include_raw: includeRaw,
      })
      if (response.code === 0 && response.data) {
        setReportContent(String(response.data))
      }
    } catch (error) {
      console.error('Failed to generate report:', error)
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

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">证据报告</h2>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">生成报告</h3>

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
                  onChange={(e) => setFormat(e.target.value)}
                >
                  <option value="markdown">Markdown</option>
                  <option value="json">JSON</option>
                  <option value="csv">CSV</option>
                </select>
              </div>

              <div>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeRaw}
                    onChange={(e) => setIncludeRaw(e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">包含原始数据</span>
                </label>
              </div>

              <div className="flex space-x-2">
                <button
                  onClick={handleGenerate}
                  disabled={!selectedCase || generating}
                  className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {generating ? '生成中...' : '生成报告'}
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
            </div>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">报告预览</h3>
            {reportContent ? (
              <pre className="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-96">{reportContent}</pre>
            ) : (
              <p className="text-gray-500 text-center py-4">生成报告后在此预览</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
