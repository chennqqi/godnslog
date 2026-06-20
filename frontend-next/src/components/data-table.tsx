'use client'

import { useState, useMemo, type ReactNode } from 'react'
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
import { ChevronUp, ChevronDown, ChevronLeft, ChevronRight, ChevronsUpDown } from 'lucide-react'

export interface Column<T> {
  key: keyof T
  label: string
  render?: (value: unknown, row: T) => ReactNode
  sortable?: boolean
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
  pageSize?: number
  selectable?: boolean
  onSelectionChange?: (selectedIds: string[]) => void
  getRowId?: (row: T) => string
}

type SortDirection = 'asc' | 'desc' | null

export function DataTable<T extends Record<string, unknown>>({
  data,
  columns,
  searchable = false,
  filterable = false,
  filterKey,
  filterOptions,
  onRowClick,
  emptyMessage = 'No data',
  pageSize = 10,
  selectable = false,
  onSelectionChange,
  getRowId,
}: DataTableProps<T>) {
  const [searchTerm, setSearchTerm] = useState('')
  /** Radix SelectItem cannot use value="" */
  const [filterValue, setFilterValue] = useState('all')
  const [sortKey, setSortKey] = useState<keyof T | null>(null)
  const [sortDirection, setSortDirection] = useState<SortDirection>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const filteredData = useMemo(() => {
    let result = data.filter((row) => {
      const matchesSearch = !searchTerm || columns.some((col) => {
        const value = row[col.key]
        return String(value).toLowerCase().includes(searchTerm.toLowerCase())
      })

      const matchesFilter =
        filterValue === 'all' || !filterKey || row[filterKey] === filterValue

      return matchesSearch && matchesFilter
    })

    if (sortKey && sortDirection) {
      result = [...result].sort((a, b) => {
        const aVal = a[sortKey]
        const bVal = b[sortKey]
        const aStr = String(aVal)
        const bStr = String(bVal)
        const cmp = aStr.localeCompare(bStr, undefined, { numeric: true })
        return sortDirection === 'asc' ? cmp : -cmp
      })
    }

    return result
  }, [data, columns, searchTerm, filterValue, filterKey, sortKey, sortDirection])

  const totalPages = Math.max(1, Math.ceil(filteredData.length / pageSize))
  const safeCurrentPage = Math.min(currentPage, totalPages)
  const paginatedData = filteredData.slice(
    (safeCurrentPage - 1) * pageSize,
    safeCurrentPage * pageSize
  )

  const handleSort = (key: keyof T) => {
    if (sortKey !== key) {
      setSortKey(key)
      setSortDirection('asc')
    } else if (sortDirection === 'asc') {
      setSortDirection('desc')
    } else if (sortDirection === 'desc') {
      setSortKey(null)
      setSortDirection(null)
    }
  }

  const handleSelectAll = () => {
    if (selectedIds.size === paginatedData.length) {
      selectedIds.clear()
    } else {
      paginatedData.forEach((row) => {
        const id = getRowId ? getRowId(row) : String(row.id)
        selectedIds.add(id)
      })
    }
    setSelectedIds(new Set(selectedIds))
    onSelectionChange?.(Array.from(selectedIds))
  }

  const handleSelectRow = (row: T) => {
    const id = getRowId ? getRowId(row) : String(row.id)
    if (selectedIds.has(id)) {
      selectedIds.delete(id)
    } else {
      selectedIds.add(id)
    }
    setSelectedIds(new Set(selectedIds))
    onSelectionChange?.(Array.from(selectedIds))
  }

  const isRowSelected = (row: T) => {
    const id = getRowId ? getRowId(row) : String(row.id)
    return selectedIds.has(id)
  }

  return (
    <div className="space-y-4">
      {(searchable || filterable) && (
        <div className="flex space-x-4">
          {searchable && (
            <Input
              placeholder="Search..."
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value)
                setCurrentPage(1)
              }}
              className="max-w-sm"
            />
          )}
          {filterable && filterKey && filterOptions && (
            <Select value={filterValue} onValueChange={(v) => { setFilterValue(v); setCurrentPage(1) }}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="All" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
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
              {selectable && (
                <TableHead className="w-12">
                  <input
                    type="checkbox"
                    checked={paginatedData.length > 0 && selectedIds.size === paginatedData.length}
                    onChange={handleSelectAll}
                    className="rounded border-gray-300"
                  />
                </TableHead>
              )}
              {columns.map((col) => (
                <TableHead
                  key={String(col.key)}
                  className={col.sortable ? 'cursor-pointer select-none hover:bg-gray-50 dark:hover:bg-gray-700' : ''}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  <div className="flex items-center gap-1">
                    {col.label}
                    {col.sortable && sortKey === col.key && sortDirection === 'asc' && <ChevronUp className="h-4 w-4" />}
                    {col.sortable && sortKey === col.key && sortDirection === 'desc' && <ChevronDown className="h-4 w-4" />}
                    {col.sortable && sortKey !== col.key && <ChevronsUpDown className="h-4 w-4 text-gray-400" />}
                  </div>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedData.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length + (selectable ? 1 : 0)} className="text-center text-gray-500 py-8">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            ) : (
              paginatedData.map((row, index) => (
                <TableRow
                  key={getRowId ? getRowId(row) : index}
                  className={`${onRowClick ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/30' : ''} ${isRowSelected(row) ? 'bg-indigo-50 dark:bg-indigo-900/20' : ''}`}
                  onClick={() => onRowClick?.(row)}
                >
                  {selectable && (
                    <TableCell className="w-12" onClick={(e) => e.stopPropagation()}>
                      <input
                        type="checkbox"
                        checked={isRowSelected(row)}
                        onChange={() => handleSelectRow(row)}
                        className="rounded border-gray-300"
                      />
                    </TableCell>
                  )}
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

      {/* Pagination */}
      {filteredData.length > pageSize && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Showing {(safeCurrentPage - 1) * pageSize + 1}-{Math.min(safeCurrentPage * pageSize, filteredData.length)} of {filteredData.length}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={safeCurrentPage <= 1}
              onClick={() => setCurrentPage(safeCurrentPage - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {safeCurrentPage} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={safeCurrentPage >= totalPages}
              onClick={() => setCurrentPage(safeCurrentPage + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
