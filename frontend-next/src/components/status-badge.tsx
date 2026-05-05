import { Badge } from '@/components/ui/badge'

export interface StatusBadgeProps {
  status: string
  variant?: 'default' | 'secondary' | 'destructive' | 'outline'
  className?: string
}

const statusVariants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  active: 'default',
  completed: 'secondary',
  archived: 'outline',
  draft: 'outline',
  deployed: 'secondary',
  hit: 'default',
  expired: 'destructive',
}

export function StatusBadge({ status, variant, className }: StatusBadgeProps) {
  const badgeVariant = variant || statusVariants[status.toLowerCase()] || 'outline'
  
  return (
    <Badge variant={badgeVariant} className={className}>
      {status}
    </Badge>
  )
}
