import { get } from '../../shared/api'
import type { Client } from './model'

export const clientApi = {
  getAll: () => get<Client[]>('/clients'),
}
