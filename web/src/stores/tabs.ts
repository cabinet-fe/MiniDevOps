import { computed, ref } from "vue";
import { defineStore } from "pinia";

export type WorkspaceTab = {
  /** Stable identity = route.path. Sub-tab query changes update this tab. */
  key: string;
  /** Latest fullPath; layout tab navigation always uses this. */
  fullPath: string;
  title: string;
  /** Vue component name for keep-alive include */
  name: string;
  closable: boolean;
};

const HOME_TAB: WorkspaceTab = {
  key: "/",
  fullPath: "/",
  title: "首页",
  name: "HomePage",
  closable: false,
};

function keepAliveNameFromRoute(route: {
  name?: string | symbol | null;
  meta: Record<string, unknown>;
}): string {
  const metaName = route.meta.keepAliveName;
  if (typeof metaName === "string" && metaName) return metaName;
  const name = route.name;
  if (typeof name === "string" && name) {
    return name
      .split("-")
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join("");
  }
  return "AnonymousPage";
}

export const useTabsStore = defineStore("tabs", () => {
  const tabs = ref<WorkspaceTab[]>([{ ...HOME_TAB }]);
  const activeKey = ref("/");

  const cachedNames = computed(() => [...new Set(tabs.value.map((t) => t.name))]);

  const tabItems = computed(() =>
    tabs.value.map((t) => ({
      key: t.key,
      name: t.title,
      closable: t.closable,
    })),
  );

  function findByKey(key: string) {
    return tabs.value.find((t) => t.key === key);
  }

  function open(tab: Omit<WorkspaceTab, "closable"> & { closable?: boolean }) {
    const existing = findByKey(tab.key);
    if (existing) {
      existing.title = tab.title;
      existing.fullPath = tab.fullPath;
      existing.name = tab.name;
      activeKey.value = existing.key;
      return;
    }
    tabs.value.push({
      key: tab.key,
      fullPath: tab.fullPath,
      title: tab.title,
      name: tab.name,
      closable: tab.closable ?? tab.key !== "/",
    });
    activeKey.value = tab.key;
  }

  function close(key: string) {
    const idx = tabs.value.findIndex((t) => t.key === key);
    if (idx < 0) return;
    const tab = tabs.value[idx]!;
    if (!tab.closable) return;

    const wasActive = activeKey.value === key;
    tabs.value.splice(idx, 1);

    if (!tabs.value.length) {
      tabs.value.push({ ...HOME_TAB });
      activeKey.value = HOME_TAB.key;
      return;
    }

    if (wasActive) {
      const next = tabs.value[Math.min(idx, tabs.value.length - 1)]!;
      activeKey.value = next.key;
    }
  }

  function syncFromRoute(
    route: {
      fullPath: string;
      path: string;
      name?: string | symbol | null;
      meta: Record<string, unknown>;
    },
    title: string,
  ) {
    if (route.path === "/login") return;
    open({
      key: route.path,
      fullPath: route.fullPath,
      title,
      name: keepAliveNameFromRoute(route),
      closable: route.path !== "/",
    });
  }

  function reset() {
    tabs.value = [{ ...HOME_TAB }];
    activeKey.value = HOME_TAB.key;
  }

  return {
    tabs,
    activeKey,
    cachedNames,
    tabItems,
    findByKey,
    open,
    close,
    syncFromRoute,
    reset,
  };
});
