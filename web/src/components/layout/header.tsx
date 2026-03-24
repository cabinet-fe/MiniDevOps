import { Link, useLocation } from 'react-router'
import {
  PanelLeft,
  ChevronRight,
  Sun,
  Moon,
  LayoutDashboard,
  FolderGit2,
  Server,
  Users,
  FileText,
  Settings,
  BookOpenText,
  KeyRound,
  BookOpen,
  Hammer,
  Layers,
  type LucideIcon,
} from 'lucide-react'
import { useTheme } from 'next-themes'
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
import { Separator } from '@/components/ui/separator'
import { useAuthStore } from '@/stores/auth-store'

const ROUTE_META: Record<string, { label: string; icon: LucideIcon }> = {
  '/': { label: '仪表盘', icon: LayoutDashboard },
  '/projects': { label: '项目', icon: FolderGit2 },
  '/environments': { label: '环境', icon: Layers },
  '/credentials': { label: '凭证', icon: KeyRound },
  '/servers': { label: '服务器', icon: Server },
  '/users': { label: '用户', icon: Users },
  '/audit-logs': { label: '审计日志', icon: FileText },
  '/settings': { label: '设置', icon: Settings },
  '/manual': { label: '项目手册', icon: BookOpenText },
  '/dictionaries': { label: '数据字典', icon: BookOpen },
}

function getBreadcrumb(pathname: string): { label: string; href?: string; icon?: LucideIcon }[] {
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) return [{ label: '仪表盘', icon: LayoutDashboard }]

  const crumbs: { label: string; href?: string; icon?: LucideIcon }[] = []
  let current = ''

  for (let i = 0; i < segments.length; i++) {
    current += '/' + segments[i]
    const key = current
    const meta = ROUTE_META[key]
    if (meta) {
      crumbs.push({
        label: meta.label,
        icon: i === 0 ? meta.icon : undefined,
        href: i < segments.length - 1 ? key : undefined,
      })
    } else if (/^\d+$/.test(segments[i])) {
      crumbs.push({ label: `#${segments[i]}` })
    } else if (segments[i] === 'builds' && segments[i + 1]) {
      crumbs.push({ label: `Build #${segments[i + 1]}`, icon: Hammer })
      i++
    } else {
      crumbs.push({ label: segments[i] })
    }
  }

  return crumbs.length > 0 ? crumbs : [{ label: '仪表盘', icon: LayoutDashboard }]
}

interface HeaderProps {
  onMenuClick?: () => void
}

export function Header({ onMenuClick }: HeaderProps) {
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const breadcrumb = getBreadcrumb(location.pathname)
  const { theme, setTheme } = useTheme()

  return (
    <header className="flex h-12 shrink-0 items-center justify-between border-b border-border/60 bg-background/80 px-4 shadow-[0_1px_3px_0_rgba(0,0,0,0.04)] backdrop-blur-md dark:shadow-[0_1px_3px_0_rgba(0,0,0,0.2)]">
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          className="size-8 md:hidden"
          onClick={onMenuClick}
        >
          <PanelLeft className="size-4" />
        </Button>
        <nav className="flex items-center gap-1.5 text-sm">
          {breadcrumb.map((crumb, i) => {
            const Icon = crumb.icon
            return (
              <span key={i} className="flex items-center gap-1.5">
                {i > 0 && (
                  <ChevronRight className="size-3 text-muted-foreground/40" />
                )}
                {crumb.href ? (
                  <Link
                    to={crumb.href}
                    className="flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-foreground"
                  >
                    {Icon && <Icon className="size-3.5 text-emerald-500/70" />}
                    {crumb.label}
                  </Link>
                ) : (
                  <span
                    className={cn(
                      'flex items-center gap-1.5',
                      i === breadcrumb.length - 1
                        ? 'font-semibold text-foreground'
                        : 'text-muted-foreground'
                    )}
                  >
                    {Icon && (
                      <Icon className={cn(
                        'size-3.5',
                        i === breadcrumb.length - 1 ? 'text-emerald-500' : 'text-emerald-500/70'
                      )} />
                    )}
                    {crumb.label}
                  </span>
                )}
              </span>
            )
          })}
        </nav>
      </div>

      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          className="size-8 text-muted-foreground hover:bg-accent hover:text-foreground"
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        >
          <Sun className="size-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute size-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
          <span className="sr-only">切换主题</span>
        </Button>
        <NotificationBell />

        <Separator orientation="vertical" className="mx-1.5 h-5" />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-8 gap-2 rounded-full pl-1 pr-2">
              <Avatar className="size-6">
                <AvatarImage src={user?.avatar} alt={user?.display_name} />
                <AvatarFallback className="bg-gradient-to-br from-emerald-500 to-teal-600 text-[10px] font-medium text-white">
                  {user?.display_name?.slice(0, 2).toUpperCase() ?? 'U'}
                </AvatarFallback>
              </Avatar>
              <span className="hidden text-sm font-medium text-foreground md:inline">
                {user?.display_name}
              </span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-52">
            <div className="flex items-center gap-2.5 p-2.5">
              <Avatar className="size-8">
                <AvatarImage src={user?.avatar} alt={user?.display_name} />
                <AvatarFallback className="bg-gradient-to-br from-emerald-500 to-teal-600 text-xs text-white">
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
