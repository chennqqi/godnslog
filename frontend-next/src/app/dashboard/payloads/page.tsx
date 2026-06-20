'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { payloadApi } from '@/lib/api-client'
import { LoadingState } from '@/components/loading-state'
import type { Payload, PayloadCreateRequest } from '@/types'
import { Button } from '@/components/ui/button'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'

const templates = [
  { id: 'ssrf-http', name: 'SSRF HTTP', template: '{{.token}}.{{.domain}}' },
  { id: 'ssrf-cloud', name: 'SSRF Cloud Metadata', template: '{{.token}}.169.254.169.254.{{.domain}}' },
  { id: 'xxe', name: 'XXE External Entity', template: 'http://{{.token}}.{{.domain}}/xxe.dtd' },
  { id: 'rce', name: 'RCE Command Injection', template: 'curl http://{{.token}}.{{.domain}}' },
  { id: 'blind-sqli', name: 'Blind SQLi DNS', template: '{{.token}}.{{.domain}}' },
]

export default function PayloadsPage() {
  const router = useRouter()
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showBatchModal, setShowBatchModal] = useState(false)
  const [previewPayload, setPreviewPayload] = useState('')
  const [selectedTemplate, setSelectedTemplate] = useState(templates[0])
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [batchCount, setBatchCount] = useState(1)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadPayloads()
  }, [router])

  const loadPayloads = async () => {
    try {
      const response = await payloadApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setPayloads(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load payloads:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreatePayload = async (e: React.FormEvent) => {
    e.preventDefault()
    const req: PayloadCreateRequest = {
      case_id: '',
      template: selectedTemplate.id,
      variables,
    }
    try {
      const response = await payloadApi.create(req)
      if (response.code === 0) {
        setShowCreateModal(false)
        setVariables({})
        loadPayloads()
      }
    } catch (error) {
      console.error('Failed to create payload:', error)
    }
  }

  const handleBatchCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (batchCount < 1 || batchCount > 100) return
    try {
      await payloadApi.batchCreate({
        case_id: '',
        template: selectedTemplate.id,
        variables,
        count: batchCount,
      })
      setShowBatchModal(false)
      loadPayloads()
    } catch (error) {
      console.error('Failed to batch create payloads:', error)
    }
  }

  const updatePreview = () => {
    let preview = selectedTemplate.template
    Object.entries(variables).forEach(([key, value]) => {
      preview = preview.replace(`{{.${key}}}`, value)
    })
    setPreviewPayload(preview)
  }

  useEffect(() => {
    updatePreview()
  }, [selectedTemplate, variables])

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const filteredPayloads = payloads.filter(p =>
    p.token.toLowerCase().includes(filter.toLowerCase()) ||
    p.template.toLowerCase().includes(filter.toLowerCase())
  )

  if (loading) {
    return <LoadingState />
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Payload Studio</h2>
        <div className="flex space-x-2">
          <Button onClick={() => setShowCreateModal(true)}>
            Create Payload
          </Button>
          <Button variant="secondary" onClick={() => setShowBatchModal(true)}>
            Batch Create
          </Button>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 shadow rounded-lg mb-4 p-4 border border-gray-200 dark:border-gray-700">
        <Input
          placeholder="Search by token or template..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className="bg-white dark:bg-gray-800 shadow rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="px-4 py-5 sm:p-6">
          {filteredPayloads.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">No payloads yet</p>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              {filteredPayloads.map((payload) => (
                <li key={payload.id} className="py-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors" onClick={() => router.push(`/dashboard/payloads/${payload.id}`)}>
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <p className="text-sm font-medium text-indigo-600 dark:text-indigo-400">{payload.template}</p>
                        <Badge variant={
                          payload.status === 'hit' ? 'default' :
                          payload.status === 'deployed' ? 'secondary' :
                          payload.status === 'expired' ? 'destructive' :
                          'outline'
                        }>
                          {payload.status}
                        </Badge>
                      </div>
                      <div className="flex items-center space-x-2 mt-1">
                        <p className="text-sm text-gray-600 dark:text-gray-400 break-all">
                          Token: {payload.token}
                        </p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => { e.stopPropagation(); copyToClipboard(payload.token) }}
                        >
                          Copy
                        </Button>
                      </div>
                      {payload.rendered_payload && (
                        <div className="flex items-center space-x-2 mt-1">
                          <p className="text-xs text-gray-500 dark:text-gray-500 break-all">
                            Payload: {payload.rendered_payload}
                          </p>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => { e.stopPropagation(); copyToClipboard(payload.rendered_payload) }}
                          >
                            Copy
                          </Button>
                        </div>
                      )}
                      {payload.case_id && (
                        <p className="text-xs text-gray-400 mt-1">
                          Case ID: {payload.case_id}
                        </p>
                      )}
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-gray-400">
                        {new Date(payload.created_at).toLocaleString()}
                      </p>
                      {payload.expires_at && (
                        <p className="text-xs text-gray-400">
                          Expires: {new Date(payload.expires_at).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Create Modal */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create Payload</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreatePayload}>
            <div className="mb-4">
              <Label htmlFor="template">Select Template</Label>
              <Select value={selectedTemplate.id} onValueChange={(value) => setSelectedTemplate(templates.find(t => t.id === value) || templates[0])}>
                <SelectTrigger id="template">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {templates.map(t => (
                    <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="mb-4">
              <Label htmlFor="template-content">Template Content</Label>
              <Input
                id="template-content"
                value={selectedTemplate.template}
                readOnly
                className="bg-gray-50 dark:bg-gray-900"
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="variables">Variables</Label>
              <Input
                id="variables"
                placeholder='{"key": "value"}'
                value={JSON.stringify(variables)}
                onChange={(e) => {
                  try {
                    setVariables(JSON.parse(e.target.value))
                  } catch {}
                }}
              />
            </div>
            <div className="mb-4">
              <Label>Preview</Label>
              <div className="p-3 bg-gray-50 dark:bg-gray-900 rounded border border-gray-200 dark:border-gray-700">
                <p className="text-sm break-all text-gray-900 dark:text-gray-100">{previewPayload}</p>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button type="submit">
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Batch Create Modal */}
      <Dialog open={showBatchModal} onOpenChange={setShowBatchModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Batch Create Payloads</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleBatchCreate}>
            <div className="mb-4">
              <Label htmlFor="batch-template">Select Template</Label>
              <Select value={selectedTemplate.id} onValueChange={(value) => setSelectedTemplate(templates.find(t => t.id === value) || templates[0])}>
                <SelectTrigger id="batch-template">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {templates.map(t => (
                    <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="mb-4">
              <Label htmlFor="batch-count">Count (1-100)</Label>
              <Input
                id="batch-count"
                type="number"
                min="1"
                max="100"
                value={batchCount}
                onChange={(e) => setBatchCount(parseInt(e.target.value) || 1)}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowBatchModal(false)}>
                Cancel
              </Button>
              <Button type="submit" variant="secondary">
                Generate
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
