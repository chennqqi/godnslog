'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { useThemeStore } from '@/stores/theme-store'
import { initTheme } from '@/stores/theme-store'
import { useI18n, type TranslationKey } from '@/lib/i18n-context'

/** Props for TopBar component */
interface TopBarProps {
  /** Callback to toggle the sidebar collapsed/expanded state */
  onToggleSidebar: () => void
  /** Current page title shown as breadcrumb */
  pageTitle?: string
}

const SunIcon = () => (
  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m8.66-13H20m-16 0H2.34M18.36 5.64l-.71.71M6.34 17.66l-.71.71M18.36 18.36l-.71-.71M6.34 6.34l-.71-.71M12 8a4 4 0 100 8 4 4 0 000-8z" />
  </svg>
)

const MoonIcon = () => (
  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
  </svg>
)

const BellIcon = () => (
  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
  </svg>
)

const MenuIcon = () => (
  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
  </svg>
)

const UserCircleIcon = () => (
  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

type ThemeMode = 'light' | 'dark' | 'system'

/** Enterprise top bar with sidebar toggle, breadcrumb, theme switcher, and user menu */
export function TopBar({ onToggleSidebar, pageTitle }: TopBarProps) {
  const router = useRouter()
  const { theme, setTheme } = useThemeStore()
  const { t } = useI18n()
  const [username, setUsername] = useState<string>('Admin')

  useEffect(() => {
    initTheme()

    try {
      const userStr = localStorage.getItem('user')
      if (userStr) {
        const user = JSON.parse(userStr)
        if (user?.username) setUsername(user.username)
      }
    } catch {
      // ignore malformed user data
    }
  }, [])

  const handleThemeChange = (mode: ThemeMode) => {
    setTheme(mode)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    router.push('/login')
  }

  const themeLabel: Record<ThemeMode, TranslationKey> = {
    light: 'topbar.theme.light',
    dark: 'topbar.theme.dark',
    system: 'topbar.theme.system',
  }

  return (
    <header className="h-14 bg-white border-b border-gray-200 flex items-center px-4 gap-3 shrink-0 dark:bg-gray-800 dark:border-gray-700">
      {/* Sidebar toggle */}
      <button
        onClick={onToggleSidebar}
        className="p-1.5 rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
        aria-label="Toggle sidebar"
      >
        <MenuIcon />
      </button>

      {/* Breadcrumb / page title */}
      <div className="flex-1 min-w-0">
        {pageTitle && (
          <h1 className="text-sm font-semibold text-gray-800 truncate dark:text-gray-100">
            {pageTitle}
          </h1>
        )}
      </div>

      {/* Right-side actions */}
      <div className="flex items-center gap-2">
        {/* Theme switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="gap-1 text-xs text-gray-500 dark:text-gray-400">
              {theme === 'dark' ? <MoonIcon /> : <SunIcon />}
              <span className="hidden sm:inline">{t(themeLabel[theme])}</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => handleThemeChange('light')}>
              <SunIcon />
              <span className="ml-2">{t('topbar.theme.light')}</span>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleThemeChange('dark')}>
              <MoonIcon />
              <span className="ml-2">{t('topbar.theme.dark')}</span>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleThemeChange('system')}>
              <span className="ml-2">{t('topbar.theme.system')}</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Notifications placeholder */}
        <button
          className="p-1.5 rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 relative"
          aria-label="Notifications"
        >
          <BellIcon />
          <span className="absolute top-1 right-1 w-1.5 h-1.5 bg-red-500 rounded-full" />
        </button>

        {/* User menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700">
              <UserCircleIcon />
              <span className="text-sm font-medium text-gray-700 hidden sm:inline dark:text-gray-200">
                {username}
              </span>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem onClick={() => router.push('/dashboard/settings')}>
              {t('topbar.profile')}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.push('/dashboard/apikeys')}>
              {t('topbar.apikeys')}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout} className="text-red-600">
              {t('topbar.signout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
