'use client'

/* eslint-disable react-hooks/set-state-in-effect */
import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi, payloadApi, scannerRunApi } from '@/lib/api-client'
import { createScannerRun, generateWebUrls, type ScannerRunInput } from '@/lib/scanner-hub'
import type { Case, Payload, ScannerRun } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

export default function ScannerHubPage() {
  const router = useRouter()
  const [cases, setCases] = useState<Case[]>([])
  const [selectedCase, setSelectedCase] = useState<string>('')
  const [target, setTarget] = useState('')
  const [template, setTemplate] = useState<'ssrf-basic' | 'xxe-basic' | 'rce-callback'>('ssrf-basic')
  const [selectedPayload, setSelectedPayload] = useState<string>('')
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)
  const [scannerRun, setScannerRun] = useState<ScannerRun | null>(null)
  const [error, setError] = useState<string>('')
  const [recentScannerRuns, setRecentScannerRuns] = useState<ScannerRun[]>([])
  const [loadingRuns, setLoadingRuns] = useState(false)

  const loadCases = useCallback(async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
      setError('加载Case列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  const loadPayloads = useCallback(async (caseId: string) => {
    try {
      const response = await payloadApi.list({ case_id: caseId, page: 1, page_size: 100 })
      if (response.data) {
        setPayloads(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load payloads:', error)
    }
  }, [])

  const loadRecentScannerRuns = useCallback(async () => {
    setLoadingRuns(true)
    try {
      const response = await scannerRunApi.list({ page: 1, page_size: 10 })
      if (response.data) {
        setRecentScannerRuns(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load scanner runs:', error)
    } finally {
      setLoadingRuns(false)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }
    loadCases()
    loadRecentScannerRuns()
  }, [router, loadCases, loadRecentScannerRuns])

  useEffect(() => {
    if (selectedCase) {
      loadPayloads(selectedCase)
    }
  }, [selectedCase, loadPayloads])

  const handleCreatePayload = async () => {
    if (!selectedCase || !template) {
      setError('请选择Case和Template')
      return
    }
    setGenerating(true)
    setError('')
    try {
      const response = await payloadApi.create({
        case_id: selectedCase,
        template: template,
        variables: {},
      })
      if (response.data && response.data.data) {
        const newPayload = response.data.data
        setSelectedPayload(newPayload.id)
        setPayloads([...payloads, newPayload])
      }
    } catch (error: unknown) {
      console.error('Failed to create payload:', error)
      setError('创建Payload失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleGenerateScannerRun = async () => {
    if (!selectedCase || !selectedPayload || !target) {
      setError('请选择Case、Payload并输入Target')
      return
    }

    const payload = payloads.find(p => p.id === selectedPayload)
    if (!payload) {
      setError('未找到选中的Payload')
      return
    }

    setGenerating(true)
    setError('')

    try {
      const input: ScannerRunInput = {
        case_id: selectedCase,
        payload_id: selectedPayload,
        token: payload.token,
        target,
        template,
        rendered_payload: payload.rendered_payload || payload.token,
        baseUrl: window.location.origin
      }

      const run = await createScannerRun(input, 'nuclei-jsonl')
      setScannerRun(run)
      loadRecentScannerRuns()
    } catch (error: unknown) {
      console.error('Failed to create scanner run:', error)
      setError('创建Scanner Run失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const webUrls = scannerRun ? generateWebUrls({
    case_id: scannerRun.case_id,
    payload_id: scannerRun.payload_id,
    token: '',
    target: scannerRun.target,
    template: scannerRun.template,
    rendered_payload: '',
    baseUrl: window.location.origin
  }) : null

  if (loading) {
    return <div className="flex items-center justify-center h-screen">加载中...</div>
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Scanner Hub</h1>
        <p className="text-muted-foreground">Nuclei 集成工作台</p>
      </div>

      <div className="grid gap-6">
        {/* Recent Scanner Runs */}
        <Card>
          <CardHeader>
            <CardTitle>最近的 Scanner Runs</CardTitle>
          </CardHeader>
          <CardContent>
            {loadingRuns ? (
              <div className="text-sm text-muted-foreground">加载中...</div>
            ) : recentScannerRuns.length === 0 ? (
              <div className="text-sm text-muted-foreground">暂无 Scanner Runs</div>
            ) : (
              <div className="space-y-2">
                {recentScannerRuns.map(run => (
                  <div
                    key={run.id}
                    className="flex items-center justify-between p-3 border rounded hover:bg-muted cursor-pointer"
                    onClick={() => router.push(`/dashboard/scanner-hub/${run.id}`)}
                  >
                    <div className="flex items-center gap-3">
                      <Badge variant={run.status === 'created' ? 'default' : 'secondary'}>
                        {run.status}
                      </Badge>
                      <div className="text-sm">
                        <div className="font-medium">{run.target}</div>
                        <div className="text-xs text-muted-foreground">
                          {run.template} · {new Date(run.created_at).toLocaleString()}
                        </div>
                      </div>
                    </div>
                    <Button size="sm" variant="ghost">
                      查看详情
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Case Selection */}
        <Card>
          <CardHeader>
            <CardTitle>选择 Case</CardTitle>
          </CardHeader>
          <CardContent>
            <Select value={selectedCase} onValueChange={setSelectedCase}>
              <SelectTrigger>
                <SelectValue placeholder="选择 Case" />
              </SelectTrigger>
              <SelectContent>
                {cases.map(c => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </CardContent>
        </Card>

        {/* Target Input */}
        <Card>
          <CardHeader>
            <CardTitle>输入 Target</CardTitle>
          </CardHeader>
          <CardContent>
            <Input
              placeholder="example.com"
              value={target}
              onChange={(e) => setTarget(e.target.value)}
            />
          </CardContent>
        </Card>

        {/* Template Selection */}
        <Card>
          <CardHeader>
            <CardTitle>选择 Template</CardTitle>
          </CardHeader>
          <CardContent>
            <Select value={template} onValueChange={(value: 'ssrf-basic' | 'xxe-basic' | 'rce-callback') => setTemplate(value)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ssrf-basic">SSRF Basic</SelectItem>
                <SelectItem value="xxe-basic">XXE Basic</SelectItem>
                <SelectItem value="rce-callback">RCE Callback</SelectItem>
              </SelectContent>
            </Select>
          </CardContent>
        </Card>

        {/* Payload Selection */}
        <Card>
          <CardHeader>
            <CardTitle>选择或创建 Payload</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Select value={selectedPayload} onValueChange={setSelectedPayload}>
              <SelectTrigger>
                <SelectValue placeholder="选择 Payload" />
              </SelectTrigger>
              <SelectContent>
                {payloads.map(p => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.token}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button onClick={handleCreatePayload} disabled={generating}>
              {generating ? '创建中...' : '创建新 Payload'}
            </Button>
          </CardContent>
        </Card>

        {/* Generate Button */}
        <Button onClick={handleGenerateScannerRun} className="w-full" size="lg" disabled={generating}>
          {generating ? '生成中...' : '生成 Scanner Run'}
        </Button>

        {error && (
          <div className="text-red-500">{error}</div>
        )}

        {/* Scanner Run Output */}
        {scannerRun && (
          <div className="grid gap-6">
            <Card>
              <CardHeader>
                <CardTitle>Token</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Input value={scannerRun.jsonl ? JSON.parse(scannerRun.jsonl).token : ''} readOnly />
                  <Button onClick={() => handleCopy(JSON.parse(scannerRun.jsonl).token)}>
                    复制
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Rendered Payload</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Input value={JSON.parse(scannerRun.jsonl).rendered_payload} readOnly />
                  <Button onClick={() => handleCopy(JSON.parse(scannerRun.jsonl).rendered_payload)}>
                    复制
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Nuclei Command</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Input value={scannerRun.command} readOnly />
                  <Button onClick={() => handleCopy(scannerRun.command)}>
                    复制
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>JSONL Preview</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Textarea value={scannerRun.jsonl} readOnly className="font-mono text-sm" />
                  <Button onClick={() => handleCopy(scannerRun.jsonl)}>
                    复制
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Scope Info */}
            <Card>
              <CardHeader>
                <CardTitle>当前 Scope</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="flex items-center gap-2">
                  <Badge>Case ID</Badge>
                  <span>{scannerRun.case_id}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Badge>Payload ID</Badge>
                  <span>{scannerRun.payload_id}</span>
                </div>
              </CardContent>
            </Card>

            {/* Navigation Links */}
            {webUrls && (
              <Card>
                <CardHeader>
                  <CardTitle>查看结果</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <Button
                    onClick={() => router.push(webUrls.interactionsUrl)}
                    className="w-full"
                    variant="outline"
                  >
                    查看 Interactions
                  </Button>
                  <Button
                    onClick={() => router.push(webUrls.evidenceUrl)}
                    className="w-full"
                    variant="outline"
                  >
                    查看 Evidence
                  </Button>
                </CardContent>
              </Card>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
