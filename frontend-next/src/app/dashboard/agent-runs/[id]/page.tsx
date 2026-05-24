'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { agentRunApi } from '@/lib/api-client'
import type { AgentRunDetail, AgentRunStatus } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

export default function AgentRunDetailPage() {
  const router = useRouter()
  const params = useParams()
  const [agentRun, setAgentRun] = useState<AgentRunDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [updatingStatus, setUpdatingStatus] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    const loadAgentRun = async () => {
      try {
        const response = await agentRunApi.get(params.id as string)
        if (response.data) {
          setAgentRun(response.data.data)
        }
      } catch (error: unknown) {
        console.error('Failed to load agent run:', error)
        setError('加载Agent Run失败')
      } finally {
        setLoading(false)
      }
    }

    loadAgentRun()
  }, [router, params.id])

  const handleUpdateStatus = async (newStatus: AgentRunStatus) => {
    if (!agentRun) return
    setUpdatingStatus(true)
    setError('')
    try {
      await agentRunApi.updateStatus(agentRun.id, { status: newStatus })
      const response = await agentRunApi.get(agentRun.id)
      if (response.data) {
        setAgentRun(response.data.data)
      }
    } catch (error: unknown) {
      console.error('Failed to update status:', error)
      setError('更新状态失败')
    } finally {
      setUpdatingStatus(false)
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center h-screen">加载中...</div>
  }

  if (error || !agentRun) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-red-500">{error || 'Agent Run未找到'}</div>
        <Button onClick={() => router.back()} className="mt-4">
          返回
        </Button>
      </div>
    )
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      created: 'bg-gray-500',
      running: 'bg-blue-500',
      waiting: 'bg-yellow-500',
      completed: 'bg-green-500',
      failed: 'bg-red-500',
      cancelled: 'bg-gray-400',
      timed_out: 'bg-orange-500',
    }
    return colors[status] || 'bg-gray-500'
  }

  const getRiskLevelColor = (riskLevel?: string) => {
    const colors: Record<string, string> = {
      low: 'bg-green-500',
      medium: 'bg-yellow-500',
      high: 'bg-orange-500',
      critical: 'bg-red-500',
    }
    return colors[riskLevel || ''] || 'bg-gray-500'
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <Button onClick={() => router.back()} variant="outline" className="mb-4">
          返回
        </Button>
        <h1 className="text-3xl font-bold">Agent Run 详情</h1>
        <p className="text-muted-foreground">ID: {agentRun.id}</p>
      </div>

      <div className="grid gap-6">
        {/* Basic Info */}
        <Card>
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge>标题</Badge>
              <span>{agentRun.title}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>状态</Badge>
              <Badge className={getStatusColor(agentRun.status)}>
                {agentRun.status}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Agent ID</Badge>
              <span>{agentRun.agent_id}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Operator ID</Badge>
              <span>{agentRun.operator_id}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Target</Badge>
              <span>{agentRun.target}</span>
            </div>
            {agentRun.case_id && (
              <div className="flex items-center gap-2">
                <Badge>Case</Badge>
                <a href={agentRun.case_url} className="text-blue-500 hover:underline">
                  {agentRun.case_id}
                </a>
              </div>
            )}
            {agentRun.payload_id && (
              <div className="flex items-center gap-2">
                <Badge>Payload</Badge>
                <a href={agentRun.payload_url} className="text-blue-500 hover:underline">
                  {agentRun.payload_id}
                </a>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Badge>创建时间</Badge>
              <span>{new Date(agentRun.created_at).toLocaleString()}</span>
            </div>
            {agentRun.started_at && (
              <div className="flex items-center gap-2">
                <Badge>开始时间</Badge>
                <span>{new Date(agentRun.started_at).toLocaleString()}</span>
              </div>
            )}
            {agentRun.ended_at && (
              <div className="flex items-center gap-2">
                <Badge>结束时间</Badge>
                <span>{new Date(agentRun.ended_at).toLocaleString()}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Badge>交互数</Badge>
              <span>{agentRun.interaction_count}</span>
            </div>
            {agentRun.last_interaction_at && (
              <div className="flex items-center gap-2">
                <Badge>最后交互</Badge>
                <span>{new Date(agentRun.last_interaction_at).toLocaleString()}</span>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Status Update */}
        <Card>
          <CardHeader>
            <CardTitle>状态更新</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex gap-2 flex-wrap">
              <Button
                onClick={() => handleUpdateStatus('created')}
                disabled={agentRun.status === 'created' || updatingStatus}
                variant={agentRun.status === 'created' ? 'default' : 'outline'}
                size="sm"
              >
                Created
              </Button>
              <Button
                onClick={() => handleUpdateStatus('running')}
                disabled={agentRun.status === 'running' || updatingStatus}
                variant={agentRun.status === 'running' ? 'default' : 'outline'}
                size="sm"
              >
                Running
              </Button>
              <Button
                onClick={() => handleUpdateStatus('waiting')}
                disabled={agentRun.status === 'waiting' || updatingStatus}
                variant={agentRun.status === 'waiting' ? 'default' : 'outline'}
                size="sm"
              >
                Waiting
              </Button>
              <Button
                onClick={() => handleUpdateStatus('completed')}
                disabled={agentRun.status === 'completed' || updatingStatus}
                variant={agentRun.status === 'completed' ? 'default' : 'outline'}
                size="sm"
              >
                Completed
              </Button>
              <Button
                onClick={() => handleUpdateStatus('failed')}
                disabled={agentRun.status === 'failed' || updatingStatus}
                variant={agentRun.status === 'failed' ? 'default' : 'outline'}
                size="sm"
              >
                Failed
              </Button>
              <Button
                onClick={() => handleUpdateStatus('cancelled')}
                disabled={agentRun.status === 'cancelled' || updatingStatus}
                variant={agentRun.status === 'cancelled' ? 'default' : 'outline'}
                size="sm"
              >
                Cancelled
              </Button>
              <Button
                onClick={() => handleUpdateStatus('timed_out')}
                disabled={agentRun.status === 'timed_out' || updatingStatus}
                variant={agentRun.status === 'timed_out' ? 'default' : 'outline'}
                size="sm"
              >
                Timed Out
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Operations */}
        <Card>
          <CardHeader>
            <CardTitle>操作历史 ({agentRun.operations.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {agentRun.operations.length === 0 ? (
              <p className="text-muted-foreground">暂无操作记录</p>
            ) : (
              <div className="space-y-4">
                {agentRun.operations.map((op) => (
                  <div key={op.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold">{op.action}</h4>
                      <div className="flex gap-2">
                        {op.risk_level && (
                          <Badge className={getRiskLevelColor(op.risk_level)}>
                            {op.risk_level}
                          </Badge>
                        )}
                        <span className="text-sm text-muted-foreground">
                          {new Date(op.started_at).toLocaleString()}
                        </span>
                      </div>
                    </div>
                    {op.request && (
                      <div className="mb-2">
                        <p className="text-sm font-medium">Request:</p>
                        <pre className="text-xs bg-muted p-2 rounded overflow-x-auto">
                          {op.request}
                        </pre>
                      </div>
                    )}
                    {op.result && (
                      <div className="mb-2">
                        <p className="text-sm font-medium">Result:</p>
                        <pre className="text-xs bg-muted p-2 rounded overflow-x-auto">
                          {op.result}
                        </pre>
                      </div>
                    )}
                    {op.error && (
                      <div>
                        <p className="text-sm font-medium text-red-500">Error:</p>
                        <p className="text-sm text-red-500">{op.error}</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Quick Links */}
        {agentRun.payload_id && (
          <Card>
            <CardHeader>
              <CardTitle>快速链接</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <a
                href={agentRun.interactions_url}
                target="_blank"
                rel="noopener noreferrer"
                className="block text-blue-500 hover:underline"
              >
                查看交互记录
              </a>
              <a
                href={agentRun.evidence_url}
                target="_blank"
                rel="noopener noreferrer"
                className="block text-blue-500 hover:underline"
              >
                查看证据
              </a>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
