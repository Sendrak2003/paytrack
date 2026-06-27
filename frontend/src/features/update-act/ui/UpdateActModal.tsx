import { useState } from 'react'
import type { Payment } from '../../../entities/payment/model'
import { paymentApi } from '../../../entities/payment'
import { Modal, Button } from '../../../shared/ui'
import { formatDate, formatMoney } from '../../../shared/lib'
import { Check } from 'lucide-react'

interface Props {
  payment: Payment
  onClose: () => void
  onSaved: (updated: Payment) => void
}

export function UpdateActModal({ payment, onClose, onSaved }: Props) {
  const act = payment.act
  const [isSent, setIsSent] = useState(act?.is_sent ?? false)
  const [isSigned, setIsSigned] = useState(act?.is_signed ?? false)
  const [comment, setComment] = useState(act?.manager_comment ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function handleSave() {
    setSaving(true)
    setError('')
    try {
      await paymentApi.updateAct(payment.id, { is_sent: isSent, is_signed: isSigned, manager_comment: comment })
      const updated = await paymentApi.getById(payment.id)
      onSaved(updated)
      onClose()
    } catch {
      setError('Ошибка сохранения. Попробуйте снова.')
    } finally {
      setSaving(false)
    }
  }

  const footer = (
    <>
      <Button variant="ghost" onClick={onClose}>Отмена</Button>
      <Button variant="primary" onClick={handleSave} disabled={saving}>
        {saving ? 'Сохранение...' : <><Check size={14} /> Сохранить</>}
      </Button>
    </>
  )

  return (
    <Modal title={`Статус акта — оплата #${payment.id}`} onClose={onClose} footer={footer}>
      <div className="act-modal-info">
        <div className="act-modal-info__title">
          <strong>{payment.project?.name ?? '—'}</strong>
          <span className="act-modal-info__stage">/ {payment.service_stage}</span>
        </div>
        <div className="act-modal-info__meta">
          {payment.legal_entity?.name ?? '—'} · {formatDate(payment.payment_date)} · <span className="amount">{formatMoney(payment.amount)}</span>
        </div>
        {payment.invoice_number && (
          <div className="act-modal-info__meta">Счёт №{payment.invoice_number} · Договор {payment.contract_number}</div>
        )}
        <div className="act-modal-info__purpose">{payment.payment_purpose}</div>
      </div>

      <div className="form-group">
        <label className="checkbox-group">
          <input type="checkbox" checked={isSent} onChange={e => { setIsSent(e.target.checked); if (!e.target.checked) setIsSigned(false) }} />
          <span>Акт отправлен клиенту</span>
          {act?.sent_at && <span className="act-modal-date">({formatDate(act.sent_at)})</span>}
        </label>
      </div>

      <div className="form-group">
        <label className="checkbox-group">
          <input type="checkbox" checked={isSigned} disabled={!isSent} onChange={e => setIsSigned(e.target.checked)} />
          <span style={{ color: !isSent ? 'var(--text-muted)' : undefined }}>Акт подписан клиентом</span>
          {act?.signed_at && <span className="act-modal-date">({formatDate(act.signed_at)})</span>}
        </label>
        {!isSent && <div className="act-modal-hint">Сначала отметьте отправку акта</div>}
      </div>

      <div className="form-group">
        <label className="form-label">Комментарий менеджера</label>
        <textarea className="form-control" value={comment} onChange={e => setComment(e.target.value)} placeholder="Например: «Клиент обещал подписать до 15 числа»" />
      </div>

      {error && <div className="act-modal-error">{error}</div>}
    </Modal>
  )
}
