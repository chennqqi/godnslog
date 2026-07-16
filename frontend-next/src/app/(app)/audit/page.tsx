import { Suspense } from 'react'
import AuditPageContent from './audit-page-content'
import { LoadingState } from '@/components/loading-state'

export default function AuditPage() {
  return (
    <Suspense fallback={<LoadingState />}>
      <AuditPageContent />
    </Suspense>
  )
}
