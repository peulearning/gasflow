import { cn } from '../../lib/utils'
import { Loader2 } from 'lucide-react'
import React from 'react'

// ── Badge de Status ─────────────────────────────────────────────────────────
const STATUS_MAP: Record<string, { label: string; bg: string; color: string }> = {
  received:    { label: 'Recebido',     bg: 'rgba(100,100,130,.14)', color: '#8888a8' },
  approved:    { label: 'Aprovado',     bg: 'rgba(59,130,246,.14)',  color: '#60a5fa' },
  separated:   { label: 'Separado',     bg: 'rgba(139,92,246,.14)',  color: '#a78bfa' },
  in_route:    { label: 'Em rota',      bg: 'rgba(245,158,11,.14)',  color: '#f59e0b' },
  delivered:   { label: 'Entregue',     bg: 'rgba(34,197,94,.14)',   color: '#22c55e' },
  cancelled:   { label: 'Cancelado',    bg: 'rgba(239,68,68,.14)',   color: '#ef4444' },
  rescheduled: { label: 'Reagendado',   bg: 'rgba(255,107,43,.14)',  color: '#ff6b2b' },
  active:      { label: 'Ativo',        bg: 'rgba(34,197,94,.14)',   color: '#22c55e' },
  blocked:     { label: 'Bloqueado',    bg: 'rgba(239,68,68,.14)',   color: '#ef4444' },
  inactive:    { label: 'Inativo',      bg: 'rgba(100,100,130,.14)', color: '#8888a8' },
  pending:     { label: 'Pendente',     bg: 'rgba(245,158,11,.14)',  color: '#f59e0b' },
  paid:        { label: 'Pago',         bg: 'rgba(34,197,94,.14)',   color: '#22c55e' },
  overdue:     { label: 'Inadimplente', bg: 'rgba(239,68,68,.14)',   color: '#ef4444' },
}

export function StatusBadge({ status }: { status: string }) {
  const s = STATUS_MAP[status] ?? { label: status, bg: 'rgba(100,100,130,.14)', color: '#8888a8' }
  return (
    <span
      className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium font-mono"
      style={{ background: s.bg, color: s.color }}
    >
      <span
        className="inline-block rounded-full flex-shrink-0"
        style={{ width: 5, height: 5, background: s.color }}
      />
      {s.label}
    </span>
  )
}

// ── Card ────────────────────────────────────────────────────────────────────
export function Card({ children, className, glow }: {
  children: React.ReactNode; className?: string; glow?: boolean
}) {
  return (
    <div
      className={cn('rounded-xl transition-all duration-200', className)}
      style={{
        background: 'var(--surface-1)',
        border: '1px solid var(--border-subtle)',
        boxShadow: '0 1px 3px rgba(0,0,0,.4)',
      }}
      onMouseEnter={e => {
        if (glow) e.currentTarget.style.boxShadow = '0 0 0 1px rgba(245,158,11,.2), 0 8px 32px rgba(245,158,11,.07)'
        else      e.currentTarget.style.borderColor = 'var(--border)'
      }}
      onMouseLeave={e => {
        if (glow) e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,.4)'
        else      e.currentTarget.style.borderColor = 'var(--border-subtle)'
      }}
    >
      {children}
    </div>
  )
}

// ── Button ──────────────────────────────────────────────────────────────────
type BtnVariant = 'primary' | 'ghost' | 'danger' | 'subtle'

export function Button({
  children, variant = 'ghost', size = 'md',
  loading, className, ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: BtnVariant; size?: 'sm' | 'md'; loading?: boolean
}) {
  const base = 'inline-flex items-center justify-center gap-2 font-medium rounded-lg transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed select-none'

  const sizes = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-4 py-2 text-sm',
  }

  const styles: Record<BtnVariant, React.CSSProperties> = {
    primary: { background: 'var(--accent)', color: '#000' },
    ghost:   { background: 'transparent', color: 'var(--text-2)', border: '1px solid var(--border)' },
    danger:  { background: 'rgba(239,68,68,.1)', color: '#ef4444', border: '1px solid rgba(239,68,68,.2)' },
    subtle:  { background: 'var(--surface-2)', color: 'var(--text-2)', border: '1px solid var(--border-subtle)' },
  }

  return (
    <button
      {...props}
      disabled={loading || props.disabled}
      className={cn(base, sizes[size], className)}
      style={styles[variant]}
      onMouseEnter={e => {
        if (variant === 'primary') e.currentTarget.style.filter = 'brightness(1.1)'
        if (variant === 'ghost' || variant === 'subtle') e.currentTarget.style.background = 'var(--surface-2)'
        if (variant === 'danger') e.currentTarget.style.background = 'rgba(239,68,68,.18)'
      }}
      onMouseLeave={e => {
        if (variant === 'primary') e.currentTarget.style.filter = 'none'
        if (variant === 'ghost')   e.currentTarget.style.background = 'transparent'
        if (variant === 'subtle')  e.currentTarget.style.background = 'var(--surface-2)'
        if (variant === 'danger')  e.currentTarget.style.background = 'rgba(239,68,68,.1)'
      }}
    >
      {loading && <Loader2 size={13} className="animate-spin" />}
      {children}
    </button>
  )
}

// ── Input ───────────────────────────────────────────────────────────────────
export const Input = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement> & { label?: string; error?: string }
>(({ label, error, className, ...props }, ref) => (
  <div className="flex flex-col gap-1.5">
    {label && (
      <label className="section-label">{label}</label>
    )}
    <input
      ref={ref}
      {...props}
      className={cn('w-full rounded-lg px-3 py-2 text-sm outline-none transition-all duration-150', className)}
      style={{
        background: 'var(--surface-2)',
        border: `1px solid ${error ? 'var(--danger)' : 'var(--border)'}`,
        color: 'var(--text)',
      }}
      onFocus={e => { e.currentTarget.style.borderColor = error ? 'var(--danger)' : 'var(--accent)'; e.currentTarget.style.boxShadow = error ? '0 0 0 3px rgba(239,68,68,.1)' : '0 0 0 3px rgba(245,158,11,.1)' }}
      onBlur={e  => { e.currentTarget.style.borderColor = error ? 'var(--danger)' : 'var(--border)'; e.currentTarget.style.boxShadow = 'none' }}
    />
    {error && <p className="text-xs" style={{ color: 'var(--danger)' }}>{error}</p>}
  </div>
))
Input.displayName = 'Input'

// ── Select ──────────────────────────────────────────────────────────────────
export const Select = React.forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement> & { label?: string }
>(({ label, className, children, ...props }, ref) => (
  <div className="flex flex-col gap-1.5">
    {label && <label className="section-label">{label}</label>}
    <select
      ref={ref}
      {...props}
      className={cn('w-full rounded-lg px-3 py-2 text-sm outline-none appearance-none transition-all duration-150', className)}
      style={{
        background: 'var(--surface-2)',
        border: '1px solid var(--border)',
        color: 'var(--text)',
        cursor: 'pointer',
      }}
    >
      {children}
    </select>
  </div>
))
Select.displayName = 'Select'

// ── Page Header ─────────────────────────────────────────────────────────────
export function PageHeader({
  title, subtitle, actions,
}: { title: string; subtitle?: string; actions?: React.ReactNode }) {
  return (
    <div
      className="sticky top-0 z-20 flex items-center justify-between px-7 py-4"
      style={{
        background: 'rgba(15,15,18,.88)',
        backdropFilter: 'blur(14px)',
        borderBottom: '1px solid var(--border-subtle)',
      }}
    >
      <div>
        <h1
          className="text-xl font-bold leading-none"
          style={{ color: 'var(--text)', letterSpacing: '-.025em' }}
        >
          {title}
        </h1>
        {subtitle && (
          <p className="text-xs mt-1" style={{ color: 'var(--text-3)' }}>{subtitle}</p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}

// ── KPI Card ────────────────────────────────────────────────────────────────
export function KPICard({
  title, value, subtitle, icon: Icon, accent = 'var(--accent)', delay = 0,
}: {
  title: string; value: string | number; subtitle?: string
  icon: React.ElementType; accent?: string; delay?: number
}) {
  return (
    <Card
      glow
      className="p-5 animate-slide-up"
      style={{ animationDelay: `${delay}ms` } as React.CSSProperties}
    >
      <div className="flex items-start justify-between mb-4">
        <div
          className="flex items-center justify-center rounded-lg"
          style={{
            width: 36, height: 36,
            background: `color-mix(in srgb, ${accent} 14%, transparent)`,
            border: `1px solid color-mix(in srgb, ${accent} 25%, transparent)`,
          }}
        >
          <Icon size={15} style={{ color: accent }} strokeWidth={2} />
        </div>
      </div>
      <p className="stat-num" style={{ color: 'var(--text)' }}>{value}</p>
      <p className="section-label mt-2">{title}</p>
      {subtitle && (
        <p className="text-xs mt-1" style={{ color: 'var(--text-3)' }}>{subtitle}</p>
      )}
    </Card>
  )
}

// ── Skeleton ─────────────────────────────────────────────────────────────────
export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('skeleton', className)} />
}

// ── Empty state ──────────────────────────────────────────────────────────────
export function Empty({ message = 'Nenhum dado encontrado' }: { message?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3">
      <div
        className="w-12 h-12 rounded-full flex items-center justify-center"
        style={{ background: 'var(--surface-2)' }}
      >
        <span className="text-2xl">📭</span>
      </div>
      <p className="text-sm" style={{ color: 'var(--text-3)' }}>{message}</p>
    </div>
  )
}

// ── Toast simples ─────────────────────────────────────────────────────────────
export function ErrorBanner({ message }: { message: string }) {
  return (
    <div
      className="flex items-center gap-3 px-4 py-3 rounded-lg text-sm mb-4 animate-fade-in"
      style={{
        background: 'rgba(239,68,68,.08)',
        border: '1px solid rgba(239,68,68,.2)',
        color: '#ef4444',
      }}
    >
      ⚠ {message}
    </div>
  )
}