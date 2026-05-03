'use client'

import { useEffect, useState } from 'react'
import { interactionApi } from '@/lib/api-client'
import type { Interaction } from '@/types'

export default function InteractionsPage() {
  const [interactions, setInteractions] = useState<Interaction[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  useEffect(() => {
    loadInteractions()
  }, [])

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
          </select>
          <input
            type="text"
            placeholder="搜索 IP、域名或token..."
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </div>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {filteredInteractions.length === 0 ? (
            <p className="text-gray-500">暂无命中记录</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {filteredInteractions.map((interaction) => (
                <li key={interaction.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <span className={`px-2 py-1 text-xs rounded font-medium ${
                          interaction.type === 'dns' ? 'bg-purple-100 text-purple-800' :
                          interaction.type === 'http' ? 'bg-blue-100 text-blue-800' :
                          interaction.type === 'smtp' ? 'bg-green-100 text-green-800' :
                          interaction.type === 'ldap' ? 'bg-yellow-100 text-yellow-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {interaction.type.toUpperCase()}
                        </span>
                        <p className="text-sm font-medium text-gray-900">{interaction.source_ip}</p>
                      </div>
                      {interaction.domain && (
                        <p className="text-sm text-gray-600 mt-1">域名: {interaction.domain}</p>
                      )}
                      {interaction.method && interaction.path && (
                        <p className="text-sm text-gray-600 mt-1">
                          {interaction.method} {interaction.path}
                        </p>
                      )}
                      {interaction.token && (
                        <p className="text-xs text-gray-500 mt-1">Token: {interaction.token}</p>
                      )}
                      {interaction.user_agent && (
                        <p className="text-xs text-gray-400 mt-1 truncate max-w-md">
                          UA: {interaction.user_agent}
                        </p>
                      )}
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-500">
                        {new Date(interaction.timestamp).toLocaleString()}
                      </p>
                    </div>
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
