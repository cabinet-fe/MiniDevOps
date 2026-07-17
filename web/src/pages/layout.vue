<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { NavItem } from "@veltra/desktop";
import {
  Books,
  DArrowLeft,
  DArrowRight,
  HouseFilled,
  Layers,
  Logout,
  Monitor,
  Setting,
  Tools,
} from "@veltra/icons/normal";
import type { DefineComponent } from "vue";

import AppBreadcrumb from "@/components/app-breadcrumb";
import AppWorkspaceTabs from "@/components/app-workspace-tabs";
import BrandLogo from "@/components/brand-logo";
import NotificationBell from "@/components/notification-bell";
import { resolveRouteTitle } from "@/composables/use-breadcrumb";
import { menuNodesToNavItems } from "@/lib/menu-nav";
import { useAuthStore } from "@/stores/auth";
import { useTabsStore } from "@/stores/tabs";

const auth = useAuthStore();
const tabsStore = useTabsStore();
const router = useRouter();
const route = useRoute();

const navCollapsed = ref(false);
const displayName = computed(() => auth.user?.display_name || auth.user?.username || "");
const nameInitial = computed(() => {
  const name = displayName.value.trim();
  return name ? name.charAt(0).toUpperCase() : "?";
});

const ROOT_ICONS: Record<string, DefineComponent> = {
  dashboard: HouseFilled as DefineComponent,
  ops: Tools as DefineComponent,
  cicd: Layers as DefineComponent,
  project: Books as DefineComponent,
  system: Setting as DefineComponent,
};

const navMenus = computed(() =>
  menuNodesToNavItems(auth.menus, ROOT_ICONS).map((item) => ({
    ...item,
    icon: item.icon || (Monitor as DefineComponent),
  })),
);

const currentPath = computed(() => route.path);

watch(
  () => [route.fullPath, auth.menus] as const,
  () => {
    tabsStore.syncFromRoute(route, resolveRouteTitle(route, auth.menus));
  },
  { immediate: true },
);

function onVisibility() {
  if (document.visibilityState === "visible") {
    void auth.refreshMe();
  }
}

onMounted(() => {
  document.addEventListener("visibilitychange", onVisibility);
});

onUnmounted(() => {
  document.removeEventListener("visibilitychange", onVisibility);
});

async function handleLogout() {
  await auth.logout();
  tabsStore.reset();
  await router.replace({ name: "login" });
}

function onNavClick(item: NavItem) {
  if (item.path && !item.disabled) {
    void router.push(item.path);
  }
}

function toggleNav() {
  navCollapsed.value = !navCollapsed.value;
}
</script>

<template>
  <div class="app-shell">
    <aside class="app-sidebar" :class="{ 'app-sidebar--collapsed': navCollapsed }">
      <div class="app-sidebar__brand">
        <BrandLogo :collapsed="navCollapsed" />
      </div>
      <u-nav
        class="app-nav"
        :menus="navMenus"
        :current-path="currentPath"
        :collapsed="navCollapsed"
        @item-click="onNavClick"
      />
      <button
        type="button"
        class="app-sidebar__fold"
        :title="navCollapsed ? '展开侧栏' : '折叠侧栏'"
        @click="toggleNav"
      >
        <u-icon :size="16">
          <DArrowRight v-if="navCollapsed" />
          <DArrowLeft v-else />
        </u-icon>
      </button>
    </aside>

    <div class="app-body">
      <!-- Thin continuous rail: crumb + quiet utilities on one height; tabs as whisper ledge -->
      <header class="app-rail">
        <div class="app-rail__bar">
          <AppBreadcrumb />
          <div class="app-rail__utils" role="group" aria-label="操作区">
            <NotificationBell />
            <span class="app-rail__identity">
              <span class="app-rail__avatar" aria-hidden="true">{{ nameInitial }}</span>
              <span class="user-name">{{ displayName }}</span>
            </span>
            <u-button text type="primary" class="app-rail__logout" @click="handleLogout">
              <u-icon :size="14">
                <Logout />
              </u-icon>
              退出
            </u-button>
          </div>
        </div>
        <div class="app-rail__tabs">
          <AppWorkspaceTabs />
        </div>
      </header>

      <main class="app-main">
        <router-view v-slot="{ Component, route: viewRoute }">
          <Transition name="fade" mode="out-in">
            <keep-alive :include="tabsStore.cachedNames">
              <component :is="Component" :key="viewRoute.fullPath" class="app-page" />
            </keep-alive>
          </Transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.app-shell {
  height: 100%;
  display: flex;
  overflow: hidden;
  background: fn.use-var(bg-color, bottom);
  color: fn.use-var(text-color, main);
}

.app-sidebar {
  --sidebar-width: 240px;

  flex-shrink: 0;
  width: var(--sidebar-width);
  min-width: var(--sidebar-width);
  max-width: var(--sidebar-width);
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: none;
  outline: none;
  background: fn.use-var(bg-color, bottom);
  box-shadow: 4px 0 24px rgb(0 0 0 / 28%);
  transition:
    width 0.2s ease,
    min-width 0.2s ease,
    max-width 0.2s ease;

  &--collapsed {
    --sidebar-width: 72px;
  }
}

.app-sidebar__brand {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  min-height: 56px;
  padding: 0 16px;
  border: none;
}

.app-sidebar--collapsed .app-sidebar__brand {
  justify-content: center;
  padding: 0 8px;
}

.app-nav {
  flex: 1;
  min-height: 0;
  width: 100%;
  overflow: hidden;
  border: none;

  :deep(*) {
    border: none;
  }
}

.app-sidebar__fold {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  margin: 0;
  padding: 0;
  border: none;
  outline: none;
  background: transparent;
  color: fn.use-var(text-color, second);
  cursor: pointer;

  &:hover {
    color: fn.use-var(color, primary);
    background: fn.use-var(bg-color, hover);
  }
}

.app-body {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: fn.use-var(bg-color, middle);
}

/* ── Thin continuous rail ──
   Single-height bar: breadcrumb left, quiet utils right; tabs whisper below */
.app-rail {
  --rail-pad-x: #{fn.use-var(gap, large)};

  position: relative;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: fn.use-var(bg-color, top);
  z-index: 2;
}

.app-rail__bar {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 44px;
  padding: 0 var(--rail-pad-x);
}

.app-rail__utils {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
}

.app-rail__identity {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  margin: 0 4px 0 2px;
  padding: 0 4px;
}

.app-rail__avatar {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: color-mix(in srgb, fn.use-var(color, primary) 28%, fn.use-var(bg-color, bottom));
  color: fn.use-var(text-color, title);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1;
}

.user-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: fn.use-var(text-color, second);
  font-size: fn.use-var(font-size-main, default);
  font-weight: 400;
}

.app-rail__logout {
  gap: 4px;
}

.app-rail__tabs {
  min-width: 0;
  padding: 0 var(--rail-pad-x) 0;
}

.app-main {
  height: 100%;
}

.app-page {
  gap: fn.use-var(gap, large);
  padding: fn.use-var(gap, large);
  height: 100%;
}
</style>
