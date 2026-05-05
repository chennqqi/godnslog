'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { caseApi } from '@/lib/api-client'
import type { Case, CaseCreateRequest, CaseUpdateRequest } from '@/types'

export function useCases(params?: { status?: string; search?: string; page?: number; page_size?: number }) {
  return useQuery({
    queryKey: ['cases', params],
    queryFn: () => caseApi.list(params),
  })
}

export function useCase(id: string) {
  return useQuery({
    queryKey: ['cases', id],
    queryFn: () => caseApi.get(id),
    enabled: id.length > 0,
  })
}

export function useCreateCase() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CaseCreateRequest) => caseApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cases'] })
    },
  })
}

export function useUpdateCase() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: CaseUpdateRequest }) => caseApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['cases'] })
      queryClient.invalidateQueries({ queryKey: ['cases', id] })
    },
  })
}

export function useDeleteCase() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => caseApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cases'] })
    },
  })
}
