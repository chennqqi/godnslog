'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { apiKeyApi } from '@/lib/api-client'
import type { APIKey } from '@/types'

export default function APIKeysPage() {
  const router = useRouter()
  const [apiKeys, setApiKeys] = useState<APIKey[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyScopes, setNewKeyScopes] = useState<string[]>(['cases:read', 'payloads:read'])
  const [editKeyName, setEditKeyName] = useState('')
  const [editKeyScopes, setEditKeyScopes] = useState<string[]>([])
  const [editKeyEnabled, setEditKeyEnabled] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadAPIKeys()
  }, [router])

  const loadAPIKeys = async () => {
    try {
      const response = await apiKeyApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setApiKeys(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load API keys:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await apiKeyApi.create({
        name: newKeyName,
        scopes: newKeyScopes,
      })
      if (response.code === 0) {
        setShowCreateModal(false)
        setNewKeyName('')
        setNewKeyScopes(['cases:read', 'payloads:read'])
        loadAPIKeys()
      }
    } catch (error) {
      console.error('Failed to create API key:', error)
    }
  }

  const handleUpdateKey = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedKey) return
    try {
      const response = await apiKeyApi.update(selectedKey.id, {
        name: editKeyName,
        scopes: editKeyScopes,
        enabled: editKeyEnabled,
      })
      if (response.code === 0) {
        setShowEditModal(false)
        setSelectedKey(null)
        loadAPIKeys()
      }
    } catch (error) {
      console.error('Failed to update API key:', error)
    }
  }

  const handleDeleteKey = async (id: string) => {
    if (!confirm('确定要删除此 API Key 吗？')) return
    try {
      const response = await apiKeyApi.delete(id)
      if (response.code === 0) {
        loadAPIKeys()
      }
    } catch (error) {
      console.error('Failed to delete API key:', error)
    }
  }

  const openEditModal = (key: APIKey) => {
    setSelectedKey(key)
    setEditKeyName(key.name)
    setEditKeyScopes(key.scopes || [])
    setEditKeyEnabled(key.enabled)
    setShowEditModal(true)
  }

  const toggleScope = (scope: string, isNew: boolean) => {
    if (isNew) {
      if (newKeyScopes.includes(scope)) {
        setNewKeyScopes(newKeyScopes.filter(s => s !== scope))
      } else {
        setNewKeyScopes([...newKeyScopes, scope])
      }
    } else {
      if (editKeyScopes.includes(scope)) {
        setEditKeyScopes(editKeyScopes.filter(s => s !== scope))
      } else {
        setEditKeyScopes([...editKeyScopes, scope])
      }
    }
  }

  const availableScopes = [
    'cases:read', 'cases:write', 'cases:delete',
    'payloads:read', 'payloads:write', 'payloads:delete',
    'interactions:read', 'interactions:delete',
    'apikeys:read', 'apikeys:write', 'apikeys:delete',
  ]

  if (loading) {
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">API Keys 管理</h2>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
        >
          创建 API Key
        </button>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {apiKeys.length === 0 ? (
            <p className="text-gray-500">暂无 API Keys</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {apiKeys.map((key) => (
                <li key={key.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <p className="text-sm font-medium text-indigo-600">{key.name}</p>
                        <span className={`px-2 py-1 text-xs rounded ${
                          key.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                        }`}>
                          {key.enabled ? '启用' : '禁用'}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mt-1 break-all">
                        Key: {key.key_masked}
                      </p>
                      <div className="mt-2">
                        <p className="text-xs text-gray-500">作用域:</p>
                        <div className="flex flex-wrap gap-1 mt-1">
                          {(key.scopes || []).map((scope) => (
                            <span key={scope} className="px-2 py-1 text-xs bg-gray-100 rounded">
                              {scope}
                            </span>
                          ))}
                        </div>
                      </div>
                      {key.expires_at && (
                        <p className="text-xs text-gray-400 mt-1">
                          过期: {new Date(key.expires_at).toLocaleString()}
                        </p>
                      )}
                    </div>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => openEditModal(key)}
                        className="text-indigo-600 hover:text-indigo-800 text-sm"
                      >
                        编辑
                      </button>
                      <button
                        onClick={() => handleDeleteKey(key.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        删除
                      </button>
                    </div>
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
            <h3 className="text-lg font-medium mb-4">创建 API Key</h3>
            <form onSubmit={handleCreateKey}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  名称
                </label>
                <input
                  type="text"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  作用域
                </label>
                <div className="space-y-2 max-h-40 overflow-auto">
                  {availableScopes.map((scope) => (
                    <label key={scope} className="flex items-center">
                      <input
                        type="checkbox"
                        checked={newKeyScopes.includes(scope)}
                        onChange={() => toggleScope(scope, true)}
                        className="mr-2"
                      />
                      <span className="text-sm text-gray-700">{scope}</span>
                    </label>
                  ))}
                </div>
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
      {showEditModal && selectedKey && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium mb-4">编辑 API Key</h3>
            <form onSubmit={handleUpdateKey}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  名称
                </label>
                <input
                  type="text"
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={editKeyName}
                  onChange={(e) => setEditKeyName(e.target.value)}
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  作用域
                </label>
                <div className="space-y-2 max-h-40 overflow-auto">
                  {availableScopes.map((scope) => (
                    <label key={scope} className="flex items-center">
                      <input
                        type="checkbox"
                        checked={editKeyScopes.includes(scope)}
                        onChange={() => toggleScope(scope, false)}
                        className="mr-2"
                      />
                      <span className="text-sm text-gray-700">{scope}</span>
                    </label>
                  ))}
                </div>
              </div>
              <div className="mb-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={editKeyEnabled}
                    onChange={(e) => setEditKeyEnabled(e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">启用</span>
                </label>
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
    </div>
  )
}
