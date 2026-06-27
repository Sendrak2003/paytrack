import type { ReactNode } from 'react'

interface Props {
  icon?: ReactNode
  message: string
}

export function EmptyState({ icon, message }: Props) {
  return (
    <div className="empty-state">
      {icon && <div className="empty-state-icon-svg">{icon}</div>}
      <div>{message}</div>
    </div>
  )
}
