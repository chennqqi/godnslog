'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { payloadApi } from '@/lib/api-client'
import type { Payload, PayloadCreateRequest } from '@/types'

const templates = [
  { id: 'ssrf-http', name: 'SSRF HTTP', template: '{{.token}}.{{.domain}}' },
  { id: 'ssrf-cloud', name: 'SSRF Cloud Metadata', template: '{{.token}}.169.254.169.254.{{.domain}}' },
  { id: 'xxe', name: 'XXE External Entity', template: 'http://{{.token}}.{{.domain}}/xxe.dtd' },
  { id: 'rce', name: 'RCE Command Injection', template: 'curl http://{{.token}}.{{.domain}}' },
  { id: 'blind-sqli', name: 'Blind SQLi DNS', template: '{{.token}}.{{.domain}}' },
]

export default function PayloadsPage() {
  const router = useRouter()
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showBatchModal, setShowBatchModal] = useState(false)
  const [previewPayload, setPreviewPayload] = useState('')
  const [selectedTemplate, setSelectedTemplate] = useState(templates[0])
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [batchCount, setBatchCount] = useState(1)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadPayloads()
  }, [router])

  const loadPayloads = async () => {
    try {
      const response = await payloadApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setPayloads(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load payloads:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreatePayload = async (e: React.FormEvent) => {
    e.preventDefault()
    const req: PayloadCreateRequest = {
      case_id: '',
      template: selectedTemplate.id,
      variables,
    }
    try {
      const response = await payloadApi.create(req)
      if (response.code === 0) {
        setShowCreateModal(false)
        setVariables({})
        loadPayloads()
      }
    } catch (error) {
      console.error('Failed to create payload:', error)
    }
  }

  const handleBatchCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    // Batch creation logic would go here
    alert('批量生成功能需要后端 API 支持')
  }

  const updatePreview = () => {
    let preview = selectedTemplate.template
    Object.entries(variables).forEach(([key, value]) => {
      preview = preview.replace(`{{.${key}}}`, value)
    })
    setPreviewPayload(preview)
  }

  useEffect(() => {
    updatePreview()
  }, [selectedTemplate, variables])

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const filteredPayloads = payloads.filter(p =>
    p.token.toLowerCase().includes(filter.toLowerCase()) ||
    p.template.toLowerCase().includes(filter.toLowerCase())
  )

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Payload Studio</h2>
        <div className="flex space-x-2">
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            创建 Payload
          </button>
          <button
            onClick={() => setShowBatchModal(true)}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            批量生成
          </button>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg mb-4 p-4">
        <input
          type="text"
          placeholder="搜索 token 或模板..."
          className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {filteredPayloads.length === 0 ? (
            <p className="text-gray-500">暂无 Payloads</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {filteredPayloads.map((payload) => (
                <li key={payload.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <p className="text-sm font-medium text-indigo-600">{payload.template}</p>
                        <span className={`px-2 py-1 text-xs rounded ${
                          payload.status === 'hit' ? 'bg-green-100 text-green-800' :
                          payload.status === 'deployed' ? 'bg-blue-100 text-blue-800' :
                          payload.status === 'expired' ? 'bg-red-100 text-red-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {payload.status}
                        </span>
                      </div>
                      <div className="flex items-center space-x-2 mt-1">
                        <p className="text-sm text-gray-600 break-all">
                          Token: {payload.token}
                        </p>
                        <button
                          onClick={() => copyToClipboard(payload.token)}
                          className="text-xs text-indigo-600 hover:text-indigo-800"
                        >
                          复制
                        </button>
                      </div>
                      {payload.rendered_payload && (
                        <div className="flex items-center space-x-2 mt-1">
                          <p className="text-xs text-gray-500 break-all">
                            Payload: {payload.rendered_payload}
                          </p>
                          <button
                            onClick={() => copyToClipboard(payload.rendered_payload)}
                            className="text-xs text-indigo-600 hover:text-indigo-800"
                          >
                            复制
                          </button>
                        </div>
                      )}
                      {payload.case_id && (
                        <p className="text-xs text-gray-400 mt-1">
                          Case ID: {payload.case_id}
                        </p>
                      )}
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-gray-400">
                        {new Date(payload.created_at).toLocaleString()}
                      </p>
                      {payload.expires_at && (
                        <p className="text-xs text-gray-400">
                          过期: {new Date(payload.expires_at).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full">
            <h3 className="text-lg font-medium mb-4">创建 Payload</h3>
            <form onSubmit={handleCreatePayload}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  选择模板
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={selectedTemplate.id}
                  onChange={(e) => setSelectedTemplate(templates.find(t => t.id === e.target.value) || templates[0])}
                >
                  {templates.map(t => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  模板内容
                </label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded bg-gray-50"
                  value={selectedTemplate.template}
                  readOnly
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  变量
                </label>
                <input
                  type="text"
                  placeholder='{"key": "value"}'
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={JSON.stringify(variables)}
                  onChange={(e) => {
                    try {
                      setVariables(JSON.parse(e.target.value))
                    } catch {}
                  }}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  预览
                </label>
                <div className="p-3 bg-gray-50 rounded border">
                  <p className="text-sm break-all">{previewPayload}</p>
                </div>
              </div>
              <div className="flex justify-end space-x-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  创建
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Batch Create Modal */}
      {showBatchModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium mb-4">批量生成 Payload</h3>
            <form onSubmit={handleBatchCreate}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  生成数量 (1-100)
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={batchCount}
                  onChange={(e) => setBatchCount(parseInt(e.target.value))}
                />
              </div>
              <div className="flex justify-end space-x-2">
                <button
                  type="button"
                  onClick={() => setShowBatchModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                >
                  生成
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
