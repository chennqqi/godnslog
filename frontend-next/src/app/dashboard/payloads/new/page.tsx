'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { payloadApi, caseApi } from '@/lib/api-client'
import type { Payload } from '@/types'

export default function NewPayloadPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [cases, setCases] = useState<any[]>([])
  const [formData, setFormData] = useState({
    case_id: '',
    template: 'ssrf',
    scenario: '',
    expires_in: 3600,
  })

  const templates = [
    { value: 'ssrf', label: 'SSRF', description: 'Server-Side Request Forgery' },
    { value: 'xxe', label: 'XXE', description: 'XML External Entity' },
    { value: 'rce', label: 'RCE', description: 'Remote Code Execution' },
    { value: 'blind_sqli', label: 'Blind SQLi', description: 'Blind SQL Injection' },
    { value: 'ssti', label: 'SSTI', description: 'Server-Side Template Injection' },
    { value: 'deserialization', label: 'Deserialization', description: 'Object Deserialization' },
    { value: 'cors', label: 'CORS/JSONP', description: 'CORS or JSONP Misconfiguration' },
    { value: 'smtp_injection', label: 'SMTP Injection', description: 'SMTP Header Injection' },
    { value: 'pdf_rendering', label: 'PDF Rendering', description: 'PDF Rendering' },
    { value: 'html_rendering', label: 'HTML Rendering', description: 'HTML Rendering' },
  ]

  useEffect(() => {
    loadCases()
  }, [])

  const loadCases = async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const createData: any = {
        template: formData.template,
        scenario: formData.scenario,
        expires_in: formData.expires_in,
      }
      if (formData.case_id) {
        createData.case_id = formData.case_id
      }
      const response = await payloadApi.create(createData)

      if (response.data) {
        const payloadData = response.data && 'data' in response.data ? response.data.data : response.data
        if (payloadData && (payloadData as Payload).id) {
          router.push(`/dashboard/payloads/${(payloadData as Payload).id}`)
        }
      }
    } catch (error) {
      console.error('Failed to create payload:', error)
      alert('创建Payload失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <button
        onClick={() => router.back()}
        className="mb-4 text-indigo-600 hover:text-indigo-800"
      >
        ← 返回
      </button>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">创建新Payload</h2>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                关联Case（可选）
              </label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                value={formData.case_id}
                onChange={(e) => setFormData({ ...formData, case_id: e.target.value })}
              >
                <option value="">选择Case</option>
                {cases.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.title}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Payload模板
              </label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                value={formData.template}
                onChange={(e) => setFormData({ ...formData, template: e.target.value })}
              >
                {templates.map((t) => (
                  <option key={t.value} value={t.value}>
                    {t.label} - {t.description}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                场景描述
              </label>
              <textarea
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                rows={3}
                value={formData.scenario}
                onChange={(e) => setFormData({ ...formData, scenario: e.target.value })}
                placeholder="描述测试场景和目标..."
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                过期时间（秒）
              </label>
              <input
                type="number"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                value={formData.expires_in}
                onChange={(e) => setFormData({ ...formData, expires_in: parseInt(e.target.value) })}
                min={60}
                max={86400}
              />
              <p className="text-xs text-gray-500 mt-1">
                默认3600秒（1小时），最大86400秒（24小时）
              </p>
            </div>

            <div className="flex space-x-4">
              <button
                type="submit"
                disabled={loading}
                className="px-6 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
              >
                {loading ? '创建中...' : '创建Payload'}
              </button>
              <button
                type="button"
                onClick={() => router.back()}
                className="px-6 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
              >
                取消
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
