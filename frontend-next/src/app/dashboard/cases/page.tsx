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
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Case Board</h2>
        <Button onClick={() => setShowCreateModal(true)}>
          创建 Case
        </Button>
      </div>

      {/* Search and Filter */}
      <div className="bg-white shadow rounded-lg mb-4 p-4">
        <div className="flex space-x-4">
          <Input
            placeholder="搜索 cases..."
            className="flex-1"
            value={searchTerm}
            onChange={(e) => { setSearchTerm(e.target.value); loadCases() }}
          />
          <Select value={statusFilter} onValueChange={(value) => { setStatusFilter(value); loadCases() }}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="所有状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">所有状态</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="archived">Archived</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Batch Operations */}
      {selectedCases.size > 0 && (
        <div className="bg-indigo-50 border border-indigo-200 rounded-lg mb-4 p-4 flex justify-between items-center">
          <span className="text-sm text-indigo-700">已选择 {selectedCases.size} 个 cases</span>
          <Button variant="destructive" size="sm" onClick={handleBatchDelete}>
            批量删除
          </Button>
        </div>
      )}

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {cases.length === 0 ? (
            <p className="text-gray-500">暂无 Cases</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              <li className="py-2 flex items-center">
                <Checkbox
                  checked={selectedCases.size === cases.length && cases.length > 0}
                  onCheckedChange={toggleSelectAll}
                  className="mr-4"
                />
                <span className="flex-1 font-medium text-gray-500">标题</span>
                <span className="w-24 text-gray-500">状态</span>
                <span className="w-32 text-gray-500">创建时间</span>
                <span className="w-24 text-gray-500">操作</span>
              </li>
              {cases.map((case_) => (
                <li
                  key={case_.id}
                  className="py-4 flex items-center hover:bg-gray-50"
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
                    <p className="text-sm font-medium text-indigo-600">{case_.title}</p>
                    <p className="text-sm text-gray-500">{case_.description}</p>
                    {case_.target && (
                      <p className="text-xs text-gray-400 mt-1">目标: {case_.target}</p>
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
                  <div className="w-24 flex space-x-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => openEditModal(case_)}
                    >
                      编辑
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => openDeleteModal(case_)}
                      className="text-red-600 hover:text-red-800"
                    >
                      删除
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
            <DialogTitle>创建 Case</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreateCase}>
            <div className="mb-4">
              <Label htmlFor="title">标题</Label>
              <Input
                id="title"
                required
                value={newCase.title}
                onChange={(e) => setNewCase({ ...newCase, title: e.target.value })}
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="description">描述</Label>
              <Textarea
                id="description"
                rows={3}
                value={newCase.description}
                onChange={(e) => setNewCase({ ...newCase, description: e.target.value })}
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="target">目标</Label>
              <Input
                id="target"
                value={newCase.target}
                onChange={(e) => setNewCase({ ...newCase, target: e.target.value })}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                取消
              </Button>
              <Button type="submit">
                创建
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Modal */}
      <Dialog open={showEditModal} onOpenChange={setShowEditModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>编辑 Case</DialogTitle>
          </DialogHeader>
          {selectedCase && (
            <form onSubmit={handleEditCase}>
              <div className="mb-4">
                <Label htmlFor="edit-title">标题</Label>
                <Input
                  id="edit-title"
                  required
                  value={editCase.title}
                  onChange={(e) => setEditCase({ ...editCase, title: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-description">描述</Label>
                <Textarea
                  id="edit-description"
                  rows={3}
                  value={editCase.description}
                  onChange={(e) => setEditCase({ ...editCase, description: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-target">目标</Label>
                <Input
                  id="edit-target"
                  value={editCase.target}
                  onChange={(e) => setEditCase({ ...editCase, target: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <Label htmlFor="edit-status">状态</Label>
                <Select value={editCase.status} onValueChange={(value) => setEditCase({ ...editCase, status: value as 'active' | 'completed' | 'archived' })}>
                  <SelectTrigger id="edit-status">
                    <SelectValue placeholder="选择状态" />
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
                  取消
                </Button>
                <Button type="submit">
                  保存
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
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除 case "{selectedCase?.title}" 吗？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteModal(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleDeleteCase}>
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
