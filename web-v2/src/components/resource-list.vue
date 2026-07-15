<script setup lang="ts" generic="T extends Record<string, any>">
import { onMounted, ref, useSlots, watch } from "vue";
import type { TableColumn } from "@veltra/desktop";

import type { PageResult } from "@/api/types";

export type ResourceListFetcher<T> = (params: {
  page: number;
  page_size: number;
  [key: string]: unknown;
}) => Promise<PageResult<T> | { items: T[] }>;

const props = withDefaults(
  defineProps<{
    fetcher: ResourceListFetcher<T>;
    columns: TableColumn[];
    rowKey?: string;
    pageSize?: number;
    filters?: Record<string, unknown>;
    /** 表格高度；默认撑满列表区，表头随 UTable sticky 固定 */
    height?: string;
  }>(),
  {
    rowKey: "id",
    pageSize: 20,
    filters: () => ({}),
    height: "100%",
  },
);

const emit = defineEmits<{
  loaded: [items: T[]];
}>();

const slots = useSlots();
const loading = ref(false);
const items = ref<Record<string, any>[]>([]);
const page = ref(1);
const pageSize = ref(props.pageSize);
const total = ref(0);
const paginated = ref(false);
const loadedOnce = ref(false);

async function load() {
  loading.value = true;
  try {
    const res = await props.fetcher({
      page: page.value,
      page_size: pageSize.value,
      ...props.filters,
    });
    const list = (res.items ?? []) as T[];
    items.value = list;
    const maybePage = res as PageResult<T>;
    paginated.value = typeof maybePage.page === "number" && typeof maybePage.total === "number";
    if (paginated.value) {
      total.value = maybePage.total;
      page.value = maybePage.page;
      pageSize.value = maybePage.page_size;
    } else {
      total.value = list.length;
    }
    emit("loaded", list);
  } finally {
    loading.value = false;
    loadedOnce.value = true;
  }
}

function refresh() {
  return load();
}

function onPageSizeChange() {
  page.value = 1;
  void load();
}

watch(
  () => props.filters,
  () => {
    page.value = 1;
    void load();
  },
  { deep: true },
);

onMounted(() => {
  void load();
});

defineExpose({ refresh, load });
</script>

<template>
  <div class="resource-list">
    <div v-if="slots.filters" class="filters">
      <slot name="filters" :reload="refresh" />
    </div>

    <div
      class="table-wrap"
      :class="{ 'is-loading': loading }"
      :style="height !== '100%' ? { height, flex: 'none' } : undefined"
    >
      <u-table :columns="columns" :data="items" :row-key="rowKey" border stripe>
        <template v-for="(_, name) in slots" :key="name" #[name]="slotData">
          <slot
            v-if="name !== 'filters' && name !== 'empty'"
            :name="name"
            v-bind="slotData || {}"
          />
        </template>
        <template #empty>
          <slot name="empty">
            <u-empty :text="loading && !loadedOnce ? '加载中…' : '暂无数据'" />
          </slot>
        </template>
      </u-table>
    </div>

    <div v-if="paginated" class="pager">
      <u-paginator
        v-model:page-number="page"
        v-model:page-size="pageSize"
        :total="total"
        @change:page-number="load"
        @change:page-size="onPageSizeChange"
      />
    </div>
  </div>
</template>

<style scoped>
.resource-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: flex-end;
  flex-shrink: 0;
}

.table-wrap {
  flex: 1;
  min-height: 280px;
  position: relative;
}

.table-wrap :deep(.u-table) {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100% !important;
}

.table-wrap.is-loading::after {
  content: "";
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.35);
  pointer-events: none;
  z-index: 1;
}

.pager {
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
</style>
