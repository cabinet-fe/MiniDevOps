import { defineStore } from "pinia";
import { computed, ref } from "vue";

import { clearTokens, loginApi, logoutApi, meApi, setTokens } from "@/api/auth";
import { getAccessToken } from "@/api/http";
import type { MenuNode, User } from "@/api/types";
import { encryptLoginPassword } from "@/lib/login-crypto";

/** Re-fetch /auth/me when menus/permissions may have changed (role edits, etc.). */
const ME_STALE_MS = 30_000;

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(getAccessToken());
  const user = ref<User | null>(null);
  const permissions = ref<string[]>([]);
  const menus = ref<MenuNode[]>([]);
  const lastMeAt = ref(0);
  let meInflight: Promise<void> | null = null;

  const isAuthenticated = computed(() => !!token.value);
  const isSuperAdmin = computed(() => !!user.value?.is_super_admin);

  async function login(username: string, password: string): Promise<void> {
    const passwordCipher = await encryptLoginPassword(password);
    const data = await loginApi(username, passwordCipher);
    setTokens(data.access_token, data.refresh_token);
    token.value = data.access_token;
    user.value = data.user;
    permissions.value = data.permissions ?? [];
    menus.value = data.menus ?? [];
    lastMeAt.value = Date.now();
    // Always hydrate from me so sidebar matches server-trimmed tree.
    await fetchMe();
  }

  async function logout(): Promise<void> {
    try {
      if (token.value) {
        await logoutApi();
      }
    } catch {
      // ignore network errors on logout
    }
    clearTokens();
    token.value = null;
    user.value = null;
    permissions.value = [];
    menus.value = [];
    lastMeAt.value = 0;
  }

  async function fetchMe(): Promise<void> {
    if (!getAccessToken()) {
      token.value = null;
      user.value = null;
      return;
    }
    if (meInflight) return meInflight;
    meInflight = (async () => {
      try {
        const me = await meApi();
        user.value = me.user;
        permissions.value = me.permissions ?? [];
        menus.value = me.menus ?? [];
        token.value = getAccessToken();
        lastMeAt.value = Date.now();
      } catch {
        clearTokens();
        token.value = null;
        user.value = null;
        permissions.value = [];
        menus.value = [];
        lastMeAt.value = 0;
      } finally {
        meInflight = null;
      }
    })();
    return meInflight;
  }

  /** Fetch me when missing or older than ME_STALE_MS. Pass force=true after admin role changes. */
  async function refreshMe(force = false): Promise<void> {
    if (!getAccessToken()) return;
    const stale = !user.value || Date.now() - lastMeAt.value >= ME_STALE_MS;
    if (force || stale) {
      await fetchMe();
    }
  }

  function clearSession(): void {
    clearTokens();
    token.value = null;
    user.value = null;
    permissions.value = [];
    menus.value = [];
    lastMeAt.value = 0;
  }

  function hasPermission(code: string): boolean {
    if (!code) return true;
    if (user.value?.is_super_admin) return true;
    return permissions.value.includes(code);
  }

  return {
    token,
    user,
    permissions,
    menus,
    lastMeAt,
    isAuthenticated,
    isSuperAdmin,
    login,
    logout,
    fetchMe,
    refreshMe,
    clearSession,
    hasPermission,
  };
});
