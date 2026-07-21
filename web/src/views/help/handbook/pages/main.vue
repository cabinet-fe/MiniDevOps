<script setup lang="ts">
defineOptions({ name: "HelpHandbook" });

import { computed, ref } from "vue";

import MarkdownViewer from "@/components/markdown-viewer";
import { handbookSections } from "@/content/handbook/manifest";

const activeKey = ref(handbookSections[0]?.key ?? "");

const activeContent = computed(
  () => handbookSections.find((s) => s.key === activeKey.value)?.content ?? "",
);

const navItems = handbookSections.map((s) => ({
  key: s.key,
  name: s.title,
}));
</script>

<template>
  <div class="handbook">
    <aside class="handbook-nav">
      <u-tabs-vertical v-model="activeKey" :items="navItems" />
    </aside>
    <u-scroll class="handbook-body">
      <MarkdownViewer :content="activeContent" />
    </u-scroll>
  </div>
</template>

<style scoped lang="scss">
.handbook {
  display: grid;
  height: 100%;
  min-height: 0;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 16px;
}

.handbook-nav {
  min-width: 0;
  padding: 8px;
  border-radius: 8px;
  background: var(--u-bg-color-top, #fff);
}

.handbook-body {
  min-width: 0;
  min-height: 0;
  height: 100%;
}

@media (max-width: 900px) {
  .handbook {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(0, 1fr);
  }
}
</style>
