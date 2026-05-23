'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { payloadApi, caseApi, interactionApi } from '@/lib/api-client'
import type { Payload, Case, Interaction } from '@/types'

export default function PayloadDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [payload, setPayload] = useState<Payload | null>(null)
  const [associatedCase, setAssociatedCase] = useState<Case | null>(null)
  const [recentInteractions, setRecentInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)

  const loadPayload = useCallback(async () => {
    try {
      const response = await payloadApi.get(params.id as string)
      let payloadData: Payload | null = null
      if (response.data && 'data' in response.data) {
        payloadData = response.data.data as Payload
      } else if (response.data) {
        payloadData = response.data as Payload
      }

      if (payloadData) {
        setPayload(payloadData)

        // Load associated case if case_id exists
        if (payloadData.case_id) {
          try {
            const caseResp = await caseApi.get(payloadData.case_id)
            if (caseResp.data && 'data' in caseResp.data) {
              setAssociatedCase(caseResp.data.data as Case)
            } else if (caseResp.data) {
              setAssociatedCase(caseResp.data as Case)
            }
          } catch (err) {
            console.error('Failed to load associated case:', err)
          }

          // Load recent interactions
          try {
            const interactionsResp = await interactionApi.list({
              payload_id: params.id as string,
              page: 1,
              page_size: 5,
            })
            if (interactionsResp.data) {
              setRecentInteractions(interactionsResp.data.items || [])
            }
          } catch (err) {
            console.error('Failed to load recent interactions:', err)
          }
        }
      }
    } catch (error) {
      console.error('Failed to load payload:', error)
    } finally {
      setLoading(false)
    }
  }, [params.id])

  useEffect(() => {
    if (params.id) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      loadPayload()
    }
  }, [params.id, loadPayload])

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

            {payload.scenario && (
              <div>
                <p className="text-sm text-gray-500">场景</p>
                <p className="text-sm font-medium">{payload.scenario}</p>
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

      {/* Associated Case */}
      {associatedCase && (
        <div className="bg-white shadow rounded-lg mb-6">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">关联 Case</h3>
            <div
              className="p-4 bg-gray-50 rounded cursor-pointer hover:bg-gray-100 transition-colors"
              onClick={() => router.push(`/dashboard/cases/${associatedCase.id}`)}
            >
              <p className="text-sm font-medium text-indigo-600">{associatedCase.title}</p>
              <p className="text-sm text-gray-500">{associatedCase.description}</p>
              {associatedCase.target && (
                <p className="text-xs text-gray-400 mt-1">Target: {associatedCase.target}</p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Recent Interactions */}
      <div className="bg-white shadow rounded-lg mb-6">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900">最近交互</h3>
            <button
              onClick={() => router.push(`/dashboard/interactions?payload_id=${payload.id}`)}
              className="text-sm text-indigo-600 hover:text-indigo-800"
            >
              查看全部 →
            </button>
          </div>
          {recentInteractions.length === 0 ? (
            <p className="text-gray-500">暂无交互记录</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {recentInteractions.map((interaction) => (
                <li key={interaction.id} className="py-3">
                  <div className="flex justify-between items-start">
                    <div>
                      <p className="text-sm font-medium text-gray-900">{interaction.type.toUpperCase()}</p>
                      <p className="text-xs text-gray-500">
                        {interaction.source_ip}
                        {interaction.domain && ` | ${interaction.domain}`}
                      </p>
                    </div>
                    <p className="text-xs text-gray-400">
                      {new Date(interaction.timestamp).toLocaleString()}
                    </p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">快速操作</h3>
          <div className="flex flex-wrap gap-3">
            <button
              onClick={() => router.push(`/dashboard/interactions?payload_id=${payload.id}`)}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
            >
              查看交互
            </button>
            {associatedCase && (
              <button
                onClick={() => router.push(`/dashboard/evidence?case_id=${associatedCase.id}`)}
                className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm"
              >
                查看证据
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
