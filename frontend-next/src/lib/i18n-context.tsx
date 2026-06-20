'use client'

import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'

export type Language = 'en-US' | 'zh-CN'

/** Full translation dictionary */
const translations = {
  'en-US': {
    // Login
    'login.title': 'GODNSLOG 2.0',
    'login.subtitle': 'Sign in to your account',
    'login.username': 'Username',
    'login.password': 'Password',
    'login.button': 'Sign In',
    'login.button.loading': 'Signing in...',
    'login.error': 'Login failed, please check your username and password',

    // Sidebar groups
    'nav.group.oast': 'OAST CORE',
    'nav.group.monitor': 'MONITOR',
    'nav.group.integrations': 'INTEGRATIONS',
    'nav.group.system': 'SYSTEM',

    // Sidebar items
    'nav.dashboard': 'Dashboard',
    'nav.cases': 'Cases',
    'nav.payloads': 'Payloads',
    'nav.agent-runs': 'Agent Runs',
    'nav.interactions': 'Interactions',
    'nav.evidence-summary': 'Evidence Summary',
    'nav.canary': 'Canary Tokens',
    'nav.rebinding': 'Rebinding Lab',
    'nav.workflow': 'Workflow',
    'nav.scanner-hub': 'Scanner Hub',
    'nav.marketplace': 'Marketplace',
    'nav.settings': 'Settings',
    'nav.users': 'Users',
    'nav.apikeys': 'API Keys',
    'nav.audit': 'Audit Log',
    'nav.docs': 'Docs',

    // Topbar
    'topbar.theme.light': 'Light',
    'topbar.theme.dark': 'Dark',
    'topbar.theme.system': 'System',
    'topbar.profile': 'Profile & Settings',
    'topbar.apikeys': 'API Keys',
    'topbar.signout': 'Sign Out',
    'topbar.notifications': 'Notifications',

    // Page titles
    'page.dashboard': 'Dashboard / Command Center',
    'page.cases': 'Cases / Case Board',
    'page.payloads.new': 'Payloads / New Payload',
    'page.payloads': 'Payloads / Payload Studio',
    'page.agent-runs': 'Agent Runs / Execution History',
    'page.interactions': 'Interactions / Timeline',
    'page.evidence-summary': 'Evidence / Summary',
    'page.canary': 'Monitor / Canary Tokens',
    'page.rebinding': 'Monitor / Rebinding Lab',
    'page.workflow': 'Monitor / Workflow',
    'page.scanner-hub': 'Integrations / Scanner Hub',
    'page.marketplace': 'Integrations / Marketplace',
    'page.settings': 'System / Settings',
    'page.users': 'System / Users',
    'page.apikeys': 'System / API Keys',
    'page.audit': 'System / Audit Log',
    'page.docs': 'System / Docs',

    // Common
    'common.loading': 'Loading...',
    'common.save': 'Save',
    'common.cancel': 'Cancel',
    'common.create': 'Create',
    'common.delete': 'Delete',
    'common.search': 'Search',
    'common.no_data': 'No data',
    'common.error': 'An error occurred',

    // Interactions
    'interactions.title': 'Interaction Timeline',
    'interactions.all': 'All Interactions',
    'interactions.live.on': '● Live',
    'interactions.live.connecting': '● Connecting...',
    'interactions.live.off': 'Live Stream: OFF',
    'interactions.new_count': 'new',
    'interactions.statistics': 'Statistics',
    'interactions.total': 'Total',
    'interactions.no_data': 'No interactions yet',
    'interactions.details': 'Details',
    'interactions.triage': 'Interaction Triage',
  },
  'zh-CN': {
    // Login
    'login.title': 'GODNSLOG 2.0',
    'login.subtitle': '登录到您的账户',
    'login.username': '用户名',
    'login.password': '密码',
    'login.button': '登录',
    'login.button.loading': '登录中...',
    'login.error': '登录失败，请检查用户名和密码',

    // Sidebar groups
    'nav.group.oast': 'OAST 核心',
    'nav.group.monitor': '监控',
    'nav.group.integrations': '集成',
    'nav.group.system': '系统',

    // Sidebar items
    'nav.dashboard': '仪表盘',
    'nav.cases': '案例',
    'nav.payloads': 'Payload',
    'nav.agent-runs': 'Agent 运行',
    'nav.interactions': '交互记录',
    'nav.evidence-summary': '证据摘要',
    'nav.canary': '金丝雀令牌',
    'nav.rebinding': '重绑定实验室',
    'nav.workflow': '工作流',
    'nav.scanner-hub': '扫描器中心',
    'nav.marketplace': '市场',
    'nav.settings': '设置',
    'nav.users': '用户',
    'nav.apikeys': 'API 密钥',
    'nav.audit': '审计日志',
    'nav.docs': '文档',

    // Topbar
    'topbar.theme.light': '浅色',
    'topbar.theme.dark': '深色',
    'topbar.theme.system': '跟随系统',
    'topbar.profile': '个人资料与设置',
    'topbar.apikeys': 'API 密钥',
    'topbar.signout': '退出登录',
    'topbar.notifications': '通知',

    // Page titles
    'page.dashboard': '仪表盘 / 指挥中心',
    'page.cases': '案例 / 案例看板',
    'page.payloads.new': 'Payload / 新建',
    'page.payloads': 'Payload / 工作室',
    'page.agent-runs': 'Agent 运行 / 执行历史',
    'page.interactions': '交互记录 / 时间线',
    'page.evidence-summary': '证据 / 摘要',
    'page.canary': '监控 / 金丝雀令牌',
    'page.rebinding': '监控 / 重绑定实验室',
    'page.workflow': '监控 / 工作流',
    'page.scanner-hub': '集成 / 扫描器中心',
    'page.marketplace': '集成 / 市场',
    'page.settings': '系统 / 设置',
    'page.users': '系统 / 用户',
    'page.apikeys': '系统 / API 密钥',
    'page.audit': '系统 / 审计日志',
    'page.docs': '系统 / 文档',

    // Common
    'common.loading': '加载中...',
    'common.save': '保存',
    'common.cancel': '取消',
    'common.create': '创建',
    'common.delete': '删除',
    'common.search': '搜索',
    'common.no_data': '暂无数据',
    'common.error': '发生错误',

    // Interactions
    'interactions.title': '交互时间线',
    'interactions.all': '全部交互',
    'interactions.live.on': '● 实时',
    'interactions.live.connecting': '● 连接中...',
    'interactions.live.off': '实时流: 关闭',
    'interactions.new_count': '条新记录',
    'interactions.statistics': '统计',
    'interactions.total': '总计',
    'interactions.no_data': '暂无交互记录',
    'interactions.details': '详情',
    'interactions.triage': '交互分诊',
  },
} as const

export type TranslationKey = keyof (typeof translations)['en-US']

interface I18nContextValue {
  lang: Language
  t: (key: TranslationKey) => string
  setLang: (lang: Language) => void
}

const I18nContext = createContext<I18nContextValue | null>(null)

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Language>('en-US')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('language') as Language
      if (saved === 'en-US' || saved === 'zh-CN') {
        setLangState(saved)
      }
    }
  }, [])

  const setLang = useCallback((newLang: Language) => {
    setLangState(newLang)
    if (typeof window !== 'undefined') {
      localStorage.setItem('language', newLang)
    }
  }, [])

  const t = useCallback(
    (key: TranslationKey): string => {
      return translations[lang]?.[key] ?? translations['en-US'][key] ?? key
    },
    [lang]
  )

  return <I18nContext.Provider value={{ lang, t, setLang }}>{children}</I18nContext.Provider>
}

export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext)
  if (!ctx) {
    // Fallback for components outside provider
    return {
      lang: 'en-US',
      t: (key: TranslationKey) => translations['en-US'][key] ?? key,
      setLang: () => {},
    }
  }
  return ctx
}

/** Get current language from localStorage (for non-React contexts) */
export function getCurrentLanguage(): Language {
  if (typeof window === 'undefined') return 'en-US'
  const saved = localStorage.getItem('language') as Language
  return saved === 'zh-CN' ? 'zh-CN' : 'en-US'
}

/** Set language in localStorage (for non-React contexts) */
export function setLanguage(lang: Language) {
  if (typeof window !== 'undefined') {
    localStorage.setItem('language', lang)
    window.location.reload()
  }
}
