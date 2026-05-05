'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export default function SettingsPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">系统设置</h2>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="general">通用设置</TabsTrigger>
          <TabsTrigger value="domain">域名设置</TabsTrigger>
          <TabsTrigger value="listener">监听配置</TabsTrigger>
          <TabsTrigger value="notification">通知设置</TabsTrigger>
          <TabsTrigger value="tokens">Token管理</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <GeneralSettings />
        </TabsContent>

        <TabsContent value="domain" className="mt-4">
          <DomainSettings />
        </TabsContent>

        <TabsContent value="listener" className="mt-4">
          <ListenerSettings />
        </TabsContent>

        <TabsContent value="notification" className="mt-4">
          <NotificationSettings />
        </TabsContent>

        <TabsContent value="tokens" className="mt-4">
          <TokenManagement />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function GeneralSettings() {
  return (
    <div className="bg-white shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">通用设置</h3>
      <div className="space-y-4">
        <div>
          <Label htmlFor="system-name">系统名称</Label>
          <Input
            id="system-name"
            defaultValue="GODNSLOG 2.0"
          />
        </div>
        <div>
          <Label htmlFor="language">语言</Label>
          <Select defaultValue="zh-CN">
            <SelectTrigger id="language">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="zh-CN">简体中文</SelectItem>
              <SelectItem value="en-US">English</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label htmlFor="timezone">时区</Label>
          <Select defaultValue="Asia/Shanghai">
            <SelectTrigger id="timezone">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="Asia/Shanghai">Asia/Shanghai</SelectItem>
              <SelectItem value="UTC">UTC</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button>保存设置</Button>
      </div>
    </div>
  )
}

function DomainSettings() {
  return (
    <div className="bg-white shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">域名设置</h3>
      <div className="space-y-4">
        <div>
          <Label htmlFor="main-domain">主域名</Label>
          <Input
            id="main-domain"
            placeholder="example.com"
          />
        </div>
        <div>
          <Label htmlFor="dns-domain">DNS域名</Label>
          <Input
            id="dns-domain"
            placeholder="dns.example.com"
          />
        </div>
        <div>
          <Label htmlFor="http-domain">HTTP域名</Label>
          <Input
            id="http-domain"
            placeholder="http.example.com"
          />
        </div>
        <Button>保存设置</Button>
      </div>
    </div>
  )
}

function ListenerSettings() {
  return (
    <div className="bg-white shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">监听配置</h3>
      <div className="space-y-4">
        <div>
          <Label htmlFor="dns-listen">DNS监听地址</Label>
          <Input
            id="dns-listen"
            defaultValue=":53"
          />
        </div>
        <div>
          <Label htmlFor="http-listen">HTTP监听地址</Label>
          <Input
            id="http-listen"
            defaultValue=":8080"
          />
        </div>
        <div>
          <Label htmlFor="https-listen">HTTPS监听地址</Label>
          <Input
            id="https-listen"
            defaultValue=":8443"
          />
        </div>
        <Button>保存设置</Button>
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
    <div className="bg-white shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">通知设置</h3>
      <div className="space-y-6">
        {notifications.map((notification) => (
          <div key={notification.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center">
                <Checkbox
                  id={`notify-${notification.id}`}
                  checked={notification.enabled}
                  onCheckedChange={(checked) => updateNotification(notification.id, 'enabled', checked)}
                  className="mr-3"
                />
                <Label htmlFor={`notify-${notification.id}`} className="font-medium text-gray-900">{notification.name}</Label>
              </div>
            </div>

            {notification.enabled && (
              <div className="space-y-4 ml-6">
                <div>
                  <Label htmlFor={`webhook-url-${notification.id}`}>Webhook URL</Label>
                  <Input
                    id={`webhook-url-${notification.id}`}
                    placeholder="https://your-webhook-url"
                    value={notification.webhook_url}
                    onChange={(e) => updateNotification(notification.id, 'webhook_url', e.target.value)}
                  />
                </div>
                <div>
                  <Label htmlFor={`webhook-body-${notification.id}`}>Webhook Body (JSON模板)</Label>
                  <Textarea
                    id={`webhook-body-${notification.id}`}
                    className="font-mono text-sm"
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

        <Button>保存设置</Button>
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
    <div className="bg-white shadow rounded-lg p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium text-gray-900">Token管理</h3>
        <Button>创建新Token</Button>
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
              <Button variant="destructive" size="sm">
                撤销
              </Button>
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
