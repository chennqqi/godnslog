'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { authApi } from '@/lib/api-client'
import type { LoginRequest } from '@/types'
import { useI18n } from '@/lib/i18n-context'
import { Spinner } from '@/components/ui/spinner'
import { useAuthStore } from '@/features/auth/store'
import { loginSchema, type LoginFormValues } from '@/features/auth/schemas/login-schema'

/** Feature key identifiers for the brand panel */
const FEATURE_KEYS = [
  { icon: 'shield', key: 'oast' },
  { icon: 'beaker', key: 'payload' },
  { icon: 'clipboard', key: 'evidence' },
  { icon: 'workflow', key: 'scanner' },
] as const

/** Icon renderer for feature highlights */
function FeatureIcon({ name }: { name: string }) {
  const icons: Record<string, React.ReactNode> = {
    shield: (
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
    ),
    beaker: (
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 3v10.17l-3 5.83h12l-3-5.83V3M9 3h6M9 3H6M15 3h3" />
    ),
    clipboard: (
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
    ),
    workflow: (
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 16V4m0 0L3 8m4-4l4 4M17 8v12m0 0l4-4m-4 4l-4-4" />
    ),
  }
  return (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      {icons[name] || icons.shield}
    </svg>
  )
}

export default function LoginPage() {
  const { setToken, setUser } = useAuthStore()
  const { t, lang, setLang } = useI18n()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  const handleLanguageChange = (newLang: 'en-US' | 'zh-CN') => {
    setLang(newLang)
  }

  const features = FEATURE_KEYS.map((f) => ({
    ...f,
    title: t(`login.feature.${f.key}.title`),
    desc: t(`login.feature.${f.key}.desc`),
  }))

  const onSubmit = async (data: LoginFormValues) => {
    setLoading(true)
    setError('')

    try {
      const response = await authApi.login(data as LoginRequest)
      if (response.code === 0 && response.data) {
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
        setToken(response.data.token)
        setUser({
          id: String(response.data.user.id),
          name: response.data.user.username,
          email: response.data.user.email,
          role: String(response.data.user.role),
        })
        window.location.href = '/'
      } else {
        setError(response.message || t('login.error'))
      }
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } }, message?: string }
      setError(error.response?.data?.message || error.message || t('login.error'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex bg-gray-50 dark:bg-gray-950">
      {/* Left brand panel - hidden on mobile */}
      <div className="hidden lg:flex lg:w-1/2 relative overflow-hidden bg-gradient-to-br from-indigo-900 via-purple-900 to-gray-900">
        {/* Decorative grid pattern */}
        <div
          className="absolute inset-0 opacity-10"
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
          }}
        />

        {/* Content */}
        <div className="relative z-10 flex flex-col justify-between p-12 text-white w-full">
          {/* Logo & title */}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-indigo-500 flex items-center justify-center font-bold text-lg shadow-lg">
              G
            </div>
            <span className="font-bold text-xl tracking-tight">GODNSLOG</span>
          </div>

          {/* Hero text */}
          <div className="space-y-6">
            <div>
              <h1 className="text-4xl font-bold leading-tight whitespace-pre-line">
                {t('login.hero.title')}
              </h1>
              <p className="mt-4 text-lg text-indigo-200 max-w-md">
                {t('login.hero.subtitle')}
              </p>
            </div>

            {/* Feature highlights */}
            <div className="grid grid-cols-2 gap-4 max-w-md">
              {features.map((f) => (
                <div
                  key={f.key}
                  className="flex items-start gap-3 p-3 rounded-lg bg-white/5 backdrop-blur-sm border border-white/10"
                >
                  <span className="text-indigo-300 shrink-0 mt-0.5">
                    <FeatureIcon name={f.icon} />
                  </span>
                  <div>
                    <p className="text-sm font-semibold text-white">{f.title}</p>
                    <p className="text-xs text-indigo-200/70 mt-0.5">{f.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center gap-4 text-sm text-indigo-300/60">
            <span>{t('login.footer.version')}</span>
            <span>·</span>
            <span>{t('login.footer.tagline')}</span>
          </div>
        </div>
      </div>

      {/* Right form panel */}
      <div className="flex-1 flex items-center justify-center p-6 sm:p-12">
        <div className="w-full max-w-md space-y-8">
          {/* Mobile logo & language switcher */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 lg:hidden">
              <div className="w-9 h-9 rounded-lg bg-indigo-600 flex items-center justify-center font-bold text-white">
                G
              </div>
              <span className="font-bold text-lg text-gray-900 dark:text-white">GODNSLOG</span>
            </div>
            <div className="flex gap-2 ml-auto">
              <button
                onClick={() => handleLanguageChange('en-US')}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${lang === 'en-US' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}`}
              >
                EN
              </button>
              <button
                onClick={() => handleLanguageChange('zh-CN')}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${lang === 'zh-CN' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}`}
              >
                中
              </button>
            </div>
          </div>

          {/* Title */}
          <div>
            <h2 className="text-3xl font-bold text-gray-900 dark:text-white">
              {t('login.title')}
            </h2>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {t('login.subtitle')}
            </p>
          </div>

          {/* Form */}
          <form className="space-y-5" onSubmit={form.handleSubmit(onSubmit)}>
            {error && (
              <div className="flex items-start gap-2 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg dark:bg-red-900/20 dark:border-red-800 dark:text-red-400">
                <svg className="w-5 h-5 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="text-sm">{error}</span>
              </div>
            )}

            {/* Username */}
            <div className="space-y-1.5">
              <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                {t('login.username')}
              </label>
              <input
                id="username"
                type="text"
                autoComplete="username"
                className="block w-full px-3.5 py-2.5 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-colors dark:bg-gray-800 dark:border-gray-700 dark:text-white dark:placeholder-gray-500 sm:text-sm"
                placeholder={t('login.username')}
                {...form.register('username')}
              />
              {form.formState.errors.username && (
                <p className="text-xs text-red-500">{form.formState.errors.username.message}</p>
              )}
            </div>

            {/* Password */}
            <div className="space-y-1.5">
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                {t('login.password')}
              </label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                className="block w-full px-3.5 py-2.5 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-colors dark:bg-gray-800 dark:border-gray-700 dark:text-white dark:placeholder-gray-500 sm:text-sm"
                placeholder={t('login.password')}
                {...form.register('password')}
              />
              {form.formState.errors.password && (
                <p className="text-xs text-red-500">{form.formState.errors.password.message}</p>
              )}
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className="w-full flex justify-center py-2.5 px-4 border border-transparent text-sm font-semibold rounded-lg text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors dark:focus:ring-offset-gray-900"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <Spinner size="sm" />
                  {t('login.button.loading')}
                </span>
              ) : (
                t('login.button')
              )}
            </button>
          </form>

          {/* Footer info */}
          <div className="text-center text-xs text-gray-400 dark:text-gray-600">
            GODNSLOG v2.0 · Self-hosted OAST Platform
          </div>
        </div>
      </div>
    </div>
  )
}
