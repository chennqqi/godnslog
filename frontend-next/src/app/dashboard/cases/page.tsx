'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'

/** Sentinel for Radix Select: empty string is reserved for clearing selection */
const STATUS_FILTER_ALL = 'all'

export default function CasesPage() {
  const router = useRouter()
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState(STATUS_FILTER_ALL)
  const [newCase, setNewCase] = useState<CaseCreateRequest>({
    title: '',
    description: '',
    target: '',
    tags: [],
  })

  const loadCases = useCallback(async () => {
    try {
      console.log('Loading cases...')
      const response = await caseApi.list({
        page: 1,
        page_size: 50,
        search: searchTerm,
        ...(statusFilter !== STATUS_FILTER_ALL ? { status: statusFilter } : {}),
      })
      console.log('Cases response:', response)
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
      setCases([])
    } finally {
      setLoading(false)
    }
  }, [searchTerm, statusFilter])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadCases()
  }, [router, loadCases])

  const handleSearchChange = (value: string) => {
    setSearchTerm(value)
  }

  const handleStatusChange = (value: string) => {
    setStatusFilter(value)
  }

  const handleCreateCase = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await caseApi.create(newCase)
      if (response.code === 0 && response.data) {
        setShowCreateModal(false)
        setNewCase({ title: '', description: '', target: '', tags: [] })
        loadCases()
      }
    } catch (error) {
      console.error('Failed to create case:', error)
    }
  }

  if (loading) {
    return <div className="text-center py-12 text-gray-500">Loading cases...</div>
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Case Board</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Manage OAST engagement cases</p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          New Case
        </Button>
      </div>

      {/* Search and Filter */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-4 border border-gray-200 dark:border-gray-700">
        <div className="flex flex-wrap gap-3">
          <Input
            placeholder="Search cases..."
            className="flex-1 min-w-[160px]"
            value={searchTerm}
            onChange={(e) => handleSearchChange(e.target.value)}
          />
          <Select value={statusFilter} onValueChange={handleStatusChange}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={STATUS_FILTER_ALL}>All statuses</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="archived">Archived</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 shadow rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="px-4 py-5 sm:p-6">
          {cases.length === 0 ? (
            <div className="text-center py-10">
              <div className="text-4xl mb-3">📂</div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">No cases yet</p>
              <p className="text-xs text-gray-400 mt-1">Create a case to start tracking OAST interactions</p>
            </div>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              <li className="py-2 flex items-center text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <span className="flex-1">Title</span>
                <span className="w-24">Status</span>
                <span className="w-32">Created</span>
              </li>
              {cases.map((case_) => (
                <li
                  key={case_.id}
                  className="py-4 flex items-center hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors cursor-pointer"
                  onClick={() => router.push(`/dashboard/cases/${case_.id}`)}
                >
                  <div className="flex-1">
                    <p className="text-sm font-medium text-indigo-600 dark:text-indigo-400">{case_.title}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">{case_.description}</p>
                    {case_.target && (
                      <p className="text-xs text-gray-400 mt-1">Target: {case_.target}</p>
                    )}
                  </div>
                  <div className="w-24">
                    <Badge variant={
                      case_.status === 'active' ? 'default' :
                      case_.status === 'completed' ? 'secondary' :
                      'outline'
                    }>
                      {case_.status}
                    </Badge>
                  </div>
                  <div className="w-32 text-xs text-gray-400">
                    {new Date(case_.created_at).toLocaleDateString()}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Create Modal */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Case</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreateCase}>
            <div className="mb-4">
              <Label htmlFor="title">Title</Label>
              <Input
                id="title"
                required
                value={newCase.title}
                onChange={(e) => setNewCase({ ...newCase, title: e.target.value })}
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                rows={3}
                value={newCase.description}
                onChange={(e) => setNewCase({ ...newCase, description: e.target.value })}
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="target">Target</Label>
              <Input
                id="target"
                placeholder="e.g. internal-api.corp.com"
                value={newCase.target}
                onChange={(e) => setNewCase({ ...newCase, target: e.target.value })}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button type="submit">
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
