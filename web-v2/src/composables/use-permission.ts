import { computed } from "vue";

import { useAuthStore } from "@/stores/auth";

/** Permission helpers for route guards and button-level checks. */
export function usePermission() {
  const auth = useAuthStore();

  const permissionSet = computed(() => new Set(auth.permissions));

  function hasPermission(code: string): boolean {
    if (!code) return true;
    if (auth.user?.is_super_admin) return true;
    return permissionSet.value.has(code);
  }

  function hasAnyPermission(...codes: string[]): boolean {
    return codes.some((c) => hasPermission(c));
  }

  return { hasPermission, hasAnyPermission, permissionSet };
}
