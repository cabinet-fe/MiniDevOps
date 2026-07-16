<script setup lang="ts">
defineOptions({ name: "CicdBuildRunDetail" });

import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message } from "@veltra/desktop";

import {
  buildRunArtifactURL,
  cancelBuildRun,
  getBuildRun,
  redeployBuildRun,
  retryBuildRun,
} from "@/api/cicd";
import { getAccessToken } from "@/api/http";
import type { BuildRun } from "@/api/types";
import BuildLogViewer, { resolveBuildLogStatus } from "@/components/build-log-viewer";
import { usePermission } from "@/composables/use-permission";
import { formatDateTime } from "@/lib/datetime";
import { JOB_STATUS_TAG, TRIGGER_TYPE_TAG, tagType, type TagType } from "@/lib/tag";

const route = useRoute();
const router = useRouter();
const { hasPermission } = usePermission();

const run = ref<BuildRun | null>(null);
const loading = ref(true);
const acting = ref(false);
const logViewerRef = ref<InstanceType<typeof BuildLogViewer> | null>(null);

let pollTimer: ReturnType<typeof setInterval> | null = null;

const STAGE_TAG: Record<string, TagType> = {
  pending: undefined,
  cloning: "primary",
  building: "primary",
  archiving: "primary",
  distributing: "warning",
  idle: "success",
};

const DISTRIBUTION_TAG: Record<string, TagType> = {
  none: undefined,
  running: "primary",
  all_success: "success",
  partial: "warning",
  all_failed: "danger",
  cancelled: "warning",
};

const canExecute = computed(() => hasPermission("cicd.build_jobs:execute"));
const runId = computed(() => Number(route.params.id));

const isLive = computed(() => {
  const s = run.value?.status;
  return s === "queued" || s === "running" || run.value?.distribution_summary === "running";
});

const logViewerStatus = computed(() =>
  resolveBuildLogStatus(run.value?.status, run.value?.distribution_summary),
);

const canCancel = computed(() => {
  if (!canExecute.value || !run.value) return false;
  return (
    run.value.status === "queued" ||
    run.value.status === "running" ||
    (run.value.status === "success" && run.value.distribution_summary === "running")
  );
});

const canRetry = computed(
  () =>
    canExecute.value &&
    !!run.value &&
    ["failed", "cancelled", "interrupted", "success"].includes(run.value.status),
);

const canRedeploy = computed(
  () => canExecute.value && run.value?.status === "success" && !!run.value.artifact_path,
);

const shortCommit = computed(() => {
  const hash = run.value?.commit_hash?.trim();
  if (!hash) return "—";
  return hash.length > 12 ? hash.slice(0, 12) : hash;
});

const snapshotText = computed(() => {
  const raw = run.value?.snapshot_json?.trim();
  if (!raw) return "（空）";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
});

async function load() {
  const id = runId.value;
  if (!id) {
    message.error("无效 ID");
    return;
  }
  try {
    run.value = await getBuildRun(id);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "加载失败");
  } finally {
    loading.value = false;
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      run.value = await getBuildRun(runId.value);
      if (!isLive.value) stopPolling();
    } catch {
      /* ignore */
    }
  }, 2500);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

async function onCancel() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    run.value = await cancelBuildRun(run.value.id);
    message.success("已取消");
  } catch (err) {
    message.error(err instanceof Error ? err.message : "取消失败");
  } finally {
    acting.value = false;
  }
}

async function onRetry() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    const next = await retryBuildRun(run.value.id);
    message.success(`已创建重试 #${next.build_number}`);
    await router.push({ name: "cicd-build-run-detail", params: { id: String(next.id) } });
  } catch (err) {
    message.error(err instanceof Error ? err.message : "重试失败");
  } finally {
    acting.value = false;
  }
}

async function onRedeploy() {
  if (!run.value || acting.value) return;
  acting.value = true;
  try {
    run.value = await redeployBuildRun(run.value.id);
    message.success("已开始重新分发");
    logViewerRef.value?.appendLine("=== Redeploy requested ===");
    logViewerRef.value?.reconnect();
    startPolling();
  } catch (err) {
    message.error(err instanceof Error ? err.message : "重新分发失败");
  } finally {
    acting.value = false;
  }
}

async function onDownloadArtifact() {
  const token = getAccessToken();
  if (!token || !run.value) return;
  try {
    const res = await fetch(buildRunArtifactURL(run.value.id), {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) {
      throw new Error((await res.text()) || `HTTP ${res.status}`);
    }
    const blob = await res.blob();
    const cd = res.headers.get("Content-Disposition") || "";
    const m = /filename="?([^"]+)"?/.exec(cd);
    const name = m?.[1] || `build-${run.value.build_number}.tar.gz`;
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = name;
    a.click();
    URL.revokeObjectURL(a.href);
  } catch (err) {
    message.error(err instanceof Error ? err.message : "下载失败");
  }
}

watch(runId, async () => {
  loading.value = true;
  stopPolling();
  await load();
  if (isLive.value) startPolling();
});

onMounted(async () => {
  await load();
  if (isLive.value) startPolling();
});

onUnmounted(() => {
  stopPolling();
});
</script>

<template>
  <div class="page">
    <div class="page-toolbar">
      <p class="risk-note">
        构建脚本与 Bedrock 以同一操作系统用户执行，无沙箱隔离；请仅让可信人员编辑脚本。
      </p>
      <div class="actions">
        <u-button v-if="canCancel" :disabled="acting" variant="outline" @click="onCancel">
          取消
        </u-button>
        <u-button v-if="canRetry" :disabled="acting" @click="onRetry">重试</u-button>
        <u-button v-if="canRedeploy" :disabled="acting" @click="onRedeploy">重新分发</u-button>
        <u-button v-if="run?.status === 'success'" variant="outline" @click="onDownloadArtifact">
          下载制品
        </u-button>
        <u-button @click="router.push({ name: 'cicd-build-runs' })">返回列表</u-button>
      </div>
    </div>

    <div v-if="loading" class="state">加载中…</div>
    <template v-else-if="run">
      <section class="panel meta-panel">
        <div class="meta-grid">
          <div class="meta-item">
            <span class="meta-label">状态</span>
            <u-tag size="small" :type="tagType(run.status, JOB_STATUS_TAG)">{{ run.status }}</u-tag>
          </div>
          <div class="meta-item">
            <span class="meta-label">阶段</span>
            <u-tag size="small" :type="tagType(run.stage, STAGE_TAG)">
              {{ run.stage || "—" }}
            </u-tag>
          </div>
          <div class="meta-item">
            <span class="meta-label">分发汇总</span>
            <u-tag size="small" :type="tagType(run.distribution_summary, DISTRIBUTION_TAG)">
              {{ run.distribution_summary || "—" }}
            </u-tag>
          </div>
          <div class="meta-item">
            <span class="meta-label">触发</span>
            <u-tag size="small" :type="tagType(run.trigger_type, TRIGGER_TYPE_TAG)">
              {{ run.trigger_type }}
            </u-tag>
          </div>
          <div class="meta-item">
            <span class="meta-label">任务 ID</span>
            <span class="meta-value">{{ run.build_job_id }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">分支</span>
            <span class="meta-value mono">{{ run.branch || "—" }}</span>
          </div>
          <div class="meta-item meta-item--wide">
            <span class="meta-label">Commit</span>
            <span class="meta-value mono" :title="run.commit_hash || undefined">
              {{ shortCommit }}
            </span>
          </div>
          <div class="meta-item">
            <span class="meta-label">创建时间</span>
            <span class="meta-value">{{ formatDateTime(run.created_at) || "—" }}</span>
          </div>
        </div>
        <p v-if="run.commit_message" class="commit-msg">{{ run.commit_message }}</p>
        <p v-if="run.error_message" class="error-msg">{{ run.error_message }}</p>
      </section>

      <section class="section">
        <h3>构建日志</h3>
        <BuildLogViewer
          ref="logViewerRef"
          :run-id="runId"
          :live="isLive"
          :status="logViewerStatus"
        />
      </section>

      <section class="section">
        <h3>部署尝试</h3>
        <div class="panel">
          <u-empty v-if="!run.deploy_attempts?.length" text="暂无部署尝试" />
          <ul v-else class="attempts">
            <li v-for="a in run.deploy_attempts" :key="a.id" class="attempt">
              <div class="attempt__main">
                <span class="mono">batch {{ a.batch_no }}</span>
                <span class="attempt__sep">·</span>
                <span>target {{ a.deploy_target_id ?? "—" }}</span>
                <u-tag size="small" :type="tagType(a.status, JOB_STATUS_TAG)">{{ a.status }}</u-tag>
              </div>
              <p v-if="a.error_message" class="attempt__error">{{ a.error_message }}</p>
            </li>
          </ul>
        </div>
      </section>

      <section class="section">
        <h3>配置快照</h3>
        <pre class="snap">{{ snapshotText }}</pre>
      </section>
    </template>
    <u-empty v-else text="执行记录不存在或无权访问" />
  </div>
</template>

<style scoped lang="scss">
@use "pkg:@veltra/styles/functions" as fn;

.page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.page-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.risk-note {
  margin: 0;
  flex: 1;
  min-width: 0;
  font-size: 12px;
  line-height: 1.5;
  color: fn.use-var(text-color, assist);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  flex-shrink: 0;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.section h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: fn.use-var(text-color, title);
}

.panel {
  min-width: 0;
  padding: fn.use-var(gap, default);
  border-radius: fn.use-var(radius, default);
  background: fn.use-var(bg-color, top);
}

.meta-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px 16px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  min-width: 0;
}

@media (min-width: 640px) {
  .meta-item--wide {
    grid-column: span 2;
  }
}

.meta-label {
  font-size: 12px;
  color: fn.use-var(text-color, assist);
}

.meta-value {
  font-size: 13px;
  font-weight: 500;
  color: fn.use-var(text-color, main);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.commit-msg {
  margin: 0;
  padding-top: 12px;
  border-top: fn.use-var(border, muted);
  font-size: 13px;
  line-height: 1.5;
  color: fn.use-var(text-color, second);
  overflow-wrap: anywhere;
}

.error-msg {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: fn.use-var(color, danger);
  overflow-wrap: anywhere;
}

.attempts {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attempt {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.attempt__main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.attempt__sep {
  color: fn.use-var(text-color, assist);
}

.attempt__error {
  margin: 0;
  font-size: 12px;
  color: fn.use-var(color, danger);
  overflow-wrap: anywhere;
}

.snap {
  margin: 0;
  padding: fn.use-var(gap, default);
  max-height: 280px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.55;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  background: fn.use-var(bg-color, top);
  border-radius: fn.use-var(radius, default);
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.state {
  opacity: 0.7;
}
</style>
