'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest, CaseUpdateRequest } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'

export default function CasesPage() {
  const router = useRouter()
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [selectedCase, setSelectedCase] = useState<Case | null>(null)
  const [selectedCases, setSelectedCases] = useState<Set<string>>(new Set())
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [newCase, setNewCase] = useState<CaseCreateRequest>({
    title: '',
    description: '',
    target: '',
    tags: [],
  })
  const [editCase, setEditCase] = useState<CaseUpdateRequest>({
    title: '',
    description: '',
    target: '',
    status: undefined,
    tags: [],
  })

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadCases()
  }, [router])

  const loadCases = async () => {
    try {
      console.log('Loading cases...')
      const response = await caseApi.list({ page: 1, page_size: 50, search: searchTerm, status: statusFilter })
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

  const handleEditCase = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedCase) return
    try {
      const response = await caseApi.update(selectedCase.id, editCase)
      if (response.code === 0) {
        setShowEditModal(false)
        setSelectedCase(null)
        loadCases()
      }
    } catch (error) {
      console.error('Failed to update case:', error)
    }
  }

  const handleDeleteCase = async () => {
    if (!selectedCase) return
    try {
      const response = await caseApi.delete(selectedCase.id)
      if (response.code === 0) {
        setShowDeleteModal(false)
        setSelectedCase(null)
        loadCases()
      }
    } catch (error) {
      console.error('Failed to delete case:', error)
    }
  }

  const handleBatchDelete = async () => {
    if (selectedCases.size === 0) return
    try {
      for (const id of selectedCases) {
        await caseApi.delete(id)
      }
      setSelectedCases(new Set())
      loadCases()
    } catch (error) {
      console.error('Failed to batch delete cases:', error)
    }
  }

  const openEditModal = (case_: Case) => {
    setSelectedCase(case_)
    setEditCase({
      title: case_.title,
      description: case_.description,
      target: case_.target || '',
      status: case_.status,
      tags: case_.tags || [],
    })
    setShowEditModal(true)
  }

  const openDeleteModal = (case_: Case) => {
    setSelectedCase(case_)
    setShowDeleteModal(true)
  }

  const toggleSelect = (id: string) => {
    const newSelected = new Set(selectedCases)
    if (newSelected.has(id)) {
      newSelected.delete(id)
    } else {
      newSelected.add(id)
    }
    setSelectedCases(newSelected)
  }

  const toggleSelectAll = () => {
    if (selectedCases.size === cases.length) {
      setSelectedCases(new Set())
    } else {
      setSelectedCases(new Set(cases.map(c => c.id)))
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
            onChange={(e) => { setSearchTerm(e.target.value); loadCases() }}
          />
          <Select value={statusFilter} onValueChange={(value) => { setStatusFilter(value); loadCases() }}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">All statuses</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="archived">Archived</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Batch Operations */}
      {selectedCases.size > 0 && (
        <div className="bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-700 rounded-lg p-4 flex justify-between items-center">
          <span className="text-sm text-indigo-700 dark:text-indigo-300">
            {selectedCases.size} case{selectedCases.size !== 1 ? 's' : ''} selected
          </span>
          <Button variant="destructive" size="sm" onClick={handleBatchDelete}>
            Delete selected
          </Button>
        </div>
      )}

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
                <Checkbox
                  checked={selectedCases.size === cases.length && cases.length > 0}
                  onCheckedChange={toggleSelectAll}
                  className="mr-4"
                />
                <span className="flex-1">Title</span>
                <span className="w-24">Status</span>
                <span className="w-32">Created</span>
                <span className="w-24">Actions</span>
              </li>
              {cases.map((case_) => (
                <li
                  key={case_.id}
                  className="py-4 flex items-center hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                >
                  <Checkbox
                    checked={selectedCases.has(case_.id)}
                    onCheckedChange={() => toggleSelect(case_.id)}
                    className="mr-4"
                  />
                  <div
                    className="flex-1 cursor-pointer"
                    onClick={() => router.push(`/dashboard/cases/${case_.id}`)}
                  >
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
                  <div className="w-24 flex space-x-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => openEditModal(case_)}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => openDeleteModal(case_)}
                      className="text-red-600 hover:text-red-800 dark:text-red-400"
                    >
                      Delete
                    </Button>
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

      {/* Edit Modal */}
      <Dialog open={showEditModal} onOpenChange={setShowEditModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Case</DialogTitle>
          </DialogHeader>
          {selectedCase && (
            <form onSubmit={handleEditCase}>
              <div className="mb-4">
                <Label htmlFor="edit-title">Title</Label>
                <Input
                  id="edit-title"
                  required
                  value={editCase.title}
                  onChange={(e) => setEditCase({ ...editCase, title: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-description">Description</Label>
                <Textarea
                  id="edit-description"
                  rows={3}
                  value={editCase.description}
                  onChange={(e) => setEditCase({ ...editCase, description: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-target">Target</Label>
                <Input
                  id="edit-target"
                  value={editCase.target}
                  onChange={(e) => setEditCase({ ...editCase, target: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-status">Status</Label>
                <Select value={editCase.status} onValueChange={(value) => setEditCase({ ...editCase, status: value as 'active' | 'completed' | 'archived' })}>
                  <SelectTrigger id="edit-status">
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                    <SelectItem value="archived">Archived</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setShowEditModal(false)}>
                  Cancel
                </Button>
                <Button type="submit">
                  Save
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Modal */}
      <Dialog open={showDeleteModal} onOpenChange={setShowDeleteModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Delete</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{selectedCase?.title}&quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteModal(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteCase}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
