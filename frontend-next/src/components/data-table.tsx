'use client'

import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export interface Column<T> {
  key: keyof T
  label: string
  render?: (value: any, row: T) => React.ReactNode
}

export interface DataTableProps<T> {
  data: T[]
  columns: Column<T>[]
  searchable?: boolean
  filterable?: boolean
  filterKey?: keyof T
  filterOptions?: { value: string; label: string }[]
  onRowClick?: (row: T) => void
  emptyMessage?: string
}

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  searchable = false,
  filterable = false,
  filterKey,
  filterOptions,
  onRowClick,
  emptyMessage = '暂无数据',
}: DataTableProps<T>) {
  const [searchTerm, setSearchTerm] = useState('')
  /** Radix SelectItem cannot use value="" */
  const [filterValue, setFilterValue] = useState('all')

  const filteredData = data.filter((row) => {
    const matchesSearch = !searchTerm || columns.some((col) => {
      const value = row[col.key]
      return String(value).toLowerCase().includes(searchTerm.toLowerCase())
    })
    
    const matchesFilter =
      filterValue === 'all' || !filterKey || row[filterKey] === filterValue
    
    return matchesSearch && matchesFilter
  })

  return (
    <div className="space-y-4">
      {(searchable || filterable) && (
        <div className="flex space-x-4">
          {searchable && (
            <Input
              placeholder="搜索..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="max-w-sm"
            />
          )}
          {filterable && filterKey && filterOptions && (
            <Select value={filterValue} onValueChange={setFilterValue}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部</SelectItem>
                {filterOptions.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((col) => (
                <TableHead key={String(col.key)}>{col.label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredData.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="text-center text-gray-500 py-8">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            ) : (
              filteredData.map((row, index) => (
                <TableRow
                  key={index}
                  className={onRowClick ? 'cursor-pointer hover:bg-gray-50' : ''}
                  onClick={() => onRowClick?.(row)}
                >
                  {columns.map((col) => (
                    <TableCell key={String(col.key)}>
                      {col.render ? col.render(row[col.key], row) : String(row[col.key])}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
