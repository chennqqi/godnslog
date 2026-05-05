import type { Settings, SettingsCreateRequest, SettingsUpdateRequest } from '@/types'

export type { Settings, SettingsCreateRequest, SettingsUpdateRequest }

export interface SettingsFormData {
  dns_domain: string
  dns_port: number
  dns_ttl: number
  http_port: number
  https_tls_cert: string
  https_tls_key: string
  enable_auth: boolean
  session_timeout: number
  enable_notification: boolean
  notification_url: string
  log_level: string
  log_retention_days: number
}
