import { z } from 'zod'

/** Zod schema for payload creation form */
export const payloadSchema = z.object({
  template: z.string().min(1, 'Template is required'),
  scenario: z.string().max(500, 'Scenario must be less than 500 characters'),
  case_id: z.string(),
  expires_in: z.number().int().min(60, 'Expiry must be at least 60 seconds').max(2592000, 'Expiry must be at most 30 days'),
})

export type PayloadFormValues = z.infer<typeof payloadSchema>

/** Zod schema for notification settings */
export const notificationSchema = z.object({
  id: z.string(),
  name: z.string(),
  enabled: z.boolean(),
  webhook_url: z.string().url('Must be a valid URL').or(z.literal('')),
  webhook_body: z.string(),
})

export type NotificationFormValues = z.infer<typeof notificationSchema>

/** Zod schema for general settings */
export const generalSettingsSchema = z.object({
  system_name: z.string().min(1, 'System name is required').max(100, 'System name must be less than 100 characters'),
  language: z.enum(['en-US', 'zh-CN']),
  timezone: z.string().min(1, 'Timezone is required'),
})

export type GeneralSettingsFormValues = z.infer<typeof generalSettingsSchema>

/** Zod schema for domain settings */
export const domainSettingsSchema = z.object({
  main_domain: z.string().min(1, 'Main domain is required').max(255, 'Main domain must be less than 255 characters'),
  dns_domain: z.string().max(255, 'DNS domain must be less than 255 characters'),
  http_domain: z.string().max(255, 'HTTP domain must be less than 255 characters'),
})

export type DomainSettingsFormValues = z.infer<typeof domainSettingsSchema>

/** Zod schema for listener settings */
export const listenerSettingsSchema = z.object({
  dns_listen: z.string().min(1, 'DNS listen address is required').max(50, 'Address must be less than 50 characters'),
  http_listen: z.string().min(1, 'HTTP listen address is required').max(50, 'Address must be less than 50 characters'),
  https_listen: z.string().max(50, 'Address must be less than 50 characters'),
})

export type ListenerSettingsFormValues = z.infer<typeof listenerSettingsSchema>
