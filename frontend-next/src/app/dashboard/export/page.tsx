'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'

export default function ExportPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [caseId, setCaseId] = useState('')
  const [payloadId, setPayloadId] = useState('')
  const [format, setFormat] = useState('json')
  const [includeRaw, setIncludeRaw] = useState(false)
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const handleExport = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v2/interactions/export', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          format,
          case_id: caseId || undefined,
          payload_id: payloadId || undefined,
          start_time: startTime ? new Date(startTime).toISOString() : undefined,
          end_time: endTime ? new Date(endTime).toISOString() : undefined,
          include_raw: includeRaw,
        }),
      })

      if (response.ok) {
        const blob = await response.blob()
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `evidence.${format}`
        document.body.appendChild(a)
        a.click()
        window.URL.revokeObjectURL(url)
        document.body.removeChild(a)
      } else {
        console.error('Export failed')
      }
    } catch (error) {
      console.error('Export error:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">证据导出</h2>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>导出配置</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Label htmlFor="format">导出格式</Label>
              <Select value={format} onValueChange={setFormat}>
                <SelectTrigger id="format">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="json">JSON</SelectItem>
                  <SelectItem value="markdown">Markdown</SelectItem>
                  <SelectItem value="csv">CSV</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="case-id">Case ID（可选）</Label>
              <Input
                id="case-id"
                placeholder="输入Case ID筛选"
                value={caseId}
                onChange={(e) => setCaseId(e.target.value)}
              />
            </div>

            <div>
              <Label htmlFor="payload-id">Payload ID（可选）</Label>
              <Input
                id="payload-id"
                placeholder="输入Payload ID筛选"
                value={payloadId}
                onChange={(e) => setPayloadId(e.target.value)}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="start-time">开始时间</Label>
                <Input
                  id="start-time"
                  type="datetime-local"
                  value={startTime}
                  onChange={(e) => setStartTime(e.target.value)}
                />
              </div>
              <div>
                <Label htmlFor="end-time">结束时间</Label>
                <Input
                  id="end-time"
                  type="datetime-local"
                  value={endTime}
                  onChange={(e) => setEndTime(e.target.value)}
                />
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="include-raw"
                checked={includeRaw}
                onCheckedChange={(checked) => setIncludeRaw(checked as boolean)}
              />
              <Label htmlFor="include-raw">包含原始数据</Label>
            </div>

            <Button onClick={handleExport} disabled={loading}>
              {loading ? '导出中...' : '开始导出'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
