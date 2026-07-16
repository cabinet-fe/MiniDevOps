import type { NavItem } from "@veltra/desktop";
import type { DefineComponent } from "vue";

import type { MenuNode } from "@/api/types";

type NavIcon = string | DefineComponent;

/** Map /auth/me MenuNode → @veltra/desktop NavItem (path ← route). */
export function menuNodesToNavItems(
  nodes: MenuNode[] | undefined | null,
  rootIcons?: Record<string, NavIcon>,
): NavItem[] {
  if (!nodes?.length) return [];
  return nodes.map((node) => toNavItem(node, rootIcons));
}

function toNavItem(node: MenuNode, rootIcons?: Record<string, NavIcon>): NavItem {
  const route = node.route || `/${node.path.replace(/\./g, "/")}`;
  const item: NavItem = {
    title: node.title,
    path: route,
  };
  if (node.icon) {
    item.icon = node.icon;
  } else if (rootIcons?.[node.path]) {
    item.icon = rootIcons[node.path];
  }
  if (node.children?.length) {
    item.children = node.children.map((child) => toNavItem(child));
  }
  return item;
}
