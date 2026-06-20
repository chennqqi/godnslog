import { z } from 'zod'

/** Re-export payload-related settings schemas from payload schema */
export {
  generalSettingsSchema,
  domainSettingsSchema,
  listenerSettingsSchema,
  notificationSchema,
  type GeneralSettingsFormValues,
  type DomainSettingsFormValues,
  type ListenerSettingsFormValues,
  type NotificationFormValues,
} from '@/features/payloads/schemas/payload-schema'
