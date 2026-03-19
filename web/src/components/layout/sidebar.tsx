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
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent } from '@/components/ui/sheet'
import { Separator } from '@/components/ui/separator'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useAuthStore } from '@/stores/auth-store'

const NAV_GROUPS = [
  {
    label: '概览',
    items: [
      { path: '/', label: '仪表盘', icon: LayoutDashboard, roles: [] as string[] },
    ],
  },
  {
    label: '资源',
    items: [
      { path: '/projects', label: '项目', icon: FolderGit2, roles: [] as string[] },
      { path: '/servers', label: '服务器', icon: Server, roles: [] as string[] },
    ],
  },
  {
    label: '管理',
    items: [
      { path: '/users', label: '用户', icon: Users, roles: ['admin'] },
      { path: '/audit-logs', label: '审计日志', icon: FileText, roles: ['admin', 'ops'] },
      { path: '/settings', label: '设置', icon: Settings, roles: ['admin'] },
    ],
  },
] as const

function NavContent({ collapsed, onClose }: { collapsed: boolean; onClose?: () => void }) {
  const location = useLocation()
  const user = useAuthStore((s) => s.user)

  const visibleGroups = NAV_GROUPS.map((group) => ({
    ...group,
    items: group.items.filter((item) => {
      if (item.roles.length === 0) return true
      return user && (item.roles as readonly string[]).includes(user.role)
    }),
  })).filter((group) => group.items.length > 0)

  return (
    <TooltipProvider delayDuration={0}>
      <nav className="flex flex-1 flex-col gap-1 px-3 py-4">
        {visibleGroups.map((group, groupIdx) => (
          <div key={group.label}>
            {groupIdx > 0 && <Separator className="my-3 bg-zinc-800/60" />}
            {!collapsed && (
              <p className="mb-2 px-3 text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
                {group.label}
              </p>
            )}
            <div className="flex flex-col gap-0.5">
              {group.items.map((item) => {
                const isActive =
                  item.path === '/'
                    ? location.pathname === '/'
                    : location.pathname.startsWith(item.path)
                const Icon = item.icon

                const linkContent = (
                  <Link
                    key={item.path}
                    to={item.path}
                    onClick={onClose}
                    className={cn(
                      'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-150',
                      collapsed && 'justify-center px-0',
                      isActive
                        ? 'bg-white/10 text-white shadow-sm'
                        : 'text-zinc-400 hover:bg-white/5 hover:text-zinc-200'
                    )}
                  >
                    <Icon className={cn('size-[18px] shrink-0', isActive && 'text-blue-400')} />
                    {!collapsed && <span>{item.label}</span>}
                  </Link>
                )

                if (collapsed) {
                  return (
                    <Tooltip key={item.path}>
                      <TooltipTrigger asChild>
                        {linkContent}
                      </TooltipTrigger>
                      <TooltipContent side="right" sideOffset={8}>
                        {item.label}
                      </TooltipContent>
                    </Tooltip>
                  )
                }

                return linkContent
              })}
            </div>
          </div>
        ))}
      </nav>
    </TooltipProvider>
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
      {/* Logo */}
      <div className={cn(
        'flex h-14 shrink-0 items-center border-b border-zinc-800/60 transition-all duration-200',
        collapsed ? 'justify-center px-2' : 'gap-2.5 px-4'
      )}>
        <div className="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-violet-600 shadow-lg shadow-blue-500/20">
          <Rocket className="size-4 text-white" />
        </div>
        {!collapsed && (
          <span className="text-[15px] font-semibold tracking-tight text-white">
            BuildFlow
          </span>
        )}
      </div>

      <NavContent collapsed={collapsed} />

      {/* Collapse toggle */}
      {onToggleCollapse && (
        <div className="border-t border-zinc-800/60 p-2">
          <Button
            variant="ghost"
            size="sm"
            className={cn(
              'h-8 w-full text-zinc-500 hover:bg-white/5 hover:text-zinc-300',
              collapsed ? 'justify-center px-0' : 'justify-start gap-2 px-3',
            )}
            onClick={onToggleCollapse}
          >
            {collapsed ? (
              <ChevronsRight className="size-4" />
            ) : (
              <>
                <ChevronsLeft className="size-4" />
                <span className="text-xs">收起侧栏</span>
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
          'hidden shrink-0 flex-col border-r border-zinc-800/60 bg-zinc-950 transition-all duration-200 ease-in-out md:flex',
          collapsed ? 'w-[60px]' : 'w-[220px]'
        )}
      >
        <div className="flex h-full flex-col">
          <SidebarContent
            collapsed={collapsed}
            onToggleCollapse={() => setCollapsed(!collapsed)}
          />
        </div>
      </aside>

      {/* Mobile: sheet overlay */}
      <Sheet open={mobileOpen} onOpenChange={onMobileOpenChange}>
        <SheetContent
          side="left"
          className="w-[260px] border-zinc-800/60 bg-zinc-950 p-0"
          showCloseButton={true}
        >
          <div className="flex h-full flex-col">
            <div className="flex h-14 shrink-0 items-center gap-2.5 border-b border-zinc-800/60 px-4">
              <div className="flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-violet-600 shadow-lg shadow-blue-500/20">
                <Rocket className="size-4 text-white" />
              </div>
              <span className="text-[15px] font-semibold tracking-tight text-white">
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
