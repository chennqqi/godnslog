'use client'

import { useQuery, useMutation } from '@tanstack/react-query'
import { evidenceApi } from '@/lib/api-client'

export function useEvidenceGenerate() {
  return useMutation({
    mutationFn: (params: { format: 'json' | 'markdown'; case_id?: string; payload_id?: string }) =>
      evidenceApi.generate(params),
  })
}

export function useEvidence(caseId?: string) {
  return useQuery({
    queryKey: ['evidence', caseId],
    queryFn: () => evidenceApi.generate({ case_id: caseId!, format: 'markdown' }),
    enabled: !!caseId,
  })
}
