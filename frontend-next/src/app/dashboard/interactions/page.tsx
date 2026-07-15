'use client'

/* eslint-disable react-hooks/set-state-in-effect */
import { useEffect, useState, Suspense, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { interactionApi } from '@/lib/api-client'
import { useInteractions, useInteractionStats } from '@/features/interactions/hooks/use-interactions'
import type { Interaction } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useInteractionStream } from '@/features/interactions/hooks/use-interaction-stream'
import { useI18n } from '@/lib/i18n-context'
import { LoadingState } from '@/components/loading-state'

/** Radix SelectItem cannot use value="" for "all types" */
const TYPE_FILTER_ALL = 'all'

function InteractionsPageContent() {
  const router = useRouter()
  const searchParams = useSearchParams()

  // Get scope from URL
  const caseId = searchParams.get('case_id') ?? undefined
  const payloadId = searchParams.get('payload_id') ?? undefined
  const typeParam = searchParams.get('type')
  const filterParam = searchParams.get('filter') ?? ''
  const startTimeParam = searchParams.get('start_time') ?? ''
  const endTimeParam = searchParams.get('end_time') ?? ''
  const viewParam = searchParams.get('view')

  const { data: interactionsData, isLoading: loading } = useInteractions({
    page: 1, page_size: 100,
    ...(caseId ? { case_id: caseId } : {}),
    ...(payloadId ? { payload_id: payloadId } : {}),
  })
  const { data: statsData } = useInteractionStats({
    ...(caseId ? { case_id: caseId } : {}),
    ...(payloadId ? { payload_id: payloadId } : {}),
  })
  const interactions = interactionsData?.data?.items ?? []
  const stats = statsData?.data ?? { total: 0, dns_count: 0, http_count: 0, smtp_count: 0, ldap_count: 0 }

  const [filter, setFilter] = useState(filterParam)
  const [typeFilter, setTypeFilter] = useState(typeParam ?? TYPE_FILTER_ALL)
  const [viewMode, setViewMode] = useState<'table' | 'timeline'>(viewParam === 'timeline' ? 'timeline' : 'table')
  const [selectedInteraction, setSelectedInteraction] = useState<Interaction | null>(null)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [liveCount, setLiveCount] = useState(0)
  const [exporting, setExporting] = useState(false)
  const [startTimeFilter, setStartTimeFilter] = useState(startTimeParam)
  const [endTimeFilter, setEndTimeFilter] = useState(endTimeParam)
  const { t } = useI18n()

  // Sync filters to URL
  const syncFiltersToURL = useCallback((overrides: Record<string, string | undefined>) => {
    const params = new URLSearchParams()
    const cId = overrides.case_id ?? caseId
    const pId = overrides.payload_id ?? payloadId
    if (cId) params.set('case_id', cId)
    if (pId) params.set('payload_id', pId)
    const f = overrides.filter ?? filter
    if (f) params.set('filter', f)
    const t = overrides.type ?? typeFilter
    if (t && t !== TYPE_FILTER_ALL) params.set('type', t)
    const st = overrides.start_time ?? startTimeFilter
    if (st) params.set('start_time', st)
    const et = overrides.end_time ?? endTimeFilter
    if (et) params.set('end_time', et)
    const v = overrides.view ?? viewMode
    if (v !== 'table') params.set('view', v)
    const qs = params.toString()
    router.replace(`/dashboard/interactions${qs ? `?${qs}` : ''}`, { scroll: false })
  }, [caseId, payloadId, filter, typeFilter, startTimeFilter, endTimeFilter, viewMode, router])

  // Set type filter from URL param if present
  useEffect(() => {
    if (typeParam) {
      setTypeFilter(typeParam)
    }
  }, [typeParam])

  // Sync filter states to URL
  useEffect(() => {
    syncFiltersToURL({})
  }, [filter, typeFilter, startTimeFilter, endTimeFilter, viewMode])

  // SSE real-time stream
  const handleNewInteraction = useCallback(() => {
    setLiveCount((c) => c + 1)
  }, [])

  const { connected: sseConnected, error: sseError } = useInteractionStream({
    caseId: caseId || undefined,
    payloadId: payloadId || undefined,
    enabled: autoRefresh,
    onInteraction: handleNewInteraction,
  })

  const clearScope = () => {
    router.push('/dashboard/interactions')
  }

  const filteredInteractions = interactions.filter(i => {
    const matchesSearch = !filter || 
      i.source_ip.includes(filter) ||
      (i.domain && i.domain.includes(filter)) ||
      (i.token && i.token.includes(filter))
    const matchesType = typeFilter === TYPE_FILTER_ALL || i.type === typeFilter
    const interactionTime = new Date(i.timestamp).getTime()
    const matchesStart = !startTimeFilter || interactionTime >= new Date(startTimeFilter).getTime()
    const matchesEnd = !endTimeFilter || interactionTime <= new Date(endTimeFilter).getTime()
    return matchesSearch && matchesType && matchesStart && matchesEnd
  })

  const handleExport = async (format: 'json' | 'csv' | 'markdown') => {
    setExporting(true)
    try {
      const response = await interactionApi.export({
        format,
        case_id: caseId,
        payload_id: payloadId,
        start_time: startTimeFilter || undefined,
        end_time: endTimeFilter || undefined,
      })
      const content = response.data as unknown as { data?: { content?: string } }
      const exportContent = content?.data?.content || ''
      if (exportContent) {
        const blob = new Blob([exportContent], {
          type: format === 'json' ? 'application/json' : format === 'csv' ? 'text/csv' : 'text/markdown',
        })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `interactions-export.${format === 'markdown' ? 'md' : format}`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
      }
    } catch (error) {
      console.error('Failed to export interactions:', error)
    } finally {
      setExporting(false)
    }
  }

  const groupedByTime = filteredInteractions.reduce((acc, interaction) => {
    const date = new Date(interaction.timestamp).toLocaleDateString()
    if (!acc[date]) {
      acc[date] = []
    }
    acc[date].push(interaction)
    return acc
  }, {} as Record<string, Interaction[]>)

  if (loading) {
    return <LoadingState />
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('interactions.title')}</h2>
          {caseId && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Case scoped: {caseId}
              <Button
                variant="ghost"
                size="sm"
                onClick={clearScope}
                className="ml-2"
              >
                Clear scope
              </Button>
            </p>
          )}
          {payloadId && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Payload scoped: {payloadId}
              <Button
                variant="ghost"
                size="sm"
                onClick={clearScope}
                className="ml-2"
              >
                Clear scope
              </Button>
            </p>
          )}
          {!caseId && !payloadId && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('interactions.all')}</p>
          )}
        </div>
        <div className="flex space-x-2">
          <Button
            variant={autoRefresh ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setAutoRefresh(!autoRefresh)
              if (!autoRefresh) setLiveCount(0)
            }}
          >
            {autoRefresh ? (sseConnected ? t('interactions.live.on') : t('interactions.live.connecting')) : t('interactions.live.off')}
          </Button>
          {autoRefresh && sseError && (
            <span className="text-xs text-amber-500">{sseError}</span>
          )}
          {autoRefresh && liveCount > 0 && (
            <span className="text-xs text-green-500">{liveCount} {t('interactions.new_count')}</span>
          )}
          <Select value={typeFilter} onValueChange={setTypeFilter}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="All types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={TYPE_FILTER_ALL}>All types</SelectItem>
              <SelectItem value="dns">DNS</SelectItem>
              <SelectItem value="http">HTTP</SelectItem>
              <SelectItem value="smtp">SMTP</SelectItem>
              <SelectItem value="ldap">LDAP</SelectItem>
              <SelectItem value="smb">SMB</SelectItem>
              <SelectItem value="ftp">FTP</SelectItem>
            </SelectContent>
          </Select>
          <Input
            type="text"
            placeholder="Search IP, domain or token..."
            className="w-64"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
          <Input
            type="datetime-local"
            className="w-48"
            value={startTimeFilter}
            onChange={(e) => setStartTimeFilter(e.target.value)}
            title="Start time filter"
          />
          <Input
            type="datetime-local"
            className="w-48"
            value={endTimeFilter}
            onChange={(e) => setEndTimeFilter(e.target.value)}
            title="End time filter"
          />
          <Button
            variant="outline"
            size="sm"
            disabled={exporting}
            onClick={() => handleExport('csv')}
          >
            {exporting ? 'Exporting...' : 'Export CSV'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={exporting}
            onClick={() => handleExport('json')}
          >
            Export JSON
          </Button>
          <Button
            onClick={() => setViewMode(viewMode === 'table' ? 'timeline' : 'table')}
          >
            {viewMode === 'table' ? 'Timeline View' : 'Table View'}
          </Button>
        </div>
      </div>

      {/* Stats Card */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>{t('interactions.statistics')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-5 gap-4">
            <div className="text-center">
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.total}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">{t('interactions.total')}</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-purple-600">{stats.dns_count}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">DNS</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-blue-600">{stats.http_count}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">HTTP</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-green-600">{stats.smtp_count}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">SMTP</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-yellow-600">{stats.ldap_count}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">LDAP</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Interactions</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredInteractions.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">
              {caseId || payloadId ? 'No interactions for this Case/Payload' : t('interactions.no_data')}
            </p>
          ) : viewMode === 'table' ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Source IP</TableHead>
                  <TableHead>Details</TableHead>
                  <TableHead>Time</TableHead>
                  <TableHead>Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredInteractions.map((interaction) => (
                  <TableRow key={interaction.id}>
                    <TableCell>
                      <Badge variant={
                        interaction.type === 'dns' ? 'default' :
                        interaction.type === 'http' ? 'secondary' :
                        interaction.type === 'smtp' ? 'outline' :
                        interaction.type === 'ldap' ? 'outline' :
                        interaction.type === 'smb' ? 'outline' :
                        interaction.type === 'ftp' ? 'destructive' :
                        'outline'
                      }>
                        {interaction.type.toUpperCase()}
                      </Badge>
                    </TableCell>
                    <TableCell>{interaction.source_ip}</TableCell>
                    <TableCell>
                      {interaction.domain && <div>Domain: {interaction.domain}</div>}
                      {interaction.method && interaction.path && (
                        <div>{interaction.method} {interaction.path}</div>
                      )}
                      {interaction.token && <div>Token: {interaction.token}</div>}
                      {interaction.user_agent && (
                        <div className="text-xs text-gray-400 truncate max-w-md">
                          UA: {interaction.user_agent}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>{new Date(interaction.timestamp).toLocaleString()}</TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedInteraction(interaction)}
                      >
                        Details
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="space-y-6">
              {Object.entries(groupedByTime).map(([date, dayInteractions]) => (
                <div key={date}>
                  <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3">{date}</h3>
                  <div className="border-l-2 border-indigo-200 pl-4 space-y-4">
                    {dayInteractions.map((interaction) => (
                      <div
                        key={interaction.id}
                        className="relative"
                      >
                        <div className="absolute -left-6 mt-1 w-4 h-4 bg-indigo-600 rounded-full"></div>
                        <div
                          className="bg-gray-50 dark:bg-gray-900 p-4 rounded cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                          onClick={() => setSelectedInteraction(interaction)}
                        >
                          <div className="flex justify-between items-start">
                            <div>
                              <Badge variant={
                                interaction.type === 'dns' ? 'default' :
                                interaction.type === 'http' ? 'secondary' :
                                interaction.type === 'smtp' ? 'outline' :
                                interaction.type === 'ldap' ? 'outline' :
                                interaction.type === 'smb' ? 'outline' :
                                interaction.type === 'ftp' ? 'destructive' :
                                'outline'
                              }>
                                {interaction.type.toUpperCase()}
                              </Badge>
                              <span className="ml-2 text-sm text-gray-600 dark:text-gray-400">{interaction.source_ip}</span>
                            </div>
                            <span className="text-xs text-gray-400">
                              {new Date(interaction.timestamp).toLocaleTimeString()}
                            </span>
                          </div>
                          {interaction.domain && (
                            <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">Domain: {interaction.domain}</p>
                          )}
                          {interaction.token && (
                            <p className="text-sm text-gray-500 dark:text-gray-400">Token: {interaction.token}</p>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Detail Drawer */}
      <Dialog open={!!selectedInteraction} onOpenChange={() => setSelectedInteraction(null)}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Interaction Triage</DialogTitle>
          </DialogHeader>
          {selectedInteraction && (
            <div className="space-y-4">
              {/* Basic Info */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Type</p>
                  <p className="text-gray-900 dark:text-gray-100">{selectedInteraction.type.toUpperCase()}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Timestamp</p>
                  <p className="text-gray-900 dark:text-gray-100">{new Date(selectedInteraction.timestamp).toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Source IP</p>
                  <p className="text-gray-900 dark:text-gray-100">{selectedInteraction.source_ip}</p>
                </div>
                {selectedInteraction.token && (
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Token</p>
                    <div className="flex items-center gap-2">
                      <p className="text-gray-900 dark:text-gray-100 break-all">{selectedInteraction.token}</p>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigator.clipboard.writeText(selectedInteraction.token!)}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                )}
              </div>

              {/* Attribution Info */}
              {(selectedInteraction.case_id || selectedInteraction.payload_id) && (
                <div className="border-t pt-4">
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">Attribution</p>
                  <div className="space-y-2">
                    {selectedInteraction.case_id && (
                      <div className="flex items-center justify-between">
                        <p className="text-sm text-gray-600 dark:text-gray-400">Case ID: {selectedInteraction.case_id}</p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => router.push(`/dashboard/cases/${selectedInteraction.case_id}`)}
                        >
                          View Case
                        </Button>
                      </div>
                    )}
                    {selectedInteraction.payload_id && (
                      <div className="flex items-center justify-between">
                        <p className="text-sm text-gray-600 dark:text-gray-400">Payload ID: {selectedInteraction.payload_id}</p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => router.push(`/dashboard/payloads/${selectedInteraction.payload_id}`)}
                        >
                          View Payload
                        </Button>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Protocol Details */}
              <div className="border-t pt-4">
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">Protocol Details</p>
                {selectedInteraction.domain && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Domain</p>
                    <div className="flex items-center gap-2">
                      <p className="text-gray-900 dark:text-gray-100">{selectedInteraction.domain}</p>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigator.clipboard.writeText(selectedInteraction.domain!)}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                )}
                {selectedInteraction.method && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Method</p>
                    <p className="text-gray-900 dark:text-gray-100">{selectedInteraction.method}</p>
                  </div>
                )}
                {selectedInteraction.path && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Path</p>
                    <div className="flex items-center gap-2">
                      <p className="text-gray-900 dark:text-gray-100 break-all">{selectedInteraction.path}</p>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigator.clipboard.writeText(selectedInteraction.path!)}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                )}
                {selectedInteraction.user_agent && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">User Agent</p>
                    <p className="text-gray-900 dark:text-gray-100 break-all text-sm">{selectedInteraction.user_agent}</p>
                  </div>
                )}
                {selectedInteraction.headers && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Headers</p>
                    <pre className="text-gray-900 dark:text-gray-100 bg-gray-50 dark:bg-gray-900 p-2 rounded text-xs overflow-auto max-h-40">{JSON.stringify(selectedInteraction.headers, null, 2)}</pre>
                  </div>
                )}
                {selectedInteraction.body && (
                  <div className="mb-2">
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Body</p>
                    <pre className="text-gray-900 dark:text-gray-100 bg-gray-50 dark:bg-gray-900 p-2 rounded text-xs overflow-auto max-h-40">{selectedInteraction.body}</pre>
                  </div>
                )}
              </div>

              {/* Quick Actions */}
              <div className="border-t pt-4">
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">Quick Actions</p>
                <div className="flex flex-wrap gap-2">
                  {selectedInteraction.case_id && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => router.push(`/dashboard/evidence?case_id=${selectedInteraction.case_id}`)}
                    >
                      Generate Evidence (Case)
                    </Button>
                  )}
                  {selectedInteraction.payload_id && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => router.push(`/dashboard/evidence?payload_id=${selectedInteraction.payload_id}`)}
                    >
                      Generate Evidence (Payload)
                    </Button>
                  )}
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default function InteractionsPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <InteractionsPageContent />
    </Suspense>
  )
}
