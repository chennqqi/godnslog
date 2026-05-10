import type { Payload, PayloadCreateRequest } from '@/types'

export type { Payload, PayloadCreateRequest }

export interface PayloadFormData {
  case_id: string
  template: string
  variables: Record<string, string>
  expected_protocol?: string
  expires_at?: string
}
