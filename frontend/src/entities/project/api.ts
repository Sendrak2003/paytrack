import { get } from '../../shared/api'
import type { ProjectSummary } from './model'

export const projectApi = {
  getAll: () => get<ProjectSummary[]>('/projects'),
}
