'use client'

import { cn } from '@/lib/utils'
import { Spinner } from '@/components/ui/spinner'
import { useI18n } from '@/lib/i18n-context'

interface LoadingStateProps {
  message?: string
  className?: string
}

/** Full-page or section-level loading placeholder */
export function LoadingState({ message, className }: LoadingStateProps) {
  const { t } = useI18n()
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 gap-3 text-gray-500 dark:text-gray-400', className)}>
      <Spinner size="lg" className="text-gray-400" />
      <span className="text-sm">{message || t('common.loading')}</span>
    </div>
  )
}
