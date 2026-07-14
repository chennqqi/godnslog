'use client'

import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts'

export interface DonutChartDataPoint {
  name: string
  value: number
  color?: string
}

export interface DonutChartProps {
  data: DonutChartDataPoint[]
  size?: number
  innerRadius?: number
  outerRadius?: number
  className?: string
}

const DEFAULT_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#f97316', '#6b7280']

export function DonutChart({
  data,
  size = 200,
  innerRadius = 60,
  outerRadius = 80,
  className,
}: DonutChartProps) {
  const hasData = data.some((d) => d.value > 0)
  if (!hasData) {
    return (
      <div
        className={`flex items-center justify-center text-sm text-gray-400 ${className ?? ''}`}
        style={{ height: size }}
      >
        No data
      </div>
    )
  }
  return (
    <div className={className} style={{ height: size }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            innerRadius={innerRadius}
            outerRadius={outerRadius}
            paddingAngle={2}
          >
            {data.map((entry, idx) => (
              <Cell
                key={entry.name}
                fill={entry.color ?? DEFAULT_COLORS[idx % DEFAULT_COLORS.length]}
              />
            ))}
          </Pie>
          <Tooltip
            formatter={(value, name) => [value, name]}
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: 4, color: '#fff' }}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}
