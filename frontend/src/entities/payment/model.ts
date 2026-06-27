import type { Project } from '../project'
import type { Client } from '../client'

export type ActStatus = 'not_sent' | 'waiting_signature' | 'closed' | 'needs_attention'

export interface Act {
  id: number
  payment_id: number
  is_sent: boolean
  sent_at: string | null
  is_signed: boolean
  signed_at: string | null
  status: ActStatus
  manager_comment: string
  created_at: string
  updated_at: string
}

export interface Payment {
  id: number
  project_id: number
  project?: Project
  legal_entity_id: number
  legal_entity?: Client
  payment_date: string
  amount: number
  payment_purpose: string
  service_stage: string
  invoice_number: string
  contract_number: string
  act?: Act | null
  created_at: string
  updated_at: string
}

export interface PaymentFilter {
  project_id?: number
  legal_entity_id?: number
  act_status?: ActStatus | ''
  service_stage?: string
  search?: string
  date_from?: string
  date_to?: string
}
