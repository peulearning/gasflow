import { Outlet } from 'react-router-dom'
import { Sidebar } from './SideBar'

export function AppLayout() {
  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--bg)' }}>
      <Sidebar />
      <main className="flex-1 overflow-y-auto overflow-x-hidden">
        <Outlet />
      </main>
    </div>
  )
}