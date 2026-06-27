import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { dashboardApi, type DashboardSummary } from '../../entities/dashboard'
import { projectApi, type ProjectSummary, PROJECT_STATUS_LABELS, DOC_STATUS_LABELS } from '../../entities/project'
import { StatCard, Badge, Card, Spinner } from '../../shared/ui'
import { formatMoney } from '../../shared/lib'

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [projects, setProjects] = useState<ProjectSummary[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([dashboardApi.getSummary(), projectApi.getAll()])
      .then(([s, p]) => { setSummary(s); setProjects(p) })
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Spinner />

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Дашборд</h1>
        <span className="page-subtitle">Обзор оплат и документооборота</span>
      </div>

      {summary && (
        <div className="summary-grid">
          <StatCard variant="primary" label="Общая сумма оплат" value={formatMoney(summary.total_payments)} sub={`${summary.total_payments_count} оплат · ${summary.total_projects} проектов`} />
          <StatCard variant="success" label="Закрыто актами" value={formatMoney(summary.closed_acts_amount)} sub={`${summary.total_payments > 0 ? Math.round(summary.closed_acts_amount / summary.total_payments * 100) : 0}% от общей суммы`} />
          <StatCard variant="warning" label="Открытые акты" value={formatMoney(summary.open_acts_amount)} sub="Требуют закрытия документов" />
          <StatCard label="Акт не отправлен" value={summary.not_sent_count} sub="оплат без отправленного акта" />
          <StatCard label="Ожидают подписи" value={<span style={{ color: 'var(--warning)' }}>{summary.waiting_signature_count}</span>} sub="акт отправлен, не подписан" />
          <StatCard variant="danger" label="Требуют внимания" value={summary.needs_attention_count} sub="просрочено / долго без ответа" />
        </div>
      )}

      <Card header={<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>Проекты <Link to="/projects" className="btn btn-ghost btn-sm">Все проекты <ArrowRight size={14} /></Link></div>}>
        {/* Desktop: table */}
        <div className="table-wrapper desktop-only">
          <table className="data-table">
            <thead>
              <tr>
                <th>Проект</th>
                <th>Клиент / ИНН</th>
                <th>Статус</th>
                <th>Оплат</th>
                <th>Сумма</th>
                <th>Закрыто</th>
                <th>Документы</th>
              </tr>
            </thead>
            <tbody>
              {projects.map(p => (
                <tr key={p.id}>
                  <td><Link to={`/payments?project_id=${p.id}`} style={{ color: 'var(--primary)', fontWeight: 500 }}>{p.name}</Link></td>
                  <td><div>{p.client_name}</div>{p.inn && <div className="text-muted text-xs">ИНН {p.inn}</div>}</td>
                  <td><Badge variant={p.status}>{PROJECT_STATUS_LABELS[p.status] || p.status}</Badge></td>
                  <td style={{ textAlign: 'center' }}>{p.payments_count}</td>
                  <td><span className="amount">{formatMoney(p.total_amount)}</span></td>
                  <td style={{ textAlign: 'center' }}><span style={{ color: 'var(--success)', fontWeight: 600 }}>{p.closed_acts_count}</span> / {p.payments_count}</td>
                  <td><Badge variant={p.doc_status}>{DOC_STATUS_LABELS[p.doc_status]}</Badge></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Mobile: project cards */}
        <div className="mobile-only project-cards">
          {projects.map(p => (
            <Link to={`/payments?project_id=${p.id}`} key={p.id} className="project-card">
              <div className="project-card-top">
                <Badge variant={p.status}>{PROJECT_STATUS_LABELS[p.status] || p.status}</Badge>
                <Badge variant={p.doc_status}>{DOC_STATUS_LABELS[p.doc_status]}</Badge>
              </div>
              <div className="project-card-name">{p.name}</div>
              <div className="project-card-row">
                <span className="project-card-label">Клиент</span>
                <span>{p.client_name}</span>
              </div>
              {p.inn && (
                <div className="project-card-row">
                  <span className="project-card-label">ИНН</span>
                  <span>{p.inn}</span>
                </div>
              )}
              <div className="project-card-row">
                <span className="project-card-label">Сумма</span>
                <span className="amount">{formatMoney(p.total_amount)}</span>
              </div>
              <div className="project-card-row">
                <span className="project-card-label">Оплат</span>
                <span>{p.payments_count}</span>
              </div>
              <div className="project-card-row">
                <span className="project-card-label">Закрыто / Всего</span>
                <span>
                  <span style={{ color: 'var(--success)', fontWeight: 600 }}>{p.closed_acts_count}</span>
                  {' / '}
                  {p.payments_count}
                </span>
              </div>
            </Link>
          ))}
        </div>
      </Card>
    </div>
  )
}
