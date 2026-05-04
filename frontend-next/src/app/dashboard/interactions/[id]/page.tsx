'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { interactionApi } from '@/lib/api-client'
import type { Interaction } from '@/types'

export default function InteractionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [interaction, setInteraction] = useState<Interaction | null>(null)
  const [loading, setLoading] = useState(true)
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (params.id) {
      loadInteraction()
    }
  }, [params.id])

  const loadInteraction = async () => {
    try {
      const response = await interactionApi.get(params.id as string)
      // Handle nested response structure
      const interactionData = response.data && 'data' in response.data ? response.data.data : response.data
      if (interactionData) {
        setInteraction(interactionData as Interaction)
      }
    } catch (error) {
      console.error('Failed to load interaction:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleExport = async (format: string) => {
    setExporting(true)
    try {
      const response = await interactionApi.export({
        ids: [params.id],
        format,
        include_raw: true,
      })
      // Handle nested response structure
      const responseData = response.data as any
      const exportData = responseData && typeof responseData === 'object' && 'data' in responseData ? responseData.data : responseData
      if (exportData) {
        const blob = new Blob([typeof exportData === 'string' ? exportData : JSON.stringify(exportData)], { type: 'text/plain' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `interaction_${params.id}.${format}`
        a.click()
        URL.revokeObjectURL(url)
      }
    } catch (error) {
      console.error('Failed to export:', error)
    } finally {
      setExporting(false)
    }
  }

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  if (!interaction) {
    return <div className="text-center py-12">Interaction 不存在</div>
  }

  return (
    <div>
      <button
        onClick={() => router.back()}
        className="mb-4 text-indigo-600 hover:text-indigo-800"
      >
        ← 返回
      </button>

      <div className="bg-white shadow rounded-lg mb-6">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-start mb-4">
            <h2 className="text-2xl font-bold text-gray-900">
              {interaction.type.toUpperCase()} Interaction
            </h2>
            <div className="flex space-x-2">
              <button
                onClick={() => handleExport('json')}
                disabled={exporting}
                className="px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm disabled:opacity-50"
              >
                {exporting ? '导出中...' : '导出JSON'}
              </button>
              <button
                onClick={() => handleExport('markdown')}
                disabled={exporting}
                className="px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm disabled:opacity-50"
              >
                {exporting ? '导出中...' : '导出Markdown'}
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-500">ID</p>
              <p className="text-sm font-medium">{interaction.id}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">时间戳</p>
              <p className="text-sm font-medium">{new Date(interaction.timestamp).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">来源IP</p>
              <p className="text-sm font-medium">{interaction.source_ip}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">类型</p>
              <p className="text-sm font-medium">{interaction.type}</p>
            </div>
            {interaction.token && (
              <div>
                <p className="text-sm text-gray-500">Token</p>
                <p className="text-sm font-medium">{interaction.token}</p>
              </div>
            )}
            {interaction.domain && (
              <div>
                <p className="text-sm text-gray-500">域名</p>
                <p className="text-sm font-medium">{interaction.domain}</p>
              </div>
            )}
            {interaction.method && (
              <div>
                <p className="text-sm text-gray-500">方法</p>
                <p className="text-sm font-medium">{interaction.method}</p>
              </div>
            )}
            {interaction.path && (
              <div>
                <p className="text-sm text-gray-500">路径</p>
                <p className="text-sm font-medium break-all">{interaction.path}</p>
              </div>
            )}
            {interaction.content_type && (
              <div>
                <p className="text-sm text-gray-500">Content-Type</p>
                <p className="text-sm font-medium">{interaction.content_type}</p>
              </div>
            )}
            {interaction.user_agent && (
              <div className="col-span-2">
                <p className="text-sm text-gray-500">User-Agent</p>
                <p className="text-sm font-medium break-all">{interaction.user_agent}</p>
              </div>
            )}
          </div>

          {interaction.headers && Object.keys(interaction.headers).length > 0 && (
            <div className="mt-6">
              <h3 className="text-lg font-medium text-gray-900 mb-2">Headers</h3>
              <pre className="bg-gray-50 p-4 rounded text-sm overflow-auto">
                {JSON.stringify(interaction.headers, null, 2)}
              </pre>
            </div>
          )}

          {interaction.body && (
            <div className="mt-6">
              <h3 className="text-lg font-medium text-gray-900 mb-2">Body</h3>
              <pre className="bg-gray-50 p-4 rounded text-sm overflow-auto max-h-96">
                {interaction.body}
              </pre>
            </div>
          )}

          {interaction.raw_data && (
            <div className="mt-6">
              <h3 className="text-lg font-medium text-gray-900 mb-2">Raw Data</h3>
              <pre className="bg-gray-50 p-4 rounded text-sm overflow-auto max-h-96">
                {interaction.raw_data}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
