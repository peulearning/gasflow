import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '../src/store/auth'
import { AppLayout }   from './components/layout/AppLayout'
import LoginPage       from '././pages/login/LoginPage'
import DashboardPage   from './pages/dashboard/Dashboardpage'
import OrdersPage      from '../src/pages/orders/OrdersPage'
import InventoryPage   from '../src/pages/inventory/Inventorypage'
import ChargesPage     from '../src/pages/charges/Chargespage'

function Guard({ children }: { children: React.ReactNode }) {
  const isAuth = useAuthStore(s => s.isAuth)
  return isAuth ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<Guard><AppLayout /></Guard>}>
        <Route index         element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="orders"    element={<OrdersPage />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="charges"   element={<ChargesPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}