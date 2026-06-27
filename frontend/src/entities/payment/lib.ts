import type { ActStatus } from './model'

/**
 * Вычисление статуса акта — бизнес-логика entity layer.
 * Зеркало серверной логики CalculateStatus (models.go).
 */
export function getActStatus(payment: { payment_date: string; act?: { is_sent: boolean; is_signed: boolean; sent_at?: string | null } | null }): ActStatus {
  const act = payment.act
  if (!act) {
    const age = Date.now() - new Date(payment.payment_date).getTime()
    return age > 30 * 24 * 60 * 60 * 1000 ? 'needs_attention' : 'not_sent'
  }
  if (act.is_sent && act.is_signed) return 'closed'
  if (act.is_sent && !act.is_signed) {
    if (act.sent_at) {
      const age = Date.now() - new Date(act.sent_at).getTime()
      if (age > 14 * 24 * 60 * 60 * 1000) return 'needs_attention'
    }
    return 'waiting_signature'
  }
  const age = Date.now() - new Date(payment.payment_date).getTime()
  return age > 30 * 24 * 60 * 60 * 1000 ? 'needs_attention' : 'not_sent'
}

export const ACT_STATUS_LABELS: Record<ActStatus, string> = {
  not_sent: 'Не отправлен',
  waiting_signature: 'Ожидает подписи',
  closed: 'Закрыт',
  needs_attention: 'Требует внимания',
}
