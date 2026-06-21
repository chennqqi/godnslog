'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  rebindingApi,
  type RebindingRule,
  type RebindingScenario,
  type RebindingSession,
} from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { useI18n } from '@/lib/i18n-context'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

/** Rebinding Lab page — DNS rebinding rule management with predefined scenarios */
export default function RebindingLabPage() {
  const router = useRouter()
  const { t } = useI18n()
  const [rules, setRules] = useState<RebindingRule[]>([])
  const [scenarios, setScenarios] = useState<RebindingScenario[]>([])
  const [sessions, setSessions] = useState<Record<string, RebindingSession[]>>({})
  const [loading, setLoading] = useState(true)
  const [scenarioDialog, setScenarioDialog] = useState<RebindingScenario | null>(null)
  const [domainInput, setDomainInput] = useState('')
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')

  const loadRules = useCallback(async () => {
    setLoading(true)
    try {
      const response = await rebindingApi.listRules({ page: 1, page_size: 100 })
      if (response.data) {
        setRules(response.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load rebinding rules:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  const loadScenarios = useCallback(async () => {
    try {
      const response = await rebindingApi.listScenarios()
      if (response.data) {
        setScenarios(response.data || [])
      }
    } catch (err) {
      console.error('Failed to load scenarios:', err)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    setTimeout(() => {
      loadRules()
      loadScenarios()
    }, 0)
  }, [router, loadRules, loadScenarios])

  const handleCreateFromScenario = async () => {
    if (!scenarioDialog || !domainInput.trim()) return
    setCreating(true)
    setError('')
    try {
      await rebindingApi.createFromScenario(scenarioDialog.name, { domain: domainInput.trim() })
      setScenarioDialog(null)
      setDomainInput('')
      loadRules()
    } catch (err) {
      console.error('Failed to create rule from scenario:', err)
      setError('Failed to create rule')
    } finally {
      setCreating(false)
    }
  }

  const handleDeleteRule = async (id: string) => {
    if (!confirm(t('rebinding.delete_confirm'))) return
    try {
      await rebindingApi.deleteRule(id)
      loadRules()
    } catch (err) {
      console.error('Failed to delete rule:', err)
    }
  }

  const handleToggleRule = async (rule: RebindingRule) => {
    try {
      await rebindingApi.updateRule(rule.id, { is_enabled: !rule.is_enabled })
      loadRules()
    } catch (err) {
      console.error('Failed to toggle rule:', err)
    }
  }

  const handleViewSessions = async (rule: RebindingRule) => {
    try {
      const response = await rebindingApi.listSessions(rule.id)
      if (response.data) {
        setSessions((prev) => ({
          ...prev,
          [rule.id]: response.data?.data?.sessions || [],
        }))
      }
    } catch (err) {
      console.error('Failed to load sessions:', err)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('rebinding.title')}</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            DNS rebinding rule management with predefined attack scenarios
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Predefined Scenarios */}
        <div className="lg:col-span-1">
          <Card className="dark:bg-gray-800 dark:border-gray-700">
            <CardHeader>
              <CardTitle className="text-sm font-semibold">{t('rebinding.scenarios')}</CardTitle>
              <CardDescription className="text-xs">
                Click a scenario to create a rule
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {scenarios.length === 0 && (
                <p className="text-xs text-gray-400">No scenarios available</p>
              )}
              {scenarios.map((sc) => (
                <div
                  key={sc.name}
                  className="p-3 border border-gray-200 dark:border-gray-600 rounded hover:border-indigo-600 cursor-pointer transition-colors"
                  onClick={() => {
                    setScenarioDialog(sc)
                    setDomainInput('')
                    setError('')
                  }}
                >
                  <p className="font-medium text-sm text-gray-900 dark:text-gray-100">{sc.name}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{sc.description}</p>
                  <div className="mt-2 flex gap-1 flex-wrap">
                    {sc.stages.map((s, i) => (
                      <Badge key={i} variant="secondary" className="text-xs">
                        {s.target_ip}
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>

        {/* Rules List */}
        <div className="lg:col-span-2">
          <Card className="dark:bg-gray-800 dark:border-gray-700">
            <CardHeader>
              <CardTitle className="text-sm font-semibold">{t('rebinding.rules')}</CardTitle>
              <CardDescription className="text-xs">
                {rules.length} rule{rules.length !== 1 ? 's' : ''} configured
              </CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              {loading ? (
                <div className="text-center py-16">
                  <p className="text-sm text-gray-500">Loading...</p>
                </div>
              ) : rules.length === 0 ? (
                <div className="text-center py-16">
                  <p className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('rebinding.no_rules')}</p>
                  <p className="text-xs text-gray-400 mt-1">Select a predefined scenario to create one</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100 dark:divide-gray-700">
                  {rules.map((rule) => (
                    <div key={rule.id} className="p-4">
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-mono text-sm text-gray-900 dark:text-gray-100">{rule.domain}</span>
                            <Badge variant={rule.is_enabled ? 'default' : 'secondary'}>
                              {rule.is_enabled ? t('rebinding.enabled') : t('rebinding.disabled')}
                            </Badge>
                          </div>
                          <div className="flex gap-1 flex-wrap mt-2">
                            {rule.stages.map((stage, i) => (
                              <div key={i} className="text-xs bg-gray-100 dark:bg-gray-700 rounded px-2 py-1">
                                <span className="font-medium">Stage {i + 1}:</span>{' '}
                                <span className="font-mono">{stage.target_ip}</span>
                                {' '}(TTL: {stage.ttl}s, Max: {stage.max_hits})
                              </div>
                            ))}
                          </div>
                        </div>
                        <div className="flex gap-2 shrink-0">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleViewSessions(rule)}
                          >
                            {t('rebinding.sessions')}
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleToggleRule(rule)}
                          >
                            {rule.is_enabled ? t('rebinding.disable') : t('rebinding.enable')}
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-red-600"
                            onClick={() => handleDeleteRule(rule.id)}
                          >
                            Delete
                          </Button>
                        </div>
                      </div>
                      {sessions[rule.id] && (
                        <div className="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
                          <p className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
                            Active Sessions ({sessions[rule.id].length})
                          </p>
                          {sessions[rule.id].length === 0 ? (
                            <p className="text-xs text-gray-400">{t('rebinding.no_sessions')}</p>
                          ) : (
                            <div className="space-y-1">
                              {sessions[rule.id].map((sess) => (
                                <div key={sess.id} className="text-xs flex gap-4">
                                  <span className="font-mono">{sess.source_ip}</span>
                                  <span>Stage: {sess.current_stage}</span>
                                  <span>Hits: {sess.hit_count}</span>
                                  <span className="text-gray-400">{sess.last_hit}</span>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Create from scenario dialog */}
      <Dialog open={!!scenarioDialog} onOpenChange={(open) => { if (!open) setScenarioDialog(null) }}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Create Rule from Scenario</DialogTitle>
            <DialogDescription>
              {scenarioDialog?.description}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <Label htmlFor="domain">Domain</Label>
              <Input
                id="domain"
                value={domainInput}
                onChange={(e) => setDomainInput(e.target.value)}
                placeholder="e.g. rebind.example.com"
              />
            </div>
            {scenarioDialog && (
              <div className="rounded border p-3 space-y-1">
                <p className="text-xs font-medium">Stages:</p>
                {scenarioDialog.stages.map((s, i) => (
                  <div key={i} className="text-xs text-gray-600 dark:text-gray-400">
                    {i + 1}. {s.target_ip} (TTL: {s.ttl}s, Max hits: {s.max_hits}) — {s.description}
                  </div>
                ))}
              </div>
            )}
            {error && <p className="text-red-500 text-sm">{error}</p>}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setScenarioDialog(null)}>Cancel</Button>
            <Button onClick={handleCreateFromScenario} disabled={creating || !domainInput.trim()}>
              {creating ? 'Creating...' : 'Create Rule'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
