import { useEffect, useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { CreditCard, RotateCcw, Pencil } from 'lucide-react'
import { paymentApi, type Payment, type PaymentFilter, getActStatus, ACT_STATUS_LABELS, ActStatusBadge, PaymentCard } from '../../entities/payment'
import { projectApi, type ProjectSummary } from '../../entities/project'
import { clientApi, type Client } from '../../entities/client'
import { UpdateActModal } from '../../features/update-act'
import { Card, Badge, Spinner, EmptyState, FilterInput, FilterSelect, Button } from '../../shared/ui'
import { formatMoney, formatDate } from '../../shared/lib'
import type { ActStatus } from '../../entities/payment/model'

const SERVICE_STAGES = ['Разработка', 'Дизайн', 'SEO', 'Реклама', 'Контент', 'Сопровождение']

export function PaymentsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [payments, setPayments] = useState<Payment[]>([])
  const [projects, setProjects] = useState<ProjectSummary[]>([])
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState(true)
  const [editPayment, setEditPayment] = useState<Payment | null>(null)

  const [filter, setFilter] = useState<PaymentFilter>({
    project_id: searchParams.get('project_id') ? Number(searchParams.get('project_id')) : undefined,
    legal_entity_id: undefined,
    act_status: (searchParams.get('act_status') as ActStatus) || '',
    service_stage: '',
    search: '',
    date_from: '',
    date_to: '',
  })

  const loadPayments = useCallback((f: PaymentFilter) => {
    setLoading(true)
    paymentApi.getAll(f).then(setPayments).finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    Promise.all([projectApi.getAll(), clientApi.getAll()]).then(([p, c]) => { setProjects(p); setClients(c) })
    loadPayments(filter)
  }, []) // eslint-disable-line

  function applyFilter(updates: Partial<PaymentFilter>) {
    const next = { ...filter, ...updates }
    setFilter(next)
    loadPayments(next)
    const params = new URLSearchParams()
    if (next.project_id) params.set('project_id', String(next.project_id))
    if (next.act_status) params.set('act_status', next.act_status)
    setSearchParams(params)
  }

  function resetFilters() {
    const empty: PaymentFilter = { project_id: undefined, legal_entity_id: undefined, act_status: '', service_stage: '', search: '', date_from: '', date_to: '' }
    setFilter(empty)
    loadPayments(empty)
    setSearchParams({})
  }

  const totalShown = payments.reduce((s, p) => s + p.amount, 0)
  const closedShown = payments.filter(p => getActStatus(p) === 'closed').reduce((s, p) => s + p.amount, 0)

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Оплаты</h1>
        <div className="page-header-stats">
          <span>Показано: <strong>{payments.length}</strong></span>
          <span>Сумма: <strong className="text-primary">{formatMoney(totalShown)}</strong></span>
          <span>Закрыто: <strong className="text-success">{formatMoney(closedShown)}</strong></span>
        </div>
      </div>

      <Card className="filters-card" bodyClassName="card-body">
        <div className="filters">
          <FilterInput withIcon placeholder="Поиск по назначению или клиенту..." value={filter.search ?? ''} onChange={e => applyFilter({ search: e.target.value })} />
          <FilterSelect value={String(filter.project_id ?? '')} onChange={e => applyFilter({ project_id: e.target.value ? Number(e.target.value) : undefined })} placeholder="Все проекты" options={projects.map(p => ({ value: String(p.id), label: p.name }))} />
          <FilterSelect value={String(filter.legal_entity_id ?? '')} onChange={e => applyFilter({ legal_entity_id: e.target.value ? Number(e.target.value) : undefined })} placeholder="Все юрлица" options={clients.map(c => ({ value: String(c.id), label: c.name }))} />
          <FilterSelect value={filter.act_status ?? ''} onChange={e => applyFilter({ act_status: e.target.value as ActStatus | '' })} options={[{ value: '', label: 'Все статусы' }, { value: 'not_sent', label: 'Не отправлен' }, { value: 'waiting_signature', label: 'Ожидает подписи' }, { value: 'closed', label: 'Закрыт' }, { value: 'needs_attention', label: 'Требует внимания' }]} />
          <FilterSelect value={filter.service_stage ?? ''} onChange={e => applyFilter({ service_stage: e.target.value })} placeholder="Все этапы" options={SERVICE_STAGES.map(s => ({ value: s, label: s }))} />
          <input type="date" className="filter-input" value={filter.date_from ?? ''} onChange={e => applyFilter({ date_from: e.target.value })} />
          <input type="date" className="filter-input" value={filter.date_to ?? ''} onChange={e => applyFilter({ date_to: e.target.value })} />
          <Button size="sm" onClick={resetFilters}><RotateCcw size={12} /> Сбросить</Button>
        </div>
      </Card>

      <Card>
        <div className="table-wrapper desktop-only">
          <table className="data-table">
            <thead><tr><th>Дата</th><th>Юрлицо</th><th>Проект</th><th>Этап</th><th>Сумма</th><th>Назначение</th><th>Счёт</th><th>Статус акта</th><th>Комментарий</th><th></th></tr></thead>
            <tbody>
              {loading && <tr><td colSpan={10}><Spinner /></td></tr>}
              {!loading && payments.length === 0 && <tr><td colSpan={10}><EmptyState icon={<CreditCard size={36} />} message="Оплаты не найдены" /></td></tr>}
              {!loading && payments.map(payment => {
                const status = payment.act?.status || getActStatus(payment)
                return (
                  <tr key={payment.id}>
                    <td className="text-nowrap">{formatDate(payment.payment_date)}</td>
                    <td><div style={{ fontWeight: 500 }}>{payment.legal_entity?.name ?? '—'}</div>{payment.legal_entity?.inn && <div className="text-muted text-xs">ИНН {payment.legal_entity.inn}</div>}</td>
                    <td style={{ maxWidth: 180 }}><div style={{ fontWeight: 500 }}>{payment.project?.name ?? '—'}</div></td>
                    <td><Badge variant="active">{payment.service_stage}</Badge></td>
                    <td><span className="amount amount-positive">{formatMoney(payment.amount)}</span></td>
                    <td className="text-muted text-xs" style={{ maxWidth: 200 }}>{payment.payment_purpose}</td>
                    <td className="text-xs">{payment.invoice_number && <div>Сч. №{payment.invoice_number}</div>}{payment.contract_number && <div className="text-muted">{payment.contract_number}</div>}</td>
                    <td><ActStatusBadge status={status} /></td>
                    <td className="text-muted text-xs" style={{ maxWidth: 140 }}>{payment.act?.manager_comment || '—'}</td>
                    <td><Button size="sm" onClick={() => setEditPayment(payment)} title="Статус акта"><Pencil size={14} /></Button></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <div className="mobile-only payment-cards">
          {loading && <Spinner />}
          {!loading && payments.length === 0 && <EmptyState icon={<CreditCard size={36} />} message="Оплаты не найдены" />}
          {!loading && payments.map(p => <PaymentCard key={p.id} payment={p} onEdit={setEditPayment} />)}
        </div>
      </Card>

      {editPayment && (
        <UpdateActModal payment={editPayment} onClose={() => setEditPayment(null)} onSaved={updated => { setPayments(prev => prev.map(p => p.id === updated.id ? updated : p)); setEditPayment(null) }} />
      )}
    </div>
  )
}
