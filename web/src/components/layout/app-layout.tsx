import { useState } from 'react'
import { Outlet } from 'react-router'
import { Sidebar } from '@/components/layout/sidebar'
import { Header } from '@/components/layout/header'

export function AppLayout() {
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-background">
      <Sidebar
        mobileOpen={mobileSidebarOpen}
        onMobileOpenChange={setMobileSidebarOpen}
      />
      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Header onMenuClick={() => setMobileSidebarOpen(true)} />
        <main className="flex-1 overflow-auto">
          <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
