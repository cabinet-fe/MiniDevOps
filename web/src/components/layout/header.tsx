import { Link, useLocation } from 'react-router'
import { PanelLeft } from 'lucide-react'
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
  '/projects/new': '新建项目',
  '/projects/edit': '编辑项目',
  '/servers': '服务器',
  '/servers/new': '新建服务器',
  '/servers/edit': '编辑服务器',
  '/users': '用户',
  '/audit-logs': '审计日志',
  '/settings': '设置',
}

function getBreadcrumb(pathname: string): string[] {
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) return ['仪表盘']

  const titles: string[] = []
  let current = ''

  for (let i = 0; i < segments.length; i++) {
    current += '/' + segments[i]
    const key = current
    if (ROUTE_TITLES[key]) {
      titles.push(ROUTE_TITLES[key])
    } else if (segments[i] === 'new') {
      titles.push('新建')
    } else if (segments[i] === 'edit') {
      titles.push('编辑')
    } else if (/^\d+$/.test(segments[i])) {
      titles.push(segments[i])
    } else if (segments[i] === 'builds' && segments[i + 1]) {
      titles.push(`Build #${segments[i + 1]}`)
      i++
    } else {
      titles.push(segments[i])
    }
  }

  return titles.length > 0 ? titles : ['仪表盘']
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
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-zinc-200 bg-white px-4 dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          className="md:hidden"
          onClick={onMenuClick}
        >
          <PanelLeft className="size-4" />
        </Button>
        <nav className="flex items-center gap-2 text-sm">
          {breadcrumb.map((part, i) => (
            <span key={i} className="flex items-center gap-2">
              {i > 0 && (
                <span className="text-zinc-400 dark:text-zinc-500">/</span>
              )}
              <span
                className={cn(
                  i === breadcrumb.length - 1
                    ? 'font-medium text-zinc-900 dark:text-white'
                    : 'text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-300'
                )}
              >
                {part}
              </span>
            </span>
          ))}
        </nav>
      </div>

      <div className="flex items-center gap-2">
        <NotificationBell />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-9 w-9 rounded-full">
              <Avatar className="size-8">
                <AvatarImage src={user?.avatar} alt={user?.display_name} />
                <AvatarFallback className="bg-zinc-200 text-zinc-700 dark:bg-zinc-700 dark:text-zinc-200">
                  {user?.display_name?.slice(0, 2).toUpperCase() ?? 'U'}
                </AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <div className="flex items-center gap-2 p-2">
              <div className="flex flex-col">
                <p className="text-sm font-medium">{user?.display_name}</p>
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
