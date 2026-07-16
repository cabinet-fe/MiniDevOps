<script setup lang="ts">
import { useBreadcrumb } from "@/composables/use-breadcrumb";

const { items, onClick } = useBreadcrumb();
</script>

<template>
  <u-breadcrumb class="app-breadcrumb" :items="items" @click="onClick" />
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.app-breadcrumb {
  min-width: 0;
  flex: 1;
  color: fn.use-var(text-color, assist);

  :deep(.u-breadcrumb) {
    font-size: fn.use-var(font-size-main, default);
  }

  :deep(.u-breadcrumb__item:last-child) {
    color: fn.use-var(text-color, title);
    font-weight: 500;
  }
}
</style>
