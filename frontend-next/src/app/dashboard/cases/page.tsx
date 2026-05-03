'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest } from '@/types'

export default function CasesPage() {
  const router = useRouter()
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newCase, setNewCase] = useState<CaseCreateRequest>({
    title: '',
    description: '',
    target: '',
    tags: [],
  })

  useEffect(() => {
    loadCases()
  }, [])

  const loadCases = async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 50 })
      if (response.data) {
        setCases(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
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

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {cases.length === 0 ? (
            <p className="text-gray-500">暂无 Cases</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {cases.map((case_) => (
                <li
                  key={case_.id}
                  className="py-4 flex justify-between items-center cursor-pointer hover:bg-gray-50"
                  onClick={() => router.push(`/dashboard/cases/${case_.id}`)}
                >
                  <div>
                    <p className="text-sm font-medium text-indigo-600">{case_.title}</p>
                    <p className="text-sm text-gray-500">{case_.description}</p>
                    {case_.target && (
                      <p className="text-xs text-gray-400 mt-1">目标: {case_.target}</p>
                    )}
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-1 text-xs rounded ${
                      case_.status === 'active' ? 'bg-green-100 text-green-800' :
                      case_.status === 'completed' ? 'bg-blue-100 text-blue-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {case_.status}
                    </span>
                    <span className="text-xs text-gray-400">
                      {new Date(case_.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

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
    </div>
  )
}
