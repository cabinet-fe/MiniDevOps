import { lazy, Suspense, useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { ThemeProvider } from 'next-themes'
import { Toaster } from '@/components/ui/sonner'
import { AppLayout } from '@/components/layout/app-layout'
import { useAuthStore } from '@/stores/auth-store'

const LoginPage = lazy(() =>
  import('@/pages/login').then((m) => ({ default: m.LoginPage })),
)
const DashboardPage = lazy(() =>
  import('@/pages/dashboard').then((m) => ({ default: m.DashboardPage })),
)
const ProjectListPage = lazy(() =>
  import('@/pages/projects/list').then((m) => ({ default: m.ProjectListPage })),
)
const ProjectDetailPage = lazy(() =>
  import('@/pages/projects/detail').then((m) => ({ default: m.ProjectDetailPage })),
)
const BuildDetailPage = lazy(() =>
  import('@/pages/builds/detail').then((m) => ({ default: m.BuildDetailPage })),
)
const ServerListPage = lazy(() =>
  import('@/pages/servers/list').then((m) => ({ default: m.ServerListPage })),
)
const UserListPage = lazy(() =>
  import('@/pages/users/list').then((m) => ({ default: m.UserListPage })),
)
const AuditLogsPage = lazy(() =>
  import('@/pages/audit-logs').then((m) => ({ default: m.AuditLogsPage })),
)
const DictionaryListPage = lazy(() =>
  import('@/pages/dictionaries/list').then((m) => ({ default: m.DictionaryListPage })),
)
const SettingsPage = lazy(() =>
  import('@/pages/settings').then((m) => ({ default: m.SettingsPage })),
)
const ProjectManualPage = lazy(() =>
  import('@/pages/project-manual').then((m) => ({ default: m.ProjectManualPage })),
)

function RouteFallback() {
  return (
    <div className="flex min-h-[50vh] w-full flex-1 items-center justify-center">
      <div
        className="size-9 animate-spin rounded-full border-2 border-muted border-t-amber-500"
        aria-label="加载中"
      />
    </div>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, fetchMe } = useAuthStore()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    fetchMe().finally(() => setChecked(true))
  }, [fetchMe])

  if (!checked) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="size-8 animate-spin rounded-full border-2 border-muted border-t-foreground" />
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
      <Route
        path="/login"
        element={
          <Suspense
            fallback={
              <div className="flex min-h-screen items-center justify-center bg-[#08090c]">
                <div className="size-9 animate-spin rounded-full border-2 border-zinc-700 border-t-amber-500" />
              </div>
            }
          >
            <LoginPage />
          </Suspense>
        }
      />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route
          index
          element={
            <Suspense fallback={<RouteFallback />}>
              <DashboardPage />
            </Suspense>
          }
        />
        <Route
          path="manual"
          element={
            <Suspense fallback={<RouteFallback />}>
              <ProjectManualPage />
            </Suspense>
          }
        />
        <Route
          path="projects"
          element={
            <Suspense fallback={<RouteFallback />}>
              <ProjectListPage />
            </Suspense>
          }
        />
        <Route
          path="projects/:id"
          element={
            <Suspense fallback={<RouteFallback />}>
              <ProjectDetailPage />
            </Suspense>
          }
        />
        <Route
          path="builds/:id"
          element={
            <Suspense fallback={<RouteFallback />}>
              <BuildDetailPage />
            </Suspense>
          }
        />
        <Route
          path="servers"
          element={
            <Suspense fallback={<RouteFallback />}>
              <ServerListPage />
            </Suspense>
          }
        />
        <Route
          path="users"
          element={
            <Suspense fallback={<RouteFallback />}>
              <UserListPage />
            </Suspense>
          }
        />
        <Route
          path="dictionaries"
          element={
            <Suspense fallback={<RouteFallback />}>
              <DictionaryListPage />
            </Suspense>
          }
        />
        <Route
          path="audit-logs"
          element={
            <Suspense fallback={<RouteFallback />}>
              <AuditLogsPage />
            </Suspense>
          }
        />
        <Route
          path="settings"
          element={
            <Suspense fallback={<RouteFallback />}>
              <SettingsPage />
            </Suspense>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="dark" enableColorScheme>
      <BrowserRouter>
        <AppRoutes />
        <Toaster position="top-right" richColors />
      </BrowserRouter>
    </ThemeProvider>
  )
}
