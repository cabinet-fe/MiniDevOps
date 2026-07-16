<script setup lang="ts" generic="T extends Record<string, any>">
import { computed, h, onMounted, ref, useSlots, watch } from "vue";
import { o } from "@cat-kit/core";
import {
  message,
  UButton,
  type TableColumn,
  type TableColumnNode,
  vLoading,
} from "@veltra/desktop";
import { ArrowDown, ArrowUp, ArrowUpdown } from "@veltra/icons/normal";

import { http } from "@/api/http";

export type ProTableQuery = Record<string, unknown>;

/** Column config; `sortable` adds a header sort control (UTable has no built-in sort). */
export type ProTableColumn = TableColumn & {
  sortable?: boolean;
};

type SortOrder = "asc" | "desc";

const props = withDefaults(
  defineProps<{
    /** API path relative to `/api/v1` (e.g. `/users`) */
    url: string;
    columns: ProTableColumn[];
    /** Fields that auto-trigger search when changed (selects, etc.) */
    autoQueryFields?: string[];
    /**
     * Path into the unwrapped response body for row data.
     * Envelope plugin unwraps `{ data }` first; default is `items`.
     */
    dataPath?: string;
    /** Enable pagination (mutually exclusive with `tree`) */
    pagination?: boolean;
    /**
     * Enable tree table. `true` uses default children key;
     * a string overrides the children field name.
     * Mutually exclusive with `pagination`.
     */
    tree?: boolean | string;
    rowKey?: string;
    pageSize?: number;
    /** Table area height; default fills remaining space */
    height?: string;
    /** Load on mount */
    immediate?: boolean;
    /** Expand all tree nodes by default */
    defaultExpandAll?: boolean;
  }>(),
  {
    autoQueryFields: () => [],
    dataPath: "items",
    pagination: false,
    tree: false,
    rowKey: "id",
    pageSize: 20,
    height: "100%",
    immediate: true,
    defaultExpandAll: false,
  },
);

const emit = defineEmits<{
  loaded: [items: T[]];
}>();

/**
 * Filter / pagination / sort query bound to the parent.
 * Sort is written as `sort: "<field>@asc" | "<field>@desc"` and omitted when cleared.
 */
const query = defineModel<ProTableQuery>("query", { default: () => ({}) });

const slots = useSlots();
const loading = ref(false);
const items = ref<T[]>([]);
const page = ref(1);
const pageSize = ref(props.pageSize);
const total = ref(0);
const loadedOnce = ref(false);

const mode = computed(() => {
  if (props.pagination && props.tree) {
    console.warn("[ProTable] `pagination` and `tree` are mutually exclusive; using pagination.");
    return "pagination" as const;
  }
  if (props.pagination) return "pagination" as const;
  if (props.tree) return "tree" as const;
  return "list" as const;
});

const tableTree = computed(() => {
  if (mode.value !== "tree") return false;
  return props.tree === true ? true : props.tree;
});

const hasFilters = computed(() => !!slots.filters);

function parseSort(value: unknown): { field: string; order: SortOrder | null } {
  if (typeof value !== "string" || !value.includes("@")) {
    return { field: "", order: null };
  }
  const at = value.lastIndexOf("@");
  const field = value.slice(0, at);
  const order = value.slice(at + 1);
  if (!field || (order !== "asc" && order !== "desc")) {
    return { field: "", order: null };
  }
  return { field, order };
}

function cycleSort(field: string) {
  const current = parseSort(query.value?.sort);
  let next: string | undefined;
  if (current.field !== field || !current.order) {
    next = `${field}@desc`;
  } else if (current.order === "desc") {
    next = `${field}@asc`;
  } else {
    next = undefined;
  }

  if (next) {
    query.value.sort = next;
  } else {
    delete query.value.sort;
  }
  void search();
}

function mapColumns(cols: ProTableColumn[]): TableColumn[] {
  const { field: sortField, order: sortOrder } = parseSort(query.value?.sort);

  return cols.map((col) => {
    const children = col.children?.length
      ? mapColumns(col.children as ProTableColumn[])
      : undefined;

    if (!col.sortable) {
      return children ? { ...col, children } : col;
    }

    const originalNameRender = col.nameRender;
    return {
      ...col,
      children,
      nameRender: (ctx: { column: TableColumnNode }) => {
        const label = originalNameRender?.(ctx) ?? col.name;
        const active = sortField === col.key && !!sortOrder;
        const Icon =
          active && sortOrder === "asc"
            ? ArrowUp
            : active && sortOrder === "desc"
              ? ArrowDown
              : ArrowUpdown;

        return h(
          "span",
          {
            class: ["pro-table__th", active && "is-sorted"],
            onClick: (e: MouseEvent) => {
              e.stopPropagation();
              cycleSort(col.key);
            },
          },
          [
            h("span", { class: "pro-table__th-label" }, label as any),
            h(UButton, {
              text: true,
              circle: true,
              icon: Icon,
              iconSize: 14,
              class: ["pro-table__sort-btn", active && "is-active"],
              // Visual indicator only; parent span owns the click cycle.
              tabindex: -1,
              style: { pointerEvents: "none" },
            }),
          ],
        );
      },
    };
  });
}

const resolvedColumns = computed(() => mapColumns(props.columns));

function cleanQuery(params: Record<string, unknown>): Record<string, string | number | boolean> {
  const out: Record<string, string | number | boolean> = {};
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") continue;
    if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") {
      out[k] = v;
    }
  }
  return out;
}

function extractItems(body: Record<string, unknown>): T[] {
  const extracted = o(body).get(props.dataPath);
  if (Array.isArray(extracted)) return extracted as T[];
  if (
    extracted &&
    typeof extracted === "object" &&
    Array.isArray((extracted as { items?: T[] }).items)
  ) {
    return (extracted as { items: T[] }).items;
  }
  return [];
}

function applyPaginationMeta(body: Record<string, unknown>) {
  if (typeof body.total === "number") total.value = body.total;
  if (typeof body.page === "number") page.value = body.page;
  if (typeof body.page_size === "number") pageSize.value = body.page_size;
}

async function load() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = { ...query.value };
    if (mode.value === "pagination") {
      params.page = page.value;
      params.page_size = pageSize.value;
    }

    // Envelope plugin unwraps `{ code, message, data }` → body is `data`.
    const { body: raw } = await http.get(props.url, { query: cleanQuery(params) });
    const body = (raw ?? {}) as Record<string, unknown>;

    const list = extractItems(body);
    items.value = list;

    if (mode.value === "pagination") {
      applyPaginationMeta(body);
    } else {
      total.value = list.length;
    }

    emit("loaded", list);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
    items.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
    loadedOnce.value = true;
  }
}

/** Reset to page 1 and fetch (manual search / Enter / submit / sort). */
function search() {
  page.value = 1;
  return load();
}

/** Re-fetch current page (after create/update/delete). */
function reload() {
  return load();
}

function onPageSizeChange() {
  page.value = 1;
  void load();
}

watch(
  () =>
    props.autoQueryFields
      .map((key) => `${key}:${JSON.stringify(query.value?.[key] ?? null)}`)
      .join("|"),
  () => {
    if (!props.autoQueryFields.length) return;
    page.value = 1;
    void load();
  },
);

watch(
  () => props.url,
  () => {
    page.value = 1;
    void load();
  },
);

onMounted(() => {
  if (props.immediate) void load();
});

defineExpose({ search, reload, load });
</script>

<template>
  <div class="pro-table">
    <form v-if="hasFilters" class="pro-table__toolbar" @submit.prevent="search">
      <div class="pro-table__filters">
        <slot name="filters" :search="search" :reload="reload" :query="query" />
      </div>
    </form>

    <div class="pro-table__panel" :style="height !== '100%' ? { height, flex: 'none' } : undefined">
      <div v-loading="loading" class="pro-table__body">
        <u-table
          :columns="resolvedColumns"
          :data="items"
          :border="false"
          :row-key="rowKey"
          :tree="tableTree"
          :default-expand-all="defaultExpandAll"
          :stripe="mode !== 'tree'"
        >
          <template v-for="(_, name) in slots" :key="name" #[name]="slotData">
            <slot
              v-if="name !== 'filters' && name !== 'empty'"
              :name="name"
              v-bind="slotData || {}"
            />
          </template>
        </u-table>
      </div>

      <div v-if="mode === 'pagination'" class="pro-table__footer">
        <u-paginator
          v-model:page-number="page"
          v-model:page-size="pageSize"
          :total="total"
          @change:page-number="load"
          @change:page-size="onPageSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.pro-table {
  display: flex;
  flex-direction: column;
  gap: fn.use-var(gap, default);
  flex: 1;
  min-height: 0;
}

.pro-table__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: flex-end;
  gap: fn.use-var(gap, default);
  padding: 0 0 fn.use-var(gap, small);
}

.pro-table__filters {
  display: flex;
  flex-wrap: wrap;
  gap: fn.use-var(gap, small);
  align-items: flex-end;
  flex: 1;
  min-width: 0;
}

.pro-table__panel {
  flex: 1;
  min-height: 280px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border-radius: fn.use-var(radius, default);
  background: fn.use-var(bg-color, top);
  overflow: hidden;
}

.pro-table__body {
  flex: 1;
  min-height: 0;
  position: relative;
}

.pro-table__body :deep(.u-table) {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100% !important;
}

.pro-table__footer {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: fn.use-var(gap, small) 0 0;
  border-top: fn.use-var(border, muted);
}

.pro-table__th {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  max-width: 100%;
  cursor: pointer;
  user-select: none;
}

.pro-table__th-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pro-table__th.is-sorted .pro-table__th-label {
  color: fn.use-var(color, primary);
}

.pro-table__sort-btn {
  flex-shrink: 0;
  color: fn.use-var(text-color, assist);
  opacity: 0.7;
}

.pro-table__sort-btn.is-active {
  color: fn.use-var(color, primary);
  opacity: 1;
}

.pro-table__th:hover .pro-table__sort-btn {
  opacity: 1;
}
</style>
