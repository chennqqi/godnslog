'use client'

import { useEffect, useState } from 'react'
import { caseApi, interactionApi } from '@/lib/api-client'
import type { Case, Interaction } from '@/types'

export default function DashboardPage() {
  const [stats, setStats] = useState({
    activeCases: 0,
    totalPayloads: 0,
    recentInteractions: 0,
  })
  const [recentCases, setRecentCases] = useState<Case[]>([])
  const [recentInteractions, setRecentInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [casesResp, interactionsResp] = await Promise.all([
        caseApi.list({ status: 'active', page: 1, page_size: 5 }),
        interactionApi.list({ page: 1, page_size: 10 }),
      ])

      if (casesResp.data) {
        setStats((prev) => ({ ...prev, activeCases: casesResp.data?.total || 0 }))
        setRecentCases(casesResp.data?.items || [])
      }

      if (interactionsResp.data) {
        setStats((prev) => ({ ...prev, recentInteractions: interactionsResp.data?.total || 0 }))
        setRecentInteractions(interactionsResp.data?.items || [])
      }
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Command Center</h2>

      {/* Stats cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-3 mb-8">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <dt className="text-sm font-medium text-gray-500 truncate">活跃 Cases</dt>
            <dd className="mt-1 text-3xl font-semibold text-gray-900">{stats.activeCases}</dd>
          </div>
        </div>
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <dt className="text-sm font-medium text-gray-500 truncate">最近命中</dt>
            <dd className="mt-1 text-3xl font-semibold text-gray-900">{stats.recentInteractions}</dd>
          </div>
        </div>
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <dt className="text-sm font-medium text-gray-500 truncate">系统状态</dt>
            <dd className="mt-1 text-3xl font-semibold text-green-600">正常</dd>
          </div>
        </div>
      </div>

      {/* Recent cases */}
      <div className="bg-white shadow rounded-lg mb-8">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">最近 Cases</h3>
          {recentCases.length === 0 ? (
            <p className="text-gray-500">暂无 Cases</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {recentCases.map((case_) => (
                <li key={case_.id} className="py-4 flex">
                  <div className="ml-3">
                    <p className="text-sm font-medium text-indigo-600">{case_.title}</p>
                    <p className="text-sm text-gray-500">{case_.description}</p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Recent interactions */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">最近命中</h3>
          {recentInteractions.length === 0 ? (
            <p className="text-gray-500">暂无命中记录</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {recentInteractions.map((interaction) => (
                <li key={interaction.id} className="py-4 flex">
                  <div className="ml-3">
                    <p className="text-sm font-medium text-gray-900">
                      {interaction.type.toUpperCase()} - {interaction.source_ip}
                    </p>
                    <p className="text-sm text-gray-500">{new Date(interaction.timestamp).toLocaleString()}</p>
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
