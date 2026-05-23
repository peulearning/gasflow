# 🚀 GasFlow

**GasFlow** é um sistema completo de gestão, vendas e distribuição. Construído sob uma arquitetura de **Monólito Modular**, o sistema utiliza mensageria assíncrona para garantir alta performance e resiliência, separando responsabilidades de forma clara entre domínios.

```
 Atenção ⚠️ : Projeto sem fins lucrativos apenas para estudos.
```

## 🛠️ Stack Tecnológico

* **Backend:** Go (Chi Router)
* **Frontend:** Next.js (React)
* **Banco de Dados:** MySQL 8
* **Mensageria:** RabbitMQ
* **Infraestrutura:** Docker & Docker Compose

---

## 🏗️ Arquitetura

O backend em Go segue os princípios de **Domain-Driven Design (DDD)** e **Clean Architecture** dentro de um modelo de Monólito Modular.

* `internal/domain`: Regras de negócio puras, entidades, Value Objects e Máquina de Estados.
* `internal/modules`: Casos de uso específicos por contexto (`clients`, `orders`, `inventory`, `billing`, `analytics`).
* `internal/infra`: Implementações técnicas (MySQL, RabbitMQ, JWT).

### Máquina de Estados (Pedidos)
O ciclo de vida do pedido é estritamente controlado por uma FSM (Finite State Machine) validada no domínio:

```text
[received] ➔ [approved] ➔ [separated] ➔ [in_route] ➔ [delivered]
      ↳ [cancelled]             ↳ [cancelled]       ↳ [rescheduled] ➔ [in_route]

```

---


# GasFlow — Frontend Architecture (React + TypeScript + Vite + Tailwind)

## Visão Geral

O frontend do GasFlow será construído com foco em:

* Performance
* Escalabilidade
* Interface moderna
* Componentização
* Responsividade
* Clean Architecture no Frontend
* Fácil manutenção
* UX moderna estilo SaaS/Admin Dashboard

A stack principal:

* React
* TypeScript
* Vite
* Tailwind CSS
* React Router DOM
* React Query (TanStack Query)
* Axios
* Zustand
* React Hook Form
* Zod
* Lucide Icons
* Framer Motion
* shadcn/ui

---

# Estrutura Recomendada do Projeto

```bash
src/
 ├── assets/
 ├── components/
 │    ├── ui/
 │    ├── layout/
 │    ├── forms/
 │    ├── cards/
 │    ├── tables/
 │    ├── charts/
 │    └── feedback/
 │
 ├── pages/
 │    ├── auth/
 │    ├── dashboard/
 │    ├── customers/
 │    ├── deliveries/
 │    ├── orders/
 │    ├── stock/
 │    ├── financial/
 │    └── settings/
 │
 ├── services/
 │    ├── api.ts
 │    ├── auth.service.ts
 │    ├── order.service.ts
 │    └── delivery.service.ts
 │
 ├── hooks/
 ├── store/
 ├── routes/
 ├── contexts/
 ├── layouts/
 ├── lib/
 ├── types/
 ├── utils/
 ├── styles/
 ├── constants/
 ├── App.tsx
 ├── main.tsx
 └── vite-env.d.ts
```

---

# Instalação Inicial

## Criar Projeto

```bash
npm create vite@latest gasflow-web -- --template react-ts
```

## Entrar na Pasta

```bash
cd gasflow-web
```

## Instalar Dependências

```bash
npm install react-router-dom axios zustand @tanstack/react-query react-hook-form zod @hookform/resolvers framer-motion lucide-react clsx tailwind-merge
```

## Tailwind CSS

```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

---

# Configuração Tailwind

## tailwind.config.ts

```ts
import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#0F172A',
        secondary: '#1E293B',
        accent: '#22C55E',
        danger: '#EF4444',
        warning: '#F59E0B',
        muted: '#94A3B8',
      },
      borderRadius: {
        xl: '1rem',
        '2xl': '1.5rem',
      },
      boxShadow: {
        soft: '0 10px 30px rgba(0,0,0,0.08)',
      },
    },
  },
  plugins: [],
} satisfies Config
```

---

# Estilo Visual do GasFlow

## Identidade Visual

O sistema deve transmitir:

* Tecnologia
* Gestão logística
* Confiabilidade
* Velocidade
* Organização operacional

## Paleta Recomendada

| Cor               | Uso               |
| ----------------- | ----------------- |
| Slate/Blue Escuro | Fundo principal   |
| Verde             | Sucesso/Gás/Fluxo |
| Branco            | Contraste         |
| Cinza Claro       | Backgrounds       |
| Vermelho          | Alertas           |
| Amarelo           | Avisos            |

---

# Layout Principal

## Estrutura do Dashboard

```txt
---------------------------------
 Sidebar | Header               |
 Sidebar |----------------------|
 Sidebar | Conteúdo             |
 Sidebar |                      |
 Sidebar |                      |
---------------------------------
```

---

# Sidebar Recomendada

## Menus

* Dashboard
* Pedidos
* Entregas
* Clientes
* Estoque
* Financeiro
* Motoristas
* Rotas
* Relatórios
* Configurações

---

# Funcionalidades Principais do Frontend

## Autenticação

* Login
* Refresh Token
* Persistência de sessão
* Controle de permissões
* Middleware de rotas privadas

## Dashboard

* KPIs
* Pedidos do dia
* Entregas pendentes
* Total faturado
* Gráficos
* Atividades recentes

## Pedidos

* CRUD
* Status em tempo real
* Busca dinâmica
* Filtros avançados
* Paginação

## Entregas

* Rastreamento
* Status da entrega
* Motorista
* Tempo estimado
* Histórico

## Estoque

* Controle de botijões
* Entradas/Saídas
* Alertas de estoque

## Financeiro

* Receitas
* Despesas
* Fluxo de caixa
* Relatórios

---

# Bibliotecas Recomendadas

## Tabelas

```bash
npm install @tanstack/react-table
```

## Gráficos

```bash
npm install recharts
```

## Toasts

```bash
npm install sonner
```

## Modais

Utilizar shadcn/ui

---

# Estrutura de Rotas

```tsx
<AuthLayout>
  /login
</AuthLayout>

<DashboardLayout>
  /dashboard
  /orders
  /deliveries
  /customers
  /stock
  /financial
</DashboardLayout>
```

---

# Organização de Componentes

## Componentes Reutilizáveis

### Cards

* KPI Card
* Financial Card
* Delivery Status Card
* Customer Card

### Inputs

* Input
* Select
* Search Input
* Currency Input
* Date Picker

### Feedback

* Toast
* Loading
* Empty State
* Error State

---

# React Query

## Motivo

Evitar:

* Requests duplicadas
* Re-renderizações desnecessárias
* Estado global excessivo

## Exemplo

```ts
export function useOrders() {
  return useQuery({
    queryKey: ['orders'],
    queryFn: orderService.getAll,
  })
}
```

---

# Zustand

## Uso

Ideal para:

* Usuário autenticado
* Sidebar state
* Tema
* Preferências

## Exemplo

```ts
import { create } from 'zustand'

interface AuthStore {
  token: string | null
  setToken: (token: string) => void
}

export const useAuthStore = create<AuthStore>((set) => ({
  token: null,
  setToken: (token) => set({ token }),
}))
```

---

# Estrutura de Services

## api.ts

```ts
import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
})
```

---

# Responsividade

## Mobile First

Sempre:

```tsx
className="grid grid-cols-1 lg:grid-cols-4 gap-4"
```

---

# UX Moderna

## Recomendações

* Skeleton loading
* Transições suaves
* Hover elegante
* Feedback visual instantâneo
* Empty states amigáveis
* Dark Mode
* Pesquisa em tempo real
* Tabelas inteligentes

---

# Componentes Importantes

## Dashboard KPI Card

```txt
+----------------------+
| Total Pedidos        |
| 1.245                |
| +12% hoje            |
+----------------------+
```

## Delivery Timeline

```txt
Pedido Criado
↓
Saiu para entrega
↓
Entregue
```

---

# Dark Mode

Recomendado fortemente.

Usar:

```bash
npm install next-themes
```

Mesmo com Vite.

---

# Tipografia Recomendada

## Fontes

* Inter
* Satoshi
* Plus Jakarta Sans

---

# Animações

## Framer Motion

Usar para:

* Entrada de páginas
* Modais
* Hover cards
* Sidebar animation
* Loading transitions

---

# Segurança Frontend

## Implementar

* Protected Routes
* Refresh Token
* Interceptors Axios
* Logout automático
* Sanitização

---

# Melhor Estrutura para Escalar

## Separar:

* UI
* Rules
* Services
* Hooks
* State
* API

Evitar:

* Lógica gigante dentro de componentes
* Requests dentro de páginas
* Componentes com mais de 300 linhas

---

# Melhor Estrutura de Página

```tsx
export function OrdersPage() {
  const { data, isLoading } = useOrders()

  if (isLoading) {
    return <OrdersSkeleton />
  }

  return (
    <div className="space-y-6">
      <OrdersHeader />
      <OrdersFilters />
      <OrdersTable data={data} />
    </div>
  )
}
```

---

# Sugestão de MVP Inicial

## Fase 1

* Login
* Dashboard
* CRUD Clientes
* CRUD Pedidos
* CRUD Entregas
* Layout Base

## Fase 2

* Financeiro
* Relatórios
* Controle de estoque
* Rastreamento

## Fase 3

* Notificações
* PWA
* Tempo real
* IA
* Analytics

---

# Design Inspiration

Referências:

* Linear
* Stripe Dashboard
* Vercel
* Notion
* Clerk
* Raycast

Visual:

* Minimalista
* Bordas suaves
* Muito espaçamento
* Glassmorphism leve
* Cards modernos
* Sombras suaves

---

# Recomendação Final

Para o GasFlow:

A melhor arquitetura será:

* Feature Based Structure
* React Query para server state
* Zustand para global state
* Tailwind + shadcn/ui
* Componentização forte
* Dashboard SaaS moderno
* Dark mode nativo
* Loading states sofisticados
* Estrutura pronta para crescer

O frontend precisa parecer:

* Sistema premium
* Plataforma logística moderna
* ERP SaaS escalável
* Aplicação enterprise
