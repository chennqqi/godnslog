'use client'

import { useEffect, useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { Timeline, TimelineItem } from '@/components/timeline'
import { LoadingState } from '@/components/loading-state'

interface Interaction {
  id: string
  type: string
  timestamp: string
  source_ip: string
  domain?: string
  path?: string
  method?: string
  decoded_data?: string
  exploit_type?: string
  confidence?: string
}

interface AttackChainDetail {
  token: string
  interaction_count: number
  protocols: string[]
  exploit_types: string[]
  first_seen: string
  last_seen: string
  confidence: string
  interactions: Interaction[]
}

const typeIcons: Record<string, string> = {
  dns: '🔍',
  http: '🌐',
  ldap: '📂',
  smtp: '📧',
  smb: '📁',
  ftp: '📄',
}

interface AttackChainDetailProps {
  token: string
  onBack: () => void
}

export function AttackChainDetail({ token, onBack }: AttackChainDetailProps) {
  const [detail, setDetail] = useState<AttackChainDetail | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    apiClient.get<AttackChainDetail>(`/api/v2/attack-chains/${encodeURIComponent(token)}`)
      .then((data) => setDetail(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [token])

  if (loading) return <LoadingState />
  if (!detail) return <div className="text-red-500">Chain not found</div>

  const timelineItems: TimelineItem[] = detail.interactions.map((interaction) => {
    const date = new Date(interaction.timestamp)
    const dateStr = date.toLocaleDateString()
    const timeStr = date.toLocaleTimeString()
    const icon = typeIcons[interaction.type] || '❓'
    const title = interaction.domain || interaction.path || interaction.source_ip

    return {
      id: interaction.id,
      date: dateStr,
      time: timeStr,
      title: `${icon} ${interaction.type.toUpperCase()}${interaction.exploit_type ? ` [${interaction.exploit_type}]` : ''}`,
      description: title,
      content: (
        <div className="text-xs space-y-1">
          <div className="text-gray-500">Source: {interaction.source_ip}</div>
          {interaction.decoded_data && (
            <div className="text-green-700 font-mono bg-green-50 p-1 rounded">
              Decoded: {interaction.decoded_data}
            </div>
          )}
          {interaction.confidence && (
            <div className={interaction.confidence === 'high' ? 'text-red-600' : 'text-yellow-600'}>
              Confidence: {interaction.confidence}
            </div>
          )}
        </div>
      ),
    }
  })

  return (
    <div className="space-y-4">
      <button onClick={onBack} className="text-sm text-indigo-600 hover:underline">
        &larr; Back to chains
      </button>

      <div className="border rounded-lg p-4 bg-gray-50">
        <h2 className="font-mono text-lg font-medium break-all">{detail.token}</h2>
        <div className="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-sm text-gray-500">
          <span>{detail.interaction_count} interactions</span>
          <span>Protocols: {detail.protocols.join(', ')}</span>
          {detail.exploit_types.length > 0 && (
            <span>Types: {detail.exploit_types.join(', ')}</span>
          )}
          <span className={detail.confidence === 'high' ? 'text-red-600 font-medium' : ''}>
            Confidence: {detail.confidence || 'unknown'}
          </span>
        </div>
      </div>

      <Timeline items={timelineItems} groupByDate={false} />
    </div>
  )
}
