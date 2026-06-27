import type { ReactNode } from 'react'

interface Props {
  children: ReactNode
  className?: string
  header?: ReactNode
  bodyClassName?: string
}

export function Card({ children, className = '', header, bodyClassName }: Props) {
  return (
    <div className={`card ${className}`.trim()}>
      {header && <div className="card-header">{header}</div>}
      {bodyClassName ? <div className={bodyClassName}>{children}</div> : children}
    </div>
  )
}
