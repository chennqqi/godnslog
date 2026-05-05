'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest, CaseUpdateRequest } from '@/types'

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
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
        >
          创建 Case
        </button>
      </div>

      {/* Search and Filter */}
      <div className="bg-white shadow rounded-lg mb-4 p-4">
        <div className="flex space-x-4">
          <input
            type="text"
            placeholder="搜索 cases..."
            className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={searchTerm}
            onChange={(e) => { setSearchTerm(e.target.value); loadCases() }}
          />
          <select
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={statusFilter}
            onChange={(e) => { setStatusFilter(e.target.value); loadCases() }}
          >
            <option value="">所有状态</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="archived">Archived</option>
          </select>
        </div>
      </div>

      {/* Batch Operations */}
      {selectedCases.size > 0 && (
        <div className="bg-indigo-50 border border-indigo-200 rounded-lg mb-4 p-4 flex justify-between items-center">
          <span className="text-sm text-indigo-700">已选择 {selectedCases.size} 个 cases</span>
          <button
            onClick={handleBatchDelete}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm"
          >
            批量删除
          </button>
        </div>
      )}

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {cases.length === 0 ? (
            <p className="text-gray-500">暂无 Cases</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              <li className="py-2 flex items-center">
                <input
                  type="checkbox"
                  checked={selectedCases.size === cases.length && cases.length > 0}
                  onChange={toggleSelectAll}
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
                  <input
                    type="checkbox"
                    checked={selectedCases.has(case_.id)}
                    onChange={() => toggleSelect(case_.id)}
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
                    <span className={`px-2 py-1 text-xs rounded ${
                      case_.status === 'active' ? 'bg-green-100 text-green-800' :
                      case_.status === 'completed' ? 'bg-blue-100 text-blue-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {case_.status}
                    </span>
                  </div>
                  <div className="w-32 text-xs text-gray-400">
                    {new Date(case_.created_at).toLocaleDateString()}
                  </div>
                  <div className="w-24 flex space-x-2">
                    <button
                      onClick={() => openEditModal(case_)}
                      className="text-indigo-600 hover:text-indigo-800 text-sm"
                    >
                      编辑
                    </button>
                    <button
                      onClick={() => openDeleteModal(case_)}
                      className="text-red-600 hover:text-red-800 text-sm"
                    >
                      删除
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium mb-4">创建 Case</h3>
            <form onSubmit={handleCreateCase}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  标题
                </label>
                <input
                  type="text"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={newCase.title}
                  onChange={(e) => setNewCase({ ...newCase, title: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  描述
                </label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  rows={3}
                  value={newCase.description}
                  onChange={(e) => setNewCase({ ...newCase, description: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  目标
                </label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={newCase.target}
                  onChange={(e) => setNewCase({ ...newCase, target: e.target.value })}
                />
              </div>
              <div className="flex justify-end space-x-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  创建
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedCase && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium mb-4">编辑 Case</h3>
            <form onSubmit={handleEditCase}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  标题
                </label>
                <input
                  type="text"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={editCase.title}
                  onChange={(e) => setEditCase({ ...editCase, title: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  描述
                </label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  rows={3}
                  value={editCase.description}
                  onChange={(e) => setEditCase({ ...editCase, description: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  目标
                </label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={editCase.target}
                  onChange={(e) => setEditCase({ ...editCase, target: e.target.value })}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  状态
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={editCase.status}
                  onChange={(e) => setEditCase({ ...editCase, status: e.target.value as 'active' | 'completed' | 'archived' })}
                >
                  <option value="active">Active</option>
                  <option value="completed">Completed</option>
                  <option value="archived">Archived</option>
                </select>
              </div>
              <div className="flex justify-end space-x-2">
                <button
                  type="button"
                  onClick={() => setShowEditModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  保存
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Modal */}
      {showDeleteModal && selectedCase && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium mb-4">确认删除</h3>
            <p className="text-gray-600 mb-4">
              确定要删除 case "{selectedCase.title}" 吗？此操作不可撤销。
            </p>
            <div className="flex justify-end space-x-2">
              <button
                onClick={() => setShowDeleteModal(false)}
                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
              >
                取消
              </button>
              <button
                onClick={handleDeleteCase}
                className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
