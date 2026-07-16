<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { message } from "@veltra/desktop";
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
import { formatDateTime } from "@/lib/datetime";

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

function hasCard(id: DashboardCardID): boolean {
  return layout.value.some((card) => card.id === id);
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
    <header class="dashboard__head">
      <div>
        <h2>仪表盘</h2>
        <p>系统信息为完整只读信息；它不会授予运维写操作权限。</p>
      </div>
      <div class="dashboard__actions">
        <u-button v-if="editing" @click="editing = false">取消</u-button>
        <u-button v-if="editing" type="primary" @click="persistLayout">保存布局</u-button>
        <u-button v-else @click="editing = true">编辑卡片</u-button>
      </div>
    </header>

    <section v-if="editing" class="dashboard__editor">
      <h3>卡片布局</h3>
      <p>仅列出当前账号可用的卡片；关闭卡片后可在此重新启用。</p>
      <div v-for="(card, index) in layout" :key="card.id" class="dashboard__editor-row">
        <label>
          <input v-model="card.visible" type="checkbox" />
          {{ cardTitles[card.id] }}
        </label>
        <span class="dashboard__editor-actions">
          <u-button text :disabled="index === 0" @click="moveCard(index, -1)">上移</u-button>
          <u-button text :disabled="index === layout.length - 1" @click="moveCard(index, 1)"
            >下移</u-button
          >
        </span>
      </div>
    </section>

    <p v-if="loading" class="dashboard__loading">正在加载仪表盘…</p>
    <section v-else class="dashboard__cards">
      <article v-for="card in visibleCards" :key="card.id" class="dashboard-card">
        <template v-if="card.id === 'build_summary' && hasCard('build_summary')">
          <h3>构建摘要</h3>
          <div class="dashboard-card__metrics">
            <span
              >运行中 <strong>{{ buildSummary?.running ?? "—" }}</strong></span
            >
            <span
              >排队 <strong>{{ buildSummary?.queued ?? "—" }}</strong></span
            >
            <span
              >成功率
              <strong>{{
                buildSummary ? `${buildSummary.success_rate.toFixed(1)}%` : "—"
              }}</strong></span
            >
          </div>
          <div v-if="buildSummary?.recent?.length" class="dashboard-card__recent">
            <button
              v-for="run in buildSummary.recent"
              :key="run.id"
              type="button"
              @click="openBuildRun(run.id)"
            >
              #{{ run.build_number }} · {{ run.status }} · {{ run.branch || "默认分支" }}
            </button>
          </div>
        </template>

        <template v-else-if="card.id === 'system_info' && hasCard('system_info')">
          <h3>系统信息</h3>
          <dl class="dashboard-card__details">
            <dt>版本</dt>
            <dd>{{ systemInfo?.version ?? "—" }}</dd>
            <dt>主机名</dt>
            <dd>{{ systemInfo?.hostname ?? "—" }}</dd>
            <dt>平台</dt>
            <dd>{{ systemInfo ? `${systemInfo.os}/${systemInfo.arch}` : "—" }}</dd>
            <dt>运行时</dt>
            <dd>{{ systemInfo?.runtime ?? "—" }}</dd>
            <dt>启动时间</dt>
            <dd>{{ formatDateTime(systemInfo?.start_time) || "—" }}</dd>
          </dl>
        </template>

        <template v-else-if="card.id === 'system_status' && hasCard('system_status')">
          <h3>系统状态</h3>
          <div class="dashboard-card__metrics">
            <span
              >健康 <strong>{{ systemStatus?.health ?? "—" }}</strong></span
            >
            <span
              >CPU
              <strong>{{ systemStatus ? `${systemStatus.cpu_usage_percent}%` : "—" }}</strong></span
            >
            <span
              >内存
              <strong>{{
                systemStatus ? `${systemStatus.memory_usage_percent}%` : "—"
              }}</strong></span
            >
          </div>
          <ul class="dashboard-card__disk">
            <li v-for="directory in systemStatus?.directories ?? []" :key="directory.path">
              {{ directory.path }}：{{ directory.used_percent }}% 已用
            </li>
          </ul>
        </template>
      </article>
      <p v-if="!visibleCards.length" class="dashboard__empty">
        当前没有可见卡片。请在编辑模式中启用卡片。
      </p>
    </section>
  </div>
</template>

<style scoped lang="scss">
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.dashboard__head,
.dashboard__actions,
.dashboard__editor-row,
.dashboard-card__metrics {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dashboard__head {
  justify-content: space-between;
}
.dashboard__head h2,
.dashboard__editor h3,
.dashboard-card h3 {
  margin: 0;
}
.dashboard__head p,
.dashboard__editor p {
  margin: 6px 0 0;
  color: #6b7280;
}
.dashboard__editor,
.dashboard-card {
  border: 1px solid var(--u-border-color, #e5e7eb);
  border-radius: 8px;
  padding: 16px;
  background: var(--u-bg-color, #fff);
}
.dashboard__editor-row {
  justify-content: space-between;
  padding: 8px 0;
  border-top: 1px solid #f0f0f0;
}
.dashboard__editor-row:first-of-type {
  border-top: 0;
}
.dashboard__cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
}
.dashboard-card {
  min-height: 160px;
}
.dashboard-card__metrics {
  margin: 18px 0;
  justify-content: space-between;
}
.dashboard-card__metrics span {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: #6b7280;
}
.dashboard-card__metrics strong {
  color: #111827;
  font-size: 20px;
}
.dashboard-card__recent {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.dashboard-card__recent button {
  border: 0;
  background: transparent;
  padding: 0;
  color: #2563eb;
  cursor: pointer;
  text-align: left;
}
.dashboard-card__details {
  display: grid;
  grid-template-columns: 88px 1fr;
  gap: 8px 12px;
  margin: 16px 0 0;
}
.dashboard-card__details dt {
  color: #6b7280;
}
.dashboard-card__details dd {
  margin: 0;
  overflow-wrap: anywhere;
}
.dashboard-card__disk {
  margin: 0;
  padding-left: 20px;
  color: #4b5563;
}
.dashboard__loading,
.dashboard__empty {
  color: #6b7280;
}
</style>
