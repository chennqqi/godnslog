'use client'

import { ReactNode } from 'react'

export interface TimelineItem {
  id: string
  date: string
  time: string
  title: string
  description?: string
  icon?: ReactNode
  content?: ReactNode
  onClick?: () => void
}

interface TimelineProps {
  items: TimelineItem[]
  groupByDate?: boolean
}

export function Timeline({ items, groupByDate = true }: TimelineProps) {
  if (groupByDate) {
    const grouped = items.reduce((acc, item) => {
      if (!acc[item.date]) {
        acc[item.date] = []
      }
      acc[item.date].push(item)
      return acc
    }, {} as Record<string, TimelineItem[]>)

    return (
      <div className="space-y-6">
        {Object.entries(grouped).map(([date, dayItems]) => (
          <div key={date}>
            <h3 className="text-lg font-medium text-gray-900 mb-3">{date}</h3>
            <div className="border-l-2 border-indigo-200 pl-4 space-y-4">
              {dayItems.map((item) => (
                <TimelineItemComponent key={item.id} item={item} />
              ))}
            </div>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="border-l-2 border-indigo-200 pl-4 space-y-4">
      {items.map((item) => (
        <TimelineItemComponent key={item.id} item={item} />
      ))}
    </div>
  )
}

function TimelineItemComponent({ item }: { item: TimelineItem }) {
  return (
    <div className="relative">
      <div className="absolute -left-6 mt-1 w-4 h-4 bg-indigo-600 rounded-full"></div>
      <div
        className="bg-gray-50 p-4 rounded cursor-pointer hover:bg-gray-100"
        onClick={item.onClick}
      >
        <div className="flex justify-between items-start mb-2">
          <div className="flex items-center space-x-2">
            {item.icon && <div className="text-indigo-600">{item.icon}</div>}
            <span className="font-medium text-gray-900">{item.title}</span>
          </div>
          <span className="text-xs text-gray-400">{item.time}</span>
        </div>
        {item.description && (
          <p className="text-sm text-gray-500">{item.description}</p>
        )}
        {item.content && <div className="mt-2">{item.content}</div>}
      </div>
    </div>
  )
}
