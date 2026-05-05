'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { interactionApi } from '@/lib/api-client'

export function useInteractions(params?: {
  case_id?: string
  payload_id?: string
  type?: string
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}) {
  return useQuery({
    queryKey: ['interactions', params],
    queryFn: () => interactionApi.list(params),
  })
}

export function useInteraction(id: string) {
  return useQuery({
    queryKey: ['interactions', id],
    queryFn: () => interactionApi.get(id),
    enabled: id.length > 0,
  })
}

export function useDeleteInteractions() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (ids: string[]) => interactionApi.delete(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interactions'] })
    },
  })
}

export function useExportInteractions() {
  return useMutation({
    mutationFn: (data: any) => interactionApi.export(data),
  })
}
