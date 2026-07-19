'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { caseApi, payloadApi, interactionApi } from '@/lib/api-client'
import { LoadingState } from '@/components/loading-state'
import type { Case, Payload, Interaction } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { caseSchema, type CaseFormValues } from '@/features/cases/schemas/case-schema'

export default function CaseDetailPage() {
  const params = useParams()
  const router = useRouter()
  const [case_, setCase] = useState<Case | null>(null)
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [interactions, setInteractions] = useState<Interaction[]>([])
  const [stats, setStats] = useState({ payload_count: 0, interaction_count: 0, hit_payload_count: 0 })
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  const form = useForm<CaseFormValues>({
    resolver: zodResolver(caseSchema),
    defaultValues: {
      title: '',
      description: '',
      target: '',
      status: 'active',
      tags: [],
    },
  })

  const loadData = useCallback(async () => {
    try {
      const [caseResp, payloadsResp, statsResp, interactionsResp] = await Promise.all([
        caseApi.get(params.id as string),
        payloadApi.list({ case_id: params.id as string, page: 1, page_size: 100 }),
        caseApi.stats(params.id as string),
        interactionApi.list({ case_id: params.id as string, page: 1, page_size: 10 }),
      ])

      const caseData = caseResp.data && 'data' in caseResp.data ? caseResp.data.data : caseResp.data
      if (caseData) {
        setCase(caseData as Case)
        form.reset({
          title: (caseData as Case).title || '',
          description: (caseData as Case).description || '',
          target: (caseData as Case).target || '',
          status: (caseData as Case).status || 'active',
          tags: (caseData as Case).tags || [],
        })
      }

      if (payloadsResp.data) {
        setPayloads(payloadsResp.data.items || [])
      }
      if (statsResp.data) {
        setStats({
          payload_count: statsResp.data.payload_count || 0,
          interaction_count: statsResp.data.interaction_count || 0,
          hit_payload_count: statsResp.data.hit_payload_count || 0,
        })
      }
      if (interactionsResp.data) {
        setInteractions(interactionsResp.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load case data:', error)
    } finally {
      setLoading(false)
    }
  }, [params.id, form])

  useEffect(() => {
    if (params.id) {
      loadData()
    }
  }, [params.id, loadData])

  const handleSaveEdit = async (values: CaseFormValues) => {
    setSaving(true)
    try {
      await caseApi.update(params.id as string, values)
      setEditing(false)
      loadData()
    } catch (error) {
      console.error('Failed to update case:', error)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <LoadingState />
  }

  if (!case_) {
    return <div className="text-center py-12 text-gray-500 dark:text-gray-400">Case not found</div>
  }

  return (
    <div className="space-y-6">
      <button
        onClick={() => router.back()}
        className="text-indigo-600 dark:text-indigo-400 hover:underline text-sm"
      >
        ← Back
      </button>

      {/* Case Header */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {editing ? (
            <form onSubmit={form.handleSubmit(handleSaveEdit)} className="space-y-4">
              <div>
                <Label htmlFor="edit-title">Title</Label>
                <Input
                  id="edit-title"
                  {...form.register('title')}
                />
                {form.formState.errors.title && (
                  <p className="text-xs text-red-500 mt-1">{form.formState.errors.title.message}</p>
                )}
              </div>
              <div>
                <Label htmlFor="edit-description">Description</Label>
                <Textarea
                  id="edit-description"
                  rows={3}
                  {...form.register('description')}
                />
              </div>
              <div>
                <Label htmlFor="edit-target">Target</Label>
                <Input
                  id="edit-target"
                  placeholder="e.g. internal-api.corp.com"
                  {...form.register('target')}
                />
              </div>
              <div>
                <Label htmlFor="edit-status">Status</Label>
                <Select
                  value={form.watch('status')}
                  onValueChange={(val) => form.setValue('status', val as 'active' | 'completed' | 'archived')}
                >
                  <SelectTrigger id="edit-status" className="w-[180px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                    <SelectItem value="archived">Archived</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex gap-2">
                <Button type="submit" disabled={saving}>
                  {saving ? 'Saving...' : 'Save'}
                </Button>
                <Button type="button" variant="outline" onClick={() => setEditing(false)}>
                  Cancel
                </Button>
              </div>
            </form>
          ) : (
            <>
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{case_.title}</h2>
                  <p className="text-gray-600 dark:text-gray-400 mt-2">{case_.description}</p>
                  {case_.target && (
                    <p className="text-sm text-gray-500 dark:text-gray-500 mt-2">Target: {case_.target}</p>
                  )}
                  {case_.tags && case_.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {case_.tags.map((tag) => (
                        <Badge key={tag} variant="outline">{tag}</Badge>
                      ))}
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <Badge variant={
                    case_.status === 'active' ? 'default' :
                    case_.status === 'completed' ? 'secondary' :
                    'outline'
                  }>
                    {case_.status}
                  </Badge>
                  <Button variant="outline" size="sm" onClick={() => setEditing(true)}>
                    Edit
                  </Button>
                </div>
              </div>
              <div className="mt-4 text-sm text-gray-500 dark:text-gray-500">
                Created at: {new Date(case_.created_at).toLocaleString()}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 uppercase">Payloads</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.payload_count}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 uppercase">Interactions</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.interaction_count}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
          <p className="text-xs text-gray-500 dark:text-gray-400 uppercase">Hit Payloads</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.hit_payload_count}</p>
        </div>
      </div>

      {/* Associated Payloads */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">Payloads</h3>
            <Button size="sm" onClick={() => router.push(`/payloads/new?case_id=${case_.id}`)}>
              Create Payload
            </Button>
          </div>
          {payloads.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-sm">No payloads yet</p>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              {payloads.map((payload) => (
                <li
                  key={payload.id}
                  className="py-4 flex justify-between items-center cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  onClick={() => router.push(`/payloads/${payload.id}`)}
                >
                  <div>
                    <p className="text-sm font-medium text-indigo-600 dark:text-indigo-400">{payload.template}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Token: {payload.token}</p>
                    <p className="text-xs text-gray-400 mt-1">Status: {payload.status}</p>
                  </div>
                  <span className="text-xs text-gray-400">
                    {new Date(payload.created_at).toLocaleDateString()}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Recent Interactions */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">Recent Interactions</h3>
            <Button size="sm" variant="outline" onClick={() => router.push(`/interactions?case_id=${case_.id}`)}>
              View All
            </Button>
          </div>
          {interactions.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-sm">No interactions yet</p>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              {interactions.map((inter) => (
                <li
                  key={inter.id}
                  className="py-3 flex justify-between items-center cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  onClick={() => router.push(`/interactions/${inter.id}`)}
                >
                  <div className="flex items-center gap-3">
                    <Badge variant={inter.type === 'dns' ? 'default' : 'secondary'}>
                      {inter.type}
                    </Badge>
                    <div>
                      <p className="text-sm text-gray-700 dark:text-gray-300">
                        {inter.domain || inter.path || inter.source_ip}
                      </p>
                      <p className="text-xs text-gray-400">Token: {inter.token}</p>
                    </div>
                  </div>
                  <span className="text-xs text-gray-400">
                    {new Date(inter.timestamp).toLocaleString()}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Quick Actions</h3>
          <div className="flex flex-wrap gap-3">
            <Button variant="outline" onClick={() => router.push(`/evidence?case_id=${case_.id}`)}>
              View Evidence
            </Button>
            <Button variant="outline" onClick={() => router.push(`/interactions?case_id=${case_.id}`)}>
              View Interactions
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
