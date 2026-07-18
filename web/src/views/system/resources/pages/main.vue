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
  <u-tabs v-model="tab" :items="tabs" class="resources-page">
    <template #resources>
      <ResourcesPanel />
    </template>

    <template #groups>
      <MenuGroupsPanel />
    </template>
  </u-tabs>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.resources-page {
  height: 100%;
}
</style>
