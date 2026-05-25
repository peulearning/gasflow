import React from "react";
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, CheckCircle, AlertTriangle, Clock, DollarSign, TrendingDown } from 'lucide-react'
import { chargesApi, type Charge } from '../../api/endpoints'
import { PageHeader, Card, StatusBadge, Button, Skeleton, Empty } from '../../components/ui'
import { fmt } from '../../lib/utils'

const TABS = [
  { key: '',        label: 'Todas'        },
  { key: 'pending', label: 'Pendentes'    },
  { key: 'paid',    label: 'Pagas'        },
  { key: 'overdue', label: 'Inadimplentes'},
]

export default function ChargesPage() {
  const qc   = useQueryClient()
  const [tab, setTab]           = useState('')
  const [paying, setPaying]     = useState<string | null>(null)
  const [successId, setSuccessId] = useState<string | null>(null)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['charges', tab],
    queryFn: () =>
      tab === 'overdue'
        ? chargesApi.overdue()
        : chargesApi.list(tab ? { status: tab, limit: '100' } : { limit: '100' }),
  })

  const payMutation = useMutation({
    mutationFn: (id: string) => chargesApi.pay(id),
    onMutate:   (id) => setPaying(id),
    onSuccess:  (_, id) => {
      setSuccessId(id)
      setTimeout(() => setSuccessId(null), 2000)
      qc.invalidateQueries({ queryKey: ['charges'] })
      qc.invalidateQueries({ queryKey: ['kpis'] })
    },
    onSettled: () => setPaying(null),
  })

  const charges = data?.data ?? []
  const total   = data?.total ?? 0

  // ── Resumo financeiro da lista atual ─────────────────────────────────────
  const totalCents    = charges.reduce((s, c) => s + amountCents(c), 0)
  const overdueCents  = charges.filter(c => c.status === 'overdue').reduce((s, c) => s + amountCents(c), 0)
  const paidCents     = charges.filter(c => c.status === 'paid').reduce((s, c) => s + amountCents(c), 0)
  const pendingCents  = charges.filter(c => c.status === 'pending').reduce((s, c) => s + amountCents(c), 0)

  return (
    <div style={{ background: 'var(--bg)', minHeight: '100%' }}>
      <PageHeader
        title="Cobranças"
        subtitle={`${total} cobrança${total !== 1 ? 's' : ''} · ${tab || 'todas'}`}
        actions={
          <Button size="sm" onClick={() => refetch()}>
            <RefreshCw size={12} /> Atualizar
          </Button>
        }
      />

      <div className="px-7 py-5 space-y-5">

        {/* ── Cards de resumo ──────────────────────────────────────────── */}
        {!isLoading && tab === '' && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 animate-fade-in">
            <SummaryCard
              label="Total em Cobranças"
              value={fmt.currency(totalCents)}
              icon={DollarSign}
              color="var(--accent)"
            />
            <SummaryCard
              label="Recebido (Pago)"
              value={fmt.currency(paidCents)}
              icon={CheckCircle}
              color="var(--success)"
            />
            <SummaryCard
              label="A Receber (Pendente)"
              value={fmt.currency(pendingCents)}
              icon={Clock}
              color="#60a5fa"
            />
            <SummaryCard
              label="Inadimplência"
              value={fmt.currency(overdueCents)}
              icon={TrendingDown}
              color="var(--danger)"
            />
          </div>
        )}

        {/* ── Tabs ─────────────────────────────────────────────────────── */}
        <div className="flex gap-2">
          {TABS.map(t => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className="px-4 py-1.5 rounded-lg text-xs font-medium transition-all duration-150"
              style={{
                background: tab === t.key ? 'rgba(245,158,11,.14)' : 'var(--surface-2)',
                border:     tab === t.key ? '1px solid rgba(245,158,11,.3)' : '1px solid var(--border-subtle)',
                color:      tab === t.key ? 'var(--accent)' : 'var(--text-3)',
                cursor: 'pointer',
              }}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* ── Alerta inadimplência ──────────────────────────────────────── */}
        {!isLoading && overdueCents > 0 && tab !== 'paid' && (
          <div
            className="flex items-center gap-3 px-4 py-3 rounded-xl text-sm animate-fade-in"
            style={{
              background: 'rgba(239,68,68,.07)',
              border:     '1px solid rgba(239,68,68,.2)',
            }}
          >
            <AlertTriangle size={15} style={{ color: 'var(--danger)', flexShrink: 0 }} />
            <span style={{ color: 'var(--danger)' }}>
              <strong>{fmt.currency(overdueCents)}</strong> em cobranças vencidas aguardando regularização.
            </span>
          </div>
        )}

        {/* ── Tabela ──────────────────────────────────────────────────── */}
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  {['ID', 'Pedido', 'Cliente', 'Valor', 'Status', 'Vencimento', 'Pago em', 'Ação'].map(h => (
                    <th key={h} className="section-label px-5 py-3 text-left font-medium whitespace-nowrap">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {isLoading
                  ? Array.from({ length: 6 }).map((_, i) => (
                      <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        {[60, 70, 120, 90, 80, 90, 90, 80].map((w, j) => (
                          <td key={j} className="px-5 py-3.5">
                            <Skeleton className="h-3.5" style={{ width: w }} />
                          </td>
                        ))}
                      </tr>
                    ))
                  : charges.length === 0
                    ? <tr><td colSpan={8}><Empty message="Nenhuma cobrança encontrada" /></td></tr>
                    : charges.map(charge => {
                        const isOverdue  = charge.status === 'overdue'
                        const isPaid     = charge.status === 'paid'
                        const isSuccess  = successId === charge.id

                        return (
                          <tr
                            key={charge.id}
                            className="transition-colors duration-100"
                            style={{
                              borderBottom: '1px solid var(--border-subtle)',
                              background: isOverdue ? 'rgba(239,68,68,.03)' : 'transparent',
                            }}
                            onMouseEnter={e => e.currentTarget.style.background = isOverdue ? 'rgba(239,68,68,.06)' : 'var(--surface-2)'}
                            onMouseLeave={e => e.currentTarget.style.background = isOverdue ? 'rgba(239,68,68,.03)' : 'transparent'}
                          >
                            <td className="px-5 py-3.5 font-mono text-xs" style={{ color: 'var(--text-3)' }}>
                              {charge.id.slice(0, 8)}…
                            </td>
                            <td className="px-5 py-3.5 font-mono text-xs" style={{ color: 'var(--text-3)' }}>
                              {charge.order_id.slice(0, 8)}…
                            </td>
                            <td className="px-5 py-3.5 font-mono text-xs" style={{ color: 'var(--text-2)' }}>
                              {charge.client_id.slice(0, 8)}
                            </td>
                            <td className="px-5 py-3.5">
                              <span
                                className="text-sm font-semibold"
                                style={{ color: isPaid ? 'var(--success)' : isOverdue ? 'var(--danger)' : 'var(--text)' }}
                              >
                                {fmt.currency(amountCents(charge))}
                              </span>
                            </td>
                            <td className="px-5 py-3.5">
                              <StatusBadge status={charge.status} />
                            </td>
                            <td className="px-5 py-3.5 font-mono text-xs whitespace-nowrap" style={{ color: isOverdue ? 'var(--danger)' : 'var(--text-3)' }}>
                              {fmt.date(charge.due_date)}
                            </td>
                            <td className="px-5 py-3.5 font-mono text-xs whitespace-nowrap" style={{ color: 'var(--text-3)' }}>
                              {charge.paid_at ? fmt.datetime(charge.paid_at) : '—'}
                            </td>
                            <td className="px-5 py-3.5">
                              {!isPaid ? (
                                <button
                                  onClick={() => payMutation.mutate(charge.id)}
                                  disabled={paying === charge.id}
                                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-150"
                                  style={{
                                    background: isSuccess ? 'rgba(34,197,94,.15)' : 'rgba(34,197,94,.1)',
                                    border:     isSuccess ? '1px solid rgba(34,197,94,.4)' : '1px solid rgba(34,197,94,.2)',
                                    color:      'var(--success)',
                                    cursor:     paying === charge.id ? 'not-allowed' : 'pointer',
                                    opacity:    paying === charge.id ? .6 : 1,
                                  }}
                                >
                                  {isSuccess
                                    ? <><CheckCircle size={11} /> Pago!</>
                                    : paying === charge.id
                                      ? 'Processando…'
                                      : <><CheckCircle size={11} /> Marcar pago</>}
                                </button>
                              ) : (
                                <span className="text-xs" style={{ color: 'var(--text-3)' }}>—</span>
                              )}
                            </td>
                          </tr>
                        )
                      })}
              </tbody>
            </table>
          </div>

          {/* Footer com totais */}
          {!isLoading && charges.length > 0 && (
            <div
              className="flex items-center justify-between px-5 py-3"
              style={{ borderTop: '1px solid var(--border-subtle)', background: 'var(--surface-2)' }}
            >
              <span className="text-xs" style={{ color: 'var(--text-3)' }}>
                {charges.length} registro{charges.length !== 1 ? 's' : ''}
              </span>
              <span className="text-sm font-semibold" style={{ color: 'var(--text)' }}>
                Total: {fmt.currency(totalCents)}
              </span>
            </div>
          )}
        </Card>

      </div>
    </div>
  )
}

/* ── Helpers ─────────────────────────────────────────────────────────────── */
function amountCents(c: Charge): number {
  // O backend pode retornar amount como objeto {cents:N} ou direto como número
  if (typeof c.amount === 'object' && c.amount !== null) return (c.amount as any).cents ?? 0
  return (c.amount as any) ?? 0
}

function SummaryCard({ label, value, icon: Icon, color }: {
  label: string; value: string; icon: React.ElementType; color: string
}) {
  return (
    <div
      className="rounded-xl p-4"
      style={{ background: 'var(--surface-1)', border: '1px solid var(--border-subtle)' }}
    >
      <div
        className="flex items-center justify-center rounded-lg mb-3"
        style={{
          width: 32, height: 32,
          background: `color-mix(in srgb, ${color} 14%, transparent)`,
          border: `1px solid color-mix(in srgb, ${color} 25%, transparent)`,
        }}
      >
        <Icon size={14} style={{ color }} />
      </div>
      <p className="text-lg font-bold" style={{ color: 'var(--text)', letterSpacing: '-.02em' }}>{value}</p>
      <p className="section-label mt-1">{label}</p>
    </div>
  )
}