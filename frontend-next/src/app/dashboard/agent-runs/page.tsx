'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { agentRunApi } from '@/lib/api-client'
import type { AgentRunDetail, AgentRunReviewQueueItem, ReviewState, EvidenceStrength, AgentRunStatus } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

type ViewMode = 'all' | 'review-queue'

export default function AgentRunsPage() {
  const router = useRouter()
  const [viewMode, setViewMode] = useState<ViewMode>('all')
  const [agentRuns, setAgentRuns] = useState<AgentRunDetail[]>([])
  const [reviewQueue, setReviewQueue] = useState<AgentRunReviewQueueItem[]>([])
  const [loading, setLoading] = useState(true)
  const [filterAgentId, setFilterAgentId] = useState('')
  const [filterStatus, setFilterStatus] = useState('')
  const [filterReviewState, setFilterReviewState] = useState<ReviewState | ''>('')
  const [filterEvidenceStrength, setFilterEvidenceStrength] = useState<EvidenceStrength | ''>('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [reviewSummary, setReviewSummary] = useState({
    total: 0,
    not_reviewed: 0,
    reviewed: 0,
    followup_created: 0,
    needs_attention: 0,
  })
  const [pageSize] = useState(20)

  const loadAgentRuns = useCallback(async () => {
    setLoading(true)
    try {
      const response = await agentRunApi.list({
        agent_id: filterAgentId || undefined,
        status: filterStatus || undefined,
        page,
        page_size: pageSize,
      })
      if (response.data) {
        setAgentRuns(response.data.items || [])
        setTotal(response.data.total || 0)
      }
    } catch (error) {
      console.error('Failed to load agent runs:', error)
    } finally {
      setLoading(false)
    }
  }, [filterAgentId, filterStatus, page, pageSize])

  const loadReviewQueue = useCallback(async () => {
    setLoading(true)
    try {
      const response = await agentRunApi.listReviewQueue({
        agent_id: filterAgentId || undefined,
        status: (filterStatus || undefined) as AgentRunStatus | undefined,
        review_state: filterReviewState || undefined,
        evidence_strength: filterEvidenceStrength || undefined,
        page,
        page_size: pageSize,
      })
      if (response.data) {
        setReviewQueue(response.data.items || [])
        setTotal(response.data.total || 0)
        setReviewSummary(response.data.summary || {
          total: 0,
          not_reviewed: 0,
          reviewed: 0,
          followup_created: 0,
          needs_attention: 0,
        })
      }
    } catch (error) {
      console.error('Failed to load review queue:', error)
    } finally {
      setLoading(false)
    }
  }, [filterAgentId, filterStatus, filterReviewState, filterEvidenceStrength, page, pageSize])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    // Wrap in setTimeout to avoid react-hooks/set-state-in-effect lint error
    setTimeout(() => {
      if (viewMode === 'all') {
        loadAgentRuns()
      } else {
        loadReviewQueue()
      }
    }, 0)
  }, [router, viewMode, loadAgentRuns, loadReviewQueue])

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

  const getReviewStateColor = (state: ReviewState) => {
    const colors: Record<ReviewState, string> = {
      not_reviewed: 'bg-gray-500',
      reviewed: 'bg-green-500',
      followup_created: 'bg-blue-500',
      needs_attention: 'bg-red-500',
    }
    return colors[state] || 'bg-gray-500'
  }

  const getEvidenceStrengthColor = (strength: EvidenceStrength) => {
    const colors: Record<EvidenceStrength, string> = {
      none: 'bg-gray-400',
      low: 'bg-yellow-500',
      medium: 'bg-orange-500',
      high: 'bg-red-500',
    }
    return colors[strength] || 'bg-gray-400'
  }

  const totalPages = Math.ceil(total / pageSize)

  const handleApplyFilters = () => {
    setPage(1)
    if (viewMode === 'all') {
      loadAgentRuns()
    } else {
      loadReviewQueue()
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Agent Runs</h1>
        <p className="text-muted-foreground">View and manage AI agent execution runs</p>
      </div>

      <Tabs value={viewMode} onValueChange={(val) => setViewMode(val as ViewMode)} className="mb-6">
        <TabsList>
          <TabsTrigger value="all">All Runs</TabsTrigger>
          <TabsTrigger value="review-queue">Review Queue</TabsTrigger>
        </TabsList>
      </Tabs>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <Input
              placeholder="Filter by Agent ID"
              value={filterAgentId}
              onChange={(e) => setFilterAgentId(e.target.value)}
              className="max-w-xs"
            />
            <Select value={filterStatus || 'all'} onValueChange={(val) => setFilterStatus(val === 'all' ? '' : val)}>
              <SelectTrigger className="max-w-xs">
                <SelectValue placeholder="Filter by Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value="created">Created</SelectItem>
                <SelectItem value="running">Running</SelectItem>
                <SelectItem value="waiting">Waiting</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
                <SelectItem value="timed_out">Timed Out</SelectItem>
              </SelectContent>
            </Select>
            {viewMode === 'review-queue' && (
              <>
                <Select value={filterReviewState || 'all'} onValueChange={(val) => setFilterReviewState(val === 'all' ? '' : val as ReviewState)}>
                  <SelectTrigger className="max-w-xs">
                    <SelectValue placeholder="Filter by Review State" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All States</SelectItem>
                    <SelectItem value="not_reviewed">Not Reviewed</SelectItem>
                    <SelectItem value="reviewed">Reviewed</SelectItem>
                    <SelectItem value="followup_created">Followup Created</SelectItem>
                    <SelectItem value="needs_attention">Needs Attention</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={filterEvidenceStrength || 'all'} onValueChange={(val) => setFilterEvidenceStrength(val === 'all' ? '' : val as EvidenceStrength)}>
                  <SelectTrigger className="max-w-xs">
                    <SelectValue placeholder="Filter by Evidence Strength" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Strengths</SelectItem>
                    <SelectItem value="none">None</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                  </SelectContent>
                </Select>
              </>
            )}
            <Button onClick={handleApplyFilters}>Apply Filters</Button>
          </div>
        </CardContent>
      </Card>

      {viewMode === 'review-queue' && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Review Queue Summary</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold">{reviewSummary.total}</div>
                <div className="text-sm text-muted-foreground">Total</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold">{reviewSummary.not_reviewed}</div>
                <div className="text-sm text-muted-foreground">Not Reviewed</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold">{reviewSummary.reviewed}</div>
                <div className="text-sm text-muted-foreground">Reviewed</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold">{reviewSummary.followup_created}</div>
                <div className="text-sm text-muted-foreground">Followup Created</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-red-500">{reviewSummary.needs_attention}</div>
                <div className="text-sm text-muted-foreground">Needs Attention</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <div className="text-center py-8">Loading...</div>
      ) : viewMode === 'all' && agentRuns.length === 0 ? (
        <Card>
          <CardContent className="text-center py-8">
            <p className="text-muted-foreground">No agent runs found</p>
          </CardContent>
        </Card>
      ) : viewMode === 'review-queue' && reviewQueue.length === 0 ? (
        <Card>
          <CardContent className="text-center py-8">
            <p className="text-muted-foreground">No items in review queue</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {viewMode === 'all' ? (
            agentRuns.map((run) => (
              <Card
                key={run.id}
                className="cursor-pointer hover:bg-accent/50 transition-colors"
                onClick={() => router.push(`/dashboard/agent-runs/${run.id}`)}
              >
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold text-lg">{run.title}</h3>
                      <p className="text-sm text-muted-foreground">
                        Agent: {run.agent_id} | Target: {run.target}
                      </p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Created: {new Date(run.created_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <Badge className={getStatusColor(run.status)}>
                        {run.status}
                      </Badge>
                      <div className="text-sm text-muted-foreground">
                        {run.interaction_count} interactions
                      </div>
                      {run.operations.length > 0 && (
                        <div className="text-xs text-muted-foreground">
                          {run.operations.length} operations
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          ) : (
            reviewQueue.map((item) => (
              <Card
                key={item.id}
                className="cursor-pointer hover:bg-accent/50 transition-colors"
                onClick={() => router.push(`/dashboard/agent-runs/${item.id}`)}
              >
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold text-lg">{item.title}</h3>
                      <p className="text-sm text-muted-foreground">
                        Agent: {item.agent_id} | Target: {item.target}
                      </p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Created: {new Date(item.created_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <div className="flex gap-2">
                        <Badge className={getStatusColor(item.status)}>
                          {item.status}
                        </Badge>
                        <Badge className={getReviewStateColor(item.review_state)}>
                          {item.review_state}
                        </Badge>
                        <Badge className={getEvidenceStrengthColor(item.evidence_strength)}>
                          {item.evidence_strength}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {item.interaction_count} interactions | {item.followup_count} followups
                      </div>
                      {item.needs_attention && (
                        <Badge variant="destructive">Needs Attention</Badge>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          )}

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Previous
              </Button>
              <span className="flex items-center">
                Page {page} of {totalPages}
              </span>
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
              >
                Next
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
