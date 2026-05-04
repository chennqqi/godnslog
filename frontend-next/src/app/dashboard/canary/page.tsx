'use client'

import { useState } from 'react'

export default function CanaryPage() {
  const [canaries, setCanaries] = useState<any[]>([
    { id: '1', type: 'dns', token: 'canary-abc123', context: '项目A-数据库配置', status: 'active', created_at: '2024-01-01' },
    { id: '2', type: 'http', token: 'canary-def456', context: '项目B-API密钥', status: 'active', created_at: '2024-01-15' },
    { id: '3', type: 'document', token: 'canary-ghi789', context: '项目C-文档模板', status: 'silent', created_at: '2024-02-01' },
  ])
  const [showCreateModal, setShowCreateModal] = useState(false)

  const canaryTypes = [
    { value: 'dns', label: 'DNS', description: 'DNS查询触发' },
    { value: 'http', label: 'HTTP', description: 'HTTP请求触发' },
    { value: 'document', label: '文档', description: '文档访问触发' },
    { value: 'config', label: '配置文件', description: '配置文件访问触发' },
    { value: 'ci', label: 'CI变量', description: 'CI环境变量访问触发' },
    { value: 'storage', label: '对象存储', description: '对象存储访问触发' },
    { value: 'email', label: '邮件地址', description: '邮件发送触发' },
  ]

  const handleCreate = (formData: any) => {
    const newCanary = {
      id: Date.now().toString(),
      ...formData,
      status: 'active',
      created_at: new Date().toISOString().split('T')[0],
    }
    setCanaries([...canaries, newCanary])
    setShowCreateModal(false)
  }

  const handleRevoke = (id: string) => {
    if (confirm('确定要撤销此Canary吗？')) {
      setCanaries(canaries.map(c => c.id === id ? { ...c, status: 'revoked' } : c))
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Canary长期监测</h2>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-6">
            <div>
              <h3 className="text-lg font-medium text-gray-900">Canary Token列表</h3>
              <p className="text-sm text-gray-500">管理长期部署的诱饵Token</p>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              创建Canary
            </button>
          </div>

          {canaries.length === 0 ? (
            <p className="text-gray-500 text-center py-8">暂无Canary Token</p>
          ) : (
            <div className="space-y-4">
              {canaries.map((canary) => (
                <div key={canary.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <span className={`px-2 py-1 text-xs rounded ${
                          canary.type === 'dns' ? 'bg-blue-100 text-blue-800' :
                          canary.type === 'http' ? 'bg-green-100 text-green-800' :
                          canary.type === 'document' ? 'bg-purple-100 text-purple-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {canary.type.toUpperCase()}
                        </span>
                        <span className={`px-2 py-1 text-xs rounded ${
                          canary.status === 'active' ? 'bg-green-100 text-green-800' :
                          canary.status === 'silent' ? 'bg-yellow-100 text-yellow-800' :
                          canary.status === 'revoked' ? 'bg-red-100 text-red-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {canary.status}
                        </span>
                      </div>
                      <p className="text-sm font-medium text-gray-900 break-all mb-1">
                        Token: {canary.token}
                      </p>
                      <p className="text-sm text-gray-600 mb-1">
                        上下文: {canary.context}
                      </p>
                      <p className="text-xs text-gray-400">
                        创建于: {canary.created_at}
                      </p>
                    </div>
                    <div className="flex space-x-2">
                      <button className="text-indigo-600 hover:text-indigo-800 text-sm">
                        编辑
                      </button>
                      <button
                        onClick={() => handleRevoke(canary.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        撤销
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {showCreateModal && (
        <CreateCanaryModal
          onClose={() => setShowCreateModal(false)}
          onSubmit={handleCreate}
          types={canaryTypes}
        />
      )}
    </div>
  )
}

function CreateCanaryModal({ onClose, onSubmit, types }: any) {
  const [formData, setFormData] = useState({
    type: 'dns',
    context: '',
    expires_in: 2592000, // 30天
    silent_window: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">创建Canary Token</h3>
        </div>
        <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Token类型
            </label>
            <select
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
            >
              {types.map((t: any) => (
                <option key={t.value} value={t.value}>
                  {t.label} - {t.description}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              上下文编码
            </label>
            <textarea
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              rows={3}
              value={formData.context}
              onChange={(e) => setFormData({ ...formData, context: e.target.value })}
              placeholder="项目、资产、投放位置、负责人等..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              过期时间（秒）
            </label>
            <input
              type="number"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={formData.expires_in}
              onChange={(e) => setFormData({ ...formData, expires_in: parseInt(e.target.value) })}
              min={86400}
              max={31536000}
            />
            <p className="text-xs text-gray-500 mt-1">
              默认30天，最大1年
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              静默窗口（可选）
            </label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={formData.silent_window}
              onChange={(e) => setFormData({ ...formData, silent_window: e.target.value })}
              placeholder="例如: 0-6,18-24 (静默时间)"
            />
            <p className="text-xs text-gray-500 mt-1">
              设置静默时间段，格式: start-end, start-end
            </p>
          </div>

          <div className="flex space-x-4 pt-4">
            <button
              type="submit"
              className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              创建
            </button>
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
            >
              取消
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
