'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { caseApi, payloadApi } from '@/lib/api-client'
import type { Case, Payload } from '@/types'

export default function CaseDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [case_, setCase] = useState<Case | null>(null)
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [stats, setStats] = useState({ payload_count: 0, interaction_count: 0, hit_payload_count: 0 })
  const [loading, setLoading] = useState(true)

  const loadData = useCallback(async () => {
    try {
      const [caseResp, payloadsResp, statsResp] = await Promise.all([
        caseApi.get(params.id as string),
        payloadApi.list({ case_id: params.id as string, page: 1, page_size: 100 }),
        caseApi.stats(params.id as string),
      ])

      // Handle nested response structure
      const caseData = caseResp.data && 'data' in caseResp.data ? caseResp.data.data : caseResp.data
      if (caseData) {
        setCase(caseData as Case)
      }

      if (payloadsResp.data) {
        setPayloads(payloadsResp.data.items)
      }
      if (statsResp.data) {
        setStats({
          payload_count: statsResp.data.payload_count || 0,
          interaction_count: statsResp.data.interaction_count || 0,
          hit_payload_count: statsResp.data.hit_payload_count || 0,
        })
      }
    } catch (error) {
      console.error('Failed to load case data:', error)
    } finally {
      setLoading(false)
    }
  }, [params.id])

  useEffect(() => {
    if (params.id) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      loadData()
    }
  }, [params.id, loadData])

  if (loading) {
    return <div className="text-center py-12">Loading...</div>
  }

  if (!case_) {
    return <div className="text-center py-12">Case not found</div>
  }

  return (
    <div>
      <button
        onClick={() => router.back()}
        className="mb-4 text-indigo-600 hover:text-indigo-800"
      >
        ← Back
      </button>

      <div className="bg-white shadow rounded-lg mb-6">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">{case_.title}</h2>
              <p className="text-gray-600 mt-2">{case_.description}</p>
              {case_.target && (
                <p className="text-sm text-gray-500 mt-2">Target: {case_.target}</p>
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
            Created at: {new Date(case_.created_at).toLocaleString()}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-white shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 uppercase">Payloads</p>
          <p className="text-2xl font-bold text-gray-900">{stats.payload_count}</p>
        </div>
        <div className="bg-white shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 uppercase">Interactions</p>
          <p className="text-2xl font-bold text-gray-900">{stats.interaction_count}</p>
        </div>
        <div className="bg-white shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 uppercase">Hit Payloads</p>
          <p className="text-2xl font-bold text-gray-900">{stats.hit_payload_count}</p>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900">Payloads</h3>
            <button
              onClick={() => router.push(`/dashboard/payloads/new?case_id=${case_.id}`)}
              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm"
            >
              Create Payload
            </button>
          </div>
          {payloads.length === 0 ? (
            <p className="text-gray-500">No payloads yet</p>
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
                      Status: {payload.status}
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

      {/* Quick Actions */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h3>
          <div className="flex flex-wrap gap-3">
            <button
              onClick={() => router.push(`/dashboard/evidence?case_id=${case_.id}`)}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm"
            >
              View Evidence
            </button>
            <button
              onClick={() => router.push(`/dashboard/interactions?case_id=${case_.id}`)}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
            >
              View Interactions
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
