'use client'

import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Calendar } from 'lucide-react'
import type { Case } from '@/types'

export interface KanbanCardProps {
  case: Case
  onClick?: (id: string) => void
}

export function KanbanCard({ case: c, onClick }: KanbanCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: c.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={() => onClick?.(c.id)}
      className="bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow"
    >
      <p className="font-medium text-sm text-gray-900 dark:text-gray-100 line-clamp-2 mb-1">
        {c.title}
      </p>
      {c.target && (
        <p className="text-xs text-gray-500 dark:text-gray-400 truncate mb-2">{c.target}</p>
      )}
      <div className="flex items-center justify-between">
        <div className="flex flex-wrap gap-1">
          {c.tags?.slice(0, 2).map((tag) => (
            <span
              key={tag}
              className="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-600 text-gray-600 dark:text-gray-300"
            >
              {tag}
            </span>
          ))}
        </div>
        <span className="flex items-center gap-1 text-xs text-gray-400">
          <Calendar className="w-3 h-3" />
          {new Date(c.created_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  )
}
