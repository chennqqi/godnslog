'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { interactionApi } from '@/lib/api-client'
import type { Interaction } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export default function InteractionsPage() {
  const router = useRouter()
  const [interactions, setInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadInteractions()
  }, [router])

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

  const filteredInteractions = interactions.filter(i => {
    const matchesSearch = !filter || 
      i.source_ip.includes(filter) ||
      (i.domain && i.domain.includes(filter)) ||
      (i.token && i.token.includes(filter))
    const matchesType = !typeFilter || i.type === typeFilter
    return matchesSearch && matchesType
  })

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Interaction Timeline</h2>
        <div className="flex space-x-2">
          <select
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
          >
            <option value="">所有类型</option>
            <option value="dns">DNS</option>
            <option value="http">HTTP</option>
            <option value="smtp">SMTP</option>
            <option value="ldap">LDAP</option>
            <option value="smb">SMB</option>
            <option value="ftp">FTP</option>
          </select>
          <Input
            type="text"
            placeholder="搜索 IP、域名或token..."
            className="w-64"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>交互记录</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredInteractions.length === 0 ? (
            <p className="text-gray-500">暂无命中记录</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>类型</TableHead>
                  <TableHead>来源IP</TableHead>
                  <TableHead>详情</TableHead>
                  <TableHead>时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredInteractions.map((interaction) => (
                  <TableRow key={interaction.id}>
                    <TableCell>
                      <span className={`px-2 py-1 text-xs rounded font-medium ${
                        interaction.type === 'dns' ? 'bg-purple-100 text-purple-800' :
                        interaction.type === 'http' ? 'bg-blue-100 text-blue-800' :
                        interaction.type === 'smtp' ? 'bg-green-100 text-green-800' :
                        interaction.type === 'ldap' ? 'bg-yellow-100 text-yellow-800' :
                        interaction.type === 'smb' ? 'bg-orange-100 text-orange-800' :
                        interaction.type === 'ftp' ? 'bg-red-100 text-red-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {interaction.type.toUpperCase()}
                      </span>
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
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
