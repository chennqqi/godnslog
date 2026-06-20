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
    'nav.listeners': 'Listeners',
    'nav.retention': 'Retention',
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
    'page.listeners': 'Monitor / Protocol Listeners',
    'page.retention': 'System / Data Retention',
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

    // Cases
    'cases.title': 'Case Board',
    'cases.subtitle': 'Manage OAST engagement cases',
    'cases.new': 'New Case',
    'cases.search': 'Search cases...',
    'cases.all_statuses': 'All statuses',
    'cases.active': 'Active',
    'cases.completed': 'Completed',
    'cases.archived': 'Archived',
    'cases.no_data': 'No cases yet',
    'cases.no_data_hint': 'Create a case to start tracking OAST interactions',
    'cases.col_title': 'Title',
    'cases.col_status': 'Status',
    'cases.col_created': 'Created',
    'cases.target': 'Target',
    'cases.description': 'Description',
    'cases.create': 'Create',
    'cases.cancel': 'Cancel',
    'cases.loading': 'Loading cases...',

    // Payloads
    'payloads.title': 'Payload Studio',
    'payloads.create': 'Create Payload',
    'payloads.batch_create': 'Batch Create',
    'payloads.search': 'Search by token or template...',
    'payloads.no_data': 'No payloads yet',
    'payloads.copy': 'Copy',
    'payloads.token': 'Token',
    'payloads.payload': 'Payload',
    'payloads.case_id': 'Case ID',
    'payloads.expires': 'Expires',
    'payloads.select_template': 'Select Template',
    'payloads.template_content': 'Template Content',
    'payloads.variables': 'Variables',
    'payloads.preview': 'Preview',
    'payloads.generate': 'Generate',
    'payloads.batch_title': 'Batch Create Payloads',
    'payloads.batch_count': 'Count (1-100)',

    // Users
    'users.title': 'User Management',
    'users.subtitle': 'Manage system users and roles',
    'users.new': 'New User',
    'users.username': 'Username',
    'users.email': 'Email',
    'users.role': 'Role',
    'users.created_at': 'Created',
    'users.actions': 'Actions',
    'users.edit': 'Edit',
    'users.delete': 'Delete',
    'users.password': 'Password',
    'users.create': 'Create User',
    'users.edit_title': 'Edit User',
    'users.delete_confirm': 'Are you sure you want to delete this user?',
    'users.delete_title': 'Delete User',
    'users.loading': 'Loading...',
    'users.no_data': 'No users found',
    'users.role.super_admin': 'Super admin',
    'users.role.admin': 'Admin',
    'users.role.user': 'User',
    'users.role.guest': 'Guest',

    // Evidence
    'evidence.title': 'Evidence Report',
    'evidence.subtitle': 'Generate and export evidence reports',
    'evidence.select_case': 'Select Case',
    'evidence.format': 'Format',
    'evidence.generate': 'Generate Report',
    'evidence.download': 'Download',
    'evidence.redact': 'Redact sensitive data',
    'evidence.strength': 'Evidence Strength',
    'evidence.confidence': 'Confidence',
    'evidence.interactions': 'Interactions',
    'evidence.sources': 'Unique Sources',
    'evidence.explainability': 'Explainability',
    'evidence.timeline': 'Timeline',

    // Settings
    'settings.title': 'System Settings',
    'settings.general': 'General',
    'settings.domain': 'Domain',
    'settings.listener': 'Listener',
    'settings.notification': 'Notification',
    'settings.api_keys': 'API Keys',
    'settings.save': 'Save Changes',
    'settings.saved': 'Settings saved successfully',

    // Common actions
    'common.edit': 'Edit',
    'common.view': 'View',
    'common.close': 'Close',
    'common.confirm': 'Confirm',
    'common.back': 'Back',
    'common.refresh': 'Refresh',
    'common.export': 'Export',
    'common.filter': 'Filter',
    'common.all': 'All',
    'common.status': 'Status',
    'common.type': 'Type',
    'common.time': 'Time',
    'common.source': 'Source',
    'common.domain': 'Domain',
    'common.protocol': 'Protocol',
    'common.method': 'Method',
    'common.path': 'Path',
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
    'nav.listeners': '协议监听',
    'nav.retention': '数据保留',
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
    'page.listeners': '监控 / 协议监听',
    'page.retention': '系统 / 数据保留',
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

    // Cases
    'cases.title': '案例看板',
    'cases.subtitle': '管理 OAST 测试案例',
    'cases.new': '新建案例',
    'cases.search': '搜索案例...',
    'cases.all_statuses': '全部状态',
    'cases.active': '进行中',
    'cases.completed': '已完成',
    'cases.archived': '已归档',
    'cases.no_data': '暂无案例',
    'cases.no_data_hint': '创建案例以开始跟踪 OAST 交互',
    'cases.col_title': '标题',
    'cases.col_status': '状态',
    'cases.col_created': '创建时间',
    'cases.target': '目标',
    'cases.description': '描述',
    'cases.create': '创建',
    'cases.cancel': '取消',
    'cases.loading': '加载案例中...',

    // Payloads
    'payloads.title': 'Payload 工作室',
    'payloads.create': '创建 Payload',
    'payloads.batch_create': '批量创建',
    'payloads.search': '按 Token 或模板搜索...',
    'payloads.no_data': '暂无 Payload',
    'payloads.copy': '复制',
    'payloads.token': 'Token',
    'payloads.payload': 'Payload',
    'payloads.case_id': '案例 ID',
    'payloads.expires': '过期时间',
    'payloads.select_template': '选择模板',
    'payloads.template_content': '模板内容',
    'payloads.variables': '变量',
    'payloads.preview': '预览',
    'payloads.generate': '生成',
    'payloads.batch_title': '批量创建 Payload',
    'payloads.batch_count': '数量 (1-100)',

    // Users
    'users.title': '用户管理',
    'users.subtitle': '管理系统用户和角色',
    'users.new': '新建用户',
    'users.username': '用户名',
    'users.email': '邮箱',
    'users.role': '角色',
    'users.created_at': '创建时间',
    'users.actions': '操作',
    'users.edit': '编辑',
    'users.delete': '删除',
    'users.password': '密码',
    'users.create': '创建用户',
    'users.edit_title': '编辑用户',
    'users.delete_confirm': '确定要删除此用户吗？',
    'users.delete_title': '删除用户',
    'users.loading': '加载中...',
    'users.no_data': '未找到用户',
    'users.role.super_admin': '超级管理员',
    'users.role.admin': '管理员',
    'users.role.user': '用户',
    'users.role.guest': '访客',

    // Evidence
    'evidence.title': '证据报告',
    'evidence.subtitle': '生成和导出证据报告',
    'evidence.select_case': '选择案例',
    'evidence.format': '格式',
    'evidence.generate': '生成报告',
    'evidence.download': '下载',
    'evidence.redact': '脱敏敏感数据',
    'evidence.strength': '证据强度',
    'evidence.confidence': '置信度',
    'evidence.interactions': '交互数',
    'evidence.sources': '唯一来源',
    'evidence.explainability': '可解释性',
    'evidence.timeline': '时间线',

    // Settings
    'settings.title': '系统设置',
    'settings.general': '通用',
    'settings.domain': '域名',
    'settings.listener': '监听器',
    'settings.notification': '通知',
    'settings.api_keys': 'API 密钥',
    'settings.save': '保存更改',
    'settings.saved': '设置保存成功',

    // Common actions
    'common.edit': '编辑',
    'common.view': '查看',
    'common.close': '关闭',
    'common.confirm': '确认',
    'common.back': '返回',
    'common.refresh': '刷新',
    'common.export': '导出',
    'common.filter': '筛选',
    'common.all': '全部',
    'common.status': '状态',
    'common.type': '类型',
    'common.time': '时间',
    'common.source': '来源',
    'common.domain': '域名',
    'common.protocol': '协议',
    'common.method': '方法',
    'common.path': '路径',
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
        const timer = setTimeout(() => setLangState(saved), 0)
        return () => clearTimeout(timer)
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
