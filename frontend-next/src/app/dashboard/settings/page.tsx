'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { settingsApi } from '@/lib/api-client'
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
import {
  generalSettingsSchema,
  domainSettingsSchema,
  listenerSettingsSchema,
  type GeneralSettingsFormValues,
  type DomainSettingsFormValues,
  type ListenerSettingsFormValues,
} from '@/features/settings/schemas/settings-schema'

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
      <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">System Settings</h2>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="domain">Domain</TabsTrigger>
          <TabsTrigger value="listener">Listener</TabsTrigger>
          <TabsTrigger value="notification">Notification</TabsTrigger>
          <TabsTrigger value="tokens">API Keys</TabsTrigger>
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
  const [saving, setSaving] = useState(false)
  const form = useForm<GeneralSettingsFormValues>({
    resolver: zodResolver(generalSettingsSchema),
    defaultValues: {
      system_name: 'GODNSLOG 2.0',
      language: 'en-US',
      timezone: 'UTC',
    },
  })

  useEffect(() => {
    settingsApi.get().then((resp) => {
      const data = resp.data as Record<string, unknown> | undefined
      if (data) {
        const settings = (data as { data?: Record<string, unknown> }).data || data
        if (settings.system_name) form.reset({ ...form.getValues(), system_name: settings.system_name as string })
        if (settings.language) form.setValue('language', settings.language as 'en-US' | 'zh-CN')
        if (settings.timezone) form.setValue('timezone', settings.timezone as string)
      }
    }).catch(() => {})
  }, [form])

  const onSubmit = async (data: GeneralSettingsFormValues) => {
    setSaving(true)
    try {
      await settingsApi.update({ general: data })
    } catch (error) {
      console.error('Failed to save general settings:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">General</h3>
      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <Label htmlFor="system-name">System Name</Label>
          <Input
            id="system-name"
            className="mt-1"
            {...form.register('system_name')}
          />
          {form.formState.errors.system_name && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.system_name.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="language">Language</Label>
          <Select defaultValue="en-US" onValueChange={(v) => form.setValue('language', v as 'en-US' | 'zh-CN')}>
            <SelectTrigger id="language">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="en-US">English</SelectItem>
              <SelectItem value="zh-CN">简体中文</SelectItem>
            </SelectContent>
          </Select>
          {form.formState.errors.language && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.language.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="timezone">Timezone</Label>
          <Select defaultValue="UTC" onValueChange={(v) => form.setValue('timezone', v)}>
            <SelectTrigger id="timezone">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="UTC">UTC</SelectItem>
              <SelectItem value="Asia/Shanghai">Asia/Shanghai</SelectItem>
            </SelectContent>
          </Select>
          {form.formState.errors.timezone && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.timezone.message}</p>
          )}
        </div>
        <Button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save Settings'}</Button>
      </form>
    </div>
  )
}

function DomainSettings() {
  const [saving, setSaving] = useState(false)
  const form = useForm<DomainSettingsFormValues>({
    resolver: zodResolver(domainSettingsSchema),
    defaultValues: {
      main_domain: '',
      dns_domain: '',
      http_domain: '',
    },
  })

  useEffect(() => {
    settingsApi.get().then((resp) => {
      const data = resp.data as Record<string, unknown> | undefined
      if (data) {
        const settings = (data as { data?: Record<string, unknown> }).data || data
        if (settings.main_domain) form.setValue('main_domain', settings.main_domain as string)
        if (settings.dns_domain) form.setValue('dns_domain', settings.dns_domain as string)
        if (settings.http_domain) form.setValue('http_domain', settings.http_domain as string)
      }
    }).catch(() => {})
  }, [form])

  const onSubmit = async (data: DomainSettingsFormValues) => {
    setSaving(true)
    try {
      await settingsApi.update({ domain: data })
    } catch (error) {
      console.error('Failed to save domain settings:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Domain</h3>
      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <Label htmlFor="main-domain">Main Domain</Label>
          <Input
            id="main-domain"
            className="mt-1"
            placeholder="example.com"
            {...form.register('main_domain')}
          />
          {form.formState.errors.main_domain && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.main_domain.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="dns-domain">DNS Domain</Label>
          <Input
            id="dns-domain"
            className="mt-1"
            placeholder="dns.example.com"
            {...form.register('dns_domain')}
          />
        </div>
        <div>
          <Label htmlFor="http-domain">HTTP Domain</Label>
          <Input
            id="http-domain"
            className="mt-1"
            placeholder="http.example.com"
            {...form.register('http_domain')}
          />
        </div>
        <Button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save Settings'}</Button>
      </form>
    </div>
  )
}

function ListenerSettings() {
  const [saving, setSaving] = useState(false)
  const form = useForm<ListenerSettingsFormValues>({
    resolver: zodResolver(listenerSettingsSchema),
    defaultValues: {
      dns_listen: ':53',
      http_listen: ':8080',
      https_listen: ':8443',
    },
  })

  useEffect(() => {
    settingsApi.get().then((resp) => {
      const data = resp.data as Record<string, unknown> | undefined
      if (data) {
        const settings = (data as { data?: Record<string, unknown> }).data || data
        if (settings.dns_listen) form.setValue('dns_listen', settings.dns_listen as string)
        if (settings.http_listen) form.setValue('http_listen', settings.http_listen as string)
        if (settings.https_listen) form.setValue('https_listen', settings.https_listen as string)
      }
    }).catch(() => {})
  }, [form])

  const onSubmit = async (data: ListenerSettingsFormValues) => {
    setSaving(true)
    try {
      await settingsApi.update({ listener: data })
    } catch (error) {
      console.error('Failed to save listener settings:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Listener</h3>
      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <Label htmlFor="dns-listen">DNS Listen Address</Label>
          <Input
            id="dns-listen"
            className="mt-1"
            {...form.register('dns_listen')}
          />
          {form.formState.errors.dns_listen && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.dns_listen.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="http-listen">HTTP Listen Address</Label>
          <Input
            id="http-listen"
            className="mt-1"
            {...form.register('http_listen')}
          />
          {form.formState.errors.http_listen && (
            <p className="text-xs text-red-500 mt-1">{form.formState.errors.http_listen.message}</p>
          )}
        </div>
        <div>
          <Label htmlFor="https-listen">HTTPS Listen Address</Label>
          <Input
            id="https-listen"
            className="mt-1"
            {...form.register('https_listen')}
          />
        </div>
        <Button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save Settings'}</Button>
      </form>
    </div>
  )
}

function NotificationSettings() {
  const [notifications, setNotifications] = useState([
    { id: 'dns', name: 'DNS Hit', enabled: true, webhook_url: '', webhook_body: '{"protocol":"dns","domain":"{{domain}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'http', name: 'HTTP Hit', enabled: true, webhook_url: '', webhook_body: '{"protocol":"http","method":"{{method}}","path":"{{path}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'smtp', name: 'SMTP Hit', enabled: false, webhook_url: '', webhook_body: '{"protocol":"smtp","from":"{{from}}","to":"{{to}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'ldap', name: 'LDAP Hit', enabled: false, webhook_url: '', webhook_body: '{"protocol":"ldap","operation":"{{operation}}","dn":"{{dn}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'ftp', name: 'FTP Hit', enabled: false, webhook_url: '', webhook_body: '{"protocol":"ftp","command":"{{command}}","ip":"{{ip}}","timestamp":"{{timestamp}}"}' },
    { id: 'payload_expire', name: 'Payload Expired', enabled: false, webhook_url: '', webhook_body: '{"event":"payload_expire","token":"{{token}}","timestamp":"{{timestamp}}"}' },
  ])

  const updateNotification = (id: string, field: string, value: unknown) => {
    setNotifications(notifications.map(n => n.id === id ? { ...n, [field]: value } : n))
  }

  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
      <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Notification</h3>
      <div className="space-y-6">
        {notifications.map((notification) => (
          <div key={notification.id} className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center">
                <Checkbox
                  id={`notify-${notification.id}`}
                  checked={notification.enabled}
                  onCheckedChange={(checked) => updateNotification(notification.id, 'enabled', checked)}
                  className="mr-3"
                />
                <Label htmlFor={`notify-${notification.id}`} className="font-medium text-gray-900 dark:text-gray-100">{notification.name}</Label>
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
                  <Label htmlFor={`webhook-body-${notification.id}`}>Webhook Body (JSON Template)</Label>
                  <Textarea
                    id={`webhook-body-${notification.id}`}
                    className="font-mono text-sm"
                    rows={4}
                    value={notification.webhook_body}
                    onChange={(e) => updateNotification(notification.id, 'webhook_body', e.target.value)}
                    placeholder='{"key": "value"}'
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Available variables: {`{{domain}}, {{ip}}, {{timestamp}}, {{method}}, {{path}}, {{from}}, {{to}}, {{operation}}, {{dn}}, {{command}}, {{token}}`}
                  </p>
                </div>
              </div>
            )}
          </div>
        ))}

        <Button>Save Settings</Button>
      </div>
    </div>
  )
}

function TokenManagement() {
  const [apiKeys] = useState([
    { id: '1', name: 'Development Key', key_prefix: 'gdl_', created_at: '2024-01-01' },
    { id: '2', name: 'Production Key', key_prefix: 'gdl_', created_at: '2024-01-15' },
  ])

  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">API Keys</h3>
        <Button>Create New Key</Button>
      </div>
      <div className="space-y-4">
        {apiKeys.map((key) => (
          <div key={key.id} className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
            <div className="flex justify-between items-start">
              <div>
                <p className="font-medium text-gray-900 dark:text-gray-100">{key.name}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">Key: {key.key_prefix}****</p>
                <p className="text-xs text-gray-400">Created: {key.created_at}</p>
              </div>
              <Button variant="destructive" size="sm">
                Revoke
              </Button>
            </div>
          </div>
        ))}
        {apiKeys.length === 0 && (
          <p className="text-gray-500 dark:text-gray-400">No API Keys</p>
        )}
      </div>
    </div>
  )
}
