import type { Client } from '../client'

export interface Project {
  id: number
  name: string
  client_id: number
  client?: Client
  status: 'active' | 'completed' | 'paused'
}

export interface ProjectSummary {
  id: number
  name: string
  client_name: string
  inn: string
  status: string
  total_amount: number
  payments_count: number
  closed_acts_count: number
  open_acts_count: number
  doc_status: 'all_closed' | 'has_open' | 'needs_attention'
}
