'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { payloadApi } from '@/lib/api-client'
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
    // Batch creation logic would go here
    alert('批量生成功能需要后端 API 支持')
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
    return <div className="text-center py-12">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Payload Studio</h2>
        <div className="flex space-x-2">
          <Button onClick={() => setShowCreateModal(true)}>
            创建 Payload
          </Button>
          <Button variant="secondary" onClick={() => setShowBatchModal(true)}>
            批量生成
          </Button>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg mb-4 p-4">
        <Input
          placeholder="搜索 token 或模板..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          {filteredPayloads.length === 0 ? (
            <p className="text-gray-500">暂无 Payloads</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {filteredPayloads.map((payload) => (
                <li key={payload.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <p className="text-sm font-medium text-indigo-600">{payload.template}</p>
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
                        <p className="text-sm text-gray-600 break-all">
                          Token: {payload.token}
                        </p>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => copyToClipboard(payload.token)}
                        >
                          复制
                        </Button>
                      </div>
                      {payload.rendered_payload && (
                        <div className="flex items-center space-x-2 mt-1">
                          <p className="text-xs text-gray-500 break-all">
                            Payload: {payload.rendered_payload}
                          </p>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => copyToClipboard(payload.rendered_payload)}
                          >
                            复制
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
                          过期: {new Date(payload.expires_at).toLocaleString()}
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
            <DialogTitle>创建 Payload</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreatePayload}>
            <div className="mb-4">
              <Label htmlFor="template">选择模板</Label>
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
              <Label htmlFor="template-content">模板内容</Label>
              <Input
                id="template-content"
                value={selectedTemplate.template}
                readOnly
                className="bg-gray-50"
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="variables">变量</Label>
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
              <Label>预览</Label>
              <div className="p-3 bg-gray-50 rounded border">
                <p className="text-sm break-all">{previewPayload}</p>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                取消
              </Button>
              <Button type="submit">
                创建
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Batch Create Modal */}
      <Dialog open={showBatchModal} onOpenChange={setShowBatchModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>批量生成 Payload</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleBatchCreate}>
            <div className="mb-4">
              <Label htmlFor="batch-count">生成数量 (1-100)</Label>
              <Input
                id="batch-count"
                type="number"
                min="1"
                max="100"
                value={batchCount}
                onChange={(e) => setBatchCount(parseInt(e.target.value))}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowBatchModal(false)}>
                取消
              </Button>
              <Button type="submit" variant="secondary">
                生成
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
