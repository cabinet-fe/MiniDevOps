import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { Toaster } from '@/components/ui/sonner'
import { AppLayout } from '@/components/layout/app-layout'
import { useAuthStore } from '@/stores/auth-store'
import { LoginPage } from '@/pages/login'
import { DashboardPage } from '@/pages/dashboard'
import { ProjectListPage } from '@/pages/projects/list'
import { ProjectDetailPage } from '@/pages/projects/detail'
import { BuildDetailPage } from '@/pages/builds/detail'
import { BuildListPage } from '@/pages/builds/list'
import { ServerListPage } from '@/pages/servers/list'
import { UserListPage } from '@/pages/users/list'
import { AuditLogsPage } from '@/pages/audit-logs'
import { SettingsPage } from '@/pages/settings'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, fetchMe } = useAuthStore()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    fetchMe().finally(() => setChecked(true))
  }, [fetchMe])

  if (!checked) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="projects" element={<ProjectListPage />} />
        <Route path="projects/:id" element={<ProjectDetailPage />} />
        <Route path="builds" element={<BuildListPage />} />
        <Route path="builds/:id" element={<BuildDetailPage />} />
        <Route path="servers" element={<ServerListPage />} />
        <Route path="users" element={<UserListPage />} />
        <Route path="audit-logs" element={<AuditLogsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
      <Toaster position="top-right" richColors />
    </BrowserRouter>
  )
}
