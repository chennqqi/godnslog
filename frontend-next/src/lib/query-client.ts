'use client'

import { QueryClient, QueryClientConfig, QueryCache, MutationCache } from '@tanstack/react-query'

function handleApiError(error: unknown) {
  if (error && typeof error === 'object' && 'response' in error) {
    const resp = error as { response?: { status?: number } }
    if (resp.response?.status === 401) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }
    }
  }
}

const queryClientConfig: QueryClientConfig = {
  queryCache: new QueryCache({
    onError: (error) => handleApiError(error),
  }),
  mutationCache: new MutationCache({
    onError: (error) => handleApiError(error),
  }),
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000,
      refetchOnWindowFocus: true,
      retry: 1,
    },
    mutations: {
      retry: 0,
    },
  },
}

export function makeQueryClient() {
  return new QueryClient(queryClientConfig)
}

let browserQueryClient: QueryClient | undefined

export function getQueryClient() {
  if (typeof window === 'undefined') {
    return makeQueryClient()
  }
  if (!browserQueryClient) {
    browserQueryClient = makeQueryClient()
  }
  return browserQueryClient
}
