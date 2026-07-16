'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { canaryApi, type CanaryToken as APICanaryToken } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

/** Canary token entry — extends API model with UI-only fields */
interface CanaryToken {
  id: string
  type: string
  token: string
  description: string
  context: string
  is_enabled: boolean
  status: 'active' | 'silent' | 'revoked'
  created_at: string
  expires_at: string
  expires_in?: number
  silent_window?: string
}

/** Type descriptor for available canary token kinds */
interface CanaryType {
  value: string
  label: string
  description: string
}

/** Status colour mapping */
const STATUS_STYLES: Record<string, string> = {
  active:  'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  silent:  'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  revoked: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
}

/** Protocol colour mapping */
const TYPE_STYLES: Record<string, string> = {
  dns:      'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  http:     'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  document: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
  config:   'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
  ci:       'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  storage:  'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
  email:    'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-400',
}

const CANARY_TYPES: CanaryType[] = [
  { value: 'dns',      label: 'DNS',          description: 'Triggers on DNS resolution' },
  { value: 'http',     label: 'HTTP',         description: 'Triggers on HTTP request' },
  { value: 'document', label: 'Document',     description: 'Triggers on document access' },
  { value: 'config',   label: 'Config File',  description: 'Triggers on config file read' },
  { value: 'ci',       label: 'CI Variable',  description: 'Triggers on CI env var access' },
  { value: 'storage',  label: 'Object Store', description: 'Triggers on storage access' },
  { value: 'email',    label: 'Email Address', description: 'Triggers on email delivery' },
]


/** Form state for creating a new canary token */
interface CreateForm {
  type: string
  context: string
  expires_in: number
  silent_window: string
}

const DEFAULT_FORM: CreateForm = {
  type: 'dns',
  context: '',
  expires_in: 2592000,
  silent_window: '',
}

/** Maps an API canary token to the UI format with computed status */
function mapApiToken(api: APICanaryToken): CanaryToken {
  let status: 'active' | 'silent' | 'revoked' = 'active'
  if (!api.is_enabled) {
    status = 'revoked'
  }
  return {
    id: api.id,
    type: api.type,
    token: api.token,
    description: api.description || '',
    context: api.context || api.description || '',
    is_enabled: api.is_enabled,
    status,
    created_at: api.created_at || '',
    expires_at: api.expires_at || '',
  }
}

/** Canary Tokens page — long-lived tripwire tokens for detecting unauthorized access */
export default function CanaryPage() {
  const router = useRouter()
  const [tokens, setTokens] = useState<CanaryToken[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [revokeTarget, setRevokeTarget] = useState<CanaryToken | null>(null)
  const [form, setForm] = useState<CreateForm>(DEFAULT_FORM)

  const loadTokens = useCallback(async () => {
    setLoading(true)
    try {
      const response = await canaryApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        const apiTokens = (response.data.items || []).map(mapApiToken)
        setTokens(apiTokens)
      }
    } catch (err) {
      console.error('Failed to load canary tokens:', err)
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
    setTimeout(() => {
      loadTokens()
    }, 0)
  }, [router, loadTokens])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    const generatedToken = `canary-${Math.random().toString(36).slice(2, 12)}`
    const expiresAt = new Date(Date.now() + form.expires_in * 1000).toISOString()
    try {
      await canaryApi.create({
        type: form.type,
        token: generatedToken,
        description: form.context,
        context: form.context,
        expires_at: expiresAt,
      })
      setForm(DEFAULT_FORM)
      setShowCreateModal(false)
      loadTokens()
    } catch (err) {
      console.error('Failed to create canary token:', err)
    }
  }

  const confirmRevoke = (t: CanaryToken) => setRevokeTarget(t)

  const executeRevoke = async () => {
    if (!revokeTarget) return
    try {
      await canaryApi.delete(revokeTarget.id)
      setTokens((prev) => prev.filter((t) => t.id !== revokeTarget.id))
      setRevokeTarget(null)
    } catch (err) {
      console.error('Failed to revoke canary token:', err)
    }
  }

  const activeCount  = tokens.filter((t) => t.status === 'active').length
  const silentCount  = tokens.filter((t) => t.status === 'silent').length
  const revokedCount = tokens.filter((t) => t.status === 'revoked').length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Canary Tokens</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Long-lived tripwire tokens for detecting unauthorized access to sensitive assets
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          New Canary Token
        </Button>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: 'Active',  value: activeCount,  cls: 'text-emerald-600' },
          { label: 'Silent',  value: silentCount,  cls: 'text-amber-600'   },
          { label: 'Revoked', value: revokedCount, cls: 'text-red-600'     },
        ].map(({ label, value, cls }) => (
          <Card key={label} className="dark:bg-gray-800 dark:border-gray-700">
            <CardContent className="p-4">
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{label}</p>
              <p className={`text-2xl font-bold mt-1 ${cls}`}>{value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Token list */}
      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardHeader className="border-b border-gray-100 dark:border-gray-700">
          <CardTitle className="text-sm font-semibold text-gray-700 dark:text-gray-300">
            Token Registry
          </CardTitle>
          <CardDescription className="text-xs">
            {tokens.length} token{tokens.length !== 1 ? 's' : ''} total
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <div className="text-center py-16">
              <p className="text-sm text-gray-500">Loading...</p>
            </div>
          ) : tokens.length === 0 ? (
            <div className="text-center py-16">
              <div className="text-4xl mb-3">🐦</div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">No canary tokens yet</p>
              <p className="text-xs text-gray-400 mt-1">Deploy tokens to sensitive assets to detect unauthorized access</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-100 dark:divide-gray-700">
              {tokens.map((t) => (
                <div key={t.id} className="flex items-start gap-4 p-4 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1.5 flex-wrap">
                      <span className={`px-2 py-0.5 text-xs font-medium rounded ${TYPE_STYLES[t.type] ?? 'bg-gray-100 text-gray-700'}`}>
                        {t.type.toUpperCase()}
                      </span>
                      <span className={`px-2 py-0.5 text-xs font-medium rounded capitalize ${STATUS_STYLES[t.status]}`}>
                        {t.status}
                      </span>
                    </div>
                    <p className="text-sm font-mono text-gray-900 dark:text-gray-100 break-all mb-1">
                      {t.token}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mb-0.5">
                      Context: {t.context}
                    </p>
                    <p className="text-xs text-gray-400">
                      Created: {t.created_at}
                      {t.silent_window && ` · Silent: ${t.silent_window}`}
                    </p>
                  </div>
                  <div className="flex gap-2 shrink-0">
                    {t.status !== 'revoked' && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-600 hover:text-red-800 dark:text-red-400"
                        onClick={() => confirmRevoke(t)}
                      >
                        Revoke
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create modal */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>New Canary Token</DialogTitle>
            <DialogDescription>
              Deploy a long-lived tripwire to a sensitive asset. You will be alerted when it is accessed.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4 pt-2">
            <div>
              <Label htmlFor="canary-type">Token type</Label>
              <Select value={form.type} onValueChange={(v) => setForm({ ...form, type: v })}>
                <SelectTrigger id="canary-type" className="mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CANARY_TYPES.map((ct) => (
                    <SelectItem key={ct.value} value={ct.value}>
                      {ct.label} — {ct.description}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="canary-context">Context / memo</Label>
              <textarea
                id="canary-context"
                className="mt-1 w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                rows={3}
                required
                placeholder="e.g. Project Alpha / AWS credentials backup / owner: security-team"
                value={form.context}
                onChange={(e) => setForm({ ...form, context: e.target.value })}
              />
              <p className="text-xs text-gray-400 mt-1">
                Describe where this token is deployed so you can act quickly when triggered.
              </p>
            </div>

            <div>
              <Label htmlFor="canary-expires">TTL (seconds)</Label>
              <Input
                id="canary-expires"
                type="number"
                className="mt-1"
                value={form.expires_in}
                min={86400}
                max={31536000}
                onChange={(e) => setForm({ ...form, expires_in: parseInt(e.target.value) || 2592000 })}
              />
              <p className="text-xs text-gray-400 mt-1">Default 30 days (2 592 000 s), max 1 year.</p>
            </div>

            <div>
              <Label htmlFor="canary-silent">Silent window (optional)</Label>
              <Input
                id="canary-silent"
                className="mt-1"
                placeholder="e.g. 0-6,18-24 (UTC hours to suppress alerts)"
                value={form.silent_window}
                onChange={(e) => setForm({ ...form, silent_window: e.target.value })}
              />
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button type="submit">Deploy token</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Revoke confirmation */}
      <Dialog open={!!revokeTarget} onOpenChange={(open) => { if (!open) setRevokeTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke Canary Token</DialogTitle>
            <DialogDescription>
              Are you sure you want to revoke <span className="font-mono font-semibold">{revokeTarget?.token}</span>?
              The token will stop accepting new interactions.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRevokeTarget(null)}>Cancel</Button>
            <Button variant="destructive" onClick={executeRevoke}>Revoke</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
