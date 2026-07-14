'use client'

import { useMemo, useState } from 'react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { useUpdateCase } from '@/features/cases/hooks/use-cases'
import type { Case } from '@/types'
import { KanbanColumn } from './kanban-column'
import { KanbanCard } from './kanban-card'

export interface KanbanBoardProps {
  cases: Case[]
  onCardClick?: (id: string) => void
  labels?: {
    active?: string
    completed?: string
    archived?: string
    empty?: string
  }
}

const COLUMNS = [
  { id: 'active', accentColor: '#3b82f6' },
  { id: 'completed', accentColor: '#10b981' },
  { id: 'archived', accentColor: '#6b7280' },
] as const

export function KanbanBoard({ cases, onCardClick, labels }: KanbanBoardProps) {
  const updateCase = useUpdateCase()
  const [activeId, setActiveId] = useState<string | null>(null)
  const [localCases, setLocalCases] = useState<Case[]>(cases)

  // Sync external cases prop into local state for optimistic updates.
  useMemo(() => {
    setLocalCases(cases)
  }, [cases])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  const activeCase = activeId ? localCases.find((c) => c.id === activeId) : null

  const grouped = useMemo(() => {
    const g: Record<string, Case[]> = { active: [], completed: [], archived: [] }
    for (const c of localCases) {
      if (g[c.status]) g[c.status].push(c)
    }
    return g
  }, [localCases])

  function handleDragStart(event: DragStartEvent) {
    setActiveId(String(event.active.id))
  }

  function handleDragEnd(event: DragEndEvent) {
    setActiveId(null)
    const { active, over } = event
    if (!over) return

    const newStatus = String(over.id)
    const movedCase = localCases.find((c) => c.id === String(active.id))
    if (!movedCase || movedCase.status === newStatus) return

    // Optimistic update.
    setLocalCases((prev) =>
      prev.map((c) => (c.id === movedCase.id ? { ...c, status: newStatus as Case['status'] } : c))
    )

    // Persist; rollback on error.
    updateCase.mutate(
      { id: movedCase.id, data: { status: newStatus as 'active' | 'completed' | 'archived' } },
      {
        onError: () => {
          setLocalCases((prev) =>
            prev.map((c) => (c.id === movedCase.id ? { ...c, status: movedCase.status } : c))
          )
        },
      }
    )
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-2">
        {COLUMNS.map((col) => (
          <KanbanColumn
            key={col.id}
            id={col.id}
            title={labels?.[col.id] ?? col.id}
            accentColor={col.accentColor}
            cases={grouped[col.id]}
            onCardClick={onCardClick}
            emptyLabel={labels?.empty}
          />
        ))}
      </div>
      <DragOverlay>
        {activeCase ? <KanbanCard case={activeCase} /> : null}
      </DragOverlay>
    </DndContext>
  )
}
