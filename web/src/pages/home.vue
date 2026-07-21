<script setup lang="ts">
defineOptions({ name: "HomePage" });

import { computed, onMounted, onUnmounted, ref } from "vue";
import { message } from "@veltra/desktop";
import { Edit, Setting } from "@veltra/icons/normal";
import { useRouter } from "vue-router";

import {
  getAgentRunSummary,
  getBuildSummary,
  getDashboardLayout,
  getSystemInfo,
  getSystemStatus,
  saveDashboardLayout,
} from "@/api/dashboard";
import type {
  AgentRunSummary,
  BuildSummary,
  DashboardCardID,
  DashboardCardLayout,
  SystemInfo,
  SystemStatus,
} from "@/api/types";
import DashboardGrid, { ensureCardGeometry } from "@/components/dashboard-grid";

const STATUS_REFRESH_MS = 30_000;

const router = useRouter();
const layout = ref<DashboardCardLayout[]>([]);
const editSnapshot = ref<DashboardCardLayout[]>([]);
const manageDraft = ref<DashboardCardLayout[]>([]);
const editing = ref(false);
const manageOpen = ref(false);
const saving = ref(false);
const loading = ref(true);
const buildSummary = ref<BuildSummary | null>(null);
const agentRunSummary = ref<AgentRunSummary | null>(null);
const systemInfo = ref<SystemInfo | null>(null);
const systemStatus = ref<SystemStatus | null>(null);
let statusTimer: ReturnType<typeof setInterval> | undefined;

const visibleCards = computed(() => layout.value.filter((card) => card.visible));

const cardTitles: Record<DashboardCardID, string> = {
  build_summary: "构建摘要",
  agent_run_summary: "智能体运行摘要",
  system_info: "系统信息",
  system_status: "系统状态",
};

function cloneCards(cards: DashboardCardLayout[]): DashboardCardLayout[] {
  return cards.map((card) => ({ ...card }));
}

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
  if (isVisible("agent_run_summary")) {
    requests.push(
      getAgentRunSummary()
        .then((result) => {
          agentRunSummary.value = result;
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
    layout.value = ensureCardGeometry(result.cards);
    await loadCardData();
    window.setTimeout(() => void refreshStatus(), 0);
  } catch (error) {
    showLoadError(error);
  } finally {
    loading.value = false;
  }
}

function enterEdit() {
  editSnapshot.value = cloneCards(layout.value);
  editing.value = true;
}

function cancelEdit() {
  layout.value = cloneCards(editSnapshot.value);
  editing.value = false;
}

async function saveEdit() {
  saving.value = true;
  try {
    const saved = await saveDashboardLayout({ cards: ensureCardGeometry(layout.value) });
    layout.value = ensureCardGeometry(saved.cards);
    editing.value = false;
    await loadCardData();
    void refreshStatus();
    message.success("仪表盘布局已保存");
  } catch (error) {
    showLoadError(error);
  } finally {
    saving.value = false;
  }
}

function openManage() {
  manageDraft.value = cloneCards(layout.value);
  manageOpen.value = true;
}

async function persistManage() {
  const next = ensureCardGeometry(
    layout.value.map((card) => {
      const draft = manageDraft.value.find((item) => item.id === card.id);
      return draft ? { ...card, visible: draft.visible } : card;
    }),
  );

  if (editing.value) {
    layout.value = next;
    manageOpen.value = false;
    await loadCardData();
    void refreshStatus();
    return;
  }

  saving.value = true;
  try {
    const saved = await saveDashboardLayout({ cards: next });
    layout.value = ensureCardGeometry(saved.cards);
    manageOpen.value = false;
    await loadCardData();
    void refreshStatus();
    message.success("卡片可见性已保存");
  } catch (error) {
    showLoadError(error);
  } finally {
    saving.value = false;
  }
}

function onGridChange(cards: DashboardCardLayout[]) {
  layout.value = ensureCardGeometry(cards);
}

function openBuildRun(id: number) {
  void router.push({ name: "cicd-build-run-detail", params: { id: String(id) } });
}

function openAgentRun(id: number) {
  void router.push({ name: "ai-run-detail", params: { id: String(id) } });
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
        <u-button text @click="cancelEdit">取消</u-button>
        <u-button type="primary" :loading="saving" @click="saveEdit">保存</u-button>
      </template>
      <template v-else>
        <u-button text type="primary" @click="enterEdit">
          <u-icon :size="14"><Edit /></u-icon>
          编辑布局
        </u-button>
      </template>
      <u-button text @click="openManage">
        <u-icon :size="14"><Setting /></u-icon>
        管理卡片
      </u-button>
    </div>

    <u-dialog v-model="manageOpen" title="管理卡片" style="width: 480px">
      <p class="dashboard__editor-hint">仅列出当前账号可用的卡片；关闭后可在此重新启用。</p>
      <div v-for="card in manageDraft" :key="card.id" class="dashboard__editor-row">
        <label class="dashboard__editor-label">
          <u-switch v-model="card.visible" />
          <span>{{ cardTitles[card.id] }}</span>
        </label>
      </div>
      <template #footer="{ close }">
        <u-button text @click="close()">取消</u-button>
        <u-button type="primary" :loading="saving" @click="persistManage">保存</u-button>
      </template>
    </u-dialog>

    <div v-if="loading" v-loading="true" class="dashboard__loading" />
    <u-scroll v-else class="dashboard__content">
      <DashboardGrid
        v-if="visibleCards.length"
        :items="layout"
        :editing="editing"
        :build-summary="buildSummary"
        :agent-run-summary="agentRunSummary"
        :system-info="systemInfo"
        :system-status="systemStatus"
        @change="onGridChange"
        @open-build-run="openBuildRun"
        @open-agent-run="openAgentRun"
      />
      <u-empty v-else class="dashboard__empty" text="当前没有可见卡片。请打开「管理卡片」启用。" />
    </u-scroll>
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.dashboard {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.dashboard__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  min-height: 32px;
}

.dashboard__content {
  flex: 1;
  min-height: 0;
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

.dashboard__loading {
  flex: 1;
  min-height: 240px;
}

.dashboard__empty {
  padding: 48px 0;
}
</style>
