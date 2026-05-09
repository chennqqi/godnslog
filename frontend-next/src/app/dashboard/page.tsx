'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi, interactionApi, payloadApi } from '@/lib/api-client'
import type { Case, Interaction } from '@/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function DashboardPage() {
  const router = useRouter()
  const [stats, setStats] = useState({
    totalCases: 0,
    activeCases: 0,
    totalPayloads: 0,
    totalInteractions: 0,
    recentInteractions: 0,
    dnsCount: 0,
    httpCount: 0,
    smtpCount: 0,
  })
  const [recentCases, setRecentCases] = useState<Case[]>([])
  const [recentInteractions, setRecentInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check authentication on client side
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    loadDashboardData()
  }, [router])

  const loadDashboardData = async () => {
    try {
      const [casesResp, interactionsResp, payloadsResp] = await Promise.all([
        caseApi.list({ page: 1, page_size: 5 }),
        interactionApi.list({ page: 1, page_size: 10 }),
        payloadApi.list({ page: 1, page_size: 1 }),
      ])

      if (casesResp.data) {
        setStats((prev) => ({
          ...prev,
          totalCases: casesResp.data?.total || 0,
          activeCases: casesResp.data?.items?.filter(c => c.status === 'active').length || 0,
        }))
        setRecentCases(casesResp.data?.items || [])
      }

      if (interactionsResp.data) {
        const interactions = interactionsResp.data?.items || []
        setStats((prev) => ({
          ...prev,
          totalInteractions: interactionsResp.data?.total || 0,
          recentInteractions: interactionsResp.data?.total || 0,
          dnsCount: interactions.filter(i => i.type === 'dns').length,
          httpCount: interactions.filter(i => i.type === 'http').length,
          smtpCount: interactions.filter(i => i.type === 'smtp').length,
        }))
        setRecentInteractions(interactions)
      }

      if (payloadsResp.data) {
        setStats((prev) => ({
          ...prev,
          totalPayloads: payloadsResp.data?.total || 0,
        }))
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
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>总 Cases</CardDescription>
            <CardTitle className="text-3xl">{stats.totalCases}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>活跃 Cases</CardDescription>
            <CardTitle className="text-3xl text-green-600">{stats.activeCases}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>总 Payloads</CardDescription>
            <CardTitle className="text-3xl">{stats.totalPayloads}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>总交互次数</CardDescription>
            <CardTitle className="text-3xl text-blue-600">{stats.totalInteractions}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      {/* Interaction type stats */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>交互类型统计</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-4 bg-purple-50 rounded">
              <p className="text-2xl font-bold text-purple-600">{stats.dnsCount}</p>
              <p className="text-sm text-gray-600">DNS</p>
            </div>
            <div className="text-center p-4 bg-blue-50 rounded">
              <p className="text-2xl font-bold text-blue-600">{stats.httpCount}</p>
              <p className="text-sm text-gray-600">HTTP</p>
            </div>
            <div className="text-center p-4 bg-green-50 rounded">
              <p className="text-2xl font-bold text-green-600">{stats.smtpCount}</p>
              <p className="text-sm text-gray-600">SMTP</p>
            </div>
          </div>
        </CardContent>
      </Card>

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
                <li key={case_.id} className="py-4 flex justify-between items-center">
                  <div className="ml-3">
                    <p className="text-sm font-medium text-indigo-600">{case_.title}</p>
                    <p className="text-sm text-gray-500">{case_.description}</p>
                  </div>
                  <span className={`px-2 py-1 text-xs rounded ${
                    case_.status === 'active' ? 'bg-green-100 text-green-800' :
                    case_.status === 'completed' ? 'bg-blue-100 text-blue-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {case_.status}
                  </span>
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
                <li key={interaction.id} className="py-4 flex justify-between items-center">
                  <div className="ml-3">
                    <p className="text-sm font-medium text-gray-900">
                      <span className={`px-2 py-1 text-xs rounded mr-2 ${
                        interaction.type === 'dns' ? 'bg-purple-100 text-purple-800' :
                        interaction.type === 'http' ? 'bg-blue-100 text-blue-800' :
                        interaction.type === 'smtp' ? 'bg-green-100 text-green-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {interaction.type.toUpperCase()}
                      </span>
                      {interaction.source_ip}
                    </p>
                    {interaction.domain && <p className="text-sm text-gray-500">域名: {interaction.domain}</p>}
                  </div>
                  <span className="text-xs text-gray-400">
                    {new Date(interaction.timestamp).toLocaleString()}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
