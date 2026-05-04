'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { t, getCurrentLanguage, Language } from '@/lib/i18n'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const [lang, setLang] = useState<Language>('en-US')

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
    setLang(getCurrentLanguage())
  }, [router])

  const handleLanguageChange = (newLang: Language) => {
    setLang(newLang)
    if (typeof window !== 'undefined') {
      localStorage.setItem('language', newLang)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <h1 className="text-xl font-bold text-gray-900">{t('dashboard.title', lang)}</h1>
              </div>
              <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                <Link href="/dashboard" className="border-indigo-500 text-gray-900 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.dashboard', lang)}
                </Link>
                <Link href="/dashboard/cases" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.cases', lang)}
                </Link>
                <Link href="/dashboard/payloads" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.payloads', lang)}
                </Link>
                <Link href="/dashboard/interactions" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.interactions', lang)}
                </Link>
                <Link href="/dashboard/workflow" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.workflow', lang)}
                </Link>
                <Link href="/dashboard/rebinding" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.rebinding', lang)}
                </Link>
                <Link href="/dashboard/canary" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.canary', lang)}
                </Link>
                <Link href="/dashboard/marketplace" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.marketplace', lang)}
                </Link>
                <Link href="/dashboard/evidence" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.evidence', lang)}
                </Link>
                <Link href="/dashboard/settings" className="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  {t('dashboard.menu.settings', lang)}
                </Link>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => handleLanguageChange('en-US')}
                className={`px-2 py-1 text-xs rounded ${lang === 'en-US' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'}`}
              >
                EN
              </button>
              <button
                onClick={() => handleLanguageChange('zh-CN')}
                className={`px-2 py-1 text-xs rounded ${lang === 'zh-CN' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'}`}
              >
                中
              </button>
              <button
                onClick={() => {
                  localStorage.removeItem('token')
                  localStorage.removeItem('user')
                  router.push('/login')
                }}
                className="text-gray-500 hover:text-gray-700 text-xs font-medium ml-2"
              >
                {t('dashboard.menu.logout', lang)}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  )
}
