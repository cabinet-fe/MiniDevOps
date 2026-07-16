<script setup lang="ts">
import type { TabItem } from "@veltra/desktop";
import { useRouter } from "vue-router";

import { useTabsStore } from "@/stores/tabs";

const tabsStore = useTabsStore();
const router = useRouter();

function activate(key: string) {
  if (key === tabsStore.activeKey) return;
  const tab = tabsStore.tabs.find((t) => t.key === key);
  if (!tab) return;
  void router.push(tab.path);
}

function handleClose(item: TabItem) {
  const closingActive = tabsStore.activeKey === item.key;
  tabsStore.close(item.key);
  if (closingActive) {
    void router.push(tabsStore.activeKey);
  }
}
</script>

<template>
  <div class="workspace-tabs">
    <u-tabs-horizontal
      :model-value="tabsStore.activeKey"
      :items="tabsStore.tabItems"
      closable
      block
      size="small"
      @update:model-value="activate"
      @close="handleClose"
    />
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.workspace-tabs {
  /* Whisper ledge under the thin rail — flush, no own bar chrome */
  flex-shrink: 0;
  min-width: 0;
  padding: 0 0 4px;
  background: transparent;

  :deep(.u-tabs-horizontal) {
    --u-tabs-header-bg: transparent;
  }
}
</style>
