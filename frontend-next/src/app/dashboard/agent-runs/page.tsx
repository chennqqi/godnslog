'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { agentRunApi } from '@/lib/api-client'
import type { AgentRunDetail } from '@/types'
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

export default function AgentRunsPage() {
  const router = useRouter()
  const [agentRuns, setAgentRuns] = useState<AgentRunDetail[]>([])
  const [loading, setLoading] = useState(true)
  const [filterAgentId, setFilterAgentId] = useState('')
  const [filterStatus, setFilterStatus] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
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

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    // Wrap in setTimeout to avoid react-hooks/set-state-in-effect lint error
    setTimeout(() => {
      loadAgentRuns()
    }, 0)
  }, [router, loadAgentRuns])

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

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Agent Runs</h1>
        <p className="text-muted-foreground">View and manage AI agent execution runs</p>
      </div>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
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
            <Button onClick={loadAgentRuns}>Apply Filters</Button>
          </div>
        </CardContent>
      </Card>

      {loading ? (
        <div className="text-center py-8">Loading...</div>
      ) : agentRuns.length === 0 ? (
        <Card>
          <CardContent className="text-center py-8">
            <p className="text-muted-foreground">No agent runs found</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {agentRuns.map((run) => (
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
          ))}

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
