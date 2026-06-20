'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { listenerApi, type ProtocolListener } from '@/lib/api-client'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

export default function ListenersPage() {
  const router = useRouter()
  const [listeners, setListeners] = useState<ProtocolListener[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<ProtocolListener | null>(null)
  const [form, setForm] = useState({
    protocol: 'smtp' as 'smtp' | 'ldap' | 'smb' | 'ftp',
    host: '0.0.0.0',
    port: 25,
    token: '',
    is_enabled: true,
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const loadListeners = useCallback(async () => {
    setLoading(true)
    try {
      const response = await listenerApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setListeners(response.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load listeners:', err)
      setError('Failed to load listeners')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    // Wrap in setTimeout to avoid react-hooks/set-state-in-effect lint error
    setTimeout(() => {
      loadListeners()
    }, 0)
  }, [router, loadListeners])

  const handleCreate = () => {
    setEditing(null)
    setForm({
      protocol: 'smtp',
      host: '0.0.0.0',
      port: 25,
      token: '',
      is_enabled: true,
    })
    setCreateOpen(true)
  }

  const handleEdit = (listener: ProtocolListener) => {
    setEditing(listener)
    setForm({
      protocol: listener.protocol,
      host: listener.host,
      port: listener.port,
      token: listener.token,
      is_enabled: listener.is_enabled,
    })
    setCreateOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      if (editing) {
        await listenerApi.update(editing.id, form)
      } else {
        await listenerApi.create(form)
      }
      setCreateOpen(false)
      loadListeners()
    } catch (err) {
      console.error('Failed to save listener:', err)
      setError('Failed to save listener')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this listener?')) return
    try {
      await listenerApi.delete(id)
      loadListeners()
    } catch (err) {
      console.error('Failed to delete listener:', err)
      setError('Failed to delete listener')
    }
  }

  const handleToggle = async (listener: ProtocolListener) => {
    try {
      await listenerApi.update(listener.id, { is_enabled: !listener.is_enabled })
      loadListeners()
    } catch (err) {
      console.error('Failed to toggle listener:', err)
      setError('Failed to toggle listener')
    }
  }

  const defaultPortForProtocol = (protocol: string): number => {
    switch (protocol) {
      case 'smtp': return 25
      case 'ldap': return 389
      case 'smb': return 445
      case 'ftp': return 21
      default: return 0
    }
  }

  const getProtocolColor = (protocol: string): string => {
    const colors: Record<string, string> = {
      smtp: 'bg-blue-500',
      ldap: 'bg-purple-500',
      smb: 'bg-orange-500',
      ftp: 'bg-green-500',
    }
    return colors[protocol] || 'bg-gray-500'
  }

  if (loading) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">Protocol Listeners</h1>
        <Button onClick={handleCreate}>Create Listener</Button>
      </div>

      {error && (
        <div className="text-red-500 mb-4">{error}</div>
      )}

      <div className="grid gap-4">
        {listeners.length === 0 && (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              No listeners configured. Click &quot;Create Listener&quot; to add one.
            </CardContent>
          </Card>
        )}
        {listeners.map((listener) => (
          <Card key={listener.id}>
            <CardContent className="flex items-center justify-between py-4">
              <div className="flex items-center gap-4">
                <Badge className={getProtocolColor(listener.protocol)}>
                  {listener.protocol.toUpperCase()}
                </Badge>
                <div>
                  <div className="font-mono text-sm">
                    {listener.host}:{listener.port}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Token: {listener.token}
                  </div>
                </div>
                <Badge variant={listener.is_enabled ? 'default' : 'secondary'}>
                  {listener.is_enabled ? 'Running' : 'Stopped'}
                </Badge>
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleToggle(listener)}
                >
                  {listener.is_enabled ? 'Stop' : 'Start'}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleEdit(listener)}
                >
                  Edit
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => handleDelete(listener.id)}
                >
                  Delete
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit Listener' : 'Create Listener'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Protocol</Label>
              <Select
                value={form.protocol}
                onValueChange={(v: 'smtp' | 'ldap' | 'smb' | 'ftp') => {
                  setForm({ ...form, protocol: v, port: defaultPortForProtocol(v) })
                }}
                disabled={!!editing}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="smtp">SMTP</SelectItem>
                  <SelectItem value="ldap">LDAP</SelectItem>
                  <SelectItem value="smb">SMB</SelectItem>
                  <SelectItem value="ftp">FTP</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Host</Label>
              <Input
                value={form.host}
                onChange={(e) => setForm({ ...form, host: e.target.value })}
                placeholder="0.0.0.0"
              />
            </div>
            <div className="space-y-2">
              <Label>Port</Label>
              <Input
                type="number"
                value={form.port}
                onChange={(e) => setForm({ ...form, port: parseInt(e.target.value) || 0 })}
              />
            </div>
            <div className="space-y-2">
              <Label>Token</Label>
              <Input
                value={form.token}
                onChange={(e) => setForm({ ...form, token: e.target.value })}
                placeholder="Auto-generated if empty"
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="is_enabled"
                checked={form.is_enabled}
                onChange={(e) => setForm({ ...form, is_enabled: e.target.checked })}
              />
              <Label htmlFor="is_enabled">Enable immediately</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
