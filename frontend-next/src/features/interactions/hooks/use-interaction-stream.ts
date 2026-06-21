'use client'

import { useEffect, useRef, useCallback, useState } from 'react'
import type { Interaction } from '@/types'

interface UseInteractionStreamOptions {
  caseId?: string
  payloadId?: string
  type?: string
  enabled: boolean
  onInteraction: (interaction: Interaction) => void
}

interface UseInteractionStreamResult {
  connected: boolean
  error: string | null
  reconnect: () => void
}

/**
 * useInteractionStream subscribes to the SSE endpoint for real-time
 * interaction updates. It uses the browser EventSource API with token
 * passed as a query parameter (since EventSource cannot set headers).
 */
export function useInteractionStream({
  caseId,
  payloadId,
  type,
  enabled,
  onInteraction,
}: UseInteractionStreamOptions): UseInteractionStreamResult {
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectRef = useRef(0)
  const onInteractionRef = useRef(onInteraction)
  const enabledRef = useRef(enabled)
  const connectRef = useRef<() => void>(() => {})

  // Keep latest callback without re-creating EventSource
  useEffect(() => {
    onInteractionRef.current = onInteraction
  }, [onInteraction])

  useEffect(() => {
    enabledRef.current = enabled
  }, [enabled])

  const connect = useCallback(() => {
    if (typeof window === 'undefined') return

    const token = localStorage.getItem('token')
    if (!token) {
      setError('No auth token')
      return
    }

    const baseUrl = process.env.NEXT_PUBLIC_API_URL || '/api/v2'
    const params = new URLSearchParams()
    params.set('token', token)
    params.set('since', new Date().toISOString())
    if (caseId) params.set('case_id', caseId)
    if (payloadId) params.set('payload_id', payloadId)
    if (type) params.set('type', type)

    const url = `${baseUrl}/interactions/stream?${params.toString()}`

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const es = new EventSource(url)
    eventSourceRef.current = es

    es.addEventListener('connected', () => {
      setConnected(true)
      setError(null)
      reconnectRef.current = 0
    })

    es.addEventListener('interaction', (event) => {
      try {
        const data = JSON.parse(event.data) as Interaction
        onInteractionRef.current(data)
      } catch {
        // Ignore parse errors
      }
    })

    es.addEventListener('heartbeat', () => {
      // Heartbeat keeps connection alive; no action needed
    })

    es.onerror = () => {
      setConnected(false)
      es.close()
      eventSourceRef.current = null

      // Exponential backoff reconnect (max 30s)
      const delay = Math.min(1000 * Math.pow(2, reconnectRef.current), 30000)
      reconnectRef.current++
      setError(`Connection lost, reconnecting in ${delay / 1000}s...`)
      setTimeout(() => {
        if (enabledRef.current) connectRef.current()
      }, delay)
    }
  }, [caseId, payloadId, type])

  useEffect(() => {
    connectRef.current = connect
  }, [connect])

  useEffect(() => {
    if (!enabled) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      const timer = setTimeout(() => {
        setConnected(false)
        setError(null)
      }, 0)
      return () => clearTimeout(timer)
    }

    const timer = setTimeout(() => connect(), 0)
    return () => {
      clearTimeout(timer)
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
    }
  }, [connect, enabled])

  const reconnect = useCallback(() => {
    reconnectRef.current = 0
    connect()
  }, [connect])

  return { connected, error, reconnect }
}
