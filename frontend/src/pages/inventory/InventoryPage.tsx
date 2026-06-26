// import React from "react";
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { RefreshCw, AlertTriangle, PackagePlus, Warehouse, X, ChevronRight } from 'lucide-react'
import { inventoryApi, type Deposit, type InventoryItem } from '../../api/endpoints'
import { PageHeader, Button, Input, Skeleton, Empty, ErrorBanner } from '../../components/ui'
import { fmt } from '../../lib/utils'

// ── Schema recebimento ────────────────────────────────────────────────────────
const receiveSchema = z.object({
  product_id: z.string().min(1, 'Produto obrigatório'),
  quantity:   z.coerce.number().min(1, 'Quantidade mínima: 1'),
})
type ReceiveForm = z.infer<typeof receiveSchema>

const PRODUCT_OPTIONS = [
  { id: 'prod-p13',  label: 'P13  — 13 kg'  },
  { id: 'prod-p20',  label: 'P20  — 20 kg'  },
  { id: 'prod-p45',  label: 'P45  — 45 kg'  },
  { id: 'prod-p190', label: 'P190 — 190 kg' },
]

export default function InventoryPage() {
  const qc = useQueryClient()
  const [activeDeposit, setActiveDeposit] = useState<Deposit | null>(null)
  const [receiveOpen, setReceiveOpen]     = useState(false)
  const [formError, setFormError]         = useState('')

  // ── Queries ──────────────────────────────────────────────────────────────
  const { data: deposits = [], isLoading: depLoading, refetch } = useQuery({
    queryKey: ['deposits'],
    queryFn:  inventoryApi.deposits,
    onSuccess: (d: Deposit[]) => { if (!activeDeposit && d.length) setActiveDeposit(d[0]) },
  } as any)

  const { data: items = [], isLoading: itemsLoading } = useQuery({
    queryKey: ['inventory-items', activeDeposit?.id],
    queryFn:  () => inventoryApi.items(activeDeposit!.id),
    enabled:  !!activeDeposit,
  })

  const { data: lowStock = [] } = useQuery({
    queryKey: ['low-stock'],
    queryFn:  inventoryApi.lowStock,
  })

  // ── Mutation recebimento ─────────────────────────────────────────────────
  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<ReceiveForm>({
    resolver: zodResolver(receiveSchema) as any,
    defaultValues: { product_id: 'prod-p13', quantity: 100 },
  })

  const receive = useMutation({
    mutationFn: ({ product_id, quantity }: ReceiveForm) =>
      inventoryApi.receive(activeDeposit!.id, product_id, quantity),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-items'] })
      qc.invalidateQueries({ queryKey: ['low-stock'] })
      qc.invalidateQueries({ queryKey: ['kpis'] })
      setReceiveOpen(false)
      reset()
      setFormError('')
    },
    onError: (e) => setFormError((e as Error).message),
  })

  const allItems = items as InventoryItem[]
  const lowItems = lowStock as InventoryItem[]

  // ── Bar de uso ───────────────────────────────────────────────────────────
  function StockBar({ item }: { item: InventoryItem }) {
    const pct = item.quantity > 0 ? Math.round((item.reserved / item.quantity) * 100) : 0
    const avail = item.quantity - item.reserved
    const isLow = avail < 10

    return (
      <div
        className="rounded-xl p-4 transition-all duration-150"
        style={{
          background: 'var(--surface-2)',
          border: `1px solid ${isLow ? 'rgba(239,68,68,.25)' : 'var(--border-subtle)'}`,
        }}
        onMouseEnter={e => { e.currentTarget.style.borderColor = isLow ? 'rgba(239,68,68,.4)' : 'var(--border)' }}
        onMouseLeave={e => { e.currentTarget.style.borderColor = isLow ? 'rgba(239,68,68,.25)' : 'var(--border-subtle)' }}
      >
        <div className="flex items-start justify-between mb-3">
          <div>
            <p className="text-sm font-semibold" style={{ color: 'var(--text)' }}>
              {item.product_id.replace('prod-', '').toUpperCase()}
            </p>
            <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>
              {fmt.number(item.quantity)} unidades totais
            </p>
          </div>
          {isLow && (
            <span
              className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium"
              style={{ background: 'rgba(239,68,68,.12)', color: '#ef4444' }}
            >
              <AlertTriangle size={10} />
              Baixo
            </span>
          )}
        </div>

        {/* Barra de estoque */}
        <div className="w-full rounded-full overflow-hidden mb-2" style={{ height: 6, background: 'var(--surface-4)' }}>
          <div
            className="h-full rounded-full transition-all duration-500"
            style={{
              width: `${100 - pct}%`,
              background: isLow
                ? 'linear-gradient(90deg, #ef4444, #ff6b2b)'
                : avail / item.quantity > 0.5
                  ? 'linear-gradient(90deg, #22c55e, #34d399)'
                  : 'linear-gradient(90deg, #f59e0b, #fbbf24)',
            }}
          />
        </div>

        <div className="flex items-center justify-between text-xs font-mono">
          <span style={{ color: 'var(--success)' }}>
            {fmt.number(avail)} disponível
          </span>
          <span style={{ color: 'var(--text-3)' }}>
            {fmt.number(item.reserved)} reservado ({pct}%)
          </span>
        </div>
      </div>
    )
  }

  return (
    <div style={{ background: 'var(--bg)', minHeight: '100%' }}>
      <PageHeader
        title="Estoque"
        subtitle={`${allItems.length} produto${allItems.length !== 1 ? 's' : ''} · ${lowItems.length} alerta${lowItems.length !== 1 ? 's' : ''} de baixo estoque`}
        actions={
          <div className="flex items-center gap-2">
            {activeDeposit && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => setReceiveOpen(true)}
              >
                <PackagePlus size={13} /> Entrada
              </Button>
            )}
            <Button size="sm" onClick={() => refetch()}>
              <RefreshCw size={12} /> Atualizar
            </Button>
          </div>
        }
      />

      <div className="px-7 py-5 flex gap-5">

        {/* ── Sidebar depósitos ──────────────────────────────────────── */}
        <div className="flex-shrink-0" style={{ width: 220 }}>
          <p className="section-label mb-3 px-1">DEPÓSITOS</p>
          <div className="space-y-1">
            {depLoading
              ? Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="rounded-lg p-3" style={{ background: 'var(--surface-1)', border: '1px solid var(--border-subtle)' }}>
                    <Skeleton className="h-3.5 w-3/4 mb-2" />
                    <Skeleton className="h-3 w-1/2" />
                  </div>
                ))
              : (deposits as Deposit[]).map(dep => {
                  const isActive = activeDeposit?.id === dep.id
                  const depLow = lowItems.filter(i => i.deposit_id === dep.id).length
                  return (
                    <button
                      key={dep.id}
                      onClick={() => setActiveDeposit(dep)}
                      className="w-full text-left rounded-xl p-3.5 transition-all duration-150"
                      style={{
                        background: isActive ? 'rgba(245,158,11,.1)'      : 'var(--surface-1)',
                        border:     isActive ? '1px solid rgba(245,158,11,.3)' : '1px solid var(--border-subtle)',
                        cursor: 'pointer',
                      }}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <div className="flex items-center gap-2">
                          <Warehouse size={13} style={{ color: isActive ? 'var(--accent)' : 'var(--text-3)', flexShrink: 0 }} />
                          <span className="text-xs font-semibold" style={{ color: isActive ? 'var(--accent)' : 'var(--text)' }}>
                            {dep.name.replace('Depósito ', '')}
                          </span>
                        </div>
                        {isActive && <ChevronRight size={11} style={{ color: 'var(--accent)' }} />}
                      </div>
                      <p className="text-xs pl-5" style={{ color: 'var(--text-3)' }}>{dep.city}</p>
                      {depLow > 0 && (
                        <div className="flex items-center gap-1 pl-5 mt-1.5">
                          <AlertTriangle size={10} style={{ color: '#ef4444' }} />
                          <span className="text-xs" style={{ color: '#ef4444' }}>{depLow} alerta{depLow > 1 ? 's' : ''}</span>
                        </div>
                      )}
                    </button>
                  )
                })}
          </div>

          {/* Alertas globais */}
          {lowItems.length > 0 && (
            <div className="mt-4">
              <p className="section-label mb-2 px-1">ALERTAS</p>
              <div
                className="rounded-xl p-3"
                style={{ background: 'rgba(239,68,68,.06)', border: '1px solid rgba(239,68,68,.2)' }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <AlertTriangle size={13} style={{ color: '#ef4444' }} />
                  <span className="text-xs font-medium" style={{ color: '#ef4444' }}>
                    {lowItems.length} item{lowItems.length > 1 ? 's' : ''} com estoque baixo
                  </span>
                </div>
                {lowItems.slice(0, 4).map(i => (
                  <p key={i.id} className="text-xs font-mono pl-5" style={{ color: 'var(--text-3)' }}>
                    {i.product_id.replace('prod-', '').toUpperCase()} — {i.quantity - i.reserved} disponível
                  </p>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* ── Grid de itens do depósito ──────────────────────────────── */}
        <div className="flex-1">
          {activeDeposit && (
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-base font-semibold" style={{ color: 'var(--text)' }}>
                  {activeDeposit.name}
                </h2>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>{activeDeposit.city}</p>
              </div>
            </div>
          )}

          {itemsLoading
            ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="rounded-xl p-4" style={{ background: 'var(--surface-2)', border: '1px solid var(--border-subtle)' }}>
                    <Skeleton className="h-4 w-1/3 mb-3" />
                    <Skeleton className="h-3 w-2/3 mb-3" />
                    <Skeleton className="h-1.5 w-full mb-2 rounded-full" />
                    <div className="flex justify-between">
                      <Skeleton className="h-3 w-1/3" />
                      <Skeleton className="h-3 w-1/3" />
                    </div>
                  </div>
                ))}
              </div>
            )
            : allItems.length === 0
              ? <Empty message="Nenhum produto neste depósito" />
              : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {allItems.map(item => <StockBar key={item.id} item={item} />)}
                </div>
              )}
        </div>
      </div>

      {/* ── Modal de entrada de estoque ────────────────────────────────── */}
      {receiveOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ background: 'rgba(0,0,0,.6)', backdropFilter: 'blur(4px)' }}
          onClick={e => { if (e.target === e.currentTarget) setReceiveOpen(false) }}
        >
          <div
            className="w-full max-w-md rounded-2xl p-7 animate-slide-up"
            style={{
              background: 'var(--surface-1)',
              border: '1px solid var(--border)',
              boxShadow: '0 24px 80px rgba(0,0,0,.6)',
            }}
          >
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-base font-semibold" style={{ color: 'var(--text)' }}>
                  Entrada de Estoque
                </h2>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>
                  {activeDeposit?.name}
                </p>
              </div>
              <button
                onClick={() => setReceiveOpen(false)}
                className="p-1.5 rounded-lg"
                style={{ background: 'var(--surface-3)', border: '1px solid var(--border-subtle)', color: 'var(--text-3)', cursor: 'pointer' }}
              >
                <X size={14} />
              </button>
            </div>

            <form onSubmit={handleSubmit((d: ReceiveForm) => receive.mutate(d))} className="space-y-4">
              {formError && <ErrorBanner message={formError} />}

              <div className="flex flex-col gap-1.5">
                <label className="section-label">PRODUTO</label>
                <select
                  {...register('product_id')}
                  className="w-full rounded-lg px-3 py-2 text-sm outline-none"
                  style={{
                    background: 'var(--surface-2)',
                    border: '1px solid var(--border)',
                    color: 'var(--text)',
                    cursor: 'pointer',
                  }}
                >
                  {PRODUCT_OPTIONS.map(p => (
                    <option key={p.id} value={p.id}>{p.label}</option>
                  ))}
                </select>
                {errors.product_id && <p className="text-xs" style={{ color: 'var(--danger)' }}>{errors.product_id.message}</p>}
              </div>

              <Input
                {...register('quantity')}
                label="QUANTIDADE"
                type="number"
                min={1}
                placeholder="100"
                error={errors.quantity?.message}
              />

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="ghost"
                  className="flex-1"
                  onClick={() => setReceiveOpen(false)}
                >
                  Cancelar
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  loading={isSubmitting || receive.isPending}
                  className="flex-1"
                >
                  Confirmar Entrada
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}