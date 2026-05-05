'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { payloadApi } from '@/lib/api-client'
import type { PayloadCreateRequest } from '@/types'

export function usePayloads(params?: { case_id?: string; status?: string; page?: number; page_size?: number }) {
  return useQuery({
    queryKey: ['payloads', params],
    queryFn: () => payloadApi.list(params),
  })
}

export function usePayload(id: string) {
  return useQuery({
    queryKey: ['payloads', id],
    queryFn: () => payloadApi.get(id),
    enabled: id.length > 0,
  })
}

export function useCreatePayload() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: PayloadCreateRequest) => payloadApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payloads'] })
    },
  })
}

export function useRevokePayload() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => payloadApi.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payloads'] })
    },
  })
}
