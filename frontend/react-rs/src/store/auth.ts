import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token:   string | null
  role:    string | null
  userId:  string | null
  isAuth:  boolean
  login:   (token: string, role: string, userId: string) => void
  logout:  () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token:  null,
      role:   null,
      userId: null,
      isAuth: false,
      login: (token, role, userId) =>
        set({ token, role, userId, isAuth: true }),
      logout: () =>
        set({ token: null, role: null, userId: null, isAuth: false }),
    }),
    { name: 'gf-auth' }
  )
)