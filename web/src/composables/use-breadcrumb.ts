import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { BreadcrumbItem } from "@veltra/desktop";

import type { MenuNode } from "@/api/types";
import { useAuthStore } from "@/stores/auth";

export type AppBreadcrumbItem = BreadcrumbItem & { path?: string };

function menuRoute(node: MenuNode): string {
  return node.route || `/${node.path.replace(/\./g, "/")}`;
}

/** Longest-prefix match through the menu tree; returns root → leaf chain. */
function matchMenuChain(nodes: MenuNode[], path: string): MenuNode[] | null {
  let best: MenuNode[] | null = null;
  let bestLen = -1;

  function walk(list: MenuNode[], chain: MenuNode[]) {
    for (const node of list) {
      const next = [...chain, node];
      const route = menuRoute(node);
      if (path === route || path.startsWith(`${route}/`)) {
        if (route.length > bestLen) {
          best = next;
          bestLen = route.length;
        }
      }
      if (node.children?.length) {
        walk(node.children, next);
      }
    }
  }

  walk(nodes, []);
  return best;
}

export function resolveRouteTitle(
  route: { path: string; meta: Record<string, unknown>; name?: string | symbol | null },
  menus: MenuNode[],
): string {
  if (route.path === "/" || route.path === "") return "首页";

  const metaTitle = route.meta.title;
  const chain = matchMenuChain(menus, route.path);
  if (chain?.length) {
    const leaf = chain[chain.length - 1]!;
    const leafRoute = menuRoute(leaf);
    if (route.path === leafRoute) return leaf.title;
    if (typeof metaTitle === "string" && metaTitle) return metaTitle;
    return leaf.title;
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

    const chain = matchMenuChain(auth.menus, path);
    if (chain?.length) {
      const crumbs: AppBreadcrumbItem[] = chain.map((node) => ({
        title: node.title,
        path: menuRoute(node),
      }));
      const leafRoute = menuRoute(chain[chain.length - 1]!);
      if (path !== leafRoute && path.startsWith(`${leafRoute}/`)) {
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
