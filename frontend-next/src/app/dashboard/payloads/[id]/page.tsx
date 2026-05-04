'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { payloadApi } from '@/lib/api-client'
import type { Payload } from '@/types'

export default function PayloadDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [payload, setPayload] = useState<Payload | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (params.id) {
      loadPayload()
    }
  }, [params.id])

  const loadPayload = async () => {
    try {
      const response = await payloadApi.get(params.id as string)
      if (response.data && 'data' in response.data) {
        setPayload(response.data.data as Payload)
      } else if (response.data) {
        setPayload(response.data as Payload)
      }
    } catch (error) {
      console.error('Failed to load payload:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleRevoke = async () => {
    if (!confirm('确定要撤销此Payload吗？')) return

    try {
      await payloadApi.revoke(params.id as string)
      alert('Payload已撤销')
      loadPayload()
    } catch (error) {
      console.error('Failed to revoke payload:', error)
      alert('撤销失败')
    }
  }

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  if (!payload) {
    return <div className="text-center py-12">Payload 不存在</div>
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
            <h2 className="text-2xl font-bold text-gray-900">{payload.template}</h2>
            <div className="flex space-x-2">
              <button
                onClick={handleRevoke}
                disabled={payload.status === 'archived' || payload.status === 'expired'}
                className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm disabled:opacity-50"
              >
                撤销
              </button>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <p className="text-sm text-gray-500">Token</p>
              <p className="text-sm font-medium break-all bg-gray-50 p-2 rounded">{payload.token}</p>
            </div>

            {payload.rendered_payload && (
              <div>
                <p className="text-sm text-gray-500">渲染Payload</p>
                <p className="text-sm font-medium break-all bg-gray-50 p-2 rounded">{payload.rendered_payload}</p>
              </div>
            )}

            {(payload as any).scenario && (
              <div>
                <p className="text-sm text-gray-500">场景</p>
                <p className="text-sm font-medium">{(payload as any).scenario}</p>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-500">状态</p>
                <span className={`px-2 py-1 text-xs rounded ${
                  payload.status === 'hit' ? 'bg-green-100 text-green-800' :
                  payload.status === 'deployed' ? 'bg-blue-100 text-blue-800' :
                  payload.status === 'archived' ? 'bg-red-100 text-red-800' :
                  payload.status === 'expired' ? 'bg-gray-100 text-gray-800' :
                  'bg-gray-100 text-gray-800'
                }`}>
                  {payload.status}
                </span>
              </div>
              <div>
                <p className="text-sm text-gray-500">创建时间</p>
                <p className="text-sm font-medium">{new Date(payload.created_at).toLocaleString()}</p>
              </div>
              {payload.expires_at && (
                <div>
                  <p className="text-sm text-gray-500">过期时间</p>
                  <p className="text-sm font-medium">{new Date(payload.expires_at).toLocaleString()}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">相关Interactions</h3>
          <p className="text-gray-500">查看关联的Interaction记录...</p>
        </div>
      </div>
    </div>
  )
}
