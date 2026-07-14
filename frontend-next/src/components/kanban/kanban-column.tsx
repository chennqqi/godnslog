'use client'

import { useDroppable } from '@dnd-kit/core'
import type { Case } from '@/types'
import { KanbanCard } from './kanban-card'

export interface KanbanColumnProps {
  id: string
  title: string
  accentColor: string
  cases: Case[]
  onCardClick?: (id: string) => void
  emptyLabel?: string
}

export function KanbanColumn({
  id,
  title,
  accentColor,
  cases,
  onCardClick,
  emptyLabel = 'No cases',
}: KanbanColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id })

  return (
    <div className="flex flex-col min-w-[260px] w-72 bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3">
      <div className="flex items-center gap-2 mb-3 px-1">
        <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: accentColor }} />
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">{title}</h3>
        <span className="ml-auto text-xs text-gray-400">{cases.length}</span>
      </div>
      <div
        ref={setNodeRef}
        className={`flex-1 space-y-2 min-h-[120px] rounded-md p-1 transition-colors ${
          isOver ? 'bg-indigo-50 dark:bg-indigo-900/30' : ''
        }`}
      >
        {cases.length === 0 ? (
          <p className="text-xs text-gray-400 text-center py-8">{emptyLabel}</p>
        ) : (
          cases.map((c) => <KanbanCard key={c.id} case={c} onClick={onCardClick} />)
        )}
      </div>
    </div>
  )
}
