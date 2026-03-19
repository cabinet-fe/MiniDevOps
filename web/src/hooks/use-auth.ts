import { useAuthStore } from '@/stores/auth-store'

export function useAuth() {
  const { user, isAuthenticated, logout } = useAuthStore()
  
  const hasRole = (...roles: string[]) => {
    return user ? roles.includes(user.role) : false
  }
  
  const isAdmin = () => hasRole('admin')
  const isOps = () => hasRole('ops', 'admin')
  
  return { user, isAuthenticated, logout, hasRole, isAdmin, isOps }
}
