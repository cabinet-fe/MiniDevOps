import { defineStore } from "pinia";
import { computed, ref } from "vue";

import { listNotifications, markAllNotificationsRead, markNotificationRead } from "@/api/system";
import type { NotificationItem } from "@/api/types";

export const useNotificationStore = defineStore("notification", () => {
  const items = ref<NotificationItem[]>([]);
  const loading = ref(false);

  const unreadCount = computed(() => items.value.filter((n) => !n.is_read).length);

  async function fetchNotifications(): Promise<void> {
    loading.value = true;
    try {
      const page = await listNotifications({ page: 1, page_size: 50 });
      items.value = page?.items ?? [];
    } finally {
      loading.value = false;
    }
  }

  function addNotification(n: NotificationItem): void {
    if (items.value.some((x) => x.id === n.id)) return;
    items.value = [n, ...items.value];
  }

  async function markRead(id: number): Promise<void> {
    await markNotificationRead(id);
    items.value = items.value.map((n) => (n.id === id ? { ...n, is_read: true } : n));
  }

  async function markAllRead(): Promise<void> {
    await markAllNotificationsRead();
    items.value = items.value.map((n) => ({ ...n, is_read: true }));
  }

  function reset(): void {
    items.value = [];
  }

  return {
    items,
    loading,
    unreadCount,
    fetchNotifications,
    addNotification,
    markRead,
    markAllRead,
    reset,
  };
});
