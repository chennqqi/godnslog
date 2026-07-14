'use client'

import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

export interface LineChartDataPoint {
  date: string
  count: number
}

export interface LineChartProps {
  data: LineChartDataPoint[]
  height?: number
  color?: string
  className?: string
}

export function LineChart({
  data,
  height = 300,
  color = '#3b82f6',
  className,
}: LineChartProps) {
  const hasData = data.some((d) => d.count > 0)
  if (!hasData) {
    return (
      <div
        className={`flex items-center justify-center text-sm text-gray-400 ${className ?? ''}`}
        style={{ height }}
      >
        No data
      </div>
    )
  }
  return (
    <div className={className} style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <RechartsLineChart data={data} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" vertical={false} />
          <XAxis
            dataKey="date"
            stroke="#9ca3af"
            tick={{ fontSize: 11 }}
            tickFormatter={(v: string) => v.slice(5)}
          />
          <YAxis stroke="#9ca3af" tick={{ fontSize: 11 }} allowDecimals={false} />
          <Tooltip
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: 4, color: '#fff' }}
          />
          <Line
            type="monotone"
            dataKey="count"
            stroke={color}
            strokeWidth={2}
            dot={{ r: 3, fill: color }}
            activeDot={{ r: 5 }}
          />
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  )
}
