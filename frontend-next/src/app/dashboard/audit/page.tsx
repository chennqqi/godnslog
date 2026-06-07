import { Suspense } from 'react'
import AuditPageContent from './audit-page-content'

export default function AuditPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <AuditPageContent />
    </Suspense>
  )
}
