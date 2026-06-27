import type { Payment } from '../model'
import { getActStatus } from '../lib'
import { ActStatusBadge } from './ActStatusBadge'
import { Badge } from '../../../shared/ui'
import { formatMoney, formatDate } from '../../../shared/lib'
import { Pencil } from 'lucide-react'

interface Props {
  payment: Payment
  onEdit: (p: Payment) => void
}

export function PaymentCard({ payment, onEdit }: Props) {
  const status = payment.act?.status || getActStatus(payment)
  return (
    <div className="payment-card">
      <div className="payment-card-top">
        <span className="amount amount-positive">{formatMoney(payment.amount)}</span>
        <ActStatusBadge status={status} />
      </div>
      <div className="payment-card-project">{payment.project?.name ?? '—'}</div>
      <div className="payment-card-row">
        <span className="payment-card-label">Юрлицо</span>
        <span>{payment.legal_entity?.name ?? '—'}</span>
      </div>
      <div className="payment-card-row">
        <span className="payment-card-label">Дата</span>
        <span>{formatDate(payment.payment_date)}</span>
      </div>
      <div className="payment-card-row">
        <span className="payment-card-label">Этап</span>
        <Badge variant="active">{payment.service_stage}</Badge>
      </div>
      {payment.invoice_number && (
        <div className="payment-card-row">
          <span className="payment-card-label">Счёт / Договор</span>
          <span>№{payment.invoice_number}{payment.contract_number ? ` · ${payment.contract_number}` : ''}</span>
        </div>
      )}
      {payment.payment_purpose && (
        <div className="payment-card-purpose">{payment.payment_purpose}</div>
      )}
      {payment.act?.manager_comment && (
        <div className="payment-card-comment">{payment.act.manager_comment}</div>
      )}
      <button className="btn btn-primary btn-sm payment-card-action" onClick={() => onEdit(payment)}>
        <Pencil size={12} /> Статус акта
      </button>
    </div>
  )
}
