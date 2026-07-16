'use client'

import { useEffect, type ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/lib/i18n-context'
import type { TranslationKey } from '@/lib/i18n-context'
import { MarkdownRenderer } from '@/components/markdown-renderer'

interface DocPageLayoutProps {
  titleKey: TranslationKey
  children?: ReactNode
  /** English markdown content */
  mdEn?: string
  /** Chinese markdown content */
  mdZh?: string
}

export function DocPageLayout({ titleKey, children, mdEn, mdZh }: DocPageLayoutProps) {
  const router = useRouter()
  const { t, lang } = useI18n()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const md = lang === 'zh-CN' && mdZh ? mdZh : mdEn

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="mb-6">
        <Button variant="ghost" size="sm" onClick={() => router.push('/docs')}>
          ← {t('docs.title')}
        </Button>
      </div>
      <h1 className="text-3xl font-bold mb-8">{t(titleKey)}</h1>
      <div className="prose prose-sm dark:prose-invert max-w-none">
        {md ? <MarkdownRenderer content={md} /> : children}
      </div>
    </div>
  )
}
