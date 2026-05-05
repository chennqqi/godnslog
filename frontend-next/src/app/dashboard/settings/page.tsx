'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'

export default function SettingsPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])
  const [activeTab, setActiveTab] = useState('general')

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">系统设置</h2>

      <div className="flex space-x-4 mb-6">
        <button
          onClick={() => setActiveTab('general')}
          className={`px-4 py-2 rounded ${
            activeTab === 'general' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          通用设置
        </button>
        <button
          onClick={() => setActiveTab('domain')}
          className={`px-4 py-2 rounded ${
            activeTab === 'domain' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          域名设置
        </button>
        <button
          onClick={() => setActiveTab('listener')}
          className={`px-4 py-2 rounded ${
            activeTab === 'listener' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          监听配置
        </button>
        <button
          onClick={() => setActiveTab('notification')}
          className={`px-4 py-2 rounded ${
            activeTab === 'notification' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          通知设置
        </button>
        <button
          onClick={() => setActiveTab('tokens')}
          className={`px-4 py-2 rounded ${
            activeTab === 'tokens' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'
          }`}
        >
          Token管理
        </button>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {activeTab === 'general' && <GeneralSettings />}
          {activeTab === 'domain' && <DomainSettings />}
          {activeTab === 'listener' && <ListenerSettings />}
          {activeTab === 'notification' && <NotificationSettings />}
          {activeTab === 'tokens' && <TokenManagement />}
        </div>
      </div>
    </div>
  )
}

function GeneralSettings() {
  return (
    <div>
      <h3 className="text-lg font-medium text-gray-900 mb-4">通用设置</h3>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            系统名称
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            defaultValue="GODNSLOG 2.0"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            语言
          </label>
          <select className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500">
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            时区
          </label>
          <select className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500">
            <option value="Asia/Shanghai">Asia/Shanghai</option>
            <option value="UTC">UTC</option>
          </select>
        </div>
        <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          保存设置
        </button>
      </div>
    </div>
  )
}

function DomainSettings() {
  return (
    <div>
      <h3 className="text-lg font-medium text-gray-900 mb-4">域名设置</h3>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            主域名
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="example.com"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            DNS域名
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="dns.example.com"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            HTTP域名
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="http.example.com"
          />
        </div>
        <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          保存设置
        </button>
      </div>
    </div>
  )
}

function ListenerSettings() {
  return (
    <div>
      <h3 className="text-lg font-medium text-gray-900 mb-4">监听配置</h3>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            DNS监听地址
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            defaultValue=":53"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            HTTP监听地址
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            defaultValue=":8080"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            HTTPS监听地址
          </label>
          <input
            type="text"
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            defaultValue=":8443"
          />
        </div>
        <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          保存设置
        </button>
      </div>
    </div>
  )
}

function NotificationSettings() {
  const [notifications, setNotifications] = useState([
    { id: 'dns', name: 'DNS命中', enabled: true, webhook_url: '', webhook_body: '{"protocol":"dns","domain":"{{domain}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'http', name: 'HTTP命中', enabled: true, webhook_url: '', webhook_body: '{"protocol":"http","method":"{{method}}","path":"{{path}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'smtp', name: 'SMTP命中', enabled: false, webhook_url: '', webhook_body: '{"protocol":"smtp","from":"{{from}}","to":"{{to}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'ldap', name: 'LDAP命中', enabled: false, webhook_url: '', webhook_body: '{"protocol":"ldap","operation":"{{operation}}","dn":"{{dn}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'ftp', name: 'FTP命中', enabled: false, webhook_url: '', webhook_body: '{"protocol":"ftp","command":"{{command}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'payload_expire', name: 'Payload过期', enabled: false, webhook_url: '', webhook_body: '{"event":"payload_expire","token":"{{token}}","timestamp":"{{timestamp}}"}' },
  ])

  const updateNotification = (id: string, field: string, value: any) => {
    setNotifications(notifications.map(n => n.id === id ? { ...n, [field]: value } : n))
  }

  return (
    <div>
      <h3 className="text-lg font-medium text-gray-900 mb-4">通知设置</h3>
      <div className="space-y-6">
        {notifications.map((notification) => (
          <div key={notification.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={notification.enabled}
                  onChange={(e) => updateNotification(notification.id, 'enabled', e.target.checked)}
                  className="mr-3"
                />
                <span className="font-medium text-gray-900">{notification.name}</span>
              </div>
            </div>

            {notification.enabled && (
              <div className="space-y-4 ml-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Webhook URL
                  </label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    placeholder="https://your-webhook-url"
                    value={notification.webhook_url}
                    onChange={(e) => updateNotification(notification.id, 'webhook_url', e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Webhook Body (JSON模板)
                  </label>
                  <textarea
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                    rows={4}
                    value={notification.webhook_body}
                    onChange={(e) => updateNotification(notification.id, 'webhook_body', e.target.value)}
                    placeholder='{"key": "value"}'
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    可用变量: {`{{domain}}, {{ip}}, {{timestamp}}, {{method}}, {{path}}, {{from}}, {{to}}, {{operation}}, {{dn}}, {{command}}, {{token}}`}
                  </p>
                </div>
              </div>
            )}
          </div>
        ))}

        <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
          保存设置
        </button>
      </div>
    </div>
  )
}

function TokenManagement() {
  const [apiKeys, setApiKeys] = useState([
    { id: '1', name: 'Development Key', key_prefix: 'gdl_', created_at: '2024-01-01' },
    { id: '2', name: 'Production Key', key_prefix: 'gdl_', created_at: '2024-01-15' },
  ])

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium text-gray-900">Token管理</h3>
        <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm">
          创建新Token
        </button>
      </div>
      <div className="space-y-4">
        {apiKeys.map((key) => (
          <div key={key.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex justify-between items-start">
              <div>
                <p className="font-medium text-gray-900">{key.name}</p>
                <p className="text-sm text-gray-500">Key: {key.key_prefix}****</p>
                <p className="text-xs text-gray-400">创建于: {key.created_at}</p>
              </div>
              <button className="text-red-600 hover:text-red-800 text-sm">
                撤销
              </button>
            </div>
          </div>
        ))}
        {apiKeys.length === 0 && (
          <p className="text-gray-500">暂无API Keys</p>
        )}
      </div>
    </div>
  )
}
