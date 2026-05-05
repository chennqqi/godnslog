import type { Payload, PayloadCreateRequest, PayloadUpdateRequest } from '@/types'

export type { Payload, PayloadCreateRequest, PayloadUpdateRequest }

export interface PayloadFormData {
  case_id: string
  template: string
  variables: Record<string, string>
  expected_protocol?: string
  expires_at?: string
}
