'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { usePayload, useRevokePayload } from '@/features/payloads/hooks/use-payloads'
import { payloadApi, caseApi, interactionApi } from '@/lib/api-client'
import type { Case, Interaction, Payload } from '@/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { LoadingState } from '@/components/loading-state'
import { useConfirmDialog } from '@/components/ui/alert-dialog'

export default function PayloadDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { data: payloadData, isLoading: loading } = usePayload(params.id as string)
  const revokePayload = useRevokePayload()
  const payload: Payload | null = (payloadData?.data as unknown as Payload) ?? null
  const [associatedCase, setAssociatedCase] = useState<Case | null>(null)
  const [recentInteractions, setRecentInteractions] = useState<Interaction[]>([])
  const [copied, setCopied] = useState(false)
  const [previewData, setPreviewData] = useState<string | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [revoking, setRevoking] = useState(false)
  const { confirm, dialogElement } = useConfirmDialog()

  const loadAssociatedData = useCallback(async (p: Payload) => {
    if (p.case_id) {
      try {
        const caseResp = await caseApi.get(p.case_id)
        if (caseResp.data && 'data' in caseResp.data) {
          setAssociatedCase(caseResp.data.data as Case)
        } else if (caseResp.data) {
          setAssociatedCase(caseResp.data as Case)
        }
      } catch (err) {
        console.error('Failed to load associated case:', err)
      }
    }

    try {
      const interactionsResp = await interactionApi.list({
        payload_id: p.id,
        page: 1,
        page_size: 5,
      })
      if (interactionsResp.data) {
        setRecentInteractions(interactionsResp.data.items || [])
      }
    } catch (err) {
      console.error('Failed to load recent interactions:', err)
    }
  }, [])

  useEffect(() => {
    if (payload) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      loadAssociatedData(payload)
    }
  }, [payload, loadAssociatedData])

  const handleCopyToken = async () => {
    if (!payload) return
    try {
      await navigator.clipboard.writeText(payload.token)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy token:', err)
    }
  }

  const handlePreview = async () => {
    if (!payload) return
    setPreviewLoading(true)
    try {
      const resp = await payloadApi.preview(payload.id)
      const data = resp.data as unknown as { data: { rendered: string } }
      setPreviewData(data?.data?.rendered || payload.rendered_payload || '')
    } catch (err) {
      console.error('Failed to preview payload:', err)
      setPreviewData(payload.rendered_payload || '')
    } finally {
      setPreviewLoading(false)
    }
  }

  const handleRevoke = async () => {
    if (!payload) return
    const ok = await confirm({
      title: 'Revoke Payload',
      description: 'Are you sure you want to revoke this payload? This action cannot be undone.',
      confirmLabel: 'Revoke',
      variant: 'destructive',
    })
    if (!ok) return
    setRevoking(true)
    try {
      await revokePayload.mutateAsync(payload.id)
    } catch (err) {
      console.error('Failed to revoke payload:', err)
    } finally {
      setRevoking(false)
    }
  }

  if (loading) {
    return <LoadingState />
  }

  if (!payload) {
    return <div className="text-center py-12 text-gray-500 dark:text-gray-400">Payload not found</div>
  }

  return (
    <>
    <div className="space-y-6">
      <button
        onClick={() => router.back()}
        className="text-indigo-600 dark:text-indigo-400 hover:underline text-sm"
      >
        ← Back
      </button>

      {/* Payload Header */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-start mb-4">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{payload.template}</h2>
            <Badge variant={
              payload.status === 'hit' ? 'default' :
              payload.status === 'deployed' ? 'secondary' :
              payload.status === 'archived' ? 'outline' :
              'outline'
            }>
              {payload.status}
            </Badge>
          </div>

          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between">
                <p className="text-sm text-gray-500 dark:text-gray-400">Token</p>
                <Button size="sm" variant="outline" onClick={handleCopyToken}>
                  {copied ? 'Copied!' : 'Copy'}
                </Button>
              </div>
              <p className="text-sm font-medium break-all bg-gray-50 dark:bg-gray-900 p-2 rounded mt-1 text-gray-900 dark:text-gray-100">{payload.token}</p>
            </div>

            {payload.rendered_payload && (
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Rendered Payload</p>
                <p className="text-sm font-medium break-all bg-gray-50 dark:bg-gray-900 p-2 rounded mt-1 text-gray-900 dark:text-gray-100">{payload.rendered_payload}</p>
              </div>
            )}

            {payload.scenario && (
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Scenario</p>
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{payload.scenario}</p>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Created</p>
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{new Date(payload.created_at).toLocaleString()}</p>
              </div>
              {payload.expires_at && (
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Expires</p>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{new Date(payload.expires_at).toLocaleString()}</p>
                </div>
              )}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-2 mt-4">
            <Button size="sm" variant="outline" onClick={handlePreview} disabled={previewLoading}>
              {previewLoading ? 'Previewing...' : 'Preview'}
            </Button>
            {payload.status !== 'archived' && (
              <Button size="sm" variant="destructive" onClick={handleRevoke} disabled={revoking}>
                {revoking ? 'Revoking...' : 'Revoke'}
              </Button>
            )}
          </div>

          {/* Preview Result */}
          {previewData !== null && (
            <div className="mt-4">
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Preview Result</p>
              <pre className="text-sm bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto text-gray-900 dark:text-gray-100 whitespace-pre-wrap break-all">{previewData}</pre>
            </div>
          )}
        </div>
      </div>

      {/* Associated Case */}
      {associatedCase && (
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Associated Case</h3>
            <div
              className="p-4 bg-gray-50 dark:bg-gray-900 rounded cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              onClick={() => router.push(`/cases/${associatedCase.id}`)}
            >
              <p className="text-sm font-medium text-indigo-600 dark:text-indigo-400">{associatedCase.title}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">{associatedCase.description}</p>
              {associatedCase.target && (
                <p className="text-xs text-gray-400 mt-1">Target: {associatedCase.target}</p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Recent Interactions */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">Recent Interactions</h3>
            <Button size="sm" variant="outline" onClick={() => router.push(`/interactions?payload_id=${payload.id}`)}>
              View All
            </Button>
          </div>
          {recentInteractions.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-sm">No interactions yet</p>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              {recentInteractions.map((interaction) => (
                <li key={interaction.id} className="py-3">
                  <div className="flex justify-between items-start">
                    <div className="flex items-center gap-3">
                      <Badge variant={interaction.type === 'dns' ? 'default' : 'secondary'}>
                        {interaction.type.toUpperCase()}
                      </Badge>
                      <div>
                        <p className="text-sm text-gray-700 dark:text-gray-300">
                          {interaction.source_ip}
                          {interaction.domain && ` | ${interaction.domain}`}
                        </p>
                      </div>
                    </div>
                    <p className="text-xs text-gray-400">
                      {new Date(interaction.timestamp).toLocaleString()}
                    </p>
                  </div>
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
            <Button variant="outline" onClick={() => router.push(`/interactions?payload_id=${payload.id}`)}>
              View Interactions
            </Button>
            <Button variant="outline" onClick={() => router.push(`/evidence?payload_id=${payload.id}`)}>
              View Evidence
            </Button>
          </div>
        </div>
      </div>
    </div>
    {dialogElement}
    </>
  )
}
