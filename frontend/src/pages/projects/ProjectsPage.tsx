import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { FolderOpen, ArrowRight } from 'lucide-react'
import { projectApi, type ProjectSummary, PROJECT_STATUS_LABELS, DOC_STATUS_LABELS } from '../../entities/project'
import { Card, Badge, Spinner, EmptyState, FilterInput } from '../../shared/ui'
import { formatMoney } from '../../shared/lib'

export function ProjectsPage() {
  const [projects, setProjects] = useState<ProjectSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => { projectApi.getAll().then(setProjects).finally(() => setLoading(false)) }, [])

  const filtered = projects.filter(p => !search || p.name.toLowerCase().includes(search.toLowerCase()) || p.client_name.toLowerCase().includes(search.toLowerCase()) || p.inn.includes(search))

  if (loading) return <Spinner />

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Проекты</h1>
        <span className="page-subtitle">{projects.length} проектов</span>
      </div>

      <div className="filters">
        <FilterInput withIcon placeholder="Поиск по проекту, клиенту, ИНН..." value={search} onChange={e => setSearch(e.target.value)} />
      </div>

      <Card>
        <div className="table-wrapper desktop-only">
          <table className="data-table">
            <thead>
              <tr><th>Проект</th><th>Клиент</th><th>ИНН</th><th>Статус</th><th>Оплат</th><th>Сумма</th><th>Закрыто / Открыто</th><th>Документы</th><th></th></tr>
            </thead>
            <tbody>
              {filtered.length === 0 && <tr><td colSpan={9}><EmptyState icon={<FolderOpen size={36} />} message="Проекты не найдены" /></td></tr>}
              {filtered.map(p => (
                <tr key={p.id}>
                  <td style={{ fontWeight: 500 }}>{p.name}</td>
                  <td>{p.client_name}</td>
                  <td className="text-muted text-xs">{p.inn || '—'}</td>
                  <td><Badge variant={p.status}>{PROJECT_STATUS_LABELS[p.status] || p.status}</Badge></td>
                  <td style={{ textAlign: 'center' }}>{p.payments_count}</td>
                  <td><span className="amount">{formatMoney(p.total_amount)}</span></td>
                  <td style={{ textAlign: 'center' }}><span style={{ color: 'var(--success)', fontWeight: 600 }}>{p.closed_acts_count}</span> / <span style={{ color: p.open_acts_count > 0 ? 'var(--warning)' : 'var(--text-muted)' }}>{p.open_acts_count}</span></td>
                  <td><Badge variant={p.doc_status}>{DOC_STATUS_LABELS[p.doc_status]}</Badge></td>
                  <td><Link to={`/payments?project_id=${p.id}`} className="btn btn-ghost btn-sm">Оплаты <ArrowRight size={12} /></Link></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="mobile-only project-cards">
          {filtered.length === 0 && <EmptyState icon={<FolderOpen size={36} />} message="Проекты не найдены" />}
          {filtered.map(p => (
            <div className="project-card" key={p.id}>
              <div className="project-card-top">
                <Badge variant={p.status}>{PROJECT_STATUS_LABELS[p.status] || p.status}</Badge>
                <Badge variant={p.doc_status}>{DOC_STATUS_LABELS[p.doc_status]}</Badge>
              </div>
              <div className="project-card-name">{p.name}</div>
              <div className="project-card-row"><span className="project-card-label">Клиент</span><span>{p.client_name}</span></div>
              {p.inn && <div className="project-card-row"><span className="project-card-label">ИНН</span><span>{p.inn}</span></div>}
              <div className="project-card-row"><span className="project-card-label">Сумма</span><span className="amount">{formatMoney(p.total_amount)}</span></div>
              <div className="project-card-row"><span className="project-card-label">Закрыто / Открыто</span><span><span style={{ color: 'var(--success)', fontWeight: 600 }}>{p.closed_acts_count}</span> / <span style={{ color: p.open_acts_count > 0 ? 'var(--warning)' : 'var(--text-muted)' }}>{p.open_acts_count}</span></span></div>
              <Link to={`/payments?project_id=${p.id}`} className="btn btn-primary btn-sm project-card-action">Оплаты <ArrowRight size={12} /></Link>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
