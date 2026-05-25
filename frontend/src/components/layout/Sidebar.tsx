import React from "react";
import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Package, Warehouse, CreditCard,
  LogOut, Flame, ChevronRight,
} from 'lucide-react'
import { cn } from '../../lib/utils'
import { useAuthStore } from '../../store/auth'

const NAV = [
  { to: '/dashboard', label: 'Dashboard',  icon: LayoutDashboard, sub: 'KPIs & métricas'    },
  { to: '/orders',    label: 'Pedidos',     icon: Package,         sub: 'Fluxo & status'     },
  { to: '/inventory', label: 'Estoque',     icon: Warehouse,       sub: 'Depósitos & saldo'  },
  { to: '/charges',   label: 'Cobranças',   icon: CreditCard,      sub: 'Financeiro'         },
]

export function Sidebar() {
  const { role, logout } = useAuthStore()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <aside
      className="flex flex-col flex-shrink-0 h-full"
      style={{
        width: 216,
        background: 'var(--surface-1)',
        borderRight: '1px solid var(--border-subtle)',
      }}
    >
      {/* ── Logo ─────────────────────────────────────────────────────── */}
      <div className="flex items-center gap-3 px-5 py-6" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
        <div
          className="flex items-center justify-center rounded-xl flex-shrink-0"
          style={{
            width: 34, height: 34,
            background: 'linear-gradient(135deg, #f59e0b 0%, #ff6b2b 100%)',
            boxShadow: '0 4px 14px rgba(245,158,11,.35)',
          }}
        >
          <Flame size={17} color="#000" strokeWidth={2.5} />
        </div>
        <div>
          <p className="font-bold text-base leading-none" style={{ color: 'var(--text)', letterSpacing: '-.02em' }}>
            GasFlow
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--text-3)', letterSpacing: '.06em' }}>
            DISTRIBUIDORA
          </p>
        </div>
      </div>

      {/* ── Status live ──────────────────────────────────────────────── */}
      <div className="flex items-center gap-2 px-5 py-2.5" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
        <span className="dot-live" />
        <span className="text-xs" style={{ color: 'var(--text-3)', letterSpacing: '.05em' }}>ONLINE</span>
      </div>

      {/* ── Nav ──────────────────────────────────────────────────────── */}
      <nav className="flex-1 px-3 py-3 space-y-0.5">
        {NAV.map(({ to, label, icon: Icon, sub }) => (
          <NavLink key={to} to={to} end={to === '/dashboard'}>
            {({ isActive }) => (
              <div
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-lg cursor-pointer transition-all duration-150',
                  isActive
                    ? 'bg-amber-dim border border-amber-border'
                    : 'hover:bg-surface-2 border border-transparent'
                )}
              >
                <Icon
                  size={15}
                  strokeWidth={isActive ? 2.5 : 2}
                  style={{ color: isActive ? 'var(--accent)' : 'var(--text-3)', flexShrink: 0 }}
                />
                <div className="flex-1 min-w-0">
                  <p
                    className="text-sm leading-tight"
                    style={{
                      color: isActive ? 'var(--accent)' : 'var(--text-2)',
                      fontWeight: isActive ? 600 : 400,
                    }}
                  >
                    {label}
                  </p>
                  <p className="text-xs mt-0.5 truncate" style={{ color: 'var(--text-3)' }}>
                    {sub}
                  </p>
                </div>
                {isActive && <ChevronRight size={11} style={{ color: 'var(--accent)', opacity: .6 }} />}
              </div>
            )}
          </NavLink>
        ))}
      </nav>

      {/* ── Footer / User ─────────────────────────────────────────────── */}
      <div className="px-3 pb-4 pt-2" style={{ borderTop: '1px solid var(--border-subtle)' }}>
        <div
          className="rounded-lg px-3 py-2.5 mb-2"
          style={{ background: 'var(--surface-2)' }}
        >
          <p className="text-xs mb-0.5" style={{ color: 'var(--text-3)' }}>Conectado como</p>
          <p
            className="text-xs font-mono font-medium uppercase tracking-wider"
            style={{ color: 'var(--accent)' }}
          >
            {role ?? '—'}
          </p>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center justify-center gap-2 w-full rounded-lg px-3 py-2 text-sm transition-all duration-150"
          style={{
            background: 'transparent',
            border: '1px solid var(--border)',
            color: 'var(--text-3)',
          }}
          onMouseEnter={e => {
            e.currentTarget.style.background = 'var(--surface-2)'
            e.currentTarget.style.color = 'var(--text-2)'
          }}
          onMouseLeave={e => {
            e.currentTarget.style.background = 'transparent'
            e.currentTarget.style.color = 'var(--text-3)'
          }}
        >
          <LogOut size={13} />
          Sair
        </button>
      </div>
    </aside>
  )
}