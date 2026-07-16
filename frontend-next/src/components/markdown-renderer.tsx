'use client'

import { useState, useMemo, type ReactNode } from 'react'

/** Minimal markdown parser that renders to React elements with copyable code blocks */
function parseMarkdown(md: string): ReactNode[] {
  const lines = md.split('\n')
  const nodes: ReactNode[] = []
  let inCodeBlock = false
  let codeBuffer: string[] = []
  let codeLang = ''
  let listItems: ReactNode[] = []
  let inList = false

  const flushList = (key: string) => {
    if (listItems.length > 0) {
      nodes.push(<ul key={key} className="list-disc pl-6 space-y-1 my-2">{listItems}</ul>)
      listItems = []
      inList = false
    }
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]

    // Code block
    if (line.startsWith('```')) {
      if (inCodeBlock) {
        nodes.push(<CodeBlock key={`cb-${i}`} code={codeBuffer.join('\n')} lang={codeLang} />)
        codeBuffer = []
        codeLang = ''
        inCodeBlock = false
      } else {
        flushList(`list-${i}`)
        inCodeBlock = true
        codeLang = line.slice(3).trim()
      }
      continue
    }
    if (inCodeBlock) {
      codeBuffer.push(line)
      continue
    }

    // Empty line
    if (line.trim() === '') {
      flushList(`list-${i}`)
      continue
    }

    // Headings
    const headingMatch = line.match(/^(#{1,6})\s+(.+)$/)
    if (headingMatch) {
      flushList(`list-${i}`)
      const level = headingMatch[1].length
      const text = headingMatch[2]
      const Tag = `h${level}` as keyof JSX.IntrinsicElements
      const size = level === 1 ? 'text-2xl font-bold mt-8 mb-4' :
                   level === 2 ? 'text-xl font-semibold mt-6 mb-3' :
                   level === 3 ? 'text-lg font-semibold mt-4 mb-2' :
                   'text-base font-medium mt-3 mb-1'
      nodes.push(<Tag key={`h-${i}`} className={size}>{renderInline(text)}</Tag>)
      continue
    }

    // Unordered list
    const listMatch = line.match(/^[-*+]\s+(.+)$/)
    if (listMatch) {
      inList = true
      listItems.push(<li key={`li-${i}`}>{renderInline(listMatch[1])}</li>)
      continue
    }

    // Ordered list
    const orderedMatch = line.match(/^\d+\.\s+(.+)$/)
    if (orderedMatch) {
      inList = true
      listItems.push(<li key={`li-${i}`}>{renderInline(orderedMatch[1])}</li>)
      continue
    }

    flushList(`list-${i}`)

    // Horizontal rule
    if (/^[-*_]{3,}$/.test(line.trim())) {
      nodes.push(<hr key={`hr-${i}`} className="my-6 border-gray-300 dark:border-gray-600" />)
      continue
    }

    // Paragraph
    nodes.push(<p key={`p-${i}`} className="mb-3 leading-relaxed">{renderInline(line)}</p>)
  }

  // Flush remaining
  if (inCodeBlock) {
    nodes.push(<CodeBlock key="cb-end" code={codeBuffer.join('\n')} lang={codeLang} />)
  }
  flushList('list-end')

  return nodes
}

/** Render inline markdown (bold, italic, code, links) */
function renderInline(text: string): ReactNode {
  const parts: ReactNode[] = []
  let remaining = text
  let idx = 0

  while (remaining.length > 0) {
    // Inline code `code`
    const codeMatch = remaining.match(/`([^`]+)`/)
    if (codeMatch && codeMatch.index !== undefined) {
      if (codeMatch.index > 0) parts.push(remaining.slice(0, codeMatch.index))
      parts.push(<code key={`c-${idx++}`} className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-pink-600 dark:text-pink-400">{codeMatch[1]}</code>)
      remaining = remaining.slice(codeMatch.index + codeMatch[0].length)
      continue
    }

    // Bold **text**
    const boldMatch = remaining.match(/\*\*([^*]+)\*\*/)
    if (boldMatch && boldMatch.index !== undefined) {
      if (boldMatch.index > 0) parts.push(remaining.slice(0, boldMatch.index))
      parts.push(<strong key={`b-${idx++}`} className="font-semibold">{boldMatch[1]}</strong>)
      remaining = remaining.slice(boldMatch.index + boldMatch[0].length)
      continue
    }

    // Link [text](url)
    const linkMatch = remaining.match(/\[([^\]]+)\]\(([^)]+)\)/)
    if (linkMatch && linkMatch.index !== undefined) {
      if (linkMatch.index > 0) parts.push(remaining.slice(0, linkMatch.index))
      parts.push(<a key={`l-${idx++}`} href={linkMatch[2]} className="text-indigo-600 dark:text-indigo-400 underline hover:no-underline" target="_blank" rel="noreferrer">{linkMatch[1]}</a>)
      remaining = remaining.slice(linkMatch.index + linkMatch[0].length)
      continue
    }

    parts.push(remaining)
    break
  }

  return <>{parts}</>
}

/** Code block with copy button */
function CodeBlock({ code, lang }: { code: string; lang: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative group my-4 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
      <div className="flex items-center justify-between px-4 py-1.5 bg-gray-100 dark:bg-gray-800 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
        <span>{lang || 'code'}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 px-2 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        >
          {copied ? (
            <>
              <svg className="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              Copied
            </>
          ) : (
            <>
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              Copy
            </>
          )}
        </button>
      </div>
      <pre className="overflow-x-auto p-4 text-sm bg-gray-50 dark:bg-gray-900">
        <code className="text-gray-800 dark:text-gray-200 font-mono">{code}</code>
      </pre>
    </div>
  )
}

/** Markdown renderer component */
export function MarkdownRenderer({ content }: { content: string }) {
  const nodes = useMemo(() => parseMarkdown(content), [content])
  return <div className="prose prose-sm dark:prose-invert max-w-none">{nodes}</div>
}
