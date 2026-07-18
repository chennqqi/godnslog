'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { scannerRunApi } from '@/lib/api-client'
import type { ScannerRunDetail } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function ScannerRunDetailPage() {
  const router = useRouter()
  const params = useParams()
  const [scannerRun, setScannerRun] = useState<ScannerRunDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [updatingStatus, setUpdatingStatus] = useState(false)
  const [backfillFormat, setBackfillFormat] = useState<'jsonl' | 'sarif'>('jsonl')
  const [backfillResults, setBackfillResults] = useState('')
  const [backfilling, setBackfilling] = useState(false)
  const [backfillResult, setBackfillResult] = useState<{ findings_count: number; associated_count: number } | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      window.location.href = '/login'
      return
    }

    const loadScannerRun = async () => {
      try {
        const response = await scannerRunApi.get(params.id as string)
        if (response.data) {
          setScannerRun(response.data.data)
        }
      } catch (error: unknown) {
        console.error('Failed to load scanner run:', error)
        setError('加载Scanner Run失败')
      } finally {
        setLoading(false)
      }
    }

    loadScannerRun()
  }, [router, params.id])

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const handleUpdateStatus = async (newStatus: 'created' | 'distributed' | 'observed' | 'evidenced') => {
    if (!scannerRun) return
    setUpdatingStatus(true)
    setError('')
    try {
      await scannerRunApi.updateStatus(scannerRun.id, { status: newStatus })
      const response = await scannerRunApi.get(scannerRun.id)
      if (response.data) {
        setScannerRun(response.data.data)
      }
    } catch (error: unknown) {
      console.error('Failed to update status:', error)
      setError('更新状态失败')
    } finally {
      setUpdatingStatus(false)
    }
  }

  const handleBackfill = async () => {
    if (!scannerRun || !backfillResults.trim()) return
    setBackfilling(true)
    setError('')
    setBackfillResult(null)
    try {
      const response = await fetch(`/api/v2/scanner-runs/${scannerRun.id}/backfill`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({ format: backfillFormat, raw_results: backfillResults }),
      })
      const data = await response.json()
      if (data.code === 0 && data.data) {
        setBackfillResult({
          findings_count: data.data.findings_count,
          associated_count: data.data.associated_count,
        })
        // Reload scanner run to update interaction count
        const detailResp = await scannerRunApi.get(scannerRun.id)
        if (detailResp.data) {
          setScannerRun(detailResp.data.data)
        }
      } else {
        setError(data.message || 'Backfill failed')
      }
    } catch (err: unknown) {
      console.error('Backfill failed:', err)
      setError('Backfill request failed')
    } finally {
      setBackfilling(false)
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center h-screen">加载中...</div>
  }

  if (error || !scannerRun) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-red-500">{error || 'Scanner Run未找到'}</div>
        <Button onClick={() => router.back()} className="mt-4">
          返回
        </Button>
      </div>
    )
  }

  const jsonlData = scannerRun.jsonl ? JSON.parse(scannerRun.jsonl) : null

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <Button onClick={() => router.back()} variant="outline" className="mb-4">
          返回
        </Button>
        <h1 className="text-3xl font-bold">Scanner Run 详情</h1>
        <p className="text-muted-foreground">ID: {scannerRun.id}</p>
      </div>

      <div className="grid gap-6">
        {/* Basic Info */}
        <Card>
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge>状态</Badge>
              <span>{scannerRun.status}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Scanner</Badge>
              <span>{scannerRun.scanner}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Delivery Method</Badge>
              <span>{scannerRun.delivery_method}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Target</Badge>
              <span>{scannerRun.target}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>Template</Badge>
              <span>{scannerRun.template}</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge>创建时间</Badge>
              <span>{new Date(scannerRun.created_at).toLocaleString()}</span>
            </div>
          </CardContent>
        </Card>

        {/* Status Update */}
        <Card>
          <CardHeader>
            <CardTitle>状态更新</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex gap-2 flex-wrap">
              <Button
                onClick={() => handleUpdateStatus('created')}
                disabled={scannerRun.status === 'created' || updatingStatus}
                variant={scannerRun.status === 'created' ? 'default' : 'outline'}
                size="sm"
              >
                Created
              </Button>
              <Button
                onClick={() => handleUpdateStatus('distributed')}
                disabled={scannerRun.status === 'distributed' || updatingStatus}
                variant={scannerRun.status === 'distributed' ? 'default' : 'outline'}
                size="sm"
              >
                Distributed
              </Button>
              <Button
                onClick={() => handleUpdateStatus('observed')}
                disabled={scannerRun.status === 'observed' || updatingStatus}
                variant={scannerRun.status === 'observed' ? 'default' : 'outline'}
                size="sm"
              >
                Observed
              </Button>
              <Button
                onClick={() => handleUpdateStatus('evidenced')}
                disabled={scannerRun.status === 'evidenced' || updatingStatus}
                variant={scannerRun.status === 'evidenced' ? 'default' : 'outline'}
                size="sm"
              >
                Evidenced
              </Button>
            </div>
            {error && (
              <div className="text-red-500 text-sm">{error}</div>
            )}
          </CardContent>
        </Card>

        {/* Scope Info */}
        <Card>
          <CardHeader>
            <CardTitle>关联信息</CardTitle>
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
            <div className="flex items-center gap-2">
              <Badge>Interaction Count</Badge>
              <span>{scannerRun.interaction_count}</span>
            </div>
            {scannerRun.last_interaction_at && (
              <div className="flex items-center gap-2">
                <Badge>Last Interaction</Badge>
                <span>{new Date(scannerRun.last_interaction_at).toLocaleString()}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Badge>Evidence Count</Badge>
              <span>{scannerRun.evidence_count}</span>
            </div>
          </CardContent>
        </Card>

        {/* Nuclei Command */}
        <Card>
          <CardHeader>
            <CardTitle>Integration Package</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Input value={scannerRun.command} readOnly className="font-mono" />
              <Button onClick={() => handleCopy(scannerRun.command)}>
                复制
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Package Hash</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Input value={scannerRun.package_hash || ''} readOnly className="font-mono text-sm" />
              <Button onClick={() => handleCopy(scannerRun.package_hash || '')}>
                复制
              </Button>
            </div>
          </CardContent>
        </Card>

        {scannerRun.package_manifest && (
          <Card>
            <CardHeader>
              <CardTitle>Package Manifest</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid gap-2 text-sm md:grid-cols-2">
                <div className="flex items-center gap-2">
                  <Badge>Schema</Badge>
                  <span>{scannerRun.package_manifest.schema_version}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Badge>Hash</Badge>
                  <span>{scannerRun.package_manifest.hash_algorithm}</span>
                </div>
              </div>
              <div className="space-y-2">
                {scannerRun.package_manifest.files.map(file => (
                  <div key={`${file.kind}-${file.name}`} className="rounded border p-3 text-sm">
                    <div className="font-mono">{file.name}</div>
                    <div className="text-muted-foreground">{file.kind} · {file.description}</div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* JSONL Record */}
        <Card>
          <CardHeader>
            <CardTitle>JSONL Record</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Textarea value={scannerRun.jsonl} readOnly className="font-mono text-sm" rows={4} />
              <Button onClick={() => handleCopy(scannerRun.jsonl)}>
                复制
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* JSONL Parsed */}
        {jsonlData && (
          <Card>
            <CardHeader>
              <CardTitle>JSONL 解析</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex items-center gap-2">
                <Badge>Token</Badge>
                <span>{jsonlData.token}</span>
                <Button onClick={() => handleCopy(jsonlData.token)} size="sm" variant="outline">
                  复制
                </Button>
              </div>
              <div className="flex items-center gap-2">
                <Badge>Rendered Payload</Badge>
                <span className="font-mono text-sm">{jsonlData.rendered_payload}</span>
                <Button onClick={() => handleCopy(jsonlData.rendered_payload)} size="sm" variant="outline">
                  复制
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Backfill Scan Results */}
        <Card>
          <CardHeader>
            <CardTitle>Backfill Scan Results</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2 items-center">
              <span className="text-sm font-medium">Format:</span>
              <Select value={backfillFormat} onValueChange={(v: 'jsonl' | 'sarif') => setBackfillFormat(v)}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="jsonl">Nuclei JSONL</SelectItem>
                  <SelectItem value="sarif">SARIF</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Textarea
              value={backfillResults}
              onChange={(e) => setBackfillResults(e.target.value)}
              placeholder="Paste Nuclei JSONL or SARIF scan output here..."
              className="font-mono text-sm"
              rows={6}
            />
            <Button
              onClick={handleBackfill}
              disabled={backfilling || !backfillResults.trim()}
            >
              {backfilling ? 'Importing...' : 'Import Results'}
            </Button>
            {backfillResult && (
              <div className="rounded border p-3 text-sm">
                <span className="font-medium">Imported {backfillResult.findings_count} findings</span>
                {' — '}
                <span>{backfillResult.associated_count} associated with interactions</span>
              </div>
            )}
            {error && (
              <div className="text-red-500 text-sm">{error}</div>
            )}
          </CardContent>
        </Card>

        {/* Navigation Links */}
        <Card>
          <CardHeader>
            <CardTitle>查看结果</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button
              onClick={() => router.push(scannerRun.interactions_url)}
              className="w-full"
              variant="outline"
            >
              查看 Interactions ({scannerRun.interaction_count})
            </Button>
            <Button
              onClick={() => router.push(scannerRun.evidence_url)}
              className="w-full"
              variant="outline"
            >
              查看 Evidence ({scannerRun.evidence_count})
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
