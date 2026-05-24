import { useQuery } from '@tanstack/react-query'
import { RefreshCw, Truck, TrendingUp, DollarSign, Warehouse, Package, Clock, AlertTriangle, Users } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell, AreaChart, Area } from 'recharts'
import { dashboardApi } from '../../api/endpoints'
import { PageHeader, KPICard, Card, StatusBadge, Skeleton, Empty, Button } from '../../components/ui'
import { fmt } from '../../lib/utils'

const COLORS = ['#f59e0b','#ff6b2b','#60a5fa','#a78bfa','#22c55e','#34d399']

const CUSTOM_TOOLTIP = ({ active, payload, label }: any) => {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-xl px-4 py-3 text-sm shadow-xl" style={{ background: 'var(--surface-3)', border: '1px solid var(--border)', minWidth: 120 }}>
      <p className="mb-1.5 font-medium" style={{ color: 'var(--text-2)' }}>{label}</p>
      {payload.map((p: any, i: number) => (
        <p key={i} style={{ color: p.color ?? 'var(--accent)' }}>{p.name}: <strong>{p.value}</strong></p>
      ))}
    </div>
  )
}

export default function DashboardPage() {
  const { data: kpi, isLoading: kpiLoading, refetch, dataUpdatedAt } = useQuery({
    queryKey: ['kpis'],
    queryFn: () => dashboardApi.kpis(),
  })
  const { data: deliveries, isLoading: delLoading } = useQuery({
    queryKey: ['deliveries-dash'],
    queryFn: () => dashboardApi.deliveries({ limit: '10' }),
  })
  const { data: drivers } = useQuery({
    queryKey: ['driver-perf'],
    queryFn: () => dashboardApi.driverPerf(),
  })
  const { data: topClients } = useQuery({
    queryKey: ['top-clients'],
    queryFn: () => dashboardApi.topClients(),
  })

  const loading = kpiLoading || delLoading
  const rows    = deliveries?.data ?? []
  const drvs    = drivers ?? []
  const tops    = topClients ?? []

  // Dados para gráfico de área a partir das entregas
  const areaData = rows.slice(0, 8).map(d => ({
    name:  fmt.dateShort(d.created_at),
    Qtd:   d.quantity,
  })).reverse()

  const lastUpdate = dataUpdatedAt ? fmt.reltime(new Date(dataUpdatedAt).toISOString()) : null

  return (
    <>
      <PageHeader
        title="Dashboard"
        subtitle={kpi ? `${fmt.date(kpi.period.from)} – ${fmt.date(kpi.period.to)}${lastUpdate ? ` · atualizado ${lastUpdate}` : ''}` : 'Carregando...'}
        actions={
          <div className="flex items-center gap-3">
            {lastUpdate && (
              <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--text-3)' }}>
                <span className="dot-live" /> Ao vivo · 30s
              </div>
            )}
            <Button size="sm" onClick={() => refetch()}>
              <RefreshCw size={12} /> Atualizar
            </Button>
          </div>
        }
      />

      <div className="px-7 py-6 space-y-6">

        {/* ── KPIs Row 1 ──────────────────────────────────────────────── */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {kpiLoading ? Array.from({length:4}).map((_,i) => (
            <div key={i} className="rounded-xl p-5 animate-pulse" style={{ background: 'var(--surface-1)', border: '1px solid var(--border-subtle)' }}>
              <Skeleton className="w-9 h-9 mb-4 rounded-lg" />
              <Skeleton className="h-8 w-3/5 mb-2" />
              <Skeleton className="h-3 w-4/5" />
            </div>
          )) : (<>
            <KPICard icon={Truck}        title="Total de Pedidos"    delay={0}   accent="#60a5fa" value={fmt.number(kpi?.deliveries.total ?? 0)} subtitle={`${kpi?.deliveries.delivered ?? 0} entregues`} />
            <KPICard icon={TrendingUp}   title="Taxa SLA"            delay={60}  accent={(kpi?.deliveries.sla_rate ?? 0) >= 90 ? '#22c55e' : '#f59e0b'} value={fmt.percent(kpi?.deliveries.sla_rate ?? 0)} subtitle={`${kpi?.deliveries.delayed ?? 0} em atraso`} />
            <KPICard icon={DollarSign}   title="Receita do Período"  delay={120} accent="#22c55e" value={fmt.currency(kpi?.billing.revenue_cents ?? 0)} subtitle={`${kpi?.billing.overdue_count ?? 0} inadimplentes`} />
            <KPICard icon={Warehouse}    title="Estoque Disponível"  delay={180} accent={(kpi?.inventory.low_stock_alerts ?? 0) > 0 ? '#ef4444' : 'var(--accent)'} value={fmt.number(kpi?.inventory.available ?? 0)} subtitle={`${kpi?.inventory.low_stock_alerts ?? 0} alertas de baixo estoque`} />
          </>)}
        </div>

        {/* ── KPIs Row 2 ──────────────────────────────────────────────── */}
        {!kpiLoading && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <KPICard icon={Package}       title="Volume Entregue"    delay={0}   accent="#a78bfa" value={fmt.kg(kpi?.deliveries.volume_kg ?? 0)} />
            <KPICard icon={Clock}         title="Reagendamentos"     delay={60}  accent="#ff6b2b" value={kpi?.deliveries.rescheduled ?? 0} subtitle="no período" />
            <KPICard icon={AlertTriangle} title="Inadimplência"      delay={120} accent="#ef4444" value={fmt.currency(kpi?.billing.overdue_amount_cents ?? 0)} subtitle={`${kpi?.billing.overdue_count ?? 0} cobranças`} />
            <KPICard icon={Users}         title="Estoque Reservado"  delay={180} accent="#f59e0b" value={fmt.number(kpi?.inventory.reserved ?? 0)} subtitle={`de ${fmt.number(kpi?.inventory.total_units ?? 0)} totais`} />
          </div>
        )}

        {/* ── Gráficos ────────────────────────────────────────────────── */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">

          {/* Motoristas */}
          <Card className="lg:col-span-2 p-5">
            <div className="flex items-center justify-between mb-5">
              <div>
                <h2 className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Performance por Motorista</h2>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>Entregas no período</p>
              </div>
            </div>
            {drvs.length === 0
              ? <Empty message="Nenhuma entrega com motorista atribuído" />
              : (
                <ResponsiveContainer width="100%" height={180}>
                  <BarChart data={drvs} margin={{ top: 0, right: 0, bottom: 0, left: -24 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" vertical={false} />
                    <XAxis dataKey="driver_name" tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false} />
                    <YAxis tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false} />
                    <Tooltip content={<CUSTOM_TOOLTIP />} cursor={{ fill: 'rgba(255,255,255,.03)' }} />
                    <Bar dataKey="delivered" name="Entregues" radius={[4,4,0,0]} maxBarSize={40}>
                      {drvs.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              )}
          </Card>

          {/* Top Clientes */}
          <Card className="p-5">
            <div className="mb-4">
              <h2 className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Top Clientes</h2>
              <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>Por receita gerada</p>
            </div>
            {tops.length === 0
              ? <Empty message="Sem dados de clientes" />
              : (
                <div className="space-y-0.5">
                  {tops.slice(0,7).map((c, i) => (
                    <div
                      key={c.client_id}
                      className="flex items-center gap-3 px-2.5 py-2 rounded-lg transition-all duration-100"
                      style={{ cursor: 'default' }}
                      onMouseEnter={e => e.currentTarget.style.background = 'var(--surface-2)'}
                      onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                    >
                      <span
                        className="text-xs font-mono w-5 text-center flex-shrink-0"
                        style={{ color: i === 0 ? 'var(--accent)' : 'var(--text-3)', fontWeight: i === 0 ? 700 : 400 }}
                      >
                        {i + 1}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p className="text-xs font-medium truncate" style={{ color: 'var(--text)' }}>{c.client_name}</p>
                        <p className="text-xs" style={{ color: 'var(--text-3)' }}>{c.total_orders} pedidos</p>
                      </div>
                      <span className="text-xs font-semibold flex-shrink-0" style={{ color: 'var(--text-2)' }}>
                        {fmt.currency(c.total_cents)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
          </Card>
        </div>

        {/* ── Volume ao longo do tempo ──────────────────────────────── */}
        {areaData.length > 1 && (
          <Card className="p-5">
            <div className="mb-5">
              <h2 className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Volume de Entregas Recentes</h2>
              <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>Quantidade entregue por data</p>
            </div>
            <ResponsiveContainer width="100%" height={140}>
              <AreaChart data={areaData} margin={{ top: 4, right: 0, bottom: 0, left: -28 }}>
                <defs>
                  <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%"   stopColor="#f59e0b" stopOpacity={0.25} />
                    <stop offset="100%" stopColor="#f59e0b" stopOpacity={0.01} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" vertical={false} />
                <XAxis dataKey="name" tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: 'var(--text-3)', fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip content={<CUSTOM_TOOLTIP />} cursor={{ stroke: 'var(--border)', strokeWidth: 1 }} />
                <Area type="monotone" dataKey="Qtd" stroke="#f59e0b" strokeWidth={2} fill="url(#areaGrad)" dot={{ r: 3, fill: '#f59e0b', strokeWidth: 0 }} />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        )}

        {/* ── Últimas entregas ─────────────────────────────────────── */}
        <Card className="overflow-hidden">
          <div
            className="flex items-center justify-between px-5 py-4"
            style={{ borderBottom: '1px solid var(--border-subtle)' }}
          >
            <div>
              <h2 className="text-sm font-semibold" style={{ color: 'var(--text)' }}>Movimentação Recente</h2>
              <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)' }}>Últimos pedidos registrados</p>
            </div>
            <a href="/orders" className="text-xs" style={{ color: 'var(--accent)', textDecoration: 'none' }}>
              Ver todos →
            </a>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm" style={{ borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  {['Cliente','Produto','Qtd','Status','Motorista','Data'].map(h => (
                    <th key={h} className="section-label text-left px-5 py-3 font-medium">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {loading
                  ? Array.from({length:5}).map((_,i) => (
                    <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      {[140,80,40,80,80,80].map((w,j) => (
                        <td key={j} className="px-5 py-3"><Skeleton className={`h-3.5`} style={{ width: w }} /></td>
                      ))}
                    </tr>
                  ))
                  : rows.length === 0
                    ? <tr><td colSpan={6}><Empty message="Nenhuma entrega no período" /></td></tr>
                    : rows.map(row => (
                      <tr
                        key={row.order_id}
                        style={{ borderBottom: '1px solid var(--border-subtle)', transition: 'background .1s' }}
                        onMouseEnter={e => e.currentTarget.style.background = 'var(--surface-2)'}
                        onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                      >
                        <td className="px-5 py-3 font-medium text-sm" style={{ color: 'var(--text)' }}>{row.client_name}</td>
                        <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--text-2)' }}>{row.product_name}</td>
                        <td className="px-5 py-3 text-xs" style={{ color: 'var(--text-2)' }}>{row.quantity}</td>
                        <td className="px-5 py-3"><StatusBadge status={row.status} /></td>
                        <td className="px-5 py-3 text-xs" style={{ color: 'var(--text-3)' }}>{row.driver_name || '—'}</td>
                        <td className="px-5 py-3 font-mono text-xs" style={{ color: 'var(--text-3)' }}>{fmt.datetime(row.created_at)}</td>
                      </tr>
                    ))}
              </tbody>
            </table>
          </div>
        </Card>

      </div>
    </>
  )
}