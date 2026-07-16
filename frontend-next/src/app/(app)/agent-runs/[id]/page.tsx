'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { agentRunApi } from '@/lib/api-client'
import type { AgentRunDetail, AgentRunStatus, AgentRunReviewPacket, AgentRunFollowupActionType, AgentRunFollowupHistoryItem, ReviewDecisionType, AgentRunReviewExportResponse, AgentRunReviewDeliveryResponse, AgentRunReviewDeliveryHistoryResponse } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogDescription } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Input } from '@/components/ui/input'

export default function AgentRunDetailPage() {
  const router = useRouter()
  const params = useParams()
  const [agentRun, setAgentRun] = useState<AgentRunDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [updatingStatus, setUpdatingStatus] = useState(false)
  const [reviewPacket, setReviewPacket] = useState<AgentRunReviewPacket | null>(null)
  const [generatingReview, setGeneratingReview] = useState(false)
  const [reviewFormat, setReviewFormat] = useState<'json' | 'markdown'>('json')
  const [followupDialogOpen, setFollowupDialogOpen] = useState(false)
  const [followupActionType, setFollowupActionType] = useState<AgentRunFollowupActionType>('recheck_evidence')
  const [followupReason, setFollowupReason] = useState('')
  const [creatingFollowup, setCreatingFollowup] = useState(false)
  const [followupHistory, setFollowupHistory] = useState<AgentRunFollowupHistoryItem[]>([])
  const [loadingFollowupHistory, setLoadingFollowupHistory] = useState(false)
  const [reviewDecisionDialogOpen, setReviewDecisionDialogOpen] = useState(false)
  const [reviewDecision, setReviewDecision] = useState<ReviewDecisionType>('accepted')
  const [reviewDecisionReason, setReviewDecisionReason] = useState('')
  const [creatingReviewDecision, setCreatingReviewDecision] = useState(false)
  const [exportDialogOpen, setExportDialogOpen] = useState(false)
  const [exportFormat, setExportFormat] = useState<'json' | 'markdown'>('json')
  const [exporting, setExporting] = useState(false)
  const [exportResult, setExportResult] = useState<AgentRunReviewExportResponse | null>(null)
  const [deliveryDialogOpen, setDeliveryDialogOpen] = useState(false)
  const [deliveryFormat, setDeliveryFormat] = useState<'json' | 'markdown'>('markdown')
  const [deliveryWebhookURL, setDeliveryWebhookURL] = useState('')
  const [deliveryHeaders, setDeliveryHeaders] = useState<Array<{ key: string; value: string }>>([])
  const [delivering, setDelivering] = useState(false)
  const [deliveryResult, setDeliveryResult] = useState<AgentRunReviewDeliveryResponse | null>(null)
  const [deliveryError, setDeliveryError] = useState('')
  const [deliveryHistory, setDeliveryHistory] = useState<AgentRunReviewDeliveryHistoryResponse | null>(null)
  const [loadingDeliveryHistory, setLoadingDeliveryHistory] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    const loadAgentRun = async () => {
      try {
        const response = await agentRunApi.get(params.id as string)
        if (response.data && response.data.data) {
          setAgentRun(response.data.data)
          // Extract review packet from agent run data if available
          if (response.data.data.review_packet) {
            setReviewPacket(response.data.data.review_packet)
          }
        }
      } catch (error: unknown) {
        console.error('Failed to load agent run:', error)
        setError('加载Agent Run失败')
      } finally {
        setLoading(false)
      }
    }

    const loadFollowupHistory = async () => {
      setLoadingFollowupHistory(true)
      try {
        const response = await agentRunApi.listFollowupHistory(params.id as string)
        if (response.data) {
          setFollowupHistory(response.data.data || [])
        }
      } catch (error: unknown) {
        console.error('Failed to load followup history:', error)
      } finally {
        setLoadingFollowupHistory(false)
      }
    }

    const loadDeliveryHistory = async () => {
      setLoadingDeliveryHistory(true)
      try {
        const response = await agentRunApi.listReviewDeliveries(params.id as string)
        if (response.data) {
          setDeliveryHistory(response.data.data)
        }
      } catch (error: unknown) {
        // Don't fail the whole page if delivery history fails to load
        console.error('Failed to load delivery history:', error)
        setDeliveryHistory(null)
      } finally {
        setLoadingDeliveryHistory(false)
      }
    }

    loadAgentRun()
    loadFollowupHistory()
    loadDeliveryHistory()
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

  const handleGenerateReview = async (format: 'json' | 'markdown') => {
    if (!agentRun) return
    setGeneratingReview(true)
    setError('')
    try {
      const response = await agentRunApi.getReview(agentRun.id, format)
      if (response.data) {
        setReviewPacket(response.data.data || response.data)
        setReviewFormat(format)
      }
    } catch (error: unknown) {
      console.error('Failed to generate review:', error)
      setError('生成Review失败')
    } finally {
      setGeneratingReview(false)
    }
  }

  const handleCreateFollowup = async () => {
    if (!agentRun || !followupReason.trim()) return
    setCreatingFollowup(true)
    setError('')
    try {
      await agentRunApi.createFollowup(agentRun.id, {
        action_type: followupActionType,
        reason: followupReason.trim(),
        review_packet_id: reviewPacket?.id,
      })
      setFollowupDialogOpen(false)
      setFollowupReason('')
      const response = await agentRunApi.get(agentRun.id)
      if (response.data) {
        setAgentRun(response.data.data)
      }
      // Refresh followup history
      const historyResponse = await agentRunApi.listFollowupHistory(agentRun.id)
      if (historyResponse.data) {
        setFollowupHistory(historyResponse.data.data || [])
      }
    } catch (error: unknown) {
      console.error('Failed to create followup:', error)
      setError('创建Follow-up失败')
    } finally {
      setCreatingFollowup(false)
    }
  }

  const handleCreateReviewDecision = async () => {
    if (!agentRun) return
    setCreatingReviewDecision(true)
    setError('')
    try {
      await agentRunApi.createReviewDecision(agentRun.id, {
        decision: reviewDecision,
        reason: reviewDecisionReason.trim(),
        review_packet_id: reviewPacket?.id,
      })
      setReviewDecisionDialogOpen(false)
      setReviewDecisionReason('')
      const response = await agentRunApi.get(agentRun.id)
      if (response.data) {
        setAgentRun(response.data.data)
      }
    } catch (error: unknown) {
      console.error('Failed to create review decision:', error)
      setError('创建Review Decision失败')
    } finally {
      setCreatingReviewDecision(false)
    }
  }

  const handleExportReview = async () => {
    if (!agentRun) return
    setExporting(true)
    setError('')
    try {
      const response = await agentRunApi.exportReview(agentRun.id, {
        format: exportFormat,
        review_packet_id: agentRun.id,
        include_audit: true,
      })
      if (response.data) {
        setExportResult(response.data.data)
        // Immediately refresh agent run to update timeline
        const agentRunResponse = await agentRunApi.get(agentRun.id)
        if (agentRunResponse.data) {
          setAgentRun(agentRunResponse.data.data)
        }
      }
    } catch (error: unknown) {
      console.error('Failed to export review:', error)
      setError('导出Review Evidence失败')
    } finally {
      setExporting(false)
    }
  }

  const handleDeliverReview = async () => {
    if (!agentRun) return
    setDelivering(true)
    setDeliveryError('')
    setDeliveryResult(null)
    try {
      // Convert headers array to map
      const headersMap: Record<string, string> = {}
      deliveryHeaders.forEach(header => {
        if (header.key && header.value) {
          headersMap[header.key] = header.value
        }
      })

      const response = await agentRunApi.deliverReview(agentRun.id, {
        format: deliveryFormat,
        review_packet_id: agentRun.id,
        webhook_url: deliveryWebhookURL,
        headers: headersMap,
        include_audit: true,
      })
      if (response.data) {
        setDeliveryResult(response.data.data)
        // Immediately refresh agent run to update timeline
        const agentRunResponse = await agentRunApi.get(agentRun.id)
        if (agentRunResponse.data) {
          setAgentRun(agentRunResponse.data.data)
        }
        // Refresh delivery history
        const deliveryHistoryResponse = await agentRunApi.listReviewDeliveries(agentRun.id)
        if (deliveryHistoryResponse.data) {
          setDeliveryHistory(deliveryHistoryResponse.data.data)
        }
      }
    } catch (error: unknown) {
      console.error('Failed to deliver review:', error)
      setDeliveryError('Delivery failed')
    } finally {
      setDelivering(false)
    }
  }

  const handleAddHeader = () => {
    setDeliveryHeaders([...deliveryHeaders, { key: '', value: '' }])
  }

  const handleRemoveHeader = (index: number) => {
    setDeliveryHeaders(deliveryHeaders.filter((_, i) => i !== index))
  }

  const handleHeaderChange = (index: number, field: 'key' | 'value', value: string) => {
    const newHeaders = [...deliveryHeaders]
    newHeaders[index][field] = value
    setDeliveryHeaders(newHeaders)
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
                        {/* Parse result for audit ref */}
                        {(() => {
                          try {
                            const resultData = JSON.parse(op.result)
                            if (resultData.audit_ref_id) {
                              return (
                                <div className="mt-2">
                                  <a
                                    href={`/dashboard/audit?resource_type=agent_run&resource_id=${agentRun.id}`}
                                    className="text-sm text-blue-500 hover:underline"
                                  >
                                    View Audit Log ({resultData.audit_ref_id})
                                  </a>
                                </div>
                              )
                            }
                          } catch {
                            // Ignore parse errors
                          }
                          return null
                        })()}
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

        {/* Follow-up History */}
        <Card>
          <CardHeader>
            <CardTitle>Follow-up History ({followupHistory.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {loadingFollowupHistory ? (
              <p className="text-muted-foreground">加载中...</p>
            ) : followupHistory.length === 0 ? (
              <p className="text-muted-foreground">暂无 Follow-up 记录</p>
            ) : (
              <div className="space-y-4">
                {followupHistory.map((item) => (
                  <div key={item.operation_id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold">{item.action_type}</h4>
                      <div className="flex gap-2">
                        {item.risk_level && (
                          <Badge className={getRiskLevelColor(item.risk_level)}>
                            {item.risk_level}
                          </Badge>
                        )}
                        <span className="text-sm text-muted-foreground">
                          {new Date(item.created_at).toLocaleString()}
                        </span>
                      </div>
                    </div>
                    {item.reason && (
                      <div className="mb-2">
                        <p className="text-sm font-medium">Reason:</p>
                        <p className="text-sm">{item.reason}</p>
                      </div>
                    )}
                    {item.review_packet_id && (
                      <div className="mb-2">
                        <p className="text-sm font-medium">Review Packet ID:</p>
                        <p className="text-sm text-muted-foreground">{item.review_packet_id}</p>
                      </div>
                    )}
                    {item.audit_ref_id && (
                      <div>
                        <p className="text-sm font-medium">Audit Ref:</p>
                        <a
                          href={`/dashboard/audit?resource_type=agent_run&resource_id=${agentRun.id}`}
                          className="text-sm text-blue-500 hover:underline"
                        >
                          {item.audit_ref_id}
                        </a>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Delivery History */}
        <Card>
          <CardHeader>
            <CardTitle>Delivery History</CardTitle>
          </CardHeader>
          <CardContent>
            {loadingDeliveryHistory ? (
              <p className="text-muted-foreground">加载中...</p>
            ) : !deliveryHistory || !deliveryHistory.items || deliveryHistory.items.length === 0 ? (
              <p className="text-muted-foreground">暂无 Delivery 记录</p>
            ) : (
              <div className="space-y-4">
                {/* Summary */}
                <div className="flex gap-4 flex-wrap">
                  <Badge variant="outline">Total: {deliveryHistory.summary.total}</Badge>
                  <Badge className="bg-green-500">Delivered: {deliveryHistory.summary.delivered}</Badge>
                  <Badge className="bg-red-500">Failed: {deliveryHistory.summary.failed}</Badge>
                  <Badge className="bg-yellow-500">Timeout: {deliveryHistory.summary.timeout}</Badge>
                </div>

                {/* History Items */}
                <div className="space-y-3">
                  {deliveryHistory.items.map((item) => (
                    <div key={item.delivery_operation_id} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="font-semibold">
                          {item.destination_host}
                        </h4>
                        <Badge className={
                          item.result === 'delivered' ? 'bg-green-500' :
                          item.result === 'failed' ? 'bg-red-500' :
                          'bg-yellow-500'
                        }>
                          {item.result}
                        </Badge>
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-sm">
                        <div>
                          <p className="font-medium">Format:</p>
                          <p className="text-muted-foreground">{item.format}</p>
                        </div>
                        {item.status_code && (
                          <div>
                            <p className="font-medium">Status Code:</p>
                            <p className="text-muted-foreground">{item.status_code}</p>
                          </div>
                        )}
                        <div>
                          <p className="font-medium">Created:</p>
                          <p className="text-muted-foreground">{new Date(item.created_at).toLocaleString()}</p>
                        </div>
                        {item.delivered_at && (
                          <div>
                            <p className="font-medium">Delivered:</p>
                            <p className="text-muted-foreground">{new Date(item.delivered_at).toLocaleString()}</p>
                          </div>
                        )}
                      </div>
                      {item.header_names && item.header_names.length > 0 && (
                        <div className="mt-2">
                          <p className="text-sm font-medium">Headers:</p>
                          <p className="text-sm text-muted-foreground">{item.header_names.join(', ')}</p>
                        </div>
                      )}
                      {item.error_summary && (
                        <div className="mt-2">
                          <p className="text-sm font-medium">Error:</p>
                          <p className="text-sm text-muted-foreground">{item.error_summary}</p>
                        </div>
                      )}
                      <div className="mt-2 flex gap-4 text-sm">
                        <div>
                          <p className="font-medium">Operation ID:</p>
                          <p className="text-muted-foreground">{item.delivery_operation_id}</p>
                        </div>
                        {item.package_hash && (
                          <div>
                            <p className="font-medium">Package Hash:</p>
                            <code className="bg-muted px-2 py-1 rounded text-xs cursor-pointer hover:bg-muted/80" 
                                  onClick={() => navigator.clipboard.writeText(item.package_hash || '')}
                                  title="Click to copy full hash">
                              {item.package_hash.substring(0, 12)}...
                            </code>
                          </div>
                        )}
                        {item.audit_ref_id && (
                          <div>
                            <p className="font-medium">Audit Ref:</p>
                            <a
                              href={`/dashboard/audit?resource_type=agent_run&resource_id=${agentRun.id}`}
                              className="text-blue-500 hover:underline"
                            >
                              {item.audit_ref_id}
                            </a>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Review Packet */}
        <Card>
          <CardHeader>
            <CardTitle>Review Packet</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2 flex-wrap">
              <Button
                onClick={() => handleGenerateReview('json')}
                disabled={generatingReview}
                variant={reviewFormat === 'json' ? 'default' : 'outline'}
                size="sm"
              >
                生成 JSON Review
              </Button>
              <Button
                onClick={() => handleGenerateReview('markdown')}
                disabled={generatingReview}
                variant={reviewFormat === 'markdown' ? 'default' : 'outline'}
                size="sm"
              >
                生成 Markdown Review
              </Button>
              {reviewPacket && (
                <Dialog open={followupDialogOpen} onOpenChange={setFollowupDialogOpen}>
                  <DialogTrigger asChild>
                    <Button variant="secondary" size="sm">
                      创建 Follow-up Action
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>创建 Follow-up Action</DialogTitle>
                      <DialogDescription>
                        为此 Agent Run 创建一个 Follow-up Action 以进行后续操作
                      </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                      <div>
                        <Label htmlFor="action-type">Action Type</Label>
                        <Select value={followupActionType} onValueChange={(value: AgentRunFollowupActionType) => setFollowupActionType(value)}>
                          <SelectTrigger id="action-type">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="recheck_evidence">Recheck Evidence</SelectItem>
                            <SelectItem value="wait_more_interactions">Wait More Interactions</SelectItem>
                            <SelectItem value="create_followup_note">Create Followup Note</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div>
                        <Label htmlFor="reason">Reason</Label>
                        <Textarea
                          id="reason"
                          placeholder="请输入原因..."
                          value={followupReason}
                          onChange={(e) => setFollowupReason(e.target.value)}
                          rows={4}
                        />
                      </div>
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="outline"
                          onClick={() => setFollowupDialogOpen(false)}
                          disabled={creatingFollowup}
                        >
                          取消
                        </Button>
                        <Button
                          onClick={handleCreateFollowup}
                          disabled={creatingFollowup || !followupReason.trim()}
                        >
                          {creatingFollowup ? '创建中...' : '创建'}
                        </Button>
                      </div>
                    </div>
                  </DialogContent>
                </Dialog>
              )}
              {reviewPacket && (
                <Dialog open={reviewDecisionDialogOpen} onOpenChange={setReviewDecisionDialogOpen}>
                  <DialogTrigger asChild>
                    <Button variant="default" size="sm">
                      记录 Review Decision
                    </Button>
                  </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>记录 Review Decision</DialogTitle>
                    <DialogDescription>
                      为此 Agent Run 记录复核结论
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="decision">Decision</Label>
                      <Select value={reviewDecision} onValueChange={(value: ReviewDecisionType) => setReviewDecision(value)}>
                        <SelectTrigger id="decision">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="accepted">Accepted</SelectItem>
                          <SelectItem value="false_positive">False Positive</SelectItem>
                          <SelectItem value="needs_manual_followup">Needs Manual Follow-up</SelectItem>
                          <SelectItem value="insufficient_evidence">Insufficient Evidence</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="decision-reason">Reason</Label>
                      <Textarea
                        id="decision-reason"
                        placeholder="请输入原因..."
                        value={reviewDecisionReason}
                        onChange={(e) => setReviewDecisionReason(e.target.value)}
                        rows={4}
                      />
                    </div>
                    <div className="flex justify-end gap-2">
                      <Button
                        variant="outline"
                        onClick={() => setReviewDecisionDialogOpen(false)}
                        disabled={creatingReviewDecision}
                      >
                        取消
                      </Button>
                      <Button
                        onClick={handleCreateReviewDecision}
                        disabled={creatingReviewDecision}
                      >
                        {creatingReviewDecision ? '记录中...' : '记录'}
                      </Button>
                    </div>
                  </div>
                </DialogContent>
              </Dialog>
              )}
              {reviewPacket && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setExportFormat('json')
                      setExportDialogOpen(true)
                    }}
                  >
                    Export JSON
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setExportFormat('markdown')
                      setExportDialogOpen(true)
                    }}
                  >
                    Export Markdown
                  </Button>
                  <Dialog open={exportDialogOpen} onOpenChange={setExportDialogOpen}>
                    <DialogContent className="max-w-3xl max-h-[80vh]">
                      <DialogHeader>
                        <DialogTitle>Export Review Evidence</DialogTitle>
                        <DialogDescription>
                          导出复核证据包 ({exportFormat.toUpperCase()})
                        </DialogDescription>
                      </DialogHeader>
                      <div className="space-y-4">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            onClick={() => setExportDialogOpen(false)}
                            disabled={exporting}
                          >
                            取消
                          </Button>
                          <Button
                            onClick={handleExportReview}
                            disabled={exporting}
                          >
                            {exporting ? '导出中...' : '导出'}
                          </Button>
                        </div>
                        {exportResult && (
                          <div className="mt-4">
                            <Label>Export Result</Label>
                            {exportResult.package_hash && (
                              <div className="mt-2 flex items-center gap-2 text-sm">
                                <span className="font-medium">Package Hash:</span>
                                <code className="bg-muted px-2 py-1 rounded text-xs cursor-pointer hover:bg-muted/80" 
                                      onClick={() => navigator.clipboard.writeText(exportResult.package_hash || '')}
                                      title="Click to copy full hash">
                                  {exportResult.package_hash.substring(0, 12)}...
                                </code>
                              </div>
                            )}
                            <ScrollArea className="h-[400px] w-full rounded-md border p-4">
                              <pre className="text-xs whitespace-pre-wrap">
                                {exportFormat === 'json' ? JSON.stringify(exportResult.package, null, 2) : exportResult.content}
                              </pre>
                            </ScrollArea>
                            {exportResult.audit_ref_id && (
                              <Button
                                variant="link"
                                className="mt-2"
                                onClick={() => router.push(`/dashboard/audit?resource_type=agent_run&resource_id=${agentRun.id}`)}
                              >
                                View Audit Log ({exportResult.audit_ref_id})
                              </Button>
                            )}
                          </div>
                        )}
                      </div>
                    </DialogContent>
                  </Dialog>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setDeliveryDialogOpen(true)}
                  >
                    Deliver to Webhook
                  </Button>
                  <Dialog open={deliveryDialogOpen} onOpenChange={setDeliveryDialogOpen}>
                    <DialogContent className="max-w-2xl">
                      <DialogHeader>
                        <DialogTitle>Deliver Review Evidence to Webhook</DialogTitle>
                        <DialogDescription>
                          将复核证据包发送到外部 Webhook
                        </DialogDescription>
                      </DialogHeader>
                      <div className="space-y-4">
                        <div>
                          <Label htmlFor="delivery-format">Format</Label>
                          <Select value={deliveryFormat} onValueChange={(value: 'json' | 'markdown') => setDeliveryFormat(value)}>
                            <SelectTrigger id="delivery-format">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="json">JSON</SelectItem>
                              <SelectItem value="markdown">Markdown</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                        <div>
                          <Label htmlFor="webhook-url">Webhook URL</Label>
                          <Input
                            id="webhook-url"
                            type="url"
                            placeholder="https://hooks.example.com/review"
                            value={deliveryWebhookURL}
                            onChange={(e) => setDeliveryWebhookURL(e.target.value)}
                          />
                        </div>
                        <div>
                          <Label>Headers (Optional)</Label>
                          <div className="space-y-2 mt-2">
                            {deliveryHeaders.map((header, index) => (
                              <div key={index} className="flex gap-2">
                                <Input
                                  placeholder="X-Custom-Header"
                                  value={header.key}
                                  onChange={(e) => handleHeaderChange(index, 'key', e.target.value)}
                                  className="flex-1"
                                />
                                <Input
                                  placeholder="value"
                                  value={header.value}
                                  onChange={(e) => handleHeaderChange(index, 'value', e.target.value)}
                                  className="flex-1"
                                />
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => handleRemoveHeader(index)}
                                >
                                  Remove
                                </Button>
                              </div>
                            ))}
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              onClick={handleAddHeader}
                            >
                              Add Header
                            </Button>
                          </div>
                          <p className="text-xs text-muted-foreground mt-1">
                            Only Content-Type and X-* headers are allowed
                          </p>
                        </div>
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            onClick={() => setDeliveryDialogOpen(false)}
                            disabled={delivering}
                          >
                            取消
                          </Button>
                          <Button
                            onClick={handleDeliverReview}
                            disabled={delivering || !deliveryWebhookURL}
                          >
                            {delivering ? '发送中...' : '发送'}
                          </Button>
                        </div>
                        {deliveryError && (
                          <div className="text-red-500 text-sm">{deliveryError}</div>
                        )}
                        {deliveryResult && (
                          <div className="mt-4 border rounded-md p-4">
                            <Label>Delivery Receipt</Label>
                            {deliveryResult.package_hash && (
                              <div className="mt-2 flex items-center gap-2 text-sm">
                                <span className="font-medium">Package Hash:</span>
                                <code className="bg-muted px-2 py-1 rounded text-xs cursor-pointer hover:bg-muted/80" 
                                      onClick={() => navigator.clipboard.writeText(deliveryResult.package_hash || '')}
                                      title="Click to copy full hash">
                                  {deliveryResult.package_hash.substring(0, 12)}...
                                </code>
                              </div>
                            )}
                            <div className="mt-2 space-y-2 text-sm">
                              <div><strong>Delivery ID:</strong> {deliveryResult.delivery_id}</div>
                              <div><strong>Format:</strong> {deliveryResult.format}</div>
                              <div><strong>Destination:</strong> {deliveryResult.destination_host}</div>
                              <div><strong>Status Code:</strong> {deliveryResult.status_code}</div>
                              <div><strong>Result:</strong> {deliveryResult.result}</div>
                              <div><strong>Delivered At:</strong> {new Date(deliveryResult.delivered_at).toLocaleString()}</div>
                              {deliveryResult.audit_ref_id && (
                                <Button
                                  variant="link"
                                  className="p-0 h-auto"
                                  onClick={() => router.push(`/dashboard/audit?resource_type=agent_run&resource_id=${agentRun.id}`)}
                                >
                                  View Audit Log ({deliveryResult.audit_ref_id})
                                </Button>
                              )}
                            </div>
                          </div>
                        )}
                      </div>
                    </DialogContent>
                  </Dialog>
                </>
              )}
              {agentRun.payload_id && (
                <Button
                  onClick={() => router.push(`/dashboard/evidence?payload_id=${agentRun.payload_id}`)}
                  variant="outline"
                  size="sm"
                >
                  查看证据
                </Button>
              )}
            </div>

            {generatingReview && (
              <p className="text-muted-foreground">生成中...</p>
            )}

            {reviewPacket && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                  <div>
                    <p className="text-sm font-medium">Evidence Strength</p>
                    {reviewPacket.evidence ? (
                      <Badge className={getRiskLevelColor(reviewPacket.evidence.evidence_strength)}>
                        {reviewPacket.evidence.evidence_strength}
                      </Badge>
                    ) : (
                      <span className="text-muted-foreground">N/A</span>
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium">Confidence</p>
                    {reviewPacket.evidence ? (
                      <span>{reviewPacket.evidence.confidence}%</span>
                    ) : (
                      <span className="text-muted-foreground">N/A</span>
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium">Interaction Count</p>
                    <span>{reviewPacket.interaction_summary.total}</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium">Unique Sources</p>
                    <span>{reviewPacket.interaction_summary.unique_sources}</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium">Generated At</p>
                    <span>{new Date(reviewPacket.generated_at).toLocaleString()}</span>
                  </div>
                </div>

                {reviewFormat === 'markdown' && reviewPacket.content && (
                  <div>
                    <p className="text-sm font-medium mb-2">Markdown Preview</p>
                    <pre className="text-xs bg-muted p-4 rounded overflow-x-auto max-h-96">
                      {reviewPacket.content}
                    </pre>
                  </div>
                )}
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
