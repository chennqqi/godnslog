'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { interactionApi } from '@/lib/api-client'
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

/** Radix SelectItem cannot use value="" for "all types" */
const TYPE_FILTER_ALL = 'all'

export default function InteractionsPage() {
  const router = useRouter()
  const [interactions, setInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState(TYPE_FILTER_ALL)
  const [viewMode, setViewMode] = useState<'table' | 'timeline'>('table')
  const [selectedInteraction, setSelectedInteraction] = useState<Interaction | null>(null)
  const [stats, setStats] = useState({ total: 0, dns_count: 0, http_count: 0, smtp_count: 0, ldap_count: 0 })
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [refreshInterval, setRefreshInterval] = useState(5000)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadInteractions()
    loadStats()
  }, [router])

  // Auto-refresh polling
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      loadInteractions()
      loadStats()
    }, refreshInterval)

    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval])

  const loadInteractions = async () => {
    try {
      const response = await interactionApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setInteractions(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load interactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await interactionApi.stats()
      if (response.code === 0 && response.data) {
        setStats({
          total: response.data.total ?? 0,
          dns_count: response.data.dns_count ?? 0,
          http_count: response.data.http_count ?? 0,
          smtp_count: response.data.smtp_count ?? 0,
          ldap_count: response.data.ldap_count ?? 0,
        })
      }
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }

  const filteredInteractions = interactions.filter(i => {
    const matchesSearch = !filter || 
      i.source_ip.includes(filter) ||
      (i.domain && i.domain.includes(filter)) ||
      (i.token && i.token.includes(filter))
    const matchesType = typeFilter === TYPE_FILTER_ALL || i.type === typeFilter
    return matchesSearch && matchesType
  })

  const groupedByTime = filteredInteractions.reduce((acc, interaction) => {
    const date = new Date(interaction.timestamp).toLocaleDateString()
    if (!acc[date]) {
      acc[date] = []
    }
    acc[date].push(interaction)
    return acc
  }, {} as Record<string, Interaction[]>)

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Interaction Timeline</h2>
        <div className="flex space-x-2">
          <Button
            variant={autoRefresh ? "default" : "outline"}
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            {autoRefresh ? "自动刷新: 开" : "自动刷新: 关"}
          </Button>
          {autoRefresh && (
            <Select value={refreshInterval.toString()} onValueChange={(v) => setRefreshInterval(parseInt(v))}>
              <SelectTrigger className="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="3000">3秒</SelectItem>
                <SelectItem value="5000">5秒</SelectItem>
                <SelectItem value="10000">10秒</SelectItem>
                <SelectItem value="30000">30秒</SelectItem>
              </SelectContent>
            </Select>
          )}
          <Select value={typeFilter} onValueChange={setTypeFilter}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="所有类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={TYPE_FILTER_ALL}>所有类型</SelectItem>
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
            placeholder="搜索 IP、域名或token..."
            className="w-64"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
          <Button
            onClick={() => setViewMode(viewMode === 'table' ? 'timeline' : 'table')}
          >
            {viewMode === 'table' ? '时间线视图' : '表格视图'}
          </Button>
        </div>
      </div>

      {/* Stats Card */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>统计信息</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-5 gap-4">
            <div className="text-center">
              <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
              <p className="text-sm text-gray-500">总数</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-purple-600">{stats.dns_count}</p>
              <p className="text-sm text-gray-500">DNS</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-blue-600">{stats.http_count}</p>
              <p className="text-sm text-gray-500">HTTP</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-green-600">{stats.smtp_count}</p>
              <p className="text-sm text-gray-500">SMTP</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-yellow-600">{stats.ldap_count}</p>
              <p className="text-sm text-gray-500">LDAP</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>交互记录</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredInteractions.length === 0 ? (
            <p className="text-gray-500">暂无命中记录</p>
          ) : viewMode === 'table' ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>类型</TableHead>
                  <TableHead>来源IP</TableHead>
                  <TableHead>详情</TableHead>
                  <TableHead>时间</TableHead>
                  <TableHead>操作</TableHead>
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
                      {interaction.domain && <div>域名: {interaction.domain}</div>}
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
                        详情
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
                  <h3 className="text-lg font-medium text-gray-900 mb-3">{date}</h3>
                  <div className="border-l-2 border-indigo-200 pl-4 space-y-4">
                    {dayInteractions.map((interaction) => (
                      <div
                        key={interaction.id}
                        className="relative"
                      >
                        <div className="absolute -left-6 mt-1 w-4 h-4 bg-indigo-600 rounded-full"></div>
                        <div
                          className="bg-gray-50 p-4 rounded cursor-pointer hover:bg-gray-100"
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
                              <span className="ml-2 text-sm text-gray-600">{interaction.source_ip}</span>
                            </div>
                            <span className="text-xs text-gray-400">
                              {new Date(interaction.timestamp).toLocaleTimeString()}
                            </span>
                          </div>
                          {interaction.domain && (
                            <p className="text-sm text-gray-500 mt-2">域名: {interaction.domain}</p>
                          )}
                          {interaction.token && (
                            <p className="text-sm text-gray-500">Token: {interaction.token}</p>
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
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>交互详情</DialogTitle>
          </DialogHeader>
          {selectedInteraction && (
            <div className="space-y-4">
              <div>
                <p className="text-sm font-medium text-gray-500">类型</p>
                <p className="text-gray-900">{selectedInteraction.type.toUpperCase()}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-gray-500">来源 IP</p>
                <p className="text-gray-900">{selectedInteraction.source_ip}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-gray-500">时间戳</p>
                <p className="text-gray-900">{new Date(selectedInteraction.timestamp).toLocaleString()}</p>
              </div>
              {selectedInteraction.domain && (
                <div>
                  <p className="text-sm font-medium text-gray-500">域名</p>
                  <p className="text-gray-900">{selectedInteraction.domain}</p>
                </div>
              )}
              {selectedInteraction.token && (
                <div>
                  <p className="text-sm font-medium text-gray-500">Token</p>
                  <p className="text-gray-900">{selectedInteraction.token}</p>
                </div>
              )}
              {selectedInteraction.method && (
                <div>
                  <p className="text-sm font-medium text-gray-500">方法</p>
                  <p className="text-gray-900">{selectedInteraction.method}</p>
                </div>
              )}
              {selectedInteraction.path && (
                <div>
                  <p className="text-sm font-medium text-gray-500">路径</p>
                  <p className="text-gray-900 break-all">{selectedInteraction.path}</p>
                </div>
              )}
              {selectedInteraction.user_agent && (
                <div>
                  <p className="text-sm font-medium text-gray-500">User Agent</p>
                  <p className="text-gray-900 break-all">{selectedInteraction.user_agent}</p>
                </div>
              )}
              {selectedInteraction.body && (
                <div>
                  <p className="text-sm font-medium text-gray-500">Body</p>
                  <pre className="text-gray-900 bg-gray-50 p-2 rounded text-xs overflow-auto max-h-40">{selectedInteraction.body}</pre>
                </div>
              )}
              {selectedInteraction.headers && (
                <div>
                  <p className="text-sm font-medium text-gray-500">Headers</p>
                  <pre className="text-gray-900 bg-gray-50 p-2 rounded text-xs overflow-auto max-h-40">{JSON.stringify(selectedInteraction.headers, null, 2)}</pre>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
