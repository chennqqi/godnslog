// Simple i18n implementation
export type Language = 'en-US' | 'zh-CN'

type TranslationKeys = {
  'login.title': string
  'login.subtitle': string
  'login.username': string
  'login.password': string
  'login.button': string
  'login.button.loading': string
  'login.error': string
  'dashboard.title': string
  'dashboard.menu.dashboard': string
  'dashboard.menu.cases': string
  'dashboard.menu.payloads': string
  'dashboard.menu.interactions': string
  'dashboard.menu.workflow': string
  'dashboard.menu.rebinding': string
  'dashboard.menu.canary': string
  'dashboard.menu.marketplace': string
  'dashboard.menu.evidence': string
  'dashboard.menu.settings': string
  'dashboard.menu.logout': string
}

export const translations: Record<Language, TranslationKeys> = {
  'en-US': {
    // Login page
    'login.title': 'GODNSLOG 2.0',
    'login.subtitle': 'Sign in to your account',
    'login.username': 'Username',
    'login.password': 'Password',
    'login.button': 'Sign In',
    'login.button.loading': 'Signing in...',
    'login.error': 'Login failed, please check your username and password',
    
    // Dashboard
    'dashboard.title': 'GODNSLOG 2.0',
    'dashboard.menu.dashboard': 'Dashboard',
    'dashboard.menu.cases': 'Cases',
    'dashboard.menu.payloads': 'Payloads',
    'dashboard.menu.interactions': 'Interactions',
    'dashboard.menu.workflow': 'Workflow',
    'dashboard.menu.rebinding': 'Rebinding',
    'dashboard.menu.canary': 'Canary',
    'dashboard.menu.marketplace': 'Marketplace',
    'dashboard.menu.evidence': 'Evidence',
    'dashboard.menu.settings': 'Settings',
    'dashboard.menu.logout': 'Logout',
  },
  'zh-CN': {
    // Login page
    'login.title': 'GODNSLOG 2.0',
    'login.subtitle': '登录到您的账户',
    'login.username': '用户名',
    'login.password': '密码',
    'login.button': '登录',
    'login.button.loading': '登录中...',
    'login.error': '登录失败，请检查用户名和密码',
    
    // Dashboard
    'dashboard.title': 'GODNSLOG 2.0',
    'dashboard.menu.dashboard': '仪表盘',
    'dashboard.menu.cases': 'Cases',
    'dashboard.menu.payloads': 'Payloads',
    'dashboard.menu.interactions': 'Interactions',
    'dashboard.menu.workflow': 'Workflow',
    'dashboard.menu.rebinding': 'Rebinding',
    'dashboard.menu.canary': 'Canary',
    'dashboard.menu.marketplace': '市场',
    'dashboard.menu.evidence': '证据报告',
    'dashboard.menu.settings': '设置',
    'dashboard.menu.logout': '登出',
  },
}

export function t<K extends keyof TranslationKeys>(key: K, lang: Language = 'en-US'): string {
  return translations[lang]?.[key] || key
}

export function getCurrentLanguage(): Language {
  if (typeof window === 'undefined') return 'en-US'
  const saved = localStorage.getItem('language') as Language
  return saved || 'en-US'
}

export function setLanguage(lang: Language) {
  if (typeof window !== 'undefined') {
    localStorage.setItem('language', lang)
    window.location.reload()
  }
}
