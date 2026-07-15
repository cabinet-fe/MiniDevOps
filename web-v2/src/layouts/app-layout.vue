<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { NavItem } from "@veltra/desktop";
import { HouseFilled, Layers, Monitor, Setting, Tools } from "@veltra/icons/normal";
import type { DefineComponent } from "vue";

import { menuNodesToNavItems } from "@/lib/menu-nav";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const displayName = computed(() => auth.user?.display_name || auth.user?.username || "");

const ROOT_ICONS: Record<string, DefineComponent> = {
  dashboard: HouseFilled as DefineComponent,
  ops: Tools as DefineComponent,
  cicd: Layers as DefineComponent,
  system: Setting as DefineComponent,
};

const navMenus = computed(() =>
  menuNodesToNavItems(auth.menus, ROOT_ICONS).map((item) => ({
    ...item,
    icon: item.icon || (Monitor as DefineComponent),
  })),
);

const currentPath = computed(() => route.path);

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
  await router.replace({ name: "login" });
}

function onNavClick(item: NavItem) {
  if (item.path && !item.disabled) {
    void router.push(item.path);
  }
}
</script>

<template>
  <div class="app-shell">
    <u-dual-nav
      class="app-nav"
      :menus="navMenus"
      :current-path="currentPath"
      @item-click="onNavClick"
    />
    <div class="app-body">
      <header class="app-header">
        <div class="brand">Bedrock</div>
        <div class="header-right">
          <span class="user-name">{{ displayName }}</span>
          <u-button text type="primary" @click="handleLogout">退出</u-button>
        </div>
      </header>
      <main class="app-main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.app-shell {
  min-height: 100vh;
  display: flex;
  background: fn.use-var(bg-color, bottom);
  color: fn.use-var(text-color, main);
}

.app-nav {
  flex-shrink: 0;
  width: 260px;
  min-width: 260px;
  max-width: 260px;
  min-height: 100vh;
}

.app-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: fn.use-var(bg-color, middle);
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 fn.use-var(gap, large);
  border-bottom: fn.use-var(border);
  background: fn.use-var(bg-color, top);
  box-shadow: fn.use-var(shadow);
}

.brand {
  font-size: fn.use-var(font-size-title, default);
  font-weight: 600;
  color: fn.use-var(text-color, title);
  letter-spacing: 0.02em;
}

.header-right {
  display: flex;
  align-items: center;
  gap: fn.use-var(gap, default);
}

.user-name {
  color: fn.use-var(text-color, second);
  font-size: fn.use-var(font-size-main, default);
}

.app-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: fn.use-var(gap, large);
  overflow: hidden;

  :deep(> .page) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
    overflow: auto;
  }
}
</style>
