'use client'

/* eslint-disable react-hooks/set-state-in-effect */
import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { caseApi, payloadApi, scannerRunApi, searchApi } from '@/lib/api-client'
import { createScannerRun, generateWebUrls, type ScannerRunInput } from '@/lib/scanner-hub'
import type { Case, Payload, ScannerAdapter, ScannerDeliveryMethod, ScannerKind, ScannerRun, SearchResultItem, SearchResult } from '@/types'
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { useI18n } from '@/lib/i18n-context'

export default function ScannerHubPage() {
  const router = useRouter()
  const { t } = useI18n()
  const [cases, setCases] = useState<Case[]>([])
  const [selectedCase, setSelectedCase] = useState<string>('')
  const [target, setTarget] = useState('')
  const [template, setTemplate] = useState<'ssrf-basic' | 'xxe-basic' | 'rce-callback'>('ssrf-basic')
  const [selectedPayload, setSelectedPayload] = useState<string>('')
  const [adapters, setAdapters] = useState<ScannerAdapter[]>([])
  const [selectedScanner, setSelectedScanner] = useState<ScannerKind>('nuclei')
  const [selectedDeliveryMethod, setSelectedDeliveryMethod] = useState<ScannerDeliveryMethod>('nuclei-jsonl')
  const [payloads, setPayloads] = useState<Payload[]>([])
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)
  const [scannerRun, setScannerRun] = useState<ScannerRun | null>(null)
  const [error, setError] = useState<string>('')
  const [recentScannerRuns, setRecentScannerRuns] = useState<ScannerRun[]>([])
  const [loadingRuns, setLoadingRuns] = useState(false)

  // Search from search engines
  const [searchQuery, setSearchQuery] = useState('')
  const [searchEngine, setSearchEngine] = useState<'zoomeye' | 'shodan' | 'fofa'>('shodan')
  const [searchResults, setSearchResults] = useState<SearchResult | null>(null)
  const [searching, setSearching] = useState(false)
  const [selectedSearchItems, setSelectedSearchItems] = useState<Set<number>>(new Set())
  const [creatingFromSearch, setCreatingFromSearch] = useState(false)
  const [showConfirmCreate, setShowConfirmCreate] = useState(false)
  const searchCardRef = useRef<HTMLDivElement>(null)

  const loadCases = useCallback(async () => {
    try {
      const response = await caseApi.list({ page: 1, page_size: 100 })
      if (response.data) {
        setCases(response.data.items || [])
      }
    } catch (error) {
      console.error('Failed to load cases:', error)
      setError(t('scanner_hub.load_cases_failed'))
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

  const loadAdapters = useCallback(async () => {
    try {
      const response = await scannerRunApi.listAdapters()
      if (response.data) {
        const items = response.data.items || []
        setAdapters(items)
        const defaultAdapter = items.find(adapter => adapter.id === selectedScanner) || items[0]
        if (defaultAdapter) {
          setSelectedScanner(defaultAdapter.id)
          setSelectedDeliveryMethod(defaultAdapter.default_method)
        }
      }
    } catch (error) {
      console.error('Failed to load scanner adapters:', error)
        setError(t('scanner_hub.load_adapters_failed'))
    }
  }, [selectedScanner])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      window.location.href = '/login'
      return
    }
    loadAdapters()
    loadCases()
    loadRecentScannerRuns()
  }, [router, loadAdapters, loadCases, loadRecentScannerRuns])

  useEffect(() => {
    if (selectedCase) {
      loadPayloads(selectedCase)
    }
  }, [selectedCase, loadPayloads])

  const handleCreatePayload = async () => {
    if (!selectedCase || !template) {
      setError(t('scanner_hub.select_case_template'))
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
        setPayloads(current => [...current.filter(payload => payload.id !== newPayload.id), newPayload])
      }
    } catch (error: unknown) {
      console.error('Failed to create payload:', error)
      setError(t('scanner_hub.create_payload_failed'))
    } finally {
      setGenerating(false)
    }
  }

  const handleGenerateScannerRun = async () => {
    if (!selectedCase || !selectedPayload || !target) {
      setError(t('scanner_hub.select_case_payload_target'))
      return
    }

    const payload = payloads.find(p => p.id === selectedPayload)
    if (!payload) {
      setError(t('scanner_hub.payload_not_found'))
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

      const run = await createScannerRun(input, selectedScanner, selectedDeliveryMethod)
      setScannerRun(run)
      loadRecentScannerRuns()
    } catch (error: unknown) {
      console.error('Failed to create scanner run:', error)
      setError(t('scanner_hub.create_run_failed'))
    } finally {
      setGenerating(false)
    }
  }

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const handleSearch = async () => {
    if (!searchQuery) return
    setSearching(true)
    setError('')
    setSearchResults(null)
    setSelectedSearchItems(new Set())
    try {
      const searchFn = searchApi[searchEngine]
      const response = await searchFn({ q: searchQuery })
      if (response.data) {
        setSearchResults(response.data as SearchResult)
        // Auto-scroll to results after search completes
        setTimeout(() => {
          searchCardRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
        }, 100)
      }
    } catch (err: unknown) {
      console.error('Search failed:', err)
      setError(t('scanner_hub.search_failed'))
    } finally {
      setSearching(false)
    }
  }

  const toggleSearchItem = (index: number) => {
    setSelectedSearchItems(prev => {
      const next = new Set(prev)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  const handleCreateFromSearch = async () => {
    if (!selectedCase || !selectedPayload || selectedSearchItems.size === 0) return
    if (!searchResults) return

    setCreatingFromSearch(true)
    setError('')
    setShowConfirmCreate(false)
    try {
      const items = Array.from(selectedSearchItems).map(i => searchResults.results[i])
      const response = await scannerRunApi.createFromSearch({
        case_id: selectedCase,
        payload_id: selectedPayload,
        source: searchEngine,
        results: items.map(item => ({
          ip: item.ip,
          port: item.port,
          protocol: item.protocol,
          hostname: item.hostname,
        })),
        scanner: selectedScanner,
        template,
        delivery_method: selectedDeliveryMethod,
      })
      if (response.data) {
        loadRecentScannerRuns()
        setSelectedSearchItems(new Set())
      }
    } catch (err: unknown) {
      console.error('Failed to create scanner runs from search:', err)
      setError(t('scanner_hub.create_from_search_failed'))
    } finally {
      setCreatingFromSearch(false)
    }
  }

  const handleScannerChange = (scanner: ScannerKind) => {
    const adapter = adapters.find(item => item.id === scanner)
    setSelectedScanner(scanner)
    if (adapter) {
      setSelectedDeliveryMethod(adapter.default_method)
    }
  }

  const selectedAdapter = adapters.find(adapter => adapter.id === selectedScanner)
  const outputLabel = scannerRun ? getScannerOutputLabel(scannerRun.scanner) : 'Integration Package'

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
    return <div className="flex items-center justify-center h-screen">{t('common.loading')}</div>
  }

  return (
    <div className="container mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">{t('scanner_hub.title')}</h1>
        <p className="text-muted-foreground">{t('scanner_hub.subtitle')}</p>
      </div>

      <div className="grid gap-6">
        {/* Recent Scanner Runs */}
        <Card>
          <CardHeader>
            <CardTitle>{t('scanner_hub.recent_runs')}</CardTitle>
          </CardHeader>
          <CardContent>
            {loadingRuns ? (
              <div className="text-sm text-muted-foreground">{t('common.loading')}</div>
            ) : recentScannerRuns.length === 0 ? (
              <div className="text-sm text-muted-foreground">{t('scanner_hub.no_runs')}</div>
            ) : (
              <div className="space-y-2">
                {recentScannerRuns.map(run => (
                  <div
                    key={run.id}
                    className="flex items-center justify-between p-3 border rounded hover:bg-muted cursor-pointer"
                    onClick={() => router.push(`/scanner-hub/${run.id}`)}
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
                      {t('scanner_hub.view_detail')}
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Search Engines */}
        <Card ref={searchCardRef}>
          <CardHeader>
            <CardTitle>{t('scanner_hub.search_engines')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <Select value={searchEngine} onValueChange={(v: 'zoomeye' | 'shodan' | 'fofa') => setSearchEngine(v)}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="shodan">Shodan</SelectItem>
                  <SelectItem value="zoomeye">ZoomEye</SelectItem>
                  <SelectItem value="fofa">Fofa</SelectItem>
                </SelectContent>
              </Select>
              <Input
                placeholder={t('scanner_hub.search_placeholder')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleSearch() }}
              />
              <Button onClick={handleSearch} disabled={searching || !searchQuery}>
                {searching ? t('common.loading') : t('scanner_hub.search')}
              </Button>
            </div>

            {searchResults && (
              <div className="space-y-3">
                <div className="text-sm text-muted-foreground">
                  {t('scanner_hub.found_results')}{searchResults.total}
                </div>
                {searchResults.results.length > 0 ? (
                  <>
                    <div className="border rounded">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b bg-muted/50">
                            <th className="p-2 w-10">
                              <input
                                type="checkbox"
                                onChange={(e) => {
                                  if (e.target.checked) {
                                    setSelectedSearchItems(new Set(searchResults.results.map((_, i) => i)))
                                  } else {
                                    setSelectedSearchItems(new Set())
                                  }
                                }}
                                checked={selectedSearchItems.size === searchResults.results.length}
                              />
                            </th>
                            <th className="p-2 text-left">IP</th>
                            <th className="p-2 text-left">{t('scanner_hub.port')}</th>
                            <th className="p-2 text-left">{t('scanner_hub.protocol')}</th>
                            <th className="p-2 text-left">{t('scanner_hub.hostname')}</th>
                            <th className="p-2 text-left">{t('scanner_hub.action')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {searchResults.results.map((item, index) => (
                            <tr key={index} className="border-b hover:bg-muted/50">
                              <td className="p-2">
                                <input
                                  type="checkbox"
                                  checked={selectedSearchItems.has(index)}
                                  onChange={() => toggleSearchItem(index)}
                                />
                              </td>
                              <td className="p-2 font-mono">{item.ip}</td>
                              <td className="p-2">{item.port}</td>
                              <td className="p-2">{item.protocol || '-'}</td>
                              <td className="p-2">{item.hostname || '-'}</td>
                              <td className="p-2">
                                <Button
                                  size="sm"
                                  variant="ghost"
                                  disabled={!selectedCase || !selectedPayload || creatingFromSearch}
                                  onClick={async () => {
                                    setSelectedSearchItems(new Set([index]))
                                    setShowConfirmCreate(true)
                                  }}
                                >
                                  {t('scanner_hub.create_run')}
                                </Button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">
                        {selectedSearchItems.size} {t('scanner_hub.selected')}
                      </span>
                      <Button
                        onClick={() => setShowConfirmCreate(true)}
                        disabled={selectedSearchItems.size === 0 || !selectedCase || !selectedPayload || creatingFromSearch}
                      >
                        {creatingFromSearch ? t('common.loading') : t('scanner_hub.create_selected_runs')}
                      </Button>
                    </div>
                  </>
                ) : (
                  <div className="text-sm text-muted-foreground py-8 text-center border rounded">
                    {t('scanner_hub.no_results')}
                  </div>
                )}
              </div>
            )}

            {/* Confirmation dialog for creating scanner runs from search results */}
            <Dialog open={showConfirmCreate} onOpenChange={setShowConfirmCreate}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>{t('scanner_hub.confirm_create_title')}</DialogTitle>
                </DialogHeader>
                <div className="py-4 text-sm">
                  <p className="text-muted-foreground">
                    {t('scanner_hub.confirm_create_desc')}
                  </p>
                  <ul className="mt-2 space-y-1">
                    {Array.from(selectedSearchItems).slice(0, 10).map(i => {
                      const item = searchResults?.results[i]
                      if (!item) return null
                      return (
                        <li key={i} className="font-mono text-xs">
                          {item.ip}:{item.port}{item.hostname ? ` (${item.hostname})` : ''}
                        </li>
                      )
                    })}
                    {selectedSearchItems.size > 10 && (
                      <li className="text-xs text-muted-foreground">
                        ...{t('scanner_hub.and_more')}{(selectedSearchItems.size - 10).toString()}
                      </li>
                    )}
                  </ul>
                  <p className="mt-3 font-medium">
                    {t('scanner_hub.confirm_create_total')}{selectedSearchItems.size}
                  </p>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setShowConfirmCreate(false)}>
                    {t('common.cancel')}
                  </Button>
                  <Button onClick={handleCreateFromSearch} disabled={creatingFromSearch}>
                    {creatingFromSearch ? t('common.loading') : t('scanner_hub.confirm_create_confirm')}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </CardContent>
        </Card>

        {/* Scanner Adapter Selection */}
        <Card>
          <CardHeader>
            <CardTitle>{t('scanner_hub.select_scanner')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Select value={selectedScanner} onValueChange={(value: ScannerKind) => handleScannerChange(value)}>
              <SelectTrigger>
                <SelectValue placeholder={t('scanner_hub.select_scanner_placeholder')} />
              </SelectTrigger>
              <SelectContent>
                {adapters.map(adapter => (
                  <SelectItem key={adapter.id} value={adapter.id}>
                    {adapter.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedAdapter && (
              <div className="rounded border p-3 text-sm">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge>{selectedAdapter.category}</Badge>
                  <Badge variant="secondary">{selectedAdapter.maturity}</Badge>
                  <span className="text-muted-foreground">{selectedAdapter.description}</span>
                </div>
                <div className="mt-3">
                  <Select value={selectedDeliveryMethod} onValueChange={(value: ScannerDeliveryMethod) => setSelectedDeliveryMethod(value)}>
                    <SelectTrigger>
                      <SelectValue placeholder={t('scanner_hub.select_delivery')} />
                    </SelectTrigger>
                    <SelectContent>
                      {selectedAdapter.supported_methods.map(method => (
                        <SelectItem key={method} value={method}>
                          {method}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}
            <div className="grid gap-2 md:grid-cols-4">
              {adapters.map(adapter => (
                <div key={adapter.id} className="rounded border p-3 text-xs">
                  <div className="font-medium">{adapter.name}</div>
                  <div className="text-muted-foreground">{adapter.default_method}</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Case Selection */}
        <Card>
          <CardHeader>
            <CardTitle>{t('scanner_hub.select_case')}</CardTitle>
          </CardHeader>
          <CardContent>
            <Select value={selectedCase} onValueChange={setSelectedCase}>
              <SelectTrigger>
                <SelectValue placeholder={cases.length === 0 ? t('scanner_hub.no_case') : t('scanner_hub.select_case_placeholder')} />
              </SelectTrigger>
              <SelectContent>
                {cases.map(c => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {cases.length === 0 && (
              <div className="mt-3 text-sm text-muted-foreground">
                {t('scanner_hub.no_case_hint')}
                <Button variant="link" className="px-0" onClick={() => router.push('/cases')}>
                  {t('scanner_hub.go_to_cases')}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Target Input */}
        <Card>
          <CardHeader>
            <CardTitle>{t('scanner_hub.input_target')}</CardTitle>
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
            <CardTitle>{t('scanner_hub.select_template')}</CardTitle>
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
            <CardTitle>{t('scanner_hub.select_payload')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Select value={selectedPayload} onValueChange={setSelectedPayload}>
              <SelectTrigger>
                <SelectValue placeholder={t('scanner_hub.select_payload_placeholder')} />
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
              {generating ? t('scanner_hub.creating') : t('scanner_hub.create_payload')}
            </Button>
          </CardContent>
        </Card>

        {/* Generate Button */}
        <Button onClick={handleGenerateScannerRun} className="w-full" size="lg" disabled={generating}>
          {generating ? t('scanner_hub.generating') : t('scanner_hub.generate_run')}
        </Button>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 flex items-start gap-3">
            <svg className="w-5 h-5 text-red-500 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
          </div>
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
                    {t('scanner_hub.copy')}
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
                    {t('scanner_hub.copy')}
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{outputLabel}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Input value={scannerRun.command} readOnly />
                  <Button onClick={() => handleCopy(scannerRun.command)}>
                    {t('scanner_hub.copy')}
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
                    {t('scanner_hub.copy')}
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

            <Card>
              <CardHeader>
                <CardTitle>JSONL Preview</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
                  <Textarea value={scannerRun.jsonl} readOnly className="font-mono text-sm" />
                  <Button onClick={() => handleCopy(scannerRun.jsonl)}>
                    {t('scanner_hub.copy')}
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Scope Info */}
            <Card>
              <CardHeader>
                <CardTitle>{t('scanner_hub.current_scope')}</CardTitle>
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
                  <CardTitle>{t('scanner_hub.view_results')}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <Button
                    onClick={() => router.push(webUrls.interactionsUrl)}
                    className="w-full"
                    variant="outline"
                  >
                    {t('scanner_hub.view_interactions')}
                  </Button>
                  <Button
                    onClick={() => router.push(webUrls.evidenceUrl)}
                    className="w-full"
                    variant="outline"
                  >
                    {t('scanner_hub.view_evidence')}
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

function getScannerOutputLabel(scanner: ScannerKind): string {
  switch (scanner) {
    case 'nuclei':
      return 'Nuclei Command'
    case 'burp':
      return 'Burp Suite Extension Package'
    case 'yakit':
      return 'Yakit/Yak Script Package'
    case 'zap':
      return 'ZAP Script Package'
    case 'xray':
    case 'rad':
      return 'Webhook Bridge Package'
    case 'postman':
    case 'apifox':
      return 'Environment Package'
    default:
      return 'Integration Package'
  }
}
