import React from "react";
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, ArrowRight, History, X, ChevronDown } from 'lucide-react'
import { ordersApi, type Order, type StatusHistory } from '../../api/endpoints'
import { PageHeader, Card, StatusBadge, Button, Select, Skeleton, Empty } from '../../components/ui'
import { fmt } from '../../lib/utils'

const ALL_STATUSES = ['', 'received', 'approved', 'separated', 'in_route', 'delivered', 'cancelled', 'rescheduled']

const NEXT_STATUS: Record<string, { to: string; label: string; danger?: boolean }[]> = {
  received:    [{ to: 'approved',   label: 'Aprovar'   }, { to: 'cancelled',  label: 'Cancelar', danger: true }],
  approved:    [{ to: 'separated',  label: 'Separar'   }, { to: 'cancelled',  label: 'Cancelar', danger: true }],
  separated:   [{ to: 'in_route',   label: 'Enviar'    }, { to: 'cancelled',  label: 'Cancelar', danger: true }],
  in_route:    [{ to: 'delivered',  label: 'Entregar'  }, { to: 'rescheduled',label: 'Reagendar' }],
  rescheduled: [{ to: 'in_route',   label: 'Reenviar'  }, { to: 'cancelled',  label: 'Cancelar', danger: true }],
}

export default function OrdersPage() {
  const qc = useQueryClient()
  const [statusFilter, setStatusFilter] = useState('')
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [histOpen, setHistOpen] = useState(false)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['orders', statusFilter],
    queryFn: () => {
      const p: Record<string, string> = { limit: '50' }
      if (statusFilter) p.status = statusFilter
      return ordersApi.list(p)
    },
  })

  const { data: history, isLoading: histLoading } = useQuery({
    queryKey: ['order-history', selectedOrder?.id],
    queryFn: () => ordersApi.history(selectedOrder!.id),
    enabled: !!selectedOrder && histOpen,
  })

  const transition = useMutation({
    mutationFn: ({ id, status, reason }: { id: string; status: string; reason?: string }) =>
      ordersApi.transition(id, status, { reason }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['orders'] })
      qc.invalidateQueries({ queryKey: ['order-history', selectedOrder?.id] })
      qc.invalidateQueries({ queryKey: ['kpis'] })
    },
  })

  function handleTransition(order: Order, to: string, needsReason: boolean) {
    let reason: string | undefined
    if (needsReason) {
      const r = window.prompt(`Motivo (${to}):`)
      if (r === null) return
      reason = r
    }
    transition.mutate({ id: order.id, status: to, reason })
  }

  function openHistory(order: Order) {
    setSelectedOrder(order)
    setHistOpen(true)
  }

  const orders = data?.data ?? []

  return (
    <div className="flex h-full" style={{ background: 'var(--bg)' }}>
      {/* ── Main ──────────────────────────────────────────────────────── */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <PageHeader
          title="Pedidos"
          subtitle={`${data?.total ?? 0} pedido${(data?.total ?? 0) !== 1 ? 's' : ''} encontrado${(data?.total ?? 0) !== 1 ? 's' : ''}`}
          actions={
            <Button size="sm" onClick={() => refetch()}>
              <RefreshCw size={12} /> Atualizar
            </Button>
          }
        />

        <div className="flex-1 overflow-y-auto px-7 py-5">
          {/* Filtros por status */}
          <div className="flex flex-wrap gap-2 mb-5">
            {ALL_STATUSES.map(s => (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className="px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-150"
                style={{
                  background:  statusFilter === s ? 'rgba(245,158,11,.14)' : 'var(--surface-2)',
                  border:      statusFilter === s ? '1px solid rgba(245,158,11,.3)' : '1px solid var(--border-subtle)',
                  color:       statusFilter === s ? 'var(--accent)' : 'var(--text-3)',
                  cursor: 'pointer',
                }}
              >
                {s === '' ? 'Todos' : <StatusBadge status={s} />}
              </button>
            ))}
          </div>

          {/* Tabela */}
          <Card className="overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm" style={{ borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    {['ID', 'Status', 'Cliente', 'Produto', 'Qtd', 'Motorista', 'Criado', 'Ações'].map(h => (
                      <th
                        key={h}
                        className="section-label px-5 py-3 text-left font-medium whitespace-nowrap"
                      >{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {isLoading
                    ? Array.from({ length: 6 }).map((_, i) => (
                        <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          {[60, 80, 120, 60, 40, 80, 100, 120].map((w, j) => (
                            <td key={j} className="px-5 py-3.5">
                              <Skeleton className="h-3.5" style={{ width: w }} />
                            </td>
                          ))}
                        </tr>
                      ))
                    : orders.length === 0
                      ? <tr><td colSpan={8}><Empty message="Nenhum pedido encontrado" /></td></tr>
                      : orders.map(order => (
                          <tr
                            key={order.id}
                            className="transition-colors duration-100"
                            style={{ borderBottom: '1px solid var(--border-subtle)' }}
                            onMouseEnter={e => e.currentTarget.style.background = 'var(--surface-2)'}
                            onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                          >
                            <td className="px-5 py-3.5 font-mono text-xs" style={{ color: 'var(--text-3)' }}>
                              {order.id.slice(0, 8)}…
                            </td>
                            <td className="px-5 py-3.5">
                              <StatusBadge status={order.status} />
                            </td>
                            <td className="px-5 py-3.5 text-xs font-mono" style={{ color: 'var(--text-2)' }}>
                              {order.client_id.slice(0, 8)}
                            </td>
                            <td className="px-5 py-3.5 text-xs font-mono" style={{ color: 'var(--text-2)' }}>
                              {order.product_id.slice(0, 8)}
                            </td>
                            <td className="px-5 py-3.5 text-sm" style={{ color: 'var(--text-2)' }}>
                              {order.quantity}
                            </td>
                            <td className="px-5 py-3.5 text-xs" style={{ color: 'var(--text-3)' }}>
                              {order.driver_id ? order.driver_id.slice(0, 8) : '—'}
                            </td>
                            <td className="px-5 py-3.5 font-mono text-xs whitespace-nowrap" style={{ color: 'var(--text-3)' }}>
                              {fmt.datetime(order.created_at)}
                            </td>
                            <td className="px-5 py-3.5">
                              <div className="flex items-center gap-1.5">
                                {(NEXT_STATUS[order.status] ?? []).map(({ to, label, danger }) => (
                                  <button
                                    key={to}
                                    onClick={() => handleTransition(order, to, danger || to === 'rescheduled')}
                                    disabled={transition.isPending}
                                    className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium transition-all duration-150"
                                    style={{
                                      background: danger ? 'rgba(239,68,68,.1)' : 'rgba(245,158,11,.1)',
                                      border:     danger ? '1px solid rgba(239,68,68,.2)' : '1px solid rgba(245,158,11,.2)',
                                      color:      danger ? '#ef4444' : 'var(--accent)',
                                      cursor:     'pointer',
                                    }}
                                  >
                                    <ArrowRight size={10} />
                                    {label}
                                  </button>
                                ))}
                                <button
                                  onClick={() => openHistory(order)}
                                  className="p-1.5 rounded-md transition-all duration-150"
                                  style={{ background: 'var(--surface-3)', border: '1px solid var(--border-subtle)', color: 'var(--text-3)', cursor: 'pointer' }}
                                  title="Ver histórico"
                                >
                                  <History size={13} />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      </div>

      {/* ── Painel de histórico (slide-in) ─────────────────────────── */}
      {histOpen && selectedOrder && (
        <div
          className="flex flex-col flex-shrink-0 overflow-y-auto animate-slide-up"
          style={{
            width: 340,
            borderLeft: '1px solid var(--border-subtle)',
            background: 'var(--surface-1)',
          }}
        >
          {/* Header painel */}
          <div
            className="flex items-center justify-between px-5 py-4 sticky top-0 z-10"
            style={{ background: 'var(--surface-1)', borderBottom: '1px solid var(--border-subtle)' }}
          >
            <div>
              <p className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Histórico</p>
              <p className="font-mono text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>
                {selectedOrder.id.slice(0, 20)}…
              </p>
            </div>
            <button
              onClick={() => setHistOpen(false)}
              className="p-1.5 rounded-lg transition-colors duration-150"
              style={{ background: 'var(--surface-3)', border: '1px solid var(--border-subtle)', color: 'var(--text-3)', cursor: 'pointer' }}
            >
              <X size={14} />
            </button>
          </div>

          <div className="px-5 py-4 space-y-5">
            {/* Info do pedido */}
            <div className="rounded-xl p-4 space-y-2.5" style={{ background: 'var(--surface-2)', border: '1px solid var(--border-subtle)' }}>
              <Row label="Status">    <StatusBadge status={selectedOrder.status} /></Row>
              <Row label="Qtd">       <span className="text-xs font-mono" style={{ color: 'var(--text-2)' }}>{selectedOrder.quantity}</span></Row>
              <Row label="Criado">    <span className="text-xs font-mono" style={{ color: 'var(--text-3)' }}>{fmt.datetime(selectedOrder.created_at)}</span></Row>
              {selectedOrder.delivered_at && (
                <Row label="Entregue"><span className="text-xs font-mono" style={{ color: 'var(--success)' }}>{fmt.datetime(selectedOrder.delivered_at)}</span></Row>
              )}
              {selectedOrder.notes && (
                <Row label="Obs"><span className="text-xs" style={{ color: 'var(--text-3)' }}>{selectedOrder.notes}</span></Row>
              )}
            </div>

            {/* Timeline */}
            <div>
              <p className="section-label mb-3">LINHA DO TEMPO</p>
              {histLoading
                ? Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex gap-3 mb-4">
                      <Skeleton className="w-2 h-2 rounded-full mt-1 flex-shrink-0" />
                      <div className="flex-1 space-y-1.5">
                        <Skeleton className="h-4 w-3/4" />
                        <Skeleton className="h-3 w-1/2" />
                      </div>
                    </div>
                  ))
                : !history?.length
                  ? <p className="text-xs" style={{ color: 'var(--text-3)' }}>Sem histórico registrado</p>
                  : (
                    <div className="relative">
                      {/* Linha vertical */}
                      <div
                        className="absolute left-[3px] top-2"
                        style={{ width: 1, height: 'calc(100% - 16px)', background: 'var(--border-subtle)' }}
                      />
                      <div className="space-y-4">
                        {history.map((h, i) => (
                          <TimelineItem key={h.id} item={h} isLast={i === history.length - 1} />
                        ))}
                      </div>
                    </div>
                  )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/* ── Sub-componentes ─────────────────────────────────────────────────────── */
function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-2">
      <span className="text-xs flex-shrink-0" style={{ color: 'var(--text-3)' }}>{label}</span>
      {children}
    </div>
  )
}

function TimelineItem({ item, isLast }: { item: StatusHistory; isLast: boolean }) {
  const dot =
    item.to_status === 'delivered' ? '#22c55e' :
    item.to_status === 'cancelled' ? '#ef4444' :
    item.to_status === 'rescheduled' ? '#ff6b2b' : 'var(--accent)'

  return (
    <div className="flex gap-3 relative">
      <div
        className="w-2 h-2 rounded-full flex-shrink-0 mt-1 z-10"
        style={{ background: dot, boxShadow: `0 0 0 3px color-mix(in srgb, ${dot} 20%, transparent)` }}
      />
      <div className={isLast ? '' : ''}>
        <div className="flex items-center gap-1.5 flex-wrap">
          {item.from_status && (
            <>
              <StatusBadge status={item.from_status} />
              <ArrowRight size={9} style={{ color: 'var(--text-3)' }} />
            </>
          )}
          <StatusBadge status={item.to_status} />
        </div>
        {item.reason && (
          <p className="text-xs mt-1" style={{ color: 'var(--text-3)' }}>{item.reason}</p>
        )}
        <p className="font-mono text-xs mt-1" style={{ color: 'var(--text-3)', fontSize: 10 }}>
          {fmt.datetime(item.created_at)}
        </p>
      </div>
    </div>
  )
}