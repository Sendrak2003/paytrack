import { Check, Clock, Circle, AlertTriangle } from 'lucide-react'
import { Badge } from '../../../shared/ui'
import type { ActStatus } from '../model'
import { ACT_STATUS_LABELS } from '../lib'

const BADGE_VARIANT: Record<ActStatus, string> = {
  not_sent: 'not_sent',
  waiting_signature: 'waiting',
  closed: 'closed',
  needs_attention: 'attention',
}

function StatusIcon({ status }: { status: ActStatus }) {
  const size = 12
  switch (status) {
    case 'closed': return <Check size={size} />
    case 'waiting_signature': return <Clock size={size} />
    case 'needs_attention': return <AlertTriangle size={size} />
    default: return <Circle size={size} />
  }
}

export function ActStatusBadge({ status }: { status: ActStatus }) {
  return (
    <Badge variant={BADGE_VARIANT[status]}>
      <StatusIcon status={status} /> {ACT_STATUS_LABELS[status]}
    </Badge>
  )
}
