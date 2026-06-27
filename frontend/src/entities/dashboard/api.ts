import { get } from '../../shared/api'
import type { DashboardSummary } from './model'

export const dashboardApi = {
  getSummary: () => get<DashboardSummary>('/dashboard/summary'),
}
