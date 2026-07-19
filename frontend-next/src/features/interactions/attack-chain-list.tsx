'use client'

import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { LoadingState } from '@/components/loading-state'
import { EmptyState } from '@/components/empty-state'

interface AttackChain {
  token: string
  interaction_count: number
  protocols: string[]
  exploit_types: string[]
  first_seen: string
  last_seen: string
  confidence: string
}

interface AttackChainListResponse {
  items: AttackChain[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

interface AttackChainListProps {
  onSelectToken?: (token: string) => void
}

const protocolColors: Record<string, string> = {
  dns: 'bg-blue-100 text-blue-800',
  http: 'bg-green-100 text-green-800',
  ldap: 'bg-purple-100 text-purple-800',
  smtp: 'bg-yellow-100 text-yellow-800',
  smb: 'bg-orange-100 text-orange-800',
  ftp: 'bg-pink-100 text-pink-800',
}

const confidenceColors: Record<string, string> = {
  high: 'text-red-600 font-medium',
  medium: 'text-yellow-600 font-medium',
  low: 'text-gray-500',
}

export function AttackChainList({ onSelectToken }: AttackChainListProps) {
  const [chains, setChains] = useState<AttackChain[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const pageSize = 20

  useEffect(() => {
    let cancelled = false
    const id = requestAnimationFrame(() => {
      setLoading(true)
      setError(null)
    })
    apiClient.get<AttackChainListResponse>(`/api/v2/attack-chains?page=${page}&page_size=${pageSize}`)
      .then((data) => {
        if (cancelled) return
        setChains(data.items || [])
        setTotalPages(data.total_pages || 1)
      })
      .catch((err: Error) => { if (!cancelled) setError(err.message) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true; cancelAnimationFrame(id) }
  }, [page])

  if (loading) return <LoadingState />
  if (error) return <div className="text-red-500">Error: {error}</div>
  if (chains.length === 0) return <EmptyState title="No Attack Chains" description="No interactions with tokens found yet." />

  return (
    <div className="space-y-3">
      {chains.map((chain) => (
        <div
          key={chain.token}
          className="border rounded-lg p-4 hover:shadow-md cursor-pointer transition-shadow"
          onClick={() => onSelectToken?.(chain.token)}
        >
          <div className="flex justify-between items-start">
            <div className="flex-1 min-w-0">
              <h3 className="font-mono text-sm font-medium truncate">{chain.token}</h3>
              <div className="flex flex-wrap gap-1.5 mt-2">
                {chain.protocols.map((p) => (
                  <span key={p} className={`text-xs px-2 py-0.5 rounded ${protocolColors[p] || 'bg-gray-100 text-gray-800'}`}>
                    {p.toUpperCase()}
                  </span>
                ))}
                {chain.exploit_types.map((et) => (
                  <span key={et} className="text-xs px-2 py-0.5 rounded bg-red-50 text-red-700 border border-red-200">
                    {et}
                  </span>
                ))}
              </div>
            </div>
            <div className="text-right text-xs ml-4">
              <span className="text-gray-400">{chain.interaction_count} hits</span>
              <div className={confidenceColors[chain.confidence] || ''}>
                {chain.confidence || 'unknown'}
              </div>
            </div>
          </div>
          <div className="text-xs text-gray-400 mt-2">
            {new Date(chain.first_seen).toLocaleString()} ~ {new Date(chain.last_seen).toLocaleString()}
          </div>
        </div>
      ))}

      {totalPages > 1 && (
        <div className="flex justify-center items-center gap-2 mt-4">
          <button
            className="px-3 py-1 border rounded text-sm disabled:opacity-50 hover:bg-gray-50"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            Previous
          </button>
          <span className="text-sm text-gray-500">
            {page} / {totalPages}
          </span>
          <button
            className="px-3 py-1 border rounded text-sm disabled:opacity-50 hover:bg-gray-50"
            disabled={page >= totalPages}
            onClick={() => setPage(page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
