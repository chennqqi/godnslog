'use client'

import { useEffect, useState } from 'react'
import { caseApi, interactionApi } from '@/lib/api-client'
import type { Case, Interaction } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

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
    loadCases()
  }, [])

  const loadCases = async () => {
    try {
      console.log('Loading dashboard data...')
      const [casesResp, interactionsResp] = await Promise.all([
        caseApi.list({ status: 'active', page: 1, page_size: 5 }),
        interactionApi.list({ page: 1, page_size: 10 }),
      ])
      console.log('Cases response:', casesResp)
      console.log('Interactions response:', interactionsResp)

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
      setStats((prev) => ({ ...prev, activeCases: 0, recentInteractions: 0 }))
      setRecentCases([])
      setRecentInteractions([])
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
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>活跃 Cases</CardDescription>
            <CardTitle className="text-3xl">{stats.activeCases}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>最近命中</CardDescription>
            <CardTitle className="text-3xl">{stats.recentInteractions}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>系统状态</CardDescription>
            <CardTitle className="text-3xl text-green-600">正常</CardTitle>
          </CardHeader>
        </Card>
      </div>

      {/* Recent cases */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>最近 Cases</CardTitle>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>

      {/* Recent interactions */}
      <Card>
        <CardHeader>
          <CardTitle>最近命中</CardTitle>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>
    </div>
  )
}
