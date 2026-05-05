'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { marketplaceApi } from '@/lib/api-client'

interface Plugin {
  id: string
  name: string
  description: string
  version: string
  author: string
  downloads: number
  rating: number
  installed: boolean
}

interface Template {
  id: string
  name: string
  description: string
  category: string
  downloads: number
  installed: boolean
}

export default function MarketplacePage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const [activeTab, setActiveTab] = useState<'plugins' | 'templates' | 'installed'>('plugins')
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [activeTab])

  const loadData = async () => {
    setLoading(true)
    try {
      if (activeTab === 'plugins') {
        const response = await marketplaceApi.listPlugins()
        if (response.data && response.data.items) {
          setPlugins(response.data.items)
        }
      } else if (activeTab === 'templates') {
        const response = await marketplaceApi.listTemplates()
        if (response.data && response.data.items) {
          setTemplates(response.data.items)
        }
      }
    } catch (error) {
      console.error('Failed to load data:', error)
    } finally {
      setLoading(false)
    }
  }

  const installPlugin = async (pluginId: string) => {
    try {
      await marketplaceApi.getPlugin(pluginId)
      setPlugins(plugins.map(p => p.id === pluginId ? { ...p, installed: true } : p))
    } catch (error) {
      console.error('Failed to install plugin:', error)
    }
  }

  const installTemplate = async (templateId: string) => {
    try {
      await marketplaceApi.getTemplate(templateId)
      setTemplates(templates.map(t => t.id === templateId ? { ...t, installed: true } : t))
    } catch (error) {
      console.error('Failed to install template:', error)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">插件和模板市场</h2>

      {/* Tab Navigation */}
      <div className="mb-6">
        <div className="flex space-x-4">
          <button
            onClick={() => setActiveTab('plugins')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'plugins'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            插件市场
          </button>
          <button
            onClick={() => setActiveTab('templates')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'templates'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            模板市场
          </button>
          <button
            onClick={() => setActiveTab('installed')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'installed'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            已安装
          </button>
        </div>
      </div>

      {/* Search */}
      <div className="mb-6">
        <input
          type="text"
          placeholder="搜索插件或模板..."
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center h-64">
          <p className="text-gray-500">加载中...</p>
        </div>
      ) : (
        <div>
          {activeTab === 'plugins' && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {plugins.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <p className="text-gray-500">暂无插件</p>
                </div>
              ) : (
                plugins.map((plugin) => (
                  <div key={plugin.id} className="bg-white shadow rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{plugin.name}</h3>
                    <p className="text-sm text-gray-600 mb-4">{plugin.description}</p>
                    <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                      <span>v{plugin.version}</span>
                      <span>by {plugin.author}</span>
                    </div>
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm">⭐ {plugin.rating}</span>
                        <span className="text-sm">↓ {plugin.downloads}</span>
                      </div>
                    </div>
                    <button
                      onClick={() => installPlugin(plugin.id)}
                      disabled={plugin.installed}
                      className={`w-full py-2 rounded-lg ${
                        plugin.installed
                          ? 'bg-gray-300 text-gray-600 cursor-not-allowed'
                          : 'bg-indigo-600 text-white hover:bg-indigo-700'
                      }`}
                    >
                      {plugin.installed ? '已安装' : '安装'}
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'templates' && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {templates.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <p className="text-gray-500">暂无模板</p>
                </div>
              ) : (
                templates.map((template) => (
                  <div key={template.id} className="bg-white shadow rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{template.name}</h3>
                    <p className="text-sm text-gray-600 mb-4">{template.description}</p>
                    <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                      <span className="px-2 py-1 bg-gray-100 rounded">{template.category}</span>
                      <span>↓ {template.downloads}</span>
                    </div>
                    <button
                      onClick={() => installTemplate(template.id)}
                      disabled={template.installed}
                      className={`w-full py-2 rounded-lg ${
                        template.installed
                          ? 'bg-gray-300 text-gray-600 cursor-not-allowed'
                          : 'bg-indigo-600 text-white hover:bg-indigo-700'
                      }`}
                    >
                      {template.installed ? '已安装' : '安装'}
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'installed' && (
            <div className="text-center py-12">
              <p className="text-gray-500">暂无已安装的插件或模板</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
