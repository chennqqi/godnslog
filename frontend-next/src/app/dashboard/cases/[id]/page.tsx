'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { caseApi, payloadApi } from '@/lib/api-client'
import type { Case, Payload } from '@/types'

export default function CaseDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [case_, setCase] = useState<Case | null>(null)
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (params.id) {
      loadData()
    }
  }, [params.id])

  const loadData = async () => {
    try {
      const [caseResp, payloadsResp] = await Promise.all([
        caseApi.get(params.id as string),
        payloadApi.list({ case_id: params.id as string, page: 1, page_size: 100 }),
      ])

      // Handle nested response structure
      const caseData = caseResp.data && 'data' in caseResp.data ? caseResp.data.data : caseResp.data
      if (caseData) {
        setCase(caseData as Case)
      }

      if (payloadsResp.data) {
        setPayloads(payloadsResp.data.items)
      }
    } catch (error) {
      console.error('Failed to load case data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  if (!case_) {
    return <div className="text-center py-12">Case 不存在</div>
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
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">{case_.title}</h2>
              <p className="text-gray-600 mt-2">{case_.description}</p>
              {case_.target && (
                <p className="text-sm text-gray-500 mt-2">目标: {case_.target}</p>
              )}
            </div>
            <span className={`px-3 py-1 text-sm rounded ${
              case_.status === 'active' ? 'bg-green-100 text-green-800' :
              case_.status === 'completed' ? 'bg-blue-100 text-blue-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {case_.status}
            </span>
          </div>
          <div className="mt-4 text-sm text-gray-500">
            创建于: {new Date(case_.created_at).toLocaleString()}
          </div>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900">Payloads</h3>
            <button
              onClick={() => router.push(`/dashboard/cases/${case_.id}/payloads/new`)}
              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm"
            >
              创建 Payload
            </button>
          </div>
          {payloads.length === 0 ? (
            <p className="text-gray-500">暂无 Payloads</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {payloads.map((payload) => (
                <li
                  key={payload.id}
                  className="py-4 flex justify-between items-center cursor-pointer hover:bg-gray-50"
                  onClick={() => router.push(`/dashboard/payloads/${payload.id}`)}
                >
                  <div>
                    <p className="text-sm font-medium text-indigo-600">{payload.template}</p>
                    <p className="text-sm text-gray-500">Token: {payload.token}</p>
                    <p className="text-xs text-gray-400 mt-1">
                      状态: {payload.status}
                    </p>
                  </div>
                  <span className="text-xs text-gray-400">
                    {new Date(payload.created_at).toLocaleDateString()}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}
