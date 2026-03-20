import { create } from 'zustand'
import { api } from '@/lib/api'

interface Notification {
  id: number
  type: string
  title: string
  message: string
  build_id: number | null
  project_id?: number
  environment_id?: number
  build_status?: string
  is_read: boolean
  created_at: string
}

interface NotificationState {
  notifications: Notification[]
  unreadCount: number
  latestNotification: Notification | null
  fetchNotifications: () => Promise<void>
  markRead: (id: number) => Promise<void>
  markAllRead: () => Promise<void>
  addNotification: (n: Notification) => void
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  unreadCount: 0,
  latestNotification: null,
  
  fetchNotifications: async () => {
    const res = await api.get<{ items: Notification[]; total: number }>('/notifications?page_size=50')
    if (res.code === 0 && res.data) {
      const items = res.data.items || []
      set({ 
        notifications: items,
        unreadCount: items.filter(n => !n.is_read).length,
        latestNotification: items[0] ?? null,
      })
    }
  },
  
  markRead: async (id) => {
    await api.put(`/notifications/${id}/read`)
    set(state => ({
      notifications: state.notifications.map(n => n.id === id ? { ...n, is_read: true } : n),
      unreadCount: Math.max(0, state.unreadCount - 1),
    }))
  },
  
  markAllRead: async () => {
    await api.put('/notifications/read-all')
    set(state => ({
      notifications: state.notifications.map(n => ({ ...n, is_read: true })),
      unreadCount: 0,
    }))
  },
  
  addNotification: (n) => {
    set(state => ({
      notifications: state.notifications.some(item => item.id === n.id)
        ? state.notifications
        : [n, ...state.notifications],
      unreadCount: state.notifications.some(item => item.id === n.id)
        ? state.unreadCount
        : state.unreadCount + (n.is_read ? 0 : 1),
      latestNotification: n,
    }))
  },
}))

export type { Notification }
