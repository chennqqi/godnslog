'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { payloadApi } from '@/lib/api-client'
import type { Payload } from '@/types'

export default function PayloadsPage() {
  const router = useRouter()
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')

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
        <input
          type="text"
          placeholder="搜索 token 或模板..."
          className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                      <p className="text-sm text-gray-600 mt-1 break-all">
                        Token: {payload.token}
                      </p>
                      {payload.rendered_payload && (
                        <p className="text-xs text-gray-500 mt-1 break-all">
                          Payload: {payload.rendered_payload}
                        </p>
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
    </div>
  )
}
