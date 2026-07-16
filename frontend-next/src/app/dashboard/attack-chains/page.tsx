'use client'

import { useState } from 'react'
import { AttackChainList } from '@/features/interactions/attack-chain-list'
import { AttackChainDetail } from '@/features/interactions/attack-chain-detail'

export default function AttackChainsPage() {
  const [selectedToken, setSelectedToken] = useState<string | null>(null)

  if (selectedToken) {
    return <AttackChainDetail token={selectedToken} onBack={() => setSelectedToken(null)} />
  }

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Attack Chains</h1>
        <p className="text-sm text-gray-500 mt-1">
          Interactions grouped by token showing the full attack timeline
        </p>
      </div>
      <AttackChainList onSelectToken={setSelectedToken} />
    </div>
  )
}
