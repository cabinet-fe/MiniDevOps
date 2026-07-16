<script setup lang="ts">
defineOptions({ name: "HomePage" });

import { computed, onMounted, onUnmounted, ref } from "vue";
import { message } from "@veltra/desktop";
import { Edit, Save } from "@veltra/icons/normal";
import { useRouter } from "vue-router";

import {
  getBuildSummary,
  getDashboardLayout,
  getSystemInfo,
  getSystemStatus,
  saveDashboardLayout,
} from "@/api/dashboard";
import type {
  BuildSummary,
  DashboardCardID,
  DashboardCardLayout,
  SystemInfo,
  SystemStatus,
} from "@/api/types";
import DashboardBuildCard from "@/components/dashboard-build-card";
import DashboardSystemInfoCard from "@/components/dashboard-system-info-card";
import DashboardSystemStatusCard from "@/components/dashboard-system-status-card";

const STATUS_REFRESH_MS = 30_000;

const router = useRouter();
const layout = ref<DashboardCardLayout[]>([]);
const editing = ref(false);
const loading = ref(true);
const buildSummary = ref<BuildSummary | null>(null);
const systemInfo = ref<SystemInfo | null>(null);
const systemStatus = ref<SystemStatus | null>(null);
let statusTimer: ReturnType<typeof setInterval> | undefined;

const visibleCards = computed(() =>
  [...layout.value].filter((card) => card.visible).sort((a, b) => a.order - b.order),
);

const cardTitles: Record<DashboardCardID, string> = {
  build_summary: "构建摘要",
  system_info: "系统信息",
  system_status: "系统状态",
};

function isVisible(id: DashboardCardID): boolean {
  return layout.value.some((card) => card.id === id && card.visible);
}

async function loadCardData() {
  const requests: Promise<void>[] = [];
  if (isVisible("build_summary")) {
    requests.push(
      getBuildSummary()
        .then((result) => {
          buildSummary.value = result;
        })
        .catch(showLoadError),
    );
  }
  if (isVisible("system_info")) {
    requests.push(
      getSystemInfo()
        .then((result) => {
          systemInfo.value = result;
        })
        .catch(showLoadError),
    );
  }
  await Promise.all(requests);
}

async function refreshStatus() {
  if (!isVisible("system_status")) return;
  try {
    systemStatus.value = await getSystemStatus();
  } catch (error) {
    showLoadError(error);
  }
}

async function loadDashboard() {
  loading.value = true;
  try {
    const result = await getDashboardLayout();
    layout.value = result.cards;
    await loadCardData();
    // Wait until layout and primary cards have rendered before collecting status.
    window.setTimeout(() => void refreshStatus(), 0);
  } catch (error) {
    showLoadError(error);
  } finally {
    loading.value = false;
  }
}

function moveCard(index: number, direction: -1 | 1) {
  const next = index + direction;
  if (next < 0 || next >= layout.value.length) return;
  const cards = [...layout.value];
  [cards[index], cards[next]] = [cards[next], cards[index]];
  layout.value = cards.map((card, order) => ({ ...card, order }));
}

async function persistLayout() {
  try {
    const saved = await saveDashboardLayout({ cards: layout.value });
    layout.value = saved.cards;
    editing.value = false;
    await loadCardData();
    void refreshStatus();
    message.success("仪表盘布局已保存");
  } catch (error) {
    showLoadError(error);
  }
}

function openBuildRun(id: number) {
  void router.push({ name: "cicd-build-run-detail", params: { id: String(id) } });
}

function showLoadError(error: unknown) {
  message.error(error instanceof Error ? error.message : "加载失败");
}

onMounted(() => {
  void loadDashboard();
  statusTimer = window.setInterval(() => void refreshStatus(), STATUS_REFRESH_MS);
});

onUnmounted(() => {
  if (statusTimer) window.clearInterval(statusTimer);
});
</script>

<template>
  <div class="dashboard">
    <div class="dashboard__toolbar">
      <template v-if="editing">
        <u-button @click="editing = false">取消</u-button>
        <u-button type="primary" @click="persistLayout">
          <u-icon :size="14"><Save /></u-icon>
          保存布局
        </u-button>
      </template>
      <u-button v-else text type="primary" @click="editing = true">
        <u-icon :size="14"><Edit /></u-icon>
        编辑卡片
      </u-button>
    </div>

    <u-card v-if="editing" class="dashboard__editor" integrate>
      <u-card-header>
        <h3 class="dashboard__editor-title">卡片布局</h3>
      </u-card-header>
      <u-card-content>
        <p class="dashboard__editor-hint">仅列出当前账号可用的卡片；关闭后可在此重新启用。</p>
        <div v-for="(card, index) in layout" :key="card.id" class="dashboard__editor-row">
          <label class="dashboard__editor-label">
            <u-switch v-model="card.visible" />
            <span>{{ cardTitles[card.id] }}</span>
          </label>
          <span class="dashboard__editor-actions">
            <u-button text :disabled="index === 0" @click="moveCard(index, -1)">上移</u-button>
            <u-button text :disabled="index === layout.length - 1" @click="moveCard(index, 1)">
              下移
            </u-button>
          </span>
        </div>
      </u-card-content>
    </u-card>

    <div v-if="loading" v-loading="true" class="dashboard__loading" />
    <section v-else class="dashboard__cards">
      <template v-for="card in visibleCards" :key="card.id">
        <DashboardBuildCard
          v-if="card.id === 'build_summary'"
          :data="buildSummary"
          @open-run="openBuildRun"
        />
        <DashboardSystemInfoCard v-else-if="card.id === 'system_info'" :data="systemInfo" />
        <DashboardSystemStatusCard v-else-if="card.id === 'system_status'" :data="systemStatus" />
      </template>
      <u-empty
        v-if="!visibleCards.length"
        class="dashboard__empty"
        text="当前没有可见卡片。请编辑并启用卡片。"
      />
    </section>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.dashboard {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  padding: fn.use-var(gap, large);
}

.dashboard__toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  min-height: 32px;
}

.dashboard__editor {
  background: color-mix(in srgb, fn.use-var(bg-color, top) 92%, transparent);
}

.dashboard__editor-title {
  margin: 0;
  color: fn.use-var(text-color, title);
  font-size: 15px;
  font-weight: 600;
}

.dashboard__editor-hint {
  margin: 0 0 12px;
  color: fn.use-var(text-color, second);
  font-size: 13px;
}

.dashboard__editor-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 0;

  & + & {
    border-top: fn.use-var(border, muted);
  }
}

.dashboard__editor-label {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: fn.use-var(text-color, main);
  cursor: pointer;
}

.dashboard__editor-actions {
  display: inline-flex;
  gap: 4px;
}

.dashboard__loading {
  flex: 1;
  min-height: 240px;
}

.dashboard__cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 20px;
  align-content: start;
  flex: 1;
}

.dashboard__empty {
  grid-column: 1 / -1;
  padding: 48px 0;
}
</style>
