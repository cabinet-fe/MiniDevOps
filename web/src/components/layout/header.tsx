import { Link, useLocation } from 'react-router'
import { PanelLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { NotificationBell } from '@/components/notification-bell'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useAuthStore } from '@/stores/auth-store'

const ROUTE_TITLES: Record<string, string> = {
  '/': '仪表盘',
  '/projects': '项目',
  '/servers': '服务器',
  '/users': '用户',
  '/audit-logs': '审计日志',
  '/settings': '设置',
}

function getBreadcrumb(pathname: string): { label: string; href?: string }[] {
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) return [{ label: '仪表盘' }]

  const crumbs: { label: string; href?: string }[] = []
  let current = ''

  for (let i = 0; i < segments.length; i++) {
    current += '/' + segments[i]
    const key = current
    if (ROUTE_TITLES[key]) {
      crumbs.push({
        label: ROUTE_TITLES[key],
        href: i < segments.length - 1 ? key : undefined,
      })
    } else if (/^\d+$/.test(segments[i])) {
      crumbs.push({ label: `#${segments[i]}` })
    } else if (segments[i] === 'builds' && segments[i + 1]) {
      crumbs.push({ label: `Build #${segments[i + 1]}` })
      i++
    } else {
      crumbs.push({ label: segments[i] })
    }
  }

  return crumbs.length > 0 ? crumbs : [{ label: '仪表盘' }]
}

interface HeaderProps {
  onMenuClick?: () => void
}

export function Header({ onMenuClick }: HeaderProps) {
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const breadcrumb = getBreadcrumb(location.pathname)

  return (
    <header className="flex h-12 shrink-0 items-center justify-between border-b border-zinc-200/80 bg-white/80 px-4 backdrop-blur-sm dark:border-zinc-800/60 dark:bg-zinc-900/80">
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          className="size-8 md:hidden"
          onClick={onMenuClick}
        >
          <PanelLeft className="size-4" />
        </Button>
        <nav className="flex items-center gap-1 text-sm">
          {breadcrumb.map((crumb, i) => (
            <span key={i} className="flex items-center gap-1">
              {i > 0 && (
                <ChevronRight className="size-3.5 text-zinc-400 dark:text-zinc-600" />
              )}
              {crumb.href ? (
                <Link
                  to={crumb.href}
                  className="text-zinc-500 transition-colors hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-300"
                >
                  {crumb.label}
                </Link>
              ) : (
                <span
                  className={cn(
                    i === breadcrumb.length - 1
                      ? 'font-medium text-zinc-900 dark:text-white'
                      : 'text-zinc-500 dark:text-zinc-400'
                  )}
                >
                  {crumb.label}
                </span>
              )}
            </span>
          ))}
        </nav>
      </div>

      <div className="flex items-center gap-1.5">
        <NotificationBell />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-8 w-8 rounded-full">
              <Avatar className="size-7">
                <AvatarImage src={user?.avatar} alt={user?.display_name} />
                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-violet-600 text-xs text-white">
                  {user?.display_name?.slice(0, 2).toUpperCase() ?? 'U'}
                </AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-52">
            <div className="flex items-center gap-2.5 p-2.5">
              <Avatar className="size-8">
                <AvatarImage src={user?.avatar} alt={user?.display_name} />
                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-violet-600 text-xs text-white">
                  {user?.display_name?.slice(0, 2).toUpperCase() ?? 'U'}
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col">
                <p className="text-sm font-medium leading-tight">{user?.display_name}</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link to="/settings">设置</Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={() => logout()}>
              退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
