import type { Interaction } from '@/types'

export type { Interaction }

export interface InteractionFilters {
  type?: string
  case_id?: string
  payload_id?: string
  start_time?: string
  end_time?: string
}
