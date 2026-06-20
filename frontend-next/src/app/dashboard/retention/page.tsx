'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { retentionApi, type RetentionPolicy, type RetentionJob } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
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

const DEFAULT_FORM: Partial<RetentionPolicy> = {
  name: '',
  description: '',
  apply_to_interactions: true,
  apply_to_cases: false,
  apply_to_payloads: false,
  apply_to_evidence: false,
  apply_to_logs: false,
  retention_days: 90,
  max_records: 0,
  archive_after_days: 0,
  archive_to_storage: '',
  delete_after_archive: false,
  run_daily: true,
  run_hourly: false,
  run_weekly: false,
  run_monthly: false,
  is_enabled: true,
}

export default function RetentionPage() {
  const router = useRouter()
  const [policies, setPolicies] = useState<RetentionPolicy[]>([])
  const [jobs, setJobs] = useState<RetentionJob[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<RetentionPolicy | null>(null)
  const [form, setForm] = useState<Partial<RetentionPolicy>>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [policyRes, jobRes] = await Promise.all([
        retentionApi.listPolicies(),
        retentionApi.listJobs(),
      ])
      if (policyRes.data) {
        setPolicies(policyRes.data.items || [])
      }
      if (jobRes.data) {
        setJobs(jobRes.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load retention data:', err)
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
      load()
    }, 0)
  }, [router, load])

  const handleCreate = () => {
    setEditing(null)
    setForm(DEFAULT_FORM)
    setCreateOpen(true)
  }

  const handleEdit = (policy: RetentionPolicy) => {
    setEditing(policy)
    setForm({ ...policy })
    setCreateOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      if (editing) {
        await retentionApi.updatePolicy(editing.id, form)
      } else {
        await retentionApi.createPolicy(form)
      }
      setCreateOpen(false)
      load()
    } catch (err) {
      console.error('Failed to save policy:', err)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this retention policy?')) return
    try {
      await retentionApi.deletePolicy(id)
      load()
    } catch (err) {
      console.error('Failed to delete policy:', err)
    }
  }

  const handleRun = async (id: string) => {
    try {
      await retentionApi.runPolicy(id)
      load()
    } catch (err) {
      console.error('Failed to run policy:', err)
    }
  }

  const handleToggle = async (policy: RetentionPolicy) => {
    try {
      await retentionApi.updatePolicy(policy.id, { is_enabled: !policy.is_enabled })
      load()
    } catch (err) {
      console.error('Failed to toggle policy:', err)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Data Retention</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage data lifecycle policies with automatic cleanup and archival
          </p>
        </div>
        <Button onClick={handleCreate}>Create Policy</Button>
      </div>

      {/* Policies */}
      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardHeader>
          <CardTitle className="text-sm font-semibold">Retention Policies</CardTitle>
          <CardDescription className="text-xs">
            {policies.length} polic{policies.length !== 1 ? 'ies' : 'y'} configured
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <div className="text-center py-16">
              <p className="text-sm text-gray-500">Loading...</p>
            </div>
          ) : policies.length === 0 ? (
            <div className="text-center py-16">
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">No retention policies yet</p>
              <p className="text-xs text-gray-400 mt-1">Create a policy to manage data lifecycle</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-100 dark:divide-gray-700">
              {policies.map((policy) => (
                <div key={policy.id} className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-sm text-gray-900 dark:text-gray-100">{policy.name}</span>
                        <Badge variant={policy.is_enabled ? 'default' : 'secondary'}>
                          {policy.is_enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                      </div>
                      {policy.description && (
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-2">{policy.description}</p>
                      )}
                      <div className="flex gap-2 flex-wrap text-xs">
                        <span className="bg-gray-100 dark:bg-gray-700 rounded px-2 py-0.5">
                          Retain: {policy.retention_days}d
                        </span>
                        {policy.max_records > 0 && (
                          <span className="bg-gray-100 dark:bg-gray-700 rounded px-2 py-0.5">
                            Max: {policy.max_records}
                          </span>
                        )}
                        {policy.apply_to_interactions && <Badge variant="outline">Interactions</Badge>}
                        {policy.apply_to_cases && <Badge variant="outline">Cases</Badge>}
                        {policy.apply_to_payloads && <Badge variant="outline">Payloads</Badge>}
                        {policy.run_daily && <Badge variant="outline">Daily</Badge>}
                        {policy.run_hourly && <Badge variant="outline">Hourly</Badge>}
                        {policy.run_weekly && <Badge variant="outline">Weekly</Badge>}
                        {policy.run_monthly && <Badge variant="outline">Monthly</Badge>}
                        {policy.last_run_at && (
                          <span className="text-gray-400">Last run: {policy.last_run_at}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <Button size="sm" variant="ghost" onClick={() => handleRun(policy.id)}>Run</Button>
                      <Button size="sm" variant="ghost" onClick={() => handleToggle(policy)}>
                        {policy.is_enabled ? 'Disable' : 'Enable'}
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => handleEdit(policy)}>Edit</Button>
                      <Button size="sm" variant="ghost" className="text-red-600" onClick={() => handleDelete(policy.id)}>Delete</Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Recent Jobs */}
      {jobs.length > 0 && (
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardHeader>
            <CardTitle className="text-sm font-semibold">Recent Jobs</CardTitle>
            <CardDescription className="text-xs">Last {jobs.length} retention executions</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-gray-100 dark:divide-gray-700">
              {jobs.slice(0, 10).map((job) => (
                <div key={job.id} className="p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Badge variant={job.status === 'completed' ? 'default' : 'secondary'}>
                      {job.status}
                    </Badge>
                    <span className="text-xs text-gray-500">
                      Processed: {job.records_processed} | Deleted: {job.records_deleted}
                    </span>
                    <span className="text-xs text-gray-400">{job.started_at}</span>
                  </div>
                  <span className="text-xs text-gray-400">{job.duration}ms</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Create/Edit Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-lg max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit Policy' : 'Create Retention Policy'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={form.name || ''} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. Default 90-day retention" />
            </div>
            <div>
              <Label htmlFor="description">Description</Label>
              <Input id="description" value={form.description || ''} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="Optional description" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="retention_days">Retention Days</Label>
                <Input id="retention_days" type="number" value={form.retention_days || 0} onChange={(e) => setForm({ ...form, retention_days: parseInt(e.target.value) || 0 })} />
              </div>
              <div>
                <Label htmlFor="max_records">Max Records (0 = unlimited)</Label>
                <Input id="max_records" type="number" value={form.max_records || 0} onChange={(e) => setForm({ ...form, max_records: parseInt(e.target.value) || 0 })} />
              </div>
            </div>
            <div>
              <Label>Apply To</Label>
              <div className="flex gap-4 mt-2 flex-wrap">
                <div className="flex items-center gap-2">
                  <Checkbox id="apply_interactions" checked={form.apply_to_interactions} onCheckedChange={(v) => setForm({ ...form, apply_to_interactions: !!v })} />
                  <Label htmlFor="apply_interactions">Interactions</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox id="apply_cases" checked={form.apply_to_cases} onCheckedChange={(v) => setForm({ ...form, apply_to_cases: !!v })} />
                  <Label htmlFor="apply_cases">Cases</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox id="apply_payloads" checked={form.apply_to_payloads} onCheckedChange={(v) => setForm({ ...form, apply_to_payloads: !!v })} />
                  <Label htmlFor="apply_payloads">Payloads</Label>
                </div>
              </div>
            </div>
            <div>
              <Label>Schedule</Label>
              <div className="flex gap-4 mt-2 flex-wrap">
                <div className="flex items-center gap-2">
                  <Checkbox id="run_hourly" checked={form.run_hourly} onCheckedChange={(v) => setForm({ ...form, run_hourly: !!v })} />
                  <Label htmlFor="run_hourly">Hourly</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox id="run_daily" checked={form.run_daily} onCheckedChange={(v) => setForm({ ...form, run_daily: !!v })} />
                  <Label htmlFor="run_daily">Daily</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox id="run_weekly" checked={form.run_weekly} onCheckedChange={(v) => setForm({ ...form, run_weekly: !!v })} />
                  <Label htmlFor="run_weekly">Weekly</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox id="run_monthly" checked={form.run_monthly} onCheckedChange={(v) => setForm({ ...form, run_monthly: !!v })} />
                  <Label htmlFor="run_monthly">Monthly</Label>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Checkbox id="is_enabled" checked={form.is_enabled} onCheckedChange={(v) => setForm({ ...form, is_enabled: !!v })} />
              <Label htmlFor="is_enabled">Enable immediately</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={saving || !form.name}>
              {saving ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
