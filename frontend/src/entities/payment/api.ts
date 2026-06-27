import { get, put } from '../../shared/api'
import type { Payment, PaymentFilter } from './model'

export const paymentApi = {
  getAll(filter: PaymentFilter = {}) {
    const params = new URLSearchParams()
    Object.entries(filter).forEach(([k, v]) => {
      if (v !== undefined && v !== '' && v !== null) params.set(k, String(v))
    })
    const qs = params.toString()
    return get<Payment[]>('/payments' + (qs ? '?' + qs : ''))
  },

  getById(id: number) {
    return get<Payment>(`/payments/${id}`)
  },

  updateAct(paymentId: number, data: { is_sent: boolean; is_signed: boolean; manager_comment: string }) {
    return put<unknown>(`/payments/${paymentId}/act`, data)
  },
}
