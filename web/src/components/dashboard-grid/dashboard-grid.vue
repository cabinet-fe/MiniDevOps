<script setup lang="ts">
defineOptions({ name: "DashboardGrid" });

import {
  h,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  render,
  useTemplateRef,
  watch,
} from "vue";
import { GridStack, type GridStackWidget } from "gridstack";
import "gridstack/dist/gridstack.min.css";

import type {
  AgentRunSummary,
  BuildSummary,
  DashboardCardID,
  DashboardCardLayout,
  SystemInfo,
  SystemStatus,
} from "@/api/types";

import DashboardWidgetHost from "./dashboard-widget-host.vue";
import {
  DASHBOARD_GRID_COLUMNS,
  geometryFromWidgets,
  toGridWidgets,
  visibleCardsSignature,
  type DashboardWidgetHostContext,
} from "./helper";

const props = defineProps<{
  items: DashboardCardLayout[];
  editing: boolean;
  buildSummary: BuildSummary | null;
  agentRunSummary: AgentRunSummary | null;
  systemInfo: SystemInfo | null;
  systemStatus: SystemStatus | null;
}>();

const emit = defineEmits<{
  change: [cards: DashboardCardLayout[]];
  openBuildRun: [id: number];
  openAgentRun: [id: number];
}>();

const gridEl = useTemplateRef<HTMLElement>("gridEl");

/** Plain instance — never put GridStack into ref/reactive (Proxy breaks internals). */
let grid: GridStack | null = null;
const mountHosts = new Map<string, HTMLElement>();
let syncingFromGrid = false;

/** Shared reactive context for renderCB-mounted hosts (stays live without remount). */
const hostCtx = reactive<DashboardWidgetHostContext>({
  editing: false,
  buildSummary: null,
  agentRunSummary: null,
  systemInfo: null,
  systemStatus: null,
  openBuildRun: (id: number) => emit("openBuildRun", id),
  openAgentRun: (id: number) => emit("openAgentRun", id),
});

function syncHostCtx() {
  hostCtx.editing = props.editing;
  hostCtx.buildSummary = props.buildSummary;
  hostCtx.agentRunSummary = props.agentRunSummary;
  hostCtx.systemInfo = props.systemInfo;
  hostCtx.systemStatus = props.systemStatus;
}

function renderWidget(el: HTMLElement, widget: GridStackWidget) {
  const id = String(widget.id ?? "") as DashboardCardID;
  if (!id) return;

  mountHosts.set(id, el);
  render(
    h(DashboardWidgetHost, {
      id,
      ctx: hostCtx,
    }),
    el,
  );
}

function unmountWidget(id: string) {
  const el = mountHosts.get(id);
  if (!el) return;
  render(null, el);
  mountHosts.delete(id);
}

function unmountAll() {
  for (const el of mountHosts.values()) {
    render(null, el);
  }
  mountHosts.clear();
}

function emitLayoutChange() {
  if (!grid) return;
  const saved = (grid.save(false) as GridStackWidget[]) ?? [];
  syncingFromGrid = true;
  emit("change", geometryFromWidgets(saved, props.items));
  void nextTick(() => {
    syncingFromGrid = false;
  });
}

function loadItems(items: DashboardCardLayout[]) {
  if (!grid) return;
  grid.load(toGridWidgets(items.filter((card) => card.visible)));
}

onMounted(() => {
  if (!gridEl.value) return;

  syncHostCtx();
  GridStack.renderCB = renderWidget;

  grid = GridStack.init(
    {
      column: DASHBOARD_GRID_COLUMNS,
      cellHeight: 80,
      margin: 12,
      animate: true,
      float: false,
      handle: ".dashboard-widget__drag",
      staticGrid: !props.editing,
      alwaysShowResizeHandle: true,
      minRow: 1,
      children: toGridWidgets(props.items.filter((card) => card.visible)),
    },
    gridEl.value,
  );

  grid.on("change", () => {
    if (!props.editing) return;
    emitLayoutChange();
  });

  grid.on("removed", (_event, items) => {
    for (const item of items) {
      if (item.id != null) unmountWidget(String(item.id));
    }
  });
});

onBeforeUnmount(() => {
  unmountAll();
  grid?.destroy(false);
  grid = null;
  if (GridStack.renderCB === renderWidget) {
    GridStack.renderCB = undefined;
  }
});

watch(
  () =>
    [
      props.editing,
      props.buildSummary,
      props.agentRunSummary,
      props.systemInfo,
      props.systemStatus,
    ] as const,
  () => {
    syncHostCtx();
  },
);

watch(
  () => props.editing,
  (editing) => {
    grid?.setStatic(!editing);
  },
);

watch(
  () => visibleCardsSignature(props.items),
  () => {
    if (syncingFromGrid || !grid) return;
    loadItems(props.items);
  },
);
</script>

<template>
  <div
    ref="gridEl"
    class="dashboard-grid grid-stack"
    :class="{ 'dashboard-grid--editing': editing }"
  />
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.dashboard-grid {
  flex: 1;
  width: 100%;
  min-height: 160px;

  :deep(.grid-stack-item-content) {
    inset: 0;
    overflow: visible;
  }

  :deep(.grid-stack-item) {
    transition:
      transform 0.18s ease,
      box-shadow 0.18s ease;
  }

  &:not(.dashboard-grid--editing) :deep(.grid-stack-item:hover) {
    transform: translateY(-2px);
  }

  &--editing :deep(.grid-stack-item-content) {
    outline: 1px dashed color-mix(in srgb, fn.use-var(color, primary) 45%, transparent);
    outline-offset: -1px;
    border-radius: fn.use-var(radius, default);
  }
}
</style>
