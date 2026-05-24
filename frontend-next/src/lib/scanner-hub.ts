/**
 * Scanner Hub Helper Functions
 * 
 * Provides utilities for generating Nuclei commands and JSONL records
 * for scanner integration with GODNSLOG.
 */

export interface ScannerRun {
  id: string
  case_id: string
  payload_id: string
  scanner: 'nuclei'
  target: string
  template: 'ssrf-basic' | 'xxe-basic' | 'rce-callback'
  delivery_method: 'nuclei-jsonl' | 'nuclei-var'
  command: string
  jsonl: string
  status: 'created' | 'distributed' | 'observed' | 'evidenced'
  created_at: string
  updated_at: string
}

export interface ScannerRunInput {
  case_id: string
  payload_id: string
  token: string
  target: string
  template: 'ssrf-basic' | 'xxe-basic' | 'rce-callback'
  rendered_payload: string
  baseUrl: string
}

function shellQuote(value: string): string {
  return `'${value.replace(/'/g, "'\\''")}'`
}

/**
 * Generate Nuclei command with template variable
 */
export function generateNucleiCommand(input: ScannerRunInput): string {
  const { target, template, rendered_payload } = input
  return [
    'nuclei',
    '-u',
    shellQuote(target),
    '-t',
    `godnslog-${template}.yaml`,
    '-var',
    shellQuote(`godnslog_payload=${rendered_payload}`),
  ].join(' ')
}

/**
 * Generate JSONL record for scanner probe
 */
export function generateJsonlRecord(input: ScannerRunInput): string {
  const { case_id, payload_id, token, target, template, rendered_payload, baseUrl } = input
  const record = {
    scanner: 'nuclei',
    case_id,
    payload_id,
    token,
    target,
    template,
    rendered_payload,
    interactions_url: `${baseUrl}/api/v2/interactions?payload_id=${payload_id}`,
    evidence_url: `${baseUrl}/dashboard/evidence?payload_id=${payload_id}`,
    created_at: new Date().toISOString()
  }
  return JSON.stringify(record)
}

/**
 * Generate web URLs for Interactions and Evidence pages
 */
export function generateWebUrls(input: ScannerRunInput): {
  interactionsUrl: string
  evidenceUrl: string
} {
  const { payload_id, baseUrl } = input
  return {
    interactionsUrl: `${baseUrl}/dashboard/interactions?payload_id=${payload_id}`,
    evidenceUrl: `${baseUrl}/dashboard/evidence?payload_id=${payload_id}`
  }
}

/**
 * Create a Scanner Run object (in-memory, not persisted)
 */
export function createScannerRun(input: ScannerRunInput, deliveryMethod: 'nuclei-jsonl' | 'nuclei-var'): ScannerRun {
  const command = generateNucleiCommand(input)
  const jsonl = generateJsonlRecord(input)
  const now = new Date().toISOString()
  
  return {
    id: `scanner-${Date.now()}`,
    case_id: input.case_id,
    payload_id: input.payload_id,
    scanner: 'nuclei',
    target: input.target,
    template: input.template,
    delivery_method: deliveryMethod,
    command,
    jsonl,
    status: 'created',
    created_at: now,
    updated_at: now
  }
}
