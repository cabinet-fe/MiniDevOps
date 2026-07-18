import type { GroupNavGroup, NavItem } from "@veltra/desktop";

import type { MenuGroupNode } from "@/api/types";

/** Map /auth/me MenuGroupNode[] → @veltra/desktop GroupNavGroup. */
export function menuGroupsToGroupNav(groups: MenuGroupNode[] | undefined | null): GroupNavGroup[] {
  if (!groups?.length) return [];
  return groups.map((group) => ({
    title: group.title,
    children: (group.children ?? []).map(
      (child): NavItem => ({
        title: child.title,
        path: child.path,
        ...(child.icon ? { icon: child.icon } : {}),
      }),
    ),
  }));
}
