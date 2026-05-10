'use client'

import { useState, useEffect, useRef } from 'react'
import { usePathname } from 'next/navigation'
import { Sidebar } from './sidebar'
import { TopBar } from './topbar'
import { cn } from '@/lib/utils'

/** Page title map keyed by route prefix */
const PAGE_TITLES: Record<string, string> = {
  '/dashboard/cases': 'Cases / Case Board',
  '/dashboard/payloads/new': 'Payloads / New Payload',
  '/dashboard/payloads': 'Payloads / Payload Studio',
  '/dashboard/interactions': 'Interactions / Timeline',
  '/dashboard/canary': 'Monitor / Canary Tokens',
  '/dashboard/rebinding': 'Monitor / Rebinding Lab',
  '/dashboard/workflow': 'Monitor / Workflow',
  '/dashboard/settings': 'System / Settings',
  '/dashboard/users': 'System / Users',
  '/dashboard/apikeys': 'System / API Keys',
  '/dashboard/audit': 'System / Audit Log',
  '/dashboard/docs': 'System / Docs',
  '/dashboard': 'Dashboard / Command Center',
}

/** Resolves the current page title from pathname */
function resolvePageTitle(pathname: string): string {
  // Longest prefix match
  const sorted = Object.keys(PAGE_TITLES).sort((a, b) => b.length - a.length)
  for (const key of sorted) {
    if (pathname === key || pathname.startsWith(key + '/')) {
      return PAGE_TITLES[key]
    }
  }
  return 'GODNSLOG'
}

/** Props for the top-level AppShell wrapper */
interface AppShellProps {
  children: React.ReactNode
}

/** Enterprise AppShell: sidebar + topbar + main content area with responsive collapse */
export function AppShell({ children }: AppShellProps) {
  const pathname = usePathname()
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)
  const overlayRef = useRef<HTMLDivElement>(null)

  // Close mobile drawer when route changes
  useEffect(() => {
    setMobileSidebarOpen(false)
  }, [pathname])

  const pageTitle = resolvePageTitle(pathname)

  const handleToggleSidebar = () => {
    // On mobile, toggle drawer; on desktop, toggle collapse
    if (window.innerWidth < 1024) {
      setMobileSidebarOpen((v) => !v)
    } else {
      setSidebarCollapsed((v) => !v)
    }
  }

  return (
    <div className="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950">
      {/* Desktop sidebar */}
      <div className="hidden lg:flex lg:shrink-0">
        <Sidebar collapsed={sidebarCollapsed} />
      </div>

      {/* Mobile sidebar overlay + drawer */}
      {mobileSidebarOpen && (
        <>
          {/* Semi-transparent backdrop */}
          <div
            ref={overlayRef}
            className="fixed inset-0 z-40 bg-black/60 lg:hidden"
            onClick={() => setMobileSidebarOpen(false)}
          />
          {/* Drawer */}
          <div className="fixed inset-y-0 left-0 z-50 lg:hidden">
            <Sidebar onClose={() => setMobileSidebarOpen(false)} />
          </div>
        </>
      )}

      {/* Right side: topbar + page content */}
      <div className={cn('flex flex-col flex-1 min-w-0 overflow-hidden')}>
        <TopBar onToggleSidebar={handleToggleSidebar} pageTitle={pageTitle} />

        {/* Scrollable main area */}
        <main className="flex-1 overflow-y-auto">
          <div className="px-4 py-6 sm:px-6 lg:px-8 max-w-screen-2xl mx-auto">
            {children}
          </div>
        </main>

        {/* Mobile bottom tab bar */}
        <MobileTabBar pathname={pathname} />
      </div>
    </div>
  )
}

/** Mobile-only bottom navigation with 4 primary tabs */
function MobileTabBar({ pathname }: { pathname: string }) {
  const tabs = [
    {
      label: 'Dashboard',
      href: '/dashboard',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
      ),
    },
    {
      label: 'OAST',
      href: '/dashboard/cases',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
    },
    {
      label: 'Monitor',
      href: '/dashboard/canary',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      ),
    },
    {
      label: 'System',
      href: '/dashboard/settings',
      icon: (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      ),
    },
  ]

  const isActiveTab = (href: string) => {
    if (href === '/dashboard') return pathname === '/dashboard'
    return pathname.startsWith(href)
  }

  return (
    <nav className="lg:hidden flex border-t border-gray-200 bg-white dark:bg-gray-800 dark:border-gray-700 shrink-0">
      {tabs.map((tab) => {
        const active = isActiveTab(tab.href)
        return (
          <a
            key={tab.href}
            href={tab.href}
            className={cn(
              'flex-1 flex flex-col items-center justify-center py-2 text-xs font-medium gap-0.5',
              active
                ? 'text-indigo-600 dark:text-indigo-400'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400'
            )}
          >
            {tab.icon}
            <span>{tab.label}</span>
          </a>
        )
      })}
    </nav>
  )
}
