'use client'

import { useState } from 'react'

export default function MarketplacePage() {
  const [activeTab, setActiveTab] = useState('plugins')
  const [plugins, setPlugins] = useState<any[]>([
    { id: '1', name: 'SSRF Scanner', type: 'listener', author: 'Security Team', downloads: 1250, rating: 4.5, is_official: true, is_published: true },
    { id: '2', name: 'Burp Integration', type: 'exporter', author: 'Community', downloads: 890, rating: 4.2, is_official: false, is_published: true },
    { id: '3', name: 'Slack Notifier', type: 'notifier', author: 'Security Team', downloads: 2100, rating: 4.8, is_official: true, is_published: true },
  ])
  const [templates, setTemplates] = useState<any[]>([
    { id: '1', name: 'AWS Metadata SSRF', type: 'payload', category: 'ssrf', downloads: 3400, rating: 4.9, is_official: true, is_published: true },
    { id: '2', name: 'Log4j RCE', type: 'payload', category: 'rce', downloads: 2800, rating: 4.7, is_official: true, is_published: true },
    { id: '3', name: 'XXE OOB', type: 'payload', category: 'xxe', downloads: 1900, rating: 4.5, is_official: false, is_published: true },
  ])

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">插件和模板市场</h2>

      <div className="flex space-x-4 mb-6">
        <button
          onClick={() => setActiveTab('plugins')}
          className={`px-4 py-2 rounded ${
            activeTab === 'plugins' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          插件市场
        </button>
        <button
          onClick={() => setActiveTab('templates')}
          className={`px-4 py-2 rounded ${
            activeTab === 'templates' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          模板市场
        </button>
        <button
          onClick={() => setActiveTab('installed')}
          className={`px-4 py-2 rounded ${
            activeTab === 'installed' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          已安装
        </button>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-6">
            <div className="flex space-x-4">
              <input
                type="text"
                placeholder="搜索..."
                className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
              <select className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">所有类型</option>
                <option value="listener">Listener</option>
                <option value="processor">Processor</option>
                <option value="notifier">Notifier</option>
                <option value="exporter">Exporter</option>
              </select>
              <label className="flex items-center">
                <input type="checkbox" className="mr-2" />
                <span className="text-sm text-gray-700">仅官方</span>
              </label>
            </div>
          </div>

          {activeTab === 'plugins' && (
            <div className="space-y-4">
              {plugins.map((plugin) => (
                <div key={plugin.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <h4 className="text-lg font-medium text-gray-900">{plugin.name}</h4>
                        {plugin.is_official && (
                          <span className="px-2 py-1 text-xs bg-indigo-100 text-indigo-800 rounded">
                            官方
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 mb-2">
                        类型: {plugin.type} | 作者: {plugin.author}
                      </p>
                      <div className="flex items-center space-x-4 text-sm text-gray-500">
                        <span>⬇ {plugin.downloads}</span>
                        <span>⭐ {plugin.rating}</span>
                      </div>
                    </div>
                    <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm">
                      安装
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {activeTab === 'templates' && (
            <div className="space-y-4">
              {templates.map((template) => (
                <div key={template.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <h4 className="text-lg font-medium text-gray-900">{template.name}</h4>
                        {template.is_official && (
                          <span className="px-2 py-1 text-xs bg-indigo-100 text-indigo-800 rounded">
                            官方
                          </span>
                        )}
                        <span className="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded">
                          {template.category}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mb-2">
                        类型: {template.type}
                      </p>
                      <div className="flex items-center space-x-4 text-sm text-gray-500">
                        <span>⬇ {template.downloads}</span>
                        <span>⭐ {template.rating}</span>
                      </div>
                    </div>
                    <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm">
                      使用
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {activeTab === 'installed' && (
            <div className="text-center py-8">
              <p className="text-gray-500">暂无已安装的插件或模板</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
