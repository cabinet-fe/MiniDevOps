import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import {
  Bell,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Info,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useWebSocket } from '@/hooks/use-websocket'
import { useNotificationStore } from '@/stores/notification-store'
import { toast } from 'sonner'

const NOTIFICATION_ICONS: Record<string, React.ElementType> = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
  build_success: CheckCircle2,
  build_failed: AlertCircle,
  build_cancelled: AlertTriangle,
  default: Bell,
}

function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) return 'Just now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffHour < 24) return `${diffHour}h ago`
  if (diffDay < 7) return `${diffDay}d ago`
  return date.toLocaleDateString()
}

export function NotificationBell() {
  const navigate = useNavigate()
  const {
    notifications,
    unreadCount,
    addNotification,
    fetchNotifications,
    markRead,
    markAllRead,
  } = useNotificationStore()

  useEffect(() => {
    fetchNotifications()
  }, [fetchNotifications])

  const hasToken = typeof window !== 'undefined' && !!localStorage.getItem('access_token')

  useWebSocket({
    url: '/ws/notifications',
    enabled: hasToken,
    onMessage: (raw) => {
      try {
        const notification = JSON.parse(raw) as {
          id: number
          type: string
          title: string
          message: string
          build_id: number | null
          project_id?: number
          environment_id?: number
          build_status?: string
          is_read?: boolean
          created_at?: string
        }

        if (!notification.id || !notification.type || !notification.title) return

        addNotification({
          ...notification,
          is_read: notification.is_read ?? false,
          created_at: notification.created_at ?? new Date().toISOString(),
        })

        if (notification.type === 'build_success') {
          toast.success(notification.title, { description: notification.message || '构建已完成' })
        } else if (notification.type === 'build_failed') {
          toast.error(notification.title, { description: notification.message || '构建失败' })
        } else {
          toast(notification.title, { description: notification.message || undefined })
        }
      } catch {
        // Ignore malformed notification payloads.
      }
    },
  })

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="size-5" />
          {unreadCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-medium text-white">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-96 p-0">
        <div className="flex items-center justify-between border-b p-3">
          <h3 className="font-semibold">站内通知</h3>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 text-xs"
              onClick={() => markAllRead()}
            >
              全部已读
            </Button>
          )}
        </div>
        <div className="max-h-[320px] overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center text-sm text-muted-foreground">
              <Bell className="mb-2 size-8 opacity-40" />
              <p>暂无通知</p>
            </div>
          ) : (
            <div className="divide-y">
              {notifications.map((n) => {
                const Icon =
                  NOTIFICATION_ICONS[n.type] ?? NOTIFICATION_ICONS.default
                return (
                  <button
                    key={n.id}
                    type="button"
                    onClick={() => {
                      if (!n.is_read) markRead(n.id)
                      if (n.build_id) {
                        navigate(`/builds/${n.build_id}`)
                      }
                    }}
                    className={cn(
                      'flex w-full items-start gap-3 p-3 text-left transition-colors hover:bg-accent',
                      !n.is_read && 'bg-muted/50'
                    )}
                  >
                    <Icon
                      className={cn(
                        'mt-0.5 size-4 shrink-0',
                        n.type === 'success' && 'text-green-600',
                        n.type === 'error' && 'text-red-600',
                        n.type === 'warning' && 'text-amber-600',
                        n.type === 'info' && 'text-blue-600'
                      )}
                    />
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium">{n.title}</p>
                      {n.message && (
                        <p className="mt-0.5 truncate text-xs text-muted-foreground">
                          {n.message}
                        </p>
                      )}
                      <p className="mt-1 text-xs text-muted-foreground">
                        {formatTimeAgo(n.created_at)}
                      </p>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
