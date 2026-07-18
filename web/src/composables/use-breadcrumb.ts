import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { BreadcrumbItem } from "@veltra/desktop";

import type { MenuGroupNode, MenuItemNode } from "@/api/types";
import { useAuthStore } from "@/stores/auth";

export type AppBreadcrumbItem = BreadcrumbItem & { path?: string };

type MenuMatch = { group: MenuGroupNode; item: MenuItemNode };

/** Longest-prefix match across two-level menus; returns group + leaf. */
function matchMenu(groups: MenuGroupNode[], path: string): MenuMatch | null {
  let best: MenuMatch | null = null;
  let bestLen = -1;

  for (const group of groups) {
    for (const item of group.children ?? []) {
      const route = item.path;
      if (!route) continue;
      if (path === route || path.startsWith(`${route}/`)) {
        if (route.length > bestLen) {
          best = { group, item };
          bestLen = route.length;
        }
      }
    }
  }
  return best;
}

export function resolveRouteTitle(
  route: { path: string; meta: Record<string, unknown>; name?: string | symbol | null },
  menus: MenuGroupNode[],
): string {
  if (route.path === "/" || route.path === "") return "首页";

  const metaTitle = route.meta.title;
  const match = matchMenu(menus, route.path);
  if (match) {
    if (route.path === match.item.path) return match.item.title;
    if (typeof metaTitle === "string" && metaTitle) return metaTitle;
    return match.item.title;
  }

  if (typeof metaTitle === "string" && metaTitle) return metaTitle;
  const name = route.name;
  return typeof name === "string" ? name : "页面";
}

export function useBreadcrumb() {
  const route = useRoute();
  const router = useRouter();
  const auth = useAuthStore();

  const items = computed<AppBreadcrumbItem[]>(() => {
    const path = route.path;
    if (path === "/" || path === "") {
      return [{ title: "首页", path: "/" }];
    }

    const match = matchMenu(auth.menus, path);
    if (match) {
      const crumbs: AppBreadcrumbItem[] = [
        { title: match.group.title },
        { title: match.item.title, path: match.item.path },
      ];
      if (path !== match.item.path && path.startsWith(`${match.item.path}/`)) {
        const detailTitle =
          typeof route.meta.title === "string" && route.meta.title ? route.meta.title : "详情";
        crumbs.push({ title: detailTitle, path });
      }
      return crumbs;
    }

    return [
      { title: "首页", path: "/" },
      { title: resolveRouteTitle(route, auth.menus), path },
    ];
  });

  function onClick(item: AppBreadcrumbItem) {
    if (!item.path || item.path === route.path) return;
    void router.push(item.path);
  }

  return { items, onClick };
}
