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

export default function ScannerRunDetailPage() {
  const router = useRouter()
  const params = useParams()
  const [scannerRun, setScannerRun] = useState<ScannerRunDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [updatingStatus, setUpdatingStatus] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
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
            <CardTitle>Nuclei Command</CardTitle>
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
