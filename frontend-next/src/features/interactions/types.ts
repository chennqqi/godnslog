import type { Interaction, ExportRequest, DeleteRequest } from '@/types'

export type { Interaction, ExportRequest, DeleteRequest }

export interface InteractionFilters {
  type?: string
  case_id?: string
  payload_id?: string
  start_time?: string
  end_time?: string
}
