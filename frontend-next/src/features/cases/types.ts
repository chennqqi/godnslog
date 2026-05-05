import type { Case, CaseCreateRequest, CaseUpdateRequest } from '@/types'

export type { Case, CaseCreateRequest, CaseUpdateRequest }

export interface CaseFormData {
  title: string
  description: string
  target: string
  status: 'active' | 'completed' | 'archived'
  tags: string[]
}
