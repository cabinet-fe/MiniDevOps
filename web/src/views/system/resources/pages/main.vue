<script setup lang="ts">
defineOptions({ name: "SystemResources" });

import { ref } from "vue";

import MenuGroupsPanel from "../components/menu-groups-panel.vue";
import ResourcesPanel from "../components/resources-panel.vue";

const tab = ref("resources");
const tabs = [
  { key: "resources", name: "菜单与功能" },
  { key: "groups", name: "菜单分组" },
];
</script>

<template>
  <div class="resources-page">
    <u-tabs v-model="tab" :items="tabs" />
    <ResourcesPanel v-if="tab === 'resources'" />
    <MenuGroupsPanel v-else-if="tab === 'groups'" />
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.resources-page {
  display: flex;
  flex-direction: column;
  gap: fn.use-var(gap, default);
  min-height: 0;
  height: 100%;
}
</style>
