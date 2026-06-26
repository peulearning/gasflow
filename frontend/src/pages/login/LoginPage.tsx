// import React from "react";
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Flame, Eye, EyeOff } from 'lucide-react'
import { authApi } from '../../api/endpoints'
import { useAuthStore } from '../../store/auth'
import { Button, Input, ErrorBanner } from '../../components/ui'

const schema = z.object({
  email:    z.string().email('E-mail inválido'),
  password: z.string().min(1, 'Senha obrigatória'),
})
type Form = z.infer<typeof schema>

const PRESETS = [
  { email: 'admin@gasflow.com',      role: 'admin'       },
  { email: 'operador@gasflow.com',   role: 'operational' },
  { email: 'financeiro@gasflow.com', role: 'financial'   },
]

export default function LoginPage() {
  const navigate   = useNavigate()
  const login      = useAuthStore(s => s.login)
  const [showPwd, setShowPwd] = useState(false)
  const [error,   setError]   = useState('')

  const { register, handleSubmit, setValue, formState: { errors, isSubmitting } } = useForm<Form>({
    resolver: zodResolver(schema),
    defaultValues: { email: 'admin@gasflow.com', password: 'password' },
  })

  async function onSubmit({ email, password }: Form) {
    setError('')
    try {
      const data = await authApi.login(email, password)
      login(data.access_token, data.role, data.user_id)
      navigate('/dashboard', { replace: true })
    } catch (e) {
      setError((e as Error).message)
    }
  }

  return (
    <div
      className="min-h-screen flex items-center justify-center p-6 relative overflow-hidden"
      style={{ background: 'var(--bg)' }}
    >
      {/* Glow de fundo */}
      <div style={{
        position: 'absolute', top: '-15%', left: '50%', transform: 'translateX(-50%)',
        width: 700, height: 500,
        background: 'radial-gradient(ellipse, rgba(245,158,11,.07) 0%, transparent 70%)',
        pointerEvents: 'none',
      }} />
      <div style={{
        position: 'absolute', bottom: '-10%', right: '-5%',
        width: 400, height: 400,
        background: 'radial-gradient(ellipse, rgba(255,107,43,.04) 0%, transparent 70%)',
        pointerEvents: 'none',
      }} />

      <div
        className="w-full max-w-sm rounded-2xl p-9 animate-slide-up relative"
        style={{
          background: 'var(--surface-1)',
          border: '1px solid var(--border-subtle)',
          boxShadow: '0 2px 4px rgba(0,0,0,.5), 0 20px 60px rgba(0,0,0,.4)',
        }}
      >
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div
            className="flex items-center justify-center rounded-2xl mb-4"
            style={{
              width: 54, height: 54,
              background: 'linear-gradient(135deg, #f59e0b 0%, #ff6b2b 100%)',
              boxShadow: '0 8px 28px rgba(245,158,11,.35)',
            }}
          >
            <Flame size={26} color="#000" strokeWidth={2.5} />
          </div>
          <h1
            className="text-2xl font-bold tracking-tight"
            style={{ color: 'var(--text)', letterSpacing: '-.03em' }}
          >
            GasFlow
          </h1>
          <p className="text-sm mt-1" style={{ color: 'var(--text-3)' }}>
            Painel de gestão da distribuidora
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && <ErrorBanner message={error} />}

          <Input
            {...register('email')}
            label="E-mail"
            type="email"
            placeholder="seu@email.com"
            error={errors.email?.message}
            autoComplete="email"
          />

          <div className="relative">
            <Input
              {...register('password')}
              label="Senha"
              type={showPwd ? 'text' : 'password'}
              placeholder="••••••••"
              error={errors.password?.message}
              autoComplete="current-password"
            />
            <button
              type="button"
              onClick={() => setShowPwd(v => !v)}
              className="absolute right-3 flex items-center"
              style={{ top: 34, color: 'var(--text-3)', background: 'none', border: 'none', cursor: 'pointer' }}
            >
              {showPwd ? <EyeOff size={15} /> : <Eye size={15} />}
            </button>
          </div>

          <Button
            type="submit"
            variant="primary"
            loading={isSubmitting}
            className="w-full py-2.5 mt-2 text-sm font-semibold"
          >
            Entrar
          </Button>
        </form>

        {/* Contas rápidas */}
        <div
          className="mt-6 rounded-xl p-4"
          style={{ background: 'var(--surface-2)', border: '1px solid var(--border-subtle)' }}
        >
          <p className="section-label mb-3">Contas de teste · senha: password</p>
          <div className="space-y-1">
            {PRESETS.map(p => (
              <button
                key={p.email}
                type="button"
                onClick={() => { setValue('email', p.email); setValue('password', 'password') }}
                className="flex items-center justify-between w-full px-3 py-2 rounded-lg text-xs transition-all duration-150"
                style={{ background: 'transparent', border: 'none', color: 'var(--text-3)', cursor: 'pointer', textAlign: 'left' }}
                onMouseEnter={e => { e.currentTarget.style.background = 'var(--surface-3)'; e.currentTarget.style.color = 'var(--accent)' }}
                onMouseLeave={e => { e.currentTarget.style.background = 'transparent';      e.currentTarget.style.color = 'var(--text-3)' }}
              >
                <span className="font-mono">{p.email}</span>
                <span
                  className="font-mono uppercase tracking-wider text-xs px-1.5 py-0.5 rounded"
                  style={{ background: 'var(--surface-3)', color: 'var(--text-3)', fontSize: 9 }}
                >
                  {p.role}
                </span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}