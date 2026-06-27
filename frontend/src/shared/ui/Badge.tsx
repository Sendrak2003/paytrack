import type { ReactNode } from 'react'

interface Props {
  variant?: string
  children: ReactNode
  className?: string
}

export function Badge({ variant, children, className = '' }: Props) {
  return (
    <span className={`badge ${variant ? `badge-${variant}` : ''} ${className}`.trim()}>
      {children}
    </span>
  )
}
