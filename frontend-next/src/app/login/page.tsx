'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { authApi } from '@/lib/api-client'
import type { LoginRequest } from '@/types'
import { t, getCurrentLanguage, Language } from '@/lib/i18n'

export default function LoginPage() {
  const router = useRouter()
  const [formData, setFormData] = useState<LoginRequest>({
    username: '',
    password: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [lang, setLang] = useState<Language>('en-US')

  useEffect(() => {
    setLang(getCurrentLanguage())
  }, [])

  const handleLanguageChange = (newLang: Language) => {
    setLang(newLang)
    if (typeof window !== 'undefined') {
      localStorage.setItem('language', newLang)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const response = await authApi.login(formData)
      console.log('Login response:', response)
      if (response.code === 0 && response.data) {
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
        console.log('Token stored:', response.data.token)
        router.push('/dashboard')
      } else {
        setError(response.message || t('login.error', lang))
      }
    } catch (err: any) {
      console.error('Login error:', err)
      setError(err.response?.data?.message || t('login.error', lang))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8">
        <div className="flex justify-between items-center">
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900 flex-1">
            {t('login.title', lang)}
          </h2>
          <div className="flex gap-2">
            <button
              onClick={() => handleLanguageChange('en-US')}
              className={`px-2 py-1 text-sm rounded ${lang === 'en-US' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              EN
            </button>
            <button
              onClick={() => handleLanguageChange('zh-CN')}
              className={`px-2 py-1 text-sm rounded ${lang === 'zh-CN' ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              中
            </button>
          </div>
        </div>
        <p className="mt-2 text-center text-sm text-gray-600">
          {t('login.subtitle', lang)}
        </p>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="username" className="sr-only">
                {t('login.username', lang)}
              </label>
              <input
                id="username"
                name="username"
                type="text"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder={t('login.username', lang)}
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              />
            </div>
            <div>
              <label htmlFor="password" className="sr-only">
                {t('login.password', lang)}
              </label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder={t('login.password', lang)}
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              />
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? t('login.button.loading', lang) : t('login.button', lang)}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
