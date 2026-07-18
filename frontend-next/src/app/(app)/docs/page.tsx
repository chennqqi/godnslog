'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/lib/i18n-context'
import type { TranslationKey } from '@/lib/i18n-context'

export default function DocsPage() {
  const router = useRouter()
  const { t } = useI18n()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      window.location.href = '/login'
    }
  }, [router])

  const docs = [
    {
      titleKey: 'docs.quick_start' as TranslationKey,
      descKey: 'docs.quick_start_desc' as TranslationKey,
      link: '/docs/quick-start',
    },
    {
      titleKey: 'docs.api' as TranslationKey,
      descKey: 'docs.api_desc' as TranslationKey,
      link: '/docs/api',
    },
    {
      titleKey: 'docs.user_guide' as TranslationKey,
      descKey: 'docs.user_guide_desc' as TranslationKey,
      link: '/docs/user-guide',
    },
    {
      titleKey: 'docs.config' as TranslationKey,
      descKey: 'docs.config_desc' as TranslationKey,
      link: '/docs/config',
    },
    {
      titleKey: 'docs.faq' as TranslationKey,
      descKey: 'docs.faq_desc' as TranslationKey,
      link: '/docs/faq',
    },
    {
      titleKey: 'docs.security' as TranslationKey,
      descKey: 'docs.security_desc' as TranslationKey,
      link: '/docs/security',
    },
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">{t('docs.title')}</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {docs.map((doc, index) => (
          <Card key={index} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle>{t(doc.titleKey)}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-500 mb-4">{t(doc.descKey)}</p>
              <Button variant="outline" className="w-full" onClick={() => router.push(doc.link)}>
                {t('docs.view')}
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="mt-8">
        <CardHeader>
          <CardTitle>{t('docs.get_help')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <h3 className="font-medium text-gray-900 mb-2">{t('docs.github_issues')}</h3>
              <p className="text-sm text-gray-500 mb-2">
                {t('docs.github_issues_desc')}
              </p>
              <Button variant="outline" size="sm">
                {t('docs.visit_github')}
              </Button>
            </div>
            <div>
              <h3 className="font-medium text-gray-900 mb-2">{t('docs.community')}</h3>
              <p className="text-sm text-gray-500 mb-2">
                {t('docs.community_desc')}
              </p>
              <Button variant="outline" size="sm">
                {t('docs.join_community')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
