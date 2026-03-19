import { useState } from 'react'
import { Link, useLocation } from 'react-router'
import {
  LayoutDashboard,
  FolderGit2,
  Server,
  Users,
  FileText,
  Settings,
  Rocket,
  PanelLeftClose,
  PanelLeft,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent } from '@/components/ui/sheet'
import { useAuthStore } from '@/stores/auth-store'

const NAV_ITEMS = [
  { path: '/', label: '仪表盘', icon: LayoutDashboard, roles: [] },
  { path: '/projects', label: '项目', icon: FolderGit2, roles: [] },
  { path: '/servers', label: '服务器', icon: Server, roles: [] },
  { path: '/users', label: '用户', icon: Users, roles: ['admin'] },
  { path: '/audit-logs', label: '审计日志', icon: FileText, roles: ['admin', 'ops'] },
  { path: '/settings', label: '设置', icon: Settings, roles: ['admin'] },
] as const

function NavContent({ collapsed, onClose }: { collapsed: boolean; onClose?: () => void }) {
  const location = useLocation()
  const user = useAuthStore((s) => s.user)

  const visibleItems = NAV_ITEMS.filter((item) => {
    if (item.roles.length === 0) return true
    return user && (item.roles as readonly string[]).includes(user.role)
  })

  return (
    <nav className="flex flex-1 flex-col gap-1 px-2 py-4">
      {visibleItems.map((item) => {
        const isActive =
          item.path === '/'
            ? location.pathname === '/'
            : location.pathname.startsWith(item.path)
        const Icon = item.icon
        return (
          <Link
            key={item.path}
            to={item.path}
            onClick={onClose}
            className={cn(
              'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
              isActive
                ? 'bg-zinc-700 text-white'
                : 'text-zinc-300 hover:bg-zinc-800 hover:text-white'
            )}
          >
            <Icon className="size-5 shrink-0" />
            {!collapsed && <span>{item.label}</span>}
          </Link>
        )
      })}
    </nav>
  )
}

function SidebarContent({
  collapsed,
  onToggleCollapse,
}: {
  collapsed: boolean
  onToggleCollapse?: () => void
}) {
  return (
    <div className="flex h-full w-full flex-col bg-zinc-950">
      <div className="flex h-14 shrink-0 items-center gap-2 border-b border-zinc-800 px-4">
        <Rocket className="size-6 text-white" />
        {!collapsed && (
          <span className="text-lg font-semibold tracking-tight text-white">
            BuildFlow
          </span>
        )}
      </div>
      <NavContent collapsed={collapsed} />
      {onToggleCollapse && (
        <div className="mt-auto border-t border-zinc-800 p-2">
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-full justify-start text-zinc-400 hover:bg-zinc-800 hover:text-white"
            onClick={onToggleCollapse}
          >
            {collapsed ? (
              <PanelLeft className="size-4" />
            ) : (
              <>
                <PanelLeftClose className="size-4" />
                <span className="ml-2 text-sm">Collapse</span>
              </>
            )}
          </Button>
        </div>
      )}
    </div>
  )
}

interface SidebarProps {
  mobileOpen?: boolean
  onMobileOpenChange?: (open: boolean) => void
}

export function Sidebar({ mobileOpen = false, onMobileOpenChange }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <>
      {/* Desktop sidebar */}
      <aside
        className={cn(
          'hidden shrink-0 flex-col border-r border-zinc-800 bg-zinc-950 transition-all duration-200 md:flex',
          collapsed ? 'w-[64px]' : 'w-[220px]'
        )}
      >
        <div className="flex h-full flex-col">
          <SidebarContent
            collapsed={collapsed}
            onToggleCollapse={() => setCollapsed(!collapsed)}
          />
        </div>
      </aside>

      {/* Mobile: sheet overlay - controlled by Header menu button */}
      <Sheet open={mobileOpen} onOpenChange={onMobileOpenChange}>
        <SheetContent
          side="left"
          className="w-[280px] border-zinc-800 bg-zinc-950 p-0"
          showCloseButton={true}
        >
          <div className="flex h-full flex-col">
            <div className="flex h-14 shrink-0 items-center gap-2 border-b border-zinc-800 px-4">
              <Rocket className="size-6 text-white" />
              <span className="text-lg font-semibold tracking-tight text-white">
                BuildFlow
              </span>
            </div>
            <NavContent
              collapsed={false}
              onClose={() => onMobileOpenChange?.(false)}
            />
          </div>
        </SheetContent>
      </Sheet>
    </>
  )
}
