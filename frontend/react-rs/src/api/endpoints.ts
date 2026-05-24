import { apiClient } from './client'

// ── Types ─────────────────────────────────────────────────────────────────────
export interface KPISummary {
  period: { from: string; to: string }
  deliveries: { total: number; delivered: number; delayed: number; rescheduled: number; sla_rate: number; volume_kg: number }
  inventory:  { total_units: number; reserved: number; available: number; low_stock_alerts: number }
  billing:    { revenue_cents: number; overdue_count: number; overdue_amount_cents: number }
}

export interface DeliveryRow {
  order_id: string; client_name: string; product_name: string
  quantity: number; status: string; driver_name: string
  scheduled_at?: string; delivered_at?: string; created_at: string
}

export interface DriverPerf {
  driver_id: string; driver_name: string
  total: number; delivered: number; delayed: number; sla_rate: number
}

export interface TopClient {
  client_id: string; client_name: string; total_orders: number; total_cents: number
}

export interface Order {
  id: string; client_id: string; address_id: string; product_id: string
  quantity: number; status: string; driver_id?: string
  scheduled_at?: string; delivered_at?: string; notes?: string
  created_at: string; updated_at: string
}

export interface StatusHistory {
  id: string; order_id: string; from_status: string; to_status: string
  changed_by: string; reason: string; created_at: string
}

export interface Client {
  id: string; name: string; document: string; phone: string
  email: string; status: string; created_at: string; addresses?: Address[]
}

export interface Address {
  id: string; client_id: string; street: string; city: string
  state: string; zipcode: string; region: string; is_primary: boolean
}

export interface InventoryItem {
  id: string; deposit_id: string; product_id: string; quantity: number; reserved: number
}

export interface Deposit { id: string; name: string; city: string }

export interface Charge {
  id: string; order_id: string; client_id: string
  amount: { cents: number }; status: string
  due_date: string; paid_at?: string; created_at: string
}

export interface Paginated<T> { data: T[] | null; total: number }

// ── Auth ─────────────────────────────────────────────────────────────────────
export const authApi = {
  login: (email: string, password: string) =>
    apiClient.post<{ access_token: string; user_id: string; role: string }>
    ('/api/auth/login', { email, password }).then(r => r.data),
}

// ── Dashboard ─────────────────────────────────────────────────────────────────
export const dashboardApi = {
  kpis: (from?: string, to?: string) => {
    const p = new URLSearchParams()
    if (from) p.set('from', from)
    if (to)   p.set('to', to)
    return apiClient.get<KPISummary>(`/api/dashboard/kpis?${p}`).then(r => r.data)
  },
  deliveries: (params: Record<string, string> = {}) =>
    apiClient.get<Paginated<DeliveryRow>>(`/api/dashboard/deliveries?${new URLSearchParams(params)}`).then(r => r.data),
  driverPerf: (from?: string, to?: string) => {
    const p = new URLSearchParams()
    if (from) p.set('from', from)
    if (to)   p.set('to', to)
    return apiClient.get<DriverPerf[]>(`/api/dashboard/driver-performance?${p}`).then(r => r.data)
  },
  topClients: () => apiClient.get<TopClient[]>('/api/dashboard/top-clients').then(r => r.data),
}

// ── Orders ────────────────────────────────────────────────────────────────────
export const ordersApi = {
  list:       (params: Record<string, string> = {}) =>
    apiClient.get<Paginated<Order>>(`/api/orders?${new URLSearchParams(params)}`).then(r => r.data),
  get:        (id: string) => apiClient.get<Order>(`/api/orders/${id}`).then(r => r.data),
  create:     (body: unknown) => apiClient.post<Order>('/api/orders', body).then(r => r.data),
  transition: (id: string, status: string, opts: { reason?: string; driver_id?: string } = {}) =>
    apiClient.patch<Order>(`/api/orders/${id}/status`, { status, ...opts }).then(r => r.data),
  history:    (id: string) => apiClient.get<StatusHistory[]>(`/api/orders/${id}/history`).then(r => r.data),
}

// ── Clients ───────────────────────────────────────────────────────────────────
export const clientsApi = {
  list:     (params: Record<string, string> = {}) =>
    apiClient.get<Paginated<Client>>(`/api/clients?${new URLSearchParams(params)}`).then(r => r.data),
  get:      (id: string) => apiClient.get<Client>(`/api/clients/${id}`).then(r => r.data),
  create:   (body: unknown) => apiClient.post<Client>('/api/clients', body).then(r => r.data),
  block:    (id: string) => apiClient.post(`/api/clients/${id}/block`, {}).then(r => r.data),
  activate: (id: string) => apiClient.post(`/api/clients/${id}/activate`, {}).then(r => r.data),
}

// ── Inventory ─────────────────────────────────────────────────────────────────
export const inventoryApi = {
  deposits:     () => apiClient.get<Deposit[]>('/api/inventory/deposits').then(r => r.data),
  items:        (depositId: string) => apiClient.get<InventoryItem[]>(`/api/inventory/deposits/${depositId}/items`).then(r => r.data),
  lowStock:     () => apiClient.get<InventoryItem[]>('/api/inventory/low-stock').then(r => r.data),
  receive:      (depositId: string, product_id: string, quantity: number) =>
    apiClient.post(`/api/inventory/deposits/${depositId}/receive`, { product_id, quantity }).then(r => r.data),
}

// ── Charges ───────────────────────────────────────────────────────────────────
export const chargesApi = {
  list:    (params: Record<string, string> = {}) =>
    apiClient.get<Paginated<Charge>>(`/api/charges?${new URLSearchParams(params)}`).then(r => r.data),
  overdue: () => apiClient.get<Paginated<Charge>>('/api/charges/overdue').then(r => r.data),
  pay:     (id: string) => apiClient.post(`/api/charges/${id}/pay`, {}).then(r => r.data),
}