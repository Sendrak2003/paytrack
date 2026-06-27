import type { ReactNode } from 'react'

interface Props {
  title: string
  onClose: () => void
  footer?: ReactNode
  children: ReactNode
}

export function Modal({ title, onClose, footer, children }: Props) {
  return (
    <div className="modal-overlay" onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div className="modal">
        <div className="modal-header">
          <span>{title}</span>
          <button className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Закрыть">&times;</button>
        </div>
        <div className="modal-body">{children}</div>
        {footer && <div className="modal-footer">{footer}</div>}
      </div>
    </div>
  )
}
