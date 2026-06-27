import type { ReactNode } from 'react'

interface Props {
  label: string
  value: ReactNode
  sub?: string
  variant?: 'primary' | 'success' | 'warning' | 'danger'
}

export function StatCard({ label, value, sub, variant }: Props) {
  return (
    <div className={`stat-card ${variant ?? ''}`}>
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      {sub && <div className="stat-sub">{sub}</div>}
    </div>
  )
}
