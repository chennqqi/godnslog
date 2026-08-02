'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { useI18n } from '@/lib/i18n-context'
import { authApi } from '@/lib/api-client'

export interface SlideCaptchaProps {
  onReady: (captchaId: string, captchaValue: number, captchaY: number) => void
  onRefresh: () => void
  invalid: boolean
  disabled?: boolean
}

const THUMB_SIZE = 48
// Backend renders the captcha background at this intrinsic width (server/captcha.go).
const IMG_WIDTH = 300

export function SlideCaptcha({ onReady, onRefresh, invalid, disabled = false }: SlideCaptchaProps) {
  const { t } = useI18n()
  const [captchaId, setCaptchaId] = useState('')
  const [imageBase64, setImageBase64] = useState('')
  const [thumbBase64, setThumbBase64] = useState('')
  const [blockDX, setBlockDX] = useState(0)
  const [blockDY, setBlockDY] = useState(0)
  const [blockWidth, setBlockWidth] = useState(0)
  const [blockHeight, setBlockHeight] = useState(0)
  // Rendered background width in display px; needed to scale the tile 1:1 with
  // the hole. Measured reactively so it is correct on the very first paint.
  const [displayedW, setDisplayedW] = useState(IMG_WIDTH)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [dragging, setDragging] = useState(false)
  const [position, setPosition] = useState(0)
  const [completed, setCompleted] = useState(false)
  const [hidden, setHidden] = useState(false)

  const trackRef = useRef<HTMLDivElement>(null)
  const imageRef = useRef<HTMLImageElement>(null)
  const startXRef = useRef(0)
  const startPosRef = useRef(0)
  const onReadyRef = useRef(onReady)

  useEffect(() => {
    onReadyRef.current = onReady
  }, [onReady])

  const loadCaptcha = useCallback(async () => {
    setLoading(true)
    setError('')
    setPosition(0)
    setCompleted(false)
    setCaptchaId('')
    setHidden(false)

    try {
      const response = await authApi.captcha()
      if (response.code === 0 && response.data) {
        if (!response.data.captcha_id) {
          setHidden(true)
          onReadyRef.current('', 0, 0)
          return
        }
        setCaptchaId(response.data.captcha_id)
        setImageBase64(response.data.image_base64)
        setThumbBase64(response.data.thumb_base64)
        setBlockDX(response.data.block_dx ?? 0)
        setBlockDY(response.data.block_dy ?? 0)
        setBlockWidth(response.data.block_width ?? 0)
        setBlockHeight(response.data.block_height ?? 0)
      } else {
        setError(response.message || t('login.captcha.error'))
      }
    } catch {
      setError(t('login.captcha.error'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    if (!imageBase64) return
    const img = imageRef.current
    if (!img) return
    const measure = () => {
      if (imageRef.current) setDisplayedW(imageRef.current.clientWidth || IMG_WIDTH)
    }
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(img)
    return () => ro.disconnect()
  }, [imageBase64])

  useEffect(() => {
    const id = requestAnimationFrame(() => loadCaptcha())
    return () => cancelAnimationFrame(id)
  }, [loadCaptcha])

  useEffect(() => {
    if (invalid) {
      const id = requestAnimationFrame(() => loadCaptcha())
      return () => cancelAnimationFrame(id)
    }
  }, [invalid, loadCaptcha])

  const handleRefresh = () => {
    loadCaptcha()
    onRefresh()
  }

  // --- Mouse drag ---
  const handleMouseDown = (e: React.MouseEvent) => {
    if (disabled || completed) return
    e.preventDefault()
    startXRef.current = e.clientX
    startPosRef.current = position
    setDragging(true)
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!dragging || disabled) return
    const delta = e.clientX - startXRef.current
    const max = trackRef.current ? trackRef.current.offsetWidth - THUMB_SIZE : 0
    setPosition(Math.max(0, Math.min(max, startPosRef.current + delta)))
  }

  const handleDragEnd = () => {
    if (!dragging) return
    setDragging(false)
    if (captchaId) {
      setCompleted(true)
      // Final tile X in image coordinates = initial display X (blockDX) + drag
      // distance, converted back to image px. When the tile visually overlaps
      // the hole, this equals the hole's X — the server answer.
      const scaled = Math.round(blockDX + position * IMG_WIDTH / displayedW)
      onReadyRef.current(captchaId, scaled, blockDY)
    }
  }

  // --- Touch drag ---
  const handleTouchStart = (e: React.TouchEvent) => {
    if (disabled || completed) return
    startXRef.current = e.touches[0].clientX
    startPosRef.current = position
    setDragging(true)
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!dragging || disabled) return
    const delta = e.touches[0].clientX - startXRef.current
    const max = trackRef.current ? trackRef.current.offsetWidth - THUMB_SIZE : 0
    setPosition(Math.max(0, Math.min(max, startPosRef.current + delta)))
  }

  // --- Loading state ---
  if (loading) {
    return (
      <div className="flex items-center justify-center h-24 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
        <span className="text-sm text-gray-400">{t('login.captcha.loading')}</span>
      </div>
    )
  }

  // --- Error state ---
  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-center">
        <p className="text-sm text-red-600 dark:text-red-400 mb-2">{error}</p>
        <button
          type="button"
          onClick={handleRefresh}
          className="text-sm text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300 font-medium"
        >
          {t('login.captcha.refresh')}
        </button>
      </div>
    )
  }

  // --- Hidden state (captcha disabled) ---
  if (hidden) return null

  // Display scale: image rendered width vs its intrinsic 300px width.
  // The tile is positioned/sized in the same scale so it matches the hole 1:1.
  const scale = displayedW / IMG_WIDTH

  return (
    <div className="space-y-2">
      {/* Captcha image area (background + puzzle piece overlay) */}
      <div className="relative w-full overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700 select-none">
        {/* Background image */}
        <img
          ref={imageRef}
          src={imageBase64}
          alt="captcha background"
          className="block w-full h-auto"
          draggable={false}
        />

        {/* Visible puzzle piece overlay that moves with the drag position */}
        <div
          className="absolute top-0 pointer-events-none"
          style={{
            left: `${blockDX * scale + position}px`,
            top: `${blockDY * scale}px`,
            width: blockWidth * scale,
            height: blockHeight * scale,
          }}
        >
          <img
            src={thumbBase64}
            alt="puzzle piece"
            className="block w-full h-full"
            draggable={false}
          />
        </div>
      </div>

      {/* Slider track (the drag interaction surface) */}
      <div
        ref={trackRef}
        className={`relative h-12 rounded-lg border select-none transition-colors ${
          disabled
            ? 'bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700 opacity-60 cursor-not-allowed'
            : completed
              ? 'bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700 cursor-default'
              : 'bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700 cursor-pointer'
        }`}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleDragEnd}
        onMouseLeave={handleDragEnd}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleDragEnd}
      >
        {/* Hint text centered in track */}
        <div className="absolute inset-0 flex items-center justify-center">
          {!completed && (
            <span className="text-sm text-gray-400 dark:text-gray-500 select-none">
              {t('login.captcha.slide_hint')}
            </span>
          )}
        </div>

        {/* Draggable slider thumb */}
        <div
          className={`absolute top-0 left-0 h-full flex items-center justify-center rounded-l-lg border-r transition-shadow ${
            dragging
              ? 'shadow-md bg-gray-50 dark:bg-gray-600 border-gray-200 dark:border-gray-500'
              : completed
                ? 'bg-white dark:bg-gray-700 border-gray-200 dark:border-gray-600'
                : 'bg-white dark:bg-gray-700 border-gray-200 dark:border-gray-600'
          }`}
          style={{
            width: `${THUMB_SIZE}px`,
            transform: `translateX(${position}px)`,
          }}
        >
          {completed ? (
            <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          ) : (
            <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          )}
        </div>
      </div>

      {/* Refresh button */}
      <div className="flex justify-end">
        <button
          type="button"
          onClick={handleRefresh}
          disabled={disabled}
          className="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {t('login.captcha.refresh')}
        </button>
      </div>
    </div>
  )
}
